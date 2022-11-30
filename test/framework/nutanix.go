package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	nutanixFeatureGateEnvVar     = "NUTANIX_PROVIDER"
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
	nutanixSSHAuthorizedKey        = "T_NUTANIX_SSH_AUTHORIZED_KEY"
	nutanixSubnetName              = "T_NUTANIX_SUBNET_NAME"

	nutanixControlPlaneEndpointIP = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar    = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar             = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar         = "T_NUTANIX_SERVICE_CIDR"

	nutanixTemplateUbuntu120Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_20"
	nutanixTemplateUbuntu121Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_21"
	nutanixTemplateUbuntu122Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_22"
	nutanixTemplateUbuntu123Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_23"
	nutanixTemplateUbuntu124Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_24"
)

var requiredNutanixEnvVars = []string{
	constants.NutanixUsernameKey,
	constants.NutanixPasswordKey,
	nutanixFeatureGateEnvVar,
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
	nutanixTemplateUbuntu120Var,
	nutanixTemplateUbuntu121Var,
	nutanixTemplateUbuntu122Var,
	nutanixTemplateUbuntu123Var,
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
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterName, api.WithNutanixPrismElementClusterName),
			api.WithNutanixStringFromEnvVar(nutanixSSHAuthorizedKey, api.WithNutanixSSHAuthorizedKey),
			api.WithNutanixStringFromEnvVar(nutanixSubnetName, api.WithNutanixSubnetName),
			api.WithNutanixBoolFromEnvVar(nutanixInsecure, api.WithNutanixInsecure),
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

func (s *Nutanix) CleanupVMs(_ string) error {
	return nil
}

func (s *Nutanix) CustomizeProviderConfig(file string) []byte {
	return s.customizeProviderConfig(file, s.fillers...)
}

func (s *Nutanix) ClusterConfigFillers() []api.ClusterFiller {
	if s.controlPlaneEndpointIP != "" {
		s.clusterFillers = append(s.clusterFillers, api.WithControlPlaneEndpointIP(s.controlPlaneEndpointIP))
	} else {
		clusterIP, err := GetIP(s.cpCidr, ClusterIPPoolEnvVar)
		if err != nil {
			s.t.Fatalf("failed to get cluster ip for test environment: %v", err)
		}
		s.clusterFillers = append(s.clusterFillers, api.WithControlPlaneEndpointIP(clusterIP))
	}

	if s.podCidr != "" {
		s.clusterFillers = append(s.clusterFillers, api.WithPodCidr(s.podCidr))
	}

	if s.serviceCidr != "" {
		s.clusterFillers = append(s.clusterFillers, api.WithServiceCidr(s.serviceCidr))
	}

	return s.clusterFillers
}

func (s *Nutanix) customizeProviderConfig(file string, fillers ...api.NutanixFiller) []byte {
	providerOutput, err := api.AutoFillNutanixProvider(file, fillers...)
	if err != nil {
		s.t.Fatalf("failed to customize provider config from file: %v", err)
	}
	return providerOutput
}

func (s *Nutanix) WithProviderUpgrade(fillers ...api.NutanixFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = s.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

// WithUbuntu120Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.20
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu120Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu120Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu121Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.21
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu121Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu121Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu122Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.22
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu122Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu122Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu123Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu123Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu124Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu124Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
}

func UpdateNutanixUbuntuTemplate120Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu120Var, api.WithNutanixMachineTemplateImageName)
}

func UpdateNutanixUbuntuTemplate121Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu121Var, api.WithNutanixMachineTemplateImageName)
}

func UpdateNutanixUbuntuTemplate122Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu122Var, api.WithNutanixMachineTemplateImageName)
}

func UpdateNutanixUbuntuTemplate123Var() api.NutanixFiller {
	return api.WithNutanixStringFromEnvVar(nutanixTemplateUbuntu123Var, api.WithNutanixMachineTemplateImageName)
}
