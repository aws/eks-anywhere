package framework

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	snowAMIIDUbuntu123   = "T_SNOW_AMIID_UBUNTU_1_23"
	snowAMIIDUbuntu124   = "T_SNOW_AMIID_UBUNTU_1_24"
	snowAMIIDUbuntu125   = "T_SNOW_AMIID_UBUNTU_1_25"
	snowAMIIDUbuntu126   = "T_SNOW_AMIID_UBUNTU_1_26"
	snowAMIIDUbuntu127   = "T_SNOW_AMIID_UBUNTU_1_27"
	snowAMIIDUbuntu128   = "T_SNOW_AMIID_UBUNTU_1_28"
	snowDevices          = "T_SNOW_DEVICES"
	snowControlPlaneCidr = "T_SNOW_CONTROL_PLANE_CIDR"
	snowPodCidr          = "T_SNOW_POD_CIDR"
	snowCredentialsFile  = "EKSA_AWS_CREDENTIALS_FILE"
	snowCertificatesFile = "EKSA_AWS_CA_BUNDLES_FILE"
	snowIPPoolIPStart    = "T_SNOW_IPPOOL_IPSTART"
	snowIPPoolIPEnd      = "T_SNOW_IPPOOL_IPEND"
	snowIPPoolGateway    = "T_SNOW_IPPOOL_GATEWAY"
	snowIPPoolSubnet     = "T_SNOW_IPPOOL_SUBNET"

	snowEc2TagPrefix = "sigs.k8s.io/cluster-api-provider-aws-snow/cluster/"
)

var requiredSnowEnvVars = []string{
	snowDevices,
	snowControlPlaneCidr,
	snowCredentialsFile,
	snowCertificatesFile,
}

type Snow struct {
	t              *testing.T
	fillers        []api.SnowFiller
	clusterFillers []api.ClusterFiller
	cpCidr         string
	podCidr        string
}

type SnowOpt func(*Snow)

func NewSnow(t *testing.T, opts ...SnowOpt) *Snow {
	checkRequiredEnvVars(t, requiredSnowEnvVars)
	s := &Snow{
		t: t,
	}

	s.cpCidr = os.Getenv(snowControlPlaneCidr)
	s.podCidr = os.Getenv(snowPodCidr)

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Snow) Name() string {
	return "snow"
}

func (s *Snow) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (s *Snow) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (s *Snow) ClusterConfigUpdates() []api.ClusterConfigFiller {
	s.t.Logf("Searching for free IP for Snow Control Plane in CIDR %s", s.cpCidr)
	ip, err := GenerateUniqueIp(s.cpCidr)
	if err != nil {
		s.t.Fatalf("failed to generate control plane ip for snow [cidr=%s]: %v", s.cpCidr, err)
	}
	s.t.Logf("Selected IP %s for Snow Control Plane", ip)

	f := make([]api.ClusterFiller, 0, len(s.clusterFillers)+2)
	f = append(f, s.clusterFillers...)
	f = append(f, api.WithControlPlaneEndpointIP(ip))

	if s.podCidr != "" {
		f = append(f, api.WithPodCidr(s.podCidr))
	}

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.SnowToConfigFiller(s.fillers...)}
}

// CleanupVMs  satisfies the test framework Provider.
func (s *Snow) CleanupVMs(clusterName string) error {
	snowDeviceIPs := strings.Split(os.Getenv(snowDevices), ",")
	s.t.Logf("Cleaning ec2 instances of %s in snow devices: %v", clusterName, snowDeviceIPs)

	var res []error
	for _, ip := range snowDeviceIPs {
		sess, err := newSession(ip)
		if err != nil {
			res = append(res, fmt.Errorf("Cannot create session to snow device: %w", err))
			continue
		}

		ec2Client := ec2.New(sess)
		// snow device doesn't support filter hitherto
		out, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{})
		if err != nil {
			res = append(res, fmt.Errorf("Cannot get ec2 instances from snow device: %w", err))
			continue
		}

		var ownedInstanceIds []*string
		for _, reservation := range out.Reservations {
			for _, instance := range reservation.Instances {
				if isNotTerminatedAndHasTag(instance, snowEc2TagPrefix+clusterName) {
					ownedInstanceIds = append(ownedInstanceIds, instance.InstanceId)
				}
			}
		}

		if len(ownedInstanceIds) != 0 {
			if _, err = ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
				InstanceIds: ownedInstanceIds,
			}); err != nil {
				res = append(res, fmt.Errorf("Cannot terminate ec2 instances from snow device: %w", err))
			} else {
				s.t.Logf("Cluster %s EC2 instances have been cleaned from device %s: %+v", clusterName, ip, ownedInstanceIds)
			}
		} else {
			s.t.Logf("No EC2 instances to cleanup for snow device: %s", ip)
		}

		cleanedKeys, err := cleanupKeypairs(ec2Client, clusterName)
		if err != nil {
			res = append(res, err)
		} else {
			s.t.Logf("KeyPairs has been cleaned: %+v", cleanedKeys)
		}

	}

	return kerrors.NewAggregate(res)
}

