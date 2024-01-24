package framework

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/retrier"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

const (
	cloudstackDomainVar                         = "T_CLOUDSTACK_DOMAIN"
	cloudstackMultiLevelDomainVar               = "T_CLOUDSTACK_MULTILEVEL_DOMAIN"
	cloudstackZoneVar                           = "T_CLOUDSTACK_ZONE"
	cloudstackZone2Var                          = "T_CLOUDSTACK_ZONE_2"
	cloudstackZone3Var                          = "T_CLOUDSTACK_ZONE_3"
	cloudstackAccountVar                        = "T_CLOUDSTACK_ACCOUNT"
	cloudstackAccountForMultiLevelDomainVar     = "T_CLOUDSTACK_ACCOUNT_FOR_MULTILEVEL_DOMAIN"
	cloudstackNetworkVar                        = "T_CLOUDSTACK_NETWORK"
	cloudstackNetwork2Var                       = "T_CLOUDSTACK_NETWORK_2"
	cloudstackNetwork3Var                       = "T_CLOUDSTACK_NETWORK_3"
	cloudstackCredentialsVar                    = "T_CLOUDSTACK_CREDENTIALS"
	cloudstackCredentials2Var                   = "T_CLOUDSTACK_CREDENTIALS_2"
	cloudstackCredentials3Var                   = "T_CLOUDSTACK_CREDENTIALS_3"
	cloudstackCredentialsForMultiLevelDomainVar = "T_CLOUDSTACK_CREDENTIALS_FOR_MULTILEVEL_DOMAIN"
	cloudstackManagementServerVar               = "T_CLOUDSTACK_MANAGEMENT_SERVER"
	cloudstackManagementServer2Var              = "T_CLOUDSTACK_MANAGEMENT_SERVER_2"
	cloudstackManagementServer3Var              = "T_CLOUDSTACK_MANAGEMENT_SERVER_3"
	cloudstackSSHAuthorizedKeyVar               = "T_CLOUDSTACK_SSH_AUTHORIZED_KEY"
	cloudstackComputeOfferingLargeVar           = "T_CLOUDSTACK_COMPUTE_OFFERING_LARGE"
	cloudstackComputeOfferingLargerVar          = "T_CLOUDSTACK_COMPUTE_OFFERING_LARGER"
	cloudStackClusterIPPoolEnvVar               = "T_CLOUDSTACK_CLUSTER_IP_POOL"
	cloudStackCidrVar                           = "T_CLOUDSTACK_CIDR"
	podCidrVar                                  = "T_CLOUDSTACK_POD_CIDR"
	serviceCidrVar                              = "T_CLOUDSTACK_SERVICE_CIDR"
	cloudstackFeatureGateEnvVar                 = "CLOUDSTACK_PROVIDER"
	cloudstackB64EncodedSecretEnvVar            = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"
)

var requiredCloudStackEnvVars = []string{
	cloudstackAccountVar,
	cloudstackAccountForMultiLevelDomainVar,
	cloudstackDomainVar,
	cloudstackMultiLevelDomainVar,
	cloudstackZoneVar,
	cloudstackZone2Var,
	cloudstackZone3Var,
	cloudstackCredentialsVar,
	cloudstackCredentials2Var,
	cloudstackCredentials3Var,
	cloudstackCredentialsForMultiLevelDomainVar,
	cloudstackAccountVar,
	cloudstackNetworkVar,
	cloudstackNetwork2Var,
	cloudstackNetwork3Var,
	cloudstackManagementServerVar,
	cloudstackManagementServer2Var,
	cloudstackManagementServer3Var,
	cloudstackSSHAuthorizedKeyVar,
	cloudstackComputeOfferingLargeVar,
	cloudstackComputeOfferingLargerVar,
	cloudStackCidrVar,
	podCidrVar,
	serviceCidrVar,
	cloudstackFeatureGateEnvVar,
	cloudstackB64EncodedSecretEnvVar,
}

type CloudStack struct {
	t                 *testing.T
	fillers           []api.CloudStackFiller
	clusterFillers    []api.ClusterFiller
	cidr              string
	podCidr           string
	serviceCidr       string
	cmkClient         *executables.Cmk
	devRelease        *releasev1.EksARelease
	templatesRegistry *templateRegistry
}

type CloudStackOpt func(*CloudStack)

func UpdateLargerCloudStackComputeOffering() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargerVar, api.WithCloudStackComputeOfferingForAllMachines)
}

// UpdateAddCloudStackAz4 add availability zone 4 to the cluster spec.
func UpdateAddCloudStackAz4() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountForMultiLevelDomainVar, cloudstackMultiLevelDomainVar, cloudstackZoneVar, cloudstackCredentialsForMultiLevelDomainVar, cloudstackNetworkVar,
		cloudstackManagementServerVar, api.WithCloudStackAz)
}

