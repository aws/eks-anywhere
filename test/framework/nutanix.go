package framework

import (
	"context"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	nutanixEndpoint              = "T_NUTANIX_ENDPOINT"
	nutanixPort                  = "T_NUTANIX_PORT"
	nutanixAdditionalTrustBundle = "T_NUTANIX_ADDITIONAL_TRUST_BUNDLE"
	nutanixInsecure              = "T_NUTANIX_INSECURE"

	nutanixMachineBootType       = "T_NUTANIX_MACHINE_BOOT_TYPE"
	nutanixMachineMemorySize     = "T_NUTANIX_MACHINE_MEMORY_SIZE"
	nutanixSystemDiskSize        = "T_NUTANIX_SYSTEMDISK_SIZE"
	nutanixMachineVCPUsPerSocket = "T_NUTANIX_MACHINE_VCPU_PER_SOCKET"
	nutanixMachineVCPUSocket     = "T_NUTANIX_MACHINE_VCPU_SOCKET"

	nutanixPrismElementClusterName = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME"
	nutanixPrismElementClusterUUID = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_UUID"
	nutanixSSHAuthorizedKey        = "T_NUTANIX_SSH_AUTHORIZED_KEY"

	nutanixSubnetName = "T_NUTANIX_SUBNET_NAME"
	nutanixSubnetUUID = "T_NUTANIX_SUBNET_UUID"

	nutanixControlPlaneEndpointIP = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar    = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar             = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar         = "T_NUTANIX_SERVICE_CIDR"

	nutanixTemplateNameUbuntu122Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_22"
	nutanixTemplateNameUbuntu123Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_23"
	nutanixTemplateNameUbuntu124Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_24"
	nutanixTemplateNameUbuntu125Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_25"
	nutanixTemplateNameUbuntu126Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_26"

	nutanixTemplateUUIDUbuntu122Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_22"
	nutanixTemplateUUIDUbuntu123Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_23"
	nutanixTemplateUUIDUbuntu124Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_24"
	nutanixTemplateUUIDUbuntu125Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_25"
	nutanixTemplateUUIDUbuntu126Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_26"
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
	nutanixPrismElementClusterUUID,
	nutanixSSHAuthorizedKey,
	nutanixSubnetName,
	nutanixSubnetUUID,
	nutanixPodCidrVar,
	nutanixServiceCidrVar,
	nutanixTemplateNameUbuntu122Var,
	nutanixTemplateNameUbuntu123Var,
	nutanixTemplateNameUbuntu124Var,
	nutanixTemplateNameUbuntu125Var,
	nutanixTemplateNameUbuntu126Var,
	nutanixTemplateUUIDUbuntu122Var,
	nutanixTemplateUUIDUbuntu123Var,
	nutanixTemplateUUIDUbuntu124Var,
	nutanixTemplateUUIDUbuntu125Var,
	nutanixTemplateUUIDUbuntu126Var,
	nutanixInsecure,
}

type Nutanix struct {
	t                      *testing.T
	fillers                []api.NutanixFiller
	clusterFillers         []api.ClusterFiller
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
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterUUID, api.WithNutanixPrismElementClusterUUID),
			api.WithNutanixStringFromEnvVar(nutanixSubnetName, api.WithNutanixSubnetName),
			api.WithNutanixStringFromEnvVar(nutanixSubnetUUID, api.WithNutanixSubnetUUID),
		},
	}

	nutanixProvider.controlPlaneEndpointIP = os.Getenv(nutanixControlPlaneEndpointIP)
	nutanixProvider.cpCidr = os.Getenv(nutanixControlPlaneCidrVar)
	nutanixProvider.podCidr = os.Getenv(nutanixPodCidrVar)
	nutanixProvider.serviceCidr = os.Getenv(nutanixServiceCidrVar)

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

// WithUbuntu122Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.22
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu122Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu122Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
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

// WithUbuntu122NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.22
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu122NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu122Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu123NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu123Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu124NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu124Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu125NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.25
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu125NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu125Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu126NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.26
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu126NutanixUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu126Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithPrismElementClusterUUID returns a NutanixOpt that adds API fillers to use a PE Cluster UUID.
func WithPrismElementClusterUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterUUID, api.WithNutanixPrismElementClusterUUID),
		)
	}
}

// WithSubnetUUID returns a NutanixOpt that adds API fillers to use a Subnet UUID.
func WithSubnetUUID() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixSubnetUUID, api.WithNutanixSubnetUUID),
		)
	}
}

// UpdateNutanixUbuntuTemplate122Var returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func UpdateNutanixUbuntuTemplate122Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu122Var, api.WithNutanixMachineTemplateImageName)
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

// ClusterStateValidations returns a list of provider specific ClusterStateValidations.
func (s *Nutanix) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}