func cleanupKeypairs(ec2Client *ec2.EC2, clusterName string) ([]*string, error) {
	out, err := ec2Client.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, err
	}

	var keyPairNames []*string
	for _, keyPair := range out.KeyPairs {
		if strings.Contains(*keyPair.KeyName, clusterName) {
			keyPairNames = append(keyPairNames, keyPair.KeyName)
		}
	}

	var errs []error
	for _, keyPairName := range keyPairNames {
		if _, err := ec2Client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
			KeyName: keyPairName,
		}); err != nil {
			errs = append(errs, err)
		}
	}
	return keyPairNames, kerrors.NewAggregate(errs)
}

func isNotTerminatedAndHasTag(instance *ec2.Instance, tag string) bool {
	if *instance.State.Name == "terminated" {
		return false
	}

	for _, t := range instance.Tags {
		if *t.Key == tag {
			return true
		}
	}
	return false
}

func newSession(ip string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String("http://" + ip + ":8008"),
		Credentials: credentials.NewSharedCredentials(os.Getenv(snowCredentialsFile), ip),
		Region:      aws.String("snow"),
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot create session to snow device: %v", err)
	}
	return sess, nil
}

func (s *Snow) WithProviderUpgrade(fillers ...api.SnowFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.SnowToConfigFiller(fillers...))
	}
}

// WithBottlerocket125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket125() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube125)
}

// WithBottlerocket126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket126() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube126)
}

// WithBottlerocket127 returns a cluster config filler that sets the kubernetes version of the cluster to 1.27
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket127() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube127)
}

// WithBottlerocket128 returns a cluster config filler that sets the kubernetes version of the cluster to 1.28
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket128() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube128)
}

// WithBottlerocketStaticIP125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket125,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP125() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube125)
}

// WithBottlerocketStaticIP126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket126,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP126() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube126)
}

// WithBottlerocketStaticIP127 returns a cluster config filler that sets the kubernetes version of the cluster to 1.27
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket127,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP127() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube127)
}

// WithBottlerocketStaticIP128 returns a cluster config filler that sets the kubernetes version of the cluster to 1.28
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket128,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP128() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube128)
}

// WithUbuntu125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu125() api.ClusterConfigFiller {
	s.t.Helper()
	return s.WithKubeVersionAndOS(anywherev1.Kube125, Ubuntu2004, nil)
}

// WithUbuntu126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu126() api.ClusterConfigFiller {
	s.t.Helper()
	return s.WithKubeVersionAndOS(anywherev1.Kube126, Ubuntu2004, nil)
}

// WithUbuntu127 returns a cluster config filler that sets the kubernetes version of the cluster to 1.27
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu127() api.ClusterConfigFiller {
	s.t.Helper()
	return s.WithKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, nil)
}

// WithUbuntu128 returns a cluster config filler that sets the kubernetes version of the cluster to 1.28
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu128() api.ClusterConfigFiller {
	s.t.Helper()
	return s.WithKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, nil)
}

func (s *Snow) withBottlerocketForKubeVersion(kubeVersion anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		s.WithKubeVersionAndOS(kubeVersion, Bottlerocket1, nil),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithSnowContainersVolumeSize(100))),
	)
}

func (s *Snow) withBottlerocketStaticIPForKubeVersion(kubeVersion anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	poolName := "pool-1"
	return api.JoinClusterConfigFillers(
		s.WithKubeVersionAndOS(kubeVersion, Bottlerocket1, nil),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithSnowContainersVolumeSize(100))),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithStaticIP(poolName))),
		api.SnowToConfigFiller(s.withIPPoolFromEnvVar(poolName)),
	)
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the correct AMI ID
// and devices for the Snow machine configs.
func (s *Snow) WithKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.ClusterConfigFiller {
	envar := fmt.Sprintf("T_SNOW_AMIID_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ReplaceAll(string(kubeVersion), ".", "_"))

	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
		api.SnowToConfigFiller(
			s.withAMIIDFromEnvVar(envar),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(osFamiliesForOS[os]),
		),
	)
}