// UpdateAddCloudStackAz3 add availability zone 3 to the cluster spec.
func UpdateAddCloudStackAz3() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZone3Var, cloudstackCredentials3Var, cloudstackNetwork3Var,
		cloudstackManagementServer3Var, api.WithCloudStackAz)
}

func UpdateAddCloudStackAz2() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZone2Var, cloudstackCredentials2Var, cloudstackNetwork2Var,
		cloudstackManagementServer2Var, api.WithCloudStackAz)
}

func UpdateAddCloudStackAz1() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZoneVar, cloudstackCredentialsVar, cloudstackNetworkVar,
		cloudstackManagementServerVar, api.WithCloudStackAz)
}

func RemoveAllCloudStackAzs() api.CloudStackFiller {
	return api.RemoveCloudStackAzs()
}

// CloudStackCredentialsAz1 returns the value of the environment variable for cloudstackCredentialsVar.
func CloudStackCredentialsAz1() string {
	return os.Getenv(cloudstackCredentialsVar)
}

func NewCloudStack(t *testing.T, opts ...CloudStackOpt) *CloudStack {
	checkRequiredEnvVars(t, requiredCloudStackEnvVars)
	cmk := buildCmk(t)
	c := &CloudStack{
		t:         t,
		cmkClient: cmk,
		fillers: []api.CloudStackFiller{
			api.RemoveCloudStackAzs(),
			api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZoneVar, cloudstackCredentialsVar, cloudstackNetworkVar,
				cloudstackManagementServerVar, api.WithCloudStackAz),
			api.WithCloudStackStringFromEnvVar(cloudstackSSHAuthorizedKeyVar, api.WithCloudStackSSHAuthorizedKey),
			api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargeVar, api.WithCloudStackComputeOfferingForAllMachines),
		},
	}

	c.cidr = os.Getenv(cloudStackCidrVar)
	c.podCidr = os.Getenv(podCidrVar)
	c.serviceCidr = os.Getenv(serviceCidrVar)
	c.templatesRegistry = &templateRegistry{cache: map[string]string{}, generator: c}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithCloudStackWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.CloudStackMachineConfigFiller) CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers, cloudStackMachineConfig(name, fillers...))

		c.clusterFillers = append(c.clusterFillers, buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup))
	}
}

// WithKubeVersionAndOS returns a CloudStack Opt that adds API fillers to use a CloudStack template for
// the specified OS family and version (default if not provided), corresponding to a particular
// Kubernetes version, in addition to configuring all machine configs to use this OS family.
func withCloudStackKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			c.templateForKubeVersionAndOS(kubeVersion, os, release),
		)
	}
}

// WithCloudStackRedhat125 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.25.
func WithCloudStackRedhat125() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube125, RedHat8, nil)
}

// WithCloudStackRedhat126 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.26.
func WithCloudStackRedhat126() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube126, RedHat8, nil)
}

// WithCloudStackRedhat127 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.27.
func WithCloudStackRedhat127() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube127, RedHat8, nil)
}

// WithCloudStackRedhat128 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.28.
func WithCloudStackRedhat128() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube128, RedHat8, nil)
}

// WithCloudStackRedhat9Kubernetes125 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.25.
func WithCloudStackRedhat9Kubernetes125() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube125, RedHat9, nil)
}

// WithCloudStackRedhat9Kubernetes126 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.26.
func WithCloudStackRedhat9Kubernetes126() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube126, RedHat9, nil)
}

// WithCloudStackRedhat9Kubernetes127 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.27.
func WithCloudStackRedhat9Kubernetes127() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube127, RedHat9, nil)
}

// WithCloudStackRedhat9Kubernetes128 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.28.
func WithCloudStackRedhat9Kubernetes128() CloudStackOpt {
	return withCloudStackKubeVersionAndOS(anywherev1.Kube128, RedHat9, nil)
}

func WithCloudStackFillers(fillers ...api.CloudStackFiller) CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers, fillers...)
	}
}

func (c *CloudStack) Name() string {
	return "cloudstack"
}

func (c *CloudStack) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (c *CloudStack) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (c *CloudStack) ClusterConfigUpdates() []api.ClusterConfigFiller {
	controlPlaneIP, err := GetIP(c.cidr, ClusterIPPoolEnvVar)
	if err != nil {
		c.t.Fatalf("failed to get cluster ip for test environment: %v", err)
	}

	f := make([]api.ClusterFiller, 0, len(c.clusterFillers)+3)
	f = append(f, c.clusterFillers...)
	f = append(f,
		api.WithPodCidr(os.Getenv(podCidrVar)),
		api.WithServiceCidr(os.Getenv(serviceCidrVar)),
		api.WithControlPlaneEndpointIP(controlPlaneIP))

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.CloudStackToConfigFiller(c.fillers...)}
}

