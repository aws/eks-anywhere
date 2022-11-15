package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	nutanixFeatureGateEnvVar     = "NUTANIX_PROVIDER"
	nutanixEndpoint              = "T_NUTANIX_ENDPOINT"
	nutanixPort                  = "T_NUTANIX_PORT"
	nutanixAdditionalTrustBundle = "T_NUTANIX_ADDITIONAL_TRUST_BUNDLE"

	nutanixMachineBootType       = "T_NUTANIX_MACHINE_BOOT_TYPE"
	nutanixMachineMemorySize     = "T_NUTANIX_MACHINE_MEMORY_SIZE"
	nutanixSystemDiskSize        = "T_NUTANIX_SYSTEMDISK_SIZE"
	nutanixMachineVCPUsPerSocket = "T_NUTANIX_MACHINE_VCPU_PER_SOCKET"
	nutanixMachineVCPUSocket     = "T_NUTANIX_MACHINE_VCPU_SOCKET"

	nutanixMachineTemplateImageName = "T_NUTANIX_MACHINE_TEMPLATE_IMAGE_NAME"
	nutanixPrismElementClusterName  = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME"
	nutanixSSHAuthorizedKey         = "T_NUTANIX_SSH_AUTHORIZED_KEY"
	nutanixSubnetName               = "T_NUTANIX_SUBNET_NAME"

	nutanixControlPlaneEndpointIP = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar    = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar             = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar         = "T_NUTANIX_SERVICE_CIDR"

	nutanixTemplateUbuntu120Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_20"
	nutanixTemplateUbuntu121Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_21"
	nutanixTemplateUbuntu122Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_22"
	nutanixTemplateUbuntu123Var = "T_NUTANIX_TEMPLATE_UBUNTU_1_23"
)

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
	requiredNutanixEnvVars := []string{
		nutanixFeatureGateEnvVar,
		nutanixEndpoint,
		nutanixPort,
		nutanixAdditionalTrustBundle,
		nutanixMachineBootType,
		nutanixMachineMemorySize,
		nutanixSystemDiskSize,
		nutanixMachineVCPUsPerSocket,
		nutanixMachineVCPUSocket,
		nutanixMachineTemplateImageName,
		nutanixPrismElementClusterName,
		nutanixSSHAuthorizedKey,
		nutanixSubnetName,
		nutanixPodCidrVar,
		nutanixServiceCidrVar,
	}
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
			api.WithNutanixStringFromEnvVar(nutanixMachineTemplateImageName, api.WithNutanixMachineTemplateImageName),
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterName, api.WithNutanixPrismElementClusterName),
			api.WithNutanixStringFromEnvVar(nutanixSSHAuthorizedKey, api.WithNutanixSSHAuthorizedKey),
			api.WithNutanixStringFromEnvVar(nutanixSubnetName, api.WithNutanixSubnetName),
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