func (s *Snow) withAMIIDFromEnvVar(envvar string) api.SnowFiller {
	val, ok := os.LookupEnv(envvar)
	if !ok {
		s.t.Log("% for Snow AMI ID is not set, leaving amiID empty which will let CAPAS select the right AMI from the ones available in the device", envvar)
		val = ""
	}

	return api.WithSnowAMIIDForAllMachines(val)
}

func (s *Snow) withIPPoolFromEnvVar(name string) api.SnowFiller {
	envVars := []string{snowIPPoolIPStart, snowIPPoolIPEnd, snowIPPoolGateway, snowIPPoolSubnet}
	checkRequiredEnvVars(s.t, envVars)
	return api.WithSnowIPPool(name, os.Getenv(snowIPPoolIPStart), os.Getenv(snowIPPoolIPEnd), os.Getenv(snowIPPoolGateway), os.Getenv(snowIPPoolSubnet))
}

func WithSnowUbuntu123() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu123, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowUbuntu124 returns SnowOpt that sets the right devices and osFamily for all SnowMachineConfigs.
// If the env var is set, this will also set the AMI ID.
// Otherwise, it will leave it empty and let CAPAS select one.
func WithSnowUbuntu124() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			s.withAMIIDFromEnvVar(snowAMIIDUbuntu124),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowUbuntu125 returns SnowOpt that sets the right devices and osFamily for all SnowMachineConfigs.
// If the env var is set, this will also set the AMI ID.
// Otherwise, it will leave it empty and let CAPAS select one.
func WithSnowUbuntu125() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu125, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowUbuntu126 returns SnowOpt that sets the right devices and osFamily for all SnowMachineConfigs.
// If the env var is set, this will also set the AMI ID.
// Otherwise, it will leave it empty and let CAPAS select one.
func WithSnowUbuntu126() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu126, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowUbuntu127 returns SnowOpt that sets the right devices and osFamily for all SnowMachineConfigs.
// If the env var is set, this will also set the AMI ID.
// Otherwise, it will leave it empty and let CAPAS select one.
func WithSnowUbuntu127() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu127, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowUbuntu128 returns SnowOpt that sets the right devices and osFamily for all SnowMachineConfigs.
// If the env var is set, this will also set the AMI ID.
// Otherwise, it will leave it empty and let CAPAS select one.
func WithSnowUbuntu128() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu128, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

// WithSnowWorkerNodeGroup stores the necessary fillers to update/create the provided worker node group with its corresponding SnowMachineConfig
// and apply the given changes to that machine config.
func WithSnowWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.SnowMachineConfigFiller) SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers, snowMachineConfig(name, fillers...))

		s.clusterFillers = append(s.clusterFillers, buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup))
	}
}

// WithWorkerNodeGroup returns a filler that updates/creates the provided worker node group with its corresponding SnowMachineConfig
// and applies the given changes to that machine config.
func (s *Snow) WithWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.SnowMachineConfigFiller) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		api.SnowToConfigFiller(snowMachineConfig(name, fillers...)),
	)
}

// WithNewWorkerNodeGroup returns a filler that updates/creates the provided worker node group with its corresponding SnowMachineConfig.
func (s *Snow) WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		api.SnowToConfigFiller(snowMachineConfig(name)),
	)
}

// WithNewSnowWorkerNodeGroup updates the test cluster Config with the fillers for an specific snow worker node group.
// It applies the fillers in WorkerNodeGroup to the named worker node group and the ones for the corresponding SnowMachineConfig.
func (s *Snow) WithNewSnowWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.SnowMachineConfigFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(
			api.SnowToConfigFiller(snowMachineConfig(name, fillers...)),
			api.ClusterToConfigFiller(buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		)
	}
}

func snowMachineConfig(name string, fillers ...api.SnowMachineConfigFiller) api.SnowFiller {
	f := make([]api.SnowMachineConfigFiller, 0, len(fillers)+2)
	f = append(f,
		api.WithSnowMachineDefaultValues(),
		api.WithSnowDevices(os.Getenv(snowDevices)),
	)
	f = append(f, fillers...)

	return api.WithSnowMachineConfig(name, f...)
}

func buildSnowWorkerNodeGroupClusterFiller(machineConfigName string, workerNodeGroup *WorkerNodeGroup) api.ClusterFiller {
	workerNodeGroup.MachineConfigKind = anywherev1.SnowMachineConfigKind
	workerNodeGroup.MachineConfigName = machineConfigName
	return workerNodeGroup.ClusterFiller()
}

func UpdateSnowUbuntuTemplate123Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu123, api.WithSnowAMIIDForAllMachines)
}

// ClusterStateValidations returns a list of provider specific validations.
func (s *Snow) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}