func (c *CloudStack) CleanupVMs(clusterName string) error {
	return cleanup.CleanUpCloudstackTestResources(context.Background(), clusterName, false)
}

func (c *CloudStack) WithProviderUpgrade(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.CloudStackToConfigFiller(fillers...))
	}
}

func (c *CloudStack) WithProviderUpgradeGit(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.CloudStackToConfigFiller(fillers...))
	}
}

func (c *CloudStack) getControlPlaneIP() (string, error) {
	value, ok := os.LookupEnv(cloudStackClusterIPPoolEnvVar)
	var clusterIP string
	var err error
	if ok && value != "" {
		clusterIP, err = PopIPFromEnv(cloudStackClusterIPPoolEnvVar)
		if err != nil {
			c.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
		}
	} else {
		clusterIP, err = GenerateUniqueIp(c.cidr)
		if err != nil {
			c.t.Fatalf("failed to generate ip for cloudstack %s: %v", c.cidr, err)
		}
	}
	return clusterIP, nil
}

func RequiredCloudstackEnvVars() []string {
	return requiredCloudStackEnvVars
}

func (c *CloudStack) WithNewCloudStackWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.CloudStackMachineConfigFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(
			api.CloudStackToConfigFiller(cloudStackMachineConfig(name, fillers...)),
			api.ClusterToConfigFiller(buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		)
	}
}

// WithNewWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding CloudStackMachineConfig to the cluster config.
func (c *CloudStack) WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.CloudStackToConfigFiller(cloudStackMachineConfig(name)),
		api.ClusterToConfigFiller(buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
	)
}

// WithWorkerNodeGroupConfiguration returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration item to the cluster config.
func (c *CloudStack) WithWorkerNodeGroupConfiguration(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.ClusterToConfigFiller(buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup))
}

func cloudStackMachineConfig(name string, fillers ...api.CloudStackMachineConfigFiller) api.CloudStackFiller {
	f := make([]api.CloudStackMachineConfigFiller, 0, len(fillers)+2)
	// Need to add these because at this point the default fillers that assign these
	// values to all machines have already ran
	f = append(f,
		api.WithCloudStackComputeOffering(os.Getenv(cloudstackComputeOfferingLargeVar)),
		api.WithCloudStackSSHKey(os.Getenv(cloudstackSSHAuthorizedKeyVar)),
	)
	f = append(f, fillers...)

	return api.WithCloudStackMachineConfig(name, f...)
}

// templateForKubeVersionAndOS returns a CloudStack filler for the given OS and Kubernetes version.
func (c *CloudStack) templateForKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.CloudStackFiller {
	var template string
	if release == nil {
		template = c.templateForDevRelease(kubeVersion, os)
	} else {
		template = c.templatesRegistry.templateForRelease(c.t, release, kubeVersion, os)
	}

	return api.WithCloudStackTemplateForAllMachines(template)
}

// Redhat125Template returns cloudstack filler for 1.25 RedHat.
func (c *CloudStack) Redhat125Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube125, RedHat8, nil)
}

// Redhat126Template returns cloudstack filler for 1.26 RedHat.
func (c *CloudStack) Redhat126Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube126, RedHat8, nil)
}

// Redhat127Template returns cloudstack filler for 1.27 RedHat.
func (c *CloudStack) Redhat127Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube127, RedHat8, nil)
}

// Redhat128Template returns cloudstack filler for 1.28 RedHat.
func (c *CloudStack) Redhat128Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube128, RedHat8, nil)
}

// Redhat9Kubernetes125Template returns cloudstack filler for 1.25 RedHat.
func (c *CloudStack) Redhat9Kubernetes125Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube125, RedHat9, nil)
}

// Redhat9Kubernetes126Template returns cloudstack filler for 1.26 RedHat.
func (c *CloudStack) Redhat9Kubernetes126Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube126, RedHat9, nil)
}

// Redhat9Kubernetes127Template returns cloudstack filler for 1.27 RedHat.
func (c *CloudStack) Redhat9Kubernetes127Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube127, RedHat9, nil)
}

// Redhat9Kubernetes128Template returns cloudstack filler for 1.28 RedHat.
func (c *CloudStack) Redhat9Kubernetes128Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS(anywherev1.Kube128, RedHat9, nil)
}

func buildCloudStackWorkerNodeGroupClusterFiller(machineConfigName string, workerNodeGroup *WorkerNodeGroup) api.ClusterFiller {
	// Set worker node group ref to cloudstack machine config
	workerNodeGroup.MachineConfigKind = anywherev1.CloudStackMachineConfigKind
	workerNodeGroup.MachineConfigName = machineConfigName
	return workerNodeGroup.ClusterFiller()
}

