package framework

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	snowAMIIDUbuntu122   = "T_SNOW_AMIID_UBUNTU_1_22"
	snowAMIIDUbuntu123   = "T_SNOW_AMIID_UBUNTU_1_23"
	snowAMIIDUbuntu124   = "T_SNOW_AMIID_UBUNTU_1_24"
	snowAMIIDUbuntu125   = "T_SNOW_AMIID_UBUNTU_1_25"
	snowAMIIDUbuntu126   = "T_SNOW_AMIID_UBUNTU_1_26"
	snowDevices          = "T_SNOW_DEVICES"
	snowControlPlaneCidr = "T_SNOW_CONTROL_PLANE_CIDR"
	snowPodCidr          = "T_SNOW_POD_CIDR"
	snowCredentialsFile  = "EKSA_AWS_CREDENTIALS_FILE"
	snowCertificatesFile = "EKSA_AWS_CA_BUNDLES_FILE"
	snowIPPoolIPStart    = "T_SNOW_IPPOOL_IPSTART"
	snowIPPoolIPEnd      = "T_SNOW_IPPOOL_IPEND"
	snowIPPoolGateway    = "T_SNOW_IPPOOL_GATEWAY"
	snowIPPoolSubnet     = "T_SNOW_IPPOOL_SUBNET"
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
func (s *Snow) CleanupVMs(_ string) error {
	return nil
}

func (s *Snow) WithProviderUpgrade(fillers ...api.SnowFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.SnowToConfigFiller(fillers...))
	}
}

// WithBottlerocket122 returns a cluster config filler that sets the kubernetes version of the cluster to 1.22
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket122() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube122)
}

// WithBottlerocket123 returns a cluster config filler that sets the kubernetes version of the cluster to 1.23
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket123() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube123)
}

// WithBottlerocket124 returns a cluster config filler that sets the kubernetes version of the cluster to 1.24
// as well as the right devices and osFamily for all SnowMachineConfigs. It also sets any
// necessary machine config default required for BR, like the container volume size. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithBottlerocket124() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketForKubeVersion(anywherev1.Kube124)
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

// WithBottlerocketStaticIP122 returns a cluster config filler that sets the kubernetes version of the cluster to 1.22
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket122,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP122() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube122)
}

// WithBottlerocketStaticIP123 returns a cluster config filler that sets the kubernetes version of the cluster to 1.23
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket123,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP123() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube123)
}

// WithBottlerocketStaticIP124 returns a cluster config filler that sets the kubernetes version of the cluster to 1.24
// as well as the right devices, osFamily and static ip config for all SnowMachineConfigs. Comparing to WithBottlerocket124,
// this method also adds a snow ip pool to support static ip configuration.
func (s *Snow) WithBottlerocketStaticIP124() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withBottlerocketStaticIPForKubeVersion(anywherev1.Kube124)
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

// WithUbuntu122 returns a cluster config filler that sets the kubernetes version of the cluster to 1.22
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu122() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withKubeVersionAndOS(anywherev1.Kube122, anywherev1.Ubuntu)
}

// WithUbuntu123 returns a cluster config filler that sets the kubernetes version of the cluster to 1.23
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu123() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withKubeVersionAndOS(anywherev1.Kube123, anywherev1.Ubuntu)
}

// WithUbuntu124 returns a cluster config filler that sets the kubernetes version of the cluster to 1.24
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu124() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withKubeVersionAndOS(anywherev1.Kube124, anywherev1.Ubuntu)
}

// WithUbuntu125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu125() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withKubeVersionAndOS(anywherev1.Kube125, anywherev1.Ubuntu)
}

// WithUbuntu126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right devices and osFamily for all SnowMachineConfigs. If the env var is set, this will
// also set the AMI ID. Otherwise, it will leave it empty and let CAPAS select one.
func (s *Snow) WithUbuntu126() api.ClusterConfigFiller {
	s.t.Helper()
	return s.withKubeVersionAndOS(anywherev1.Kube126, anywherev1.Ubuntu)
}

func (s *Snow) withBottlerocketForKubeVersion(kubeVersion anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		s.withKubeVersionAndOS(kubeVersion, anywherev1.Bottlerocket),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithSnowContainersVolumeSize(100))),
	)
}

func (s *Snow) withBottlerocketStaticIPForKubeVersion(kubeVersion anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	poolName := "pool-1"
	return api.JoinClusterConfigFillers(
		s.withKubeVersionAndOS(kubeVersion, anywherev1.Bottlerocket),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithSnowContainersVolumeSize(100))),
		api.SnowToConfigFiller(api.WithChangeForAllSnowMachines(api.WithStaticIP(poolName))),
		api.SnowToConfigFiller(s.withIPPoolFromEnvVar(poolName)),
	)
}

func (s *Snow) withKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, osFamily anywherev1.OSFamily) api.ClusterConfigFiller {
	envar := fmt.Sprintf("T_SNOW_AMIID_%s_%s", strings.ToUpper(string(osFamily)), strings.ReplaceAll(string(kubeVersion), ".", "_"))

	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
		api.SnowToConfigFiller(
			s.withAMIIDFromEnvVar(envar),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(osFamily),
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

func WithSnowUbuntu122() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu122, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
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

func UpdateSnowUbuntuTemplate122Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu122, api.WithSnowAMIIDForAllMachines)
}

func UpdateSnowUbuntuTemplate123Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu123, api.WithSnowAMIIDForAllMachines)
}

// ClusterStateValidations returns a list of provider specific validations.
func (s *Snow) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}
