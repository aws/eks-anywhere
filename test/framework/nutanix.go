package framework

import (
	"context"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/nutanix"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	nutanixEndpoint                 = "T_NUTANIX_ENDPOINT"
	nutanixPort                     = "T_NUTANIX_PORT"
	nutanixAdditionalTrustBundle    = "T_NUTANIX_ADDITIONAL_TRUST_BUNDLE"
	nutanixInsecure                 = "T_NUTANIX_INSECURE"
	nutanixMachineBootType          = "T_NUTANIX_MACHINE_BOOT_TYPE"
	nutanixMachineMemorySize        = "T_NUTANIX_MACHINE_MEMORY_SIZE"
	nutanixSystemDiskSize           = "T_NUTANIX_SYSTEMDISK_SIZE"
	nutanixMachineVCPUsPerSocket    = "T_NUTANIX_MACHINE_VCPU_PER_SOCKET"
	nutanixMachineVCPUSocket        = "T_NUTANIX_MACHINE_VCPU_SOCKET"
	nutanixPrismElementClusterName  = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME"
	nutanixSSHAuthorizedKey         = "T_NUTANIX_SSH_AUTHORIZED_KEY"
	nutanixSubnetName               = "T_NUTANIX_SUBNET_NAME"
	nutanixControlPlaneEndpointIP   = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar      = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar               = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar           = "T_NUTANIX_SERVICE_CIDR"
	nutanixTemplateNameUbuntu123Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_23"
	nutanixTemplateNameUbuntu124Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_24"
	nutanixTemplateNameUbuntu125Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_25"
	nutanixTemplateNameUbuntu126Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_26"
	nutanixTemplateNameUbuntu127Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_27"
)

var requiredNutanixEnvVars = []string{
	constants.EksaNutanixUsernameKey,
	constants.EksaNutanixPasswordKey,
	nutanixEndpoint,
	nutanixPort,
	nutanixAdditionalTrustBundle,
	nutanixMachineBootType,
	nutanixMachineMemorySize,
	nutanixSystemDiskSize,
	nutanixMachineVCPUsPerSocket,
	nutanixMachineVCPUSocket,
	nutanixPrismElementClusterName,
	nutanixSSHAuthorizedKey,
	nutanixSubnetName,
	nutanixPodCidrVar,
	nutanixServiceCidrVar,
	nutanixTemplateNameUbuntu123Var,
	nutanixTemplateNameUbuntu124Var,
	nutanixTemplateNameUbuntu125Var,
	nutanixTemplateNameUbuntu126Var,
	nutanixTemplateNameUbuntu127Var,
	nutanixInsecure,
}

type Nutanix struct {
	t                      *testing.T
	fillers                []api.NutanixFiller
	clusterFillers         []api.ClusterFiller
	client                 nutanix.PrismClient
	controlPlaneEndpointIP string
	cpCidr                 string
	podCidr                string
	serviceCidr            string
}

type NutanixOpt func(*Nutanix)

func NewNutanix(t *testing.T, opts ...NutanixOpt) *Nutanix {
	checkRequiredEnvVars(t, requiredNutanixEnvVars)

	nutanixProvider := &Nutanix{
		t: t,
		fillers: []api.NutanixFiller{
			api.WithNutanixStringFromEnvVar(nutanixEndpoint, api.WithNutanixEndpoint),
			api.WithNutanixIntFromEnvVar(nutanixPort, api.WithNutanixPort),
			api.WithNutanixStringFromEnvVar(nutanixAdditionalTrustBundle, api.WithNutanixAdditionalTrustBundle),
			api.WithNutanixStringFromEnvVar(nutanixMachineMemorySize, api.WithNutanixMachineMemorySize),
			api.WithNutanixStringFromEnvVar(nutanixSystemDiskSize, api.WithNutanixMachineSystemDiskSize),
			api.WithNutanixInt32FromEnvVar(nutanixMachineVCPUsPerSocket, api.WithNutanixMachineVCPUsPerSocket),
			api.WithNutanixInt32FromEnvVar(nutanixMachineVCPUSocket, api.WithNutanixMachineVCPUSocket),
			api.WithNutanixStringFromEnvVar(nutanixSSHAuthorizedKey, api.WithNutanixSSHAuthorizedKey),
			api.WithNutanixBoolFromEnvVar(nutanixInsecure, api.WithNutanixInsecure),
			// Assumption: generated clusterconfig by nutanix provider sets name as id type by default.
			// for uuid specific id type, we will set it thru each specific test so that current CI
			// works as is with name id type for following resources
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterName, api.WithNutanixPrismElementClusterName),
			api.WithNutanixStringFromEnvVar(nutanixSubnetName, api.WithNutanixSubnetName),
		},
	}

	nutanixProvider.controlPlaneEndpointIP = os.Getenv(nutanixControlPlaneEndpointIP)
	nutanixProvider.cpCidr = os.Getenv(nutanixControlPlaneCidrVar)
	nutanixProvider.podCidr = os.Getenv(nutanixPodCidrVar)
	nutanixProvider.serviceCidr = os.Getenv(nutanixServiceCidrVar)
	client, err := nutanix.NewPrismClient(os.Getenv(nutanixEndpoint), os.Getenv(nutanixPort), true)
	if err != nil {
		t.Fatalf("Failed to initialize Nutanix Prism Client: %v", err)
	}
	nutanixProvider.client = client

	for _, opt := range opts {
		opt(nutanixProvider)
	}

	return nutanixProvider
}