// ClusterStateValidations returns a list of provider specific validations.
func (c *CloudStack) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{
		clusterf.RetriableStateValidation(
			retrier.NewWithMaxRetries(60, 5*time.Second),
			validations.ValidateAvailabilityZones,
		),
	}
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right template for all
// cloudstack machine configs.
func (c *CloudStack) WithKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
		api.CloudStackToConfigFiller(
			c.templateForKubeVersionAndOS(kubeVersion, os, release),
		),
	)
}

// WithRedhat125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat125() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube125, RedHat8, nil)
}

// WithRedhat126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat126() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube126, RedHat8, nil)
}

// WithRedhat127 returns a cluster config filler that sets the kubernetes version of the cluster to 1.27
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat127() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube127, RedHat8, nil)
}

// WithRedhat128 returns a cluster config filler that sets the kubernetes version of the cluster to 1.28
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat128() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube128, RedHat8, nil)
}

// WithRedhat9Kubernetes125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat9Kubernetes125() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube125, RedHat9, nil)
}

// WithRedhat9Kubernetes126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat9Kubernetes126() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube126, RedHat9, nil)
}

// WithRedhat9Kubernetes127 returns a cluster config filler that sets the kubernetes version of the cluster to 1.27
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat9Kubernetes127() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube127, RedHat9, nil)
}

// WithRedhat9Kubernetes128 returns a cluster config filler that sets the kubernetes version of the cluster to 1.28
// as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhat9Kubernetes128() api.ClusterConfigFiller {
	return c.WithKubeVersionAndOS(anywherev1.Kube128, RedHat9, nil)
}

// WithRedhatVersion returns a cluster config filler that sets the kubernetes version of the cluster to the k8s
// version provider, as well as the right redhat template for all CloudStackMachineConfigs.
func (c *CloudStack) WithRedhatVersion(version anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	switch version {
	case anywherev1.Kube125:
		return c.WithRedhat125()
	case anywherev1.Kube126:
		return c.WithRedhat126()
	case anywherev1.Kube127:
		return c.WithRedhat127()
	case anywherev1.Kube128:
		return c.WithRedhat128()
	default:
		return nil
	}
}

func (c *CloudStack) getDevRelease() *releasev1.EksARelease {
	c.t.Helper()
	if c.devRelease == nil {
		latestRelease, err := getLatestDevRelease()
		if err != nil {
			c.t.Fatal(err)
		}
		c.devRelease = latestRelease
	}

	return c.devRelease
}

func (c *CloudStack) templateForDevRelease(kubeVersion anywherev1.KubernetesVersion, os OS) string {
	c.t.Helper()
	return c.templatesRegistry.templateForRelease(c.t, c.getDevRelease(), kubeVersion, os)
}

// envVarForTemplate Looks for explicit configuration through an env var: "T_CLOUDSTACK_TEMPLATE_{osFamily}_{eks-d version}"
// eg: T_CLOUDSTACK_TEMPLATE_REDHAT_KUBERNETES_1_23_EKS_22.
func (c *CloudStack) envVarForTemplate(os OS, eksDName string) string {
	return fmt.Sprintf("T_CLOUDSTACK_TEMPLATE_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ToUpper(strings.ReplaceAll(eksDName, "-", "_")))
}

// defaultNameForTemplate looks for a template: "{eks-d version}-{osFamily}"
// eg: kubernetes-1-23-eks-22-redhat.
func (c *CloudStack) defaultNameForTemplate(os OS, eksDName string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(eksDName), strings.ToLower(string(os)))
}

// defaultEnvVarForTemplate returns the value of the default template env vars: "T_CLOUDSTACK_TEMPLATE_{osFamily}_{kubeVersion}"
// eg. T_CLOUDSTACK_TEMPLATE_REDHAT_1_23.
func (c *CloudStack) defaultEnvVarForTemplate(os OS, kubeVersion anywherev1.KubernetesVersion) string {
	if osFamiliesForOS[os] == anywherev1.Bottlerocket {
		os = OS(strings.ReplaceAll(string(os), "bottlerocket", "br"))
	}
	return fmt.Sprintf("T_CLOUDSTACK_TEMPLATE_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ReplaceAll(string(kubeVersion), ".", "_"))
}

// searchTemplate returns template name if the given template exists in the datacenter.
func (c *CloudStack) searchTemplate(ctx context.Context, template string) (string, error) {
	profile, ok := os.LookupEnv(cloudstackCredentialsVar)
	if !ok {
		return "", fmt.Errorf("Required environment variable for CloudStack not set: %s", cloudstackCredentialsVar)
	}
	templateResource := v1alpha1.CloudStackResourceIdentifier{
		Name: template,
	}
	template, err := c.cmkClient.SearchTemplate(context.Background(), profile, templateResource)
	if err != nil {
		return "", err
	}
	return template, nil
}