// RequiredNutanixEnvVars returns a list of environment variables needed for Nutanix tests.
func RequiredNutanixEnvVars() []string {
	return requiredNutanixEnvVars
}

func (s *Nutanix) Name() string {
	return "nutanix"
}

func (s *Nutanix) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (s *Nutanix) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// CleanupVMs satisfies the test framework Provider.
func (s *Nutanix) CleanupVMs(clustername string) error {
	return cleanup.NutanixTestResourcesCleanup(context.Background(), clustername, os.Getenv(nutanixEndpoint), os.Getenv(nutanixPort), true, true)
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (s *Nutanix) ClusterConfigUpdates() []api.ClusterConfigFiller {
	f := make([]api.ClusterFiller, 0, len(s.clusterFillers)+3)
	f = append(f, s.clusterFillers...)
	if s.controlPlaneEndpointIP != "" {
		f = append(f, api.WithControlPlaneEndpointIP(s.controlPlaneEndpointIP))
	} else {
		clusterIP, err := GetIP(s.cpCidr, ClusterIPPoolEnvVar)
		if err != nil {
			s.t.Fatalf("failed to get cluster ip for test environment: %v", err)
		}
		f = append(f, api.WithControlPlaneEndpointIP(clusterIP))
	}

	if s.podCidr != "" {
		f = append(f, api.WithPodCidr(s.podCidr))
	}

	if s.serviceCidr != "" {
		f = append(f, api.WithServiceCidr(s.serviceCidr))
	}

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.NutanixToConfigFiller(s.fillers...)}
}

func (s *Nutanix) WithProviderUpgrade(fillers ...api.NutanixFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.NutanixToConfigFiller(fillers...))
	}
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right template for all
// nutanix machine configs.
func (s *Nutanix) WithKubeVersionAndOS(osFamily anywherev1.OSFamily, kubeVersion anywherev1.KubernetesVersion, release *releasev1.EksARelease) api.ClusterConfigFiller {
	// TODO: Update tests to use this
	panic("Not implemented for Nutanix yet")
}

// WithNewWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding NutanixMachineConfig to the cluster config.
func (s *Nutanix) WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	// TODO: Implement for Nutanix provider
	panic("Not implemented for Nutanix yet")
}

// WithUbuntu123Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu123Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu124Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu124Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu125Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.25
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu125Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu125Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu126Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.26
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu126Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu126Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu127Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.27
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu127Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu127Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu123NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixTemplateNameUbuntu123Var)
		v.fillers = append(v.fillers, v.withUbuntuNutanixUUID(name)...)
	}
}

// WithUbuntu124NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixTemplateNameUbuntu124Var)
		v.fillers = append(v.fillers, v.withUbuntuNutanixUUID(name)...)
	}
}

// WithUbuntu125NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.25
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu125NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixTemplateNameUbuntu125Var)
		v.fillers = append(v.fillers, v.withUbuntuNutanixUUID(name)...)
	}
}

// WithUbuntu126NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.26
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu126NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixTemplateNameUbuntu126Var)
		v.fillers = append(v.fillers, v.withUbuntuNutanixUUID(name)...)
	}
}

// WithUbuntu127NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.27
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu127NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixTemplateNameUbuntu127Var)
		v.fillers = append(v.fillers, v.withUbuntuNutanixUUID(name)...)
	}
}

func (s *Nutanix) withUbuntuNutanixUUID(name string) []api.NutanixFiller {
	uuid, err := s.client.GetImageUUIDFromName(context.Background(), name)
	if err != nil {
		s.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
	}
	return append([]api.NutanixFiller{},
		api.WithNutanixMachineTemplateImageUUID(*uuid),
		api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
	)
}

// WithPrismElementClusterUUID returns a NutanixOpt that adds API fillers to use a PE Cluster UUID.
func WithPrismElementClusterUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixPrismElementClusterName)
		uuid, err := v.client.GetClusterUUIDFromName(context.Background(), name)
		if err != nil {
			v.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
		}
		v.fillers = append(v.fillers, api.WithNutanixPrismElementClusterUUID(*uuid))
	}
}

// WithNutanixSubnetUUID returns a NutanixOpt that adds API fillers to use a Subnet UUID.
func WithNutanixSubnetUUID() NutanixOpt {
	return func(v *Nutanix) {
		name := os.Getenv(nutanixSubnetName)
		uuid, err := v.client.GetSubnetUUIDFromName(context.Background(), name)
		if err != nil {
			v.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
		}
		v.fillers = append(v.fillers, api.WithNutanixSubnetUUID(*uuid))
	}
}

// UpdateNutanixUbuntuTemplate123Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate123Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu123Var, api.WithNutanixMachineTemplateImageName)
}

// UpdateNutanixUbuntuTemplate124Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate124Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu124Var, api.WithNutanixMachineTemplateImageName)
}

// UpdateNutanixUbuntuTemplate125Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate125Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu125Var, api.WithNutanixMachineTemplateImageName)
}

// UpdateNutanixUbuntuTemplate126Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate126Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu126Var, api.WithNutanixMachineTemplateImageName)
}

// UpdateNutanixUbuntuTemplate127Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate127Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu127Var, api.WithNutanixMachineTemplateImageName)
}

// ClusterStateValidations returns a list of provider specific ClusterStateValidations.
func (s *Nutanix) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}
