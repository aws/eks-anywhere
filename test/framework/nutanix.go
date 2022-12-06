package framework

import (
	"fmt"
	"os"
	"strings"
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

	nutanixPrismElementClusterIDType = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_ID_TYPE"
	nutanixPrismElementClusterName   = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME"
	nutanixPrismElementClusterUUID   = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_UUID"
	nutanixSSHAuthorizedKey          = "T_NUTANIX_SSH_AUTHORIZED_KEY"
	nutanixSubnetIDType              = "T_NUTANIX_SUBNET_ID_TYPE"
	nutanixSubnetName                = "T_NUTANIX_SUBNET_NAME"
	nutanixSubnetUUID                = "T_NUTANIX_SUBNET_UUID"

	nutanixControlPlaneEndpointIP = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar    = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar             = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar         = "T_NUTANIX_SERVICE_CIDR"

	nutanixMachineTemplateIDTypeVar = "T_NUTANIX_MACHINE_TEMPLATE_ID_TYPE"

	nutanixTemplateNameUbuntu121Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_21"
	nutanixTemplateNameUbuntu122Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_22"
	nutanixTemplateNameUbuntu123Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_23"
	nutanixTemplateNameUbuntu124Var = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_24"

	nutanixTemplateUUIDUbuntu121Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_21"
	nutanixTemplateUUIDUbuntu122Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_22"
	nutanixTemplateUUIDUbuntu123Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_23"
	nutanixTemplateUUIDUbuntu124Var = "T_NUTANIX_TEMPLATE_UUID_UBUNTU_1_24"
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
	nutanixPrismElementClusterIDType,
	nutanixPrismElementClusterName,
	nutanixPrismElementClusterUUID,
	nutanixSSHAuthorizedKey,
	nutanixSubnetIDType,
	nutanixSubnetName,
	nutanixSubnetUUID,
	nutanixPodCidrVar,
	nutanixServiceCidrVar,
	nutanixMachineTemplateIDTypeVar,
	nutanixTemplateNameUbuntu121Var,
	nutanixTemplateNameUbuntu122Var,
	nutanixTemplateNameUbuntu123Var,
	nutanixTemplateNameUbuntu124Var,
	nutanixTemplateUUIDUbuntu121Var,
	nutanixTemplateUUIDUbuntu122Var,
	nutanixTemplateUUIDUbuntu123Var,
	nutanixTemplateUUIDUbuntu124Var,
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

// WithUbuntu121Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.21
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu121Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers, GetUbuntu121NutanixFillers()...)
	}
}

// WithUbuntu122Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.22
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu122Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers, GetUbuntu122NutanixFillers()...)
	}
}

// WithUbuntu123Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers, GetUbuntu123NutanixFillers()...)
	}
}

// WithUbuntu124Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124Nutanix() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers, GetUbuntu124NutanixFillers()...)
	}
}

// GetUbuntu121NutanixFillers returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func GetUbuntu121NutanixFillers() []api.NutanixFiller {
	var fillers []api.NutanixFiller
	nutanixMachineTemplateIDType := strings.Trim(os.Getenv(nutanixMachineTemplateIDTypeVar), "\"")
	nutanixMachineTemplateIDType = strings.Trim(nutanixMachineTemplateIDType, "'")
	if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierName) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu121Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	} else if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierUUID) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu121Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
	return fillers
}

// GetUbuntu122NutanixFillers returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func GetUbuntu122NutanixFillers() []api.NutanixFiller {
	var fillers []api.NutanixFiller
	nutanixMachineTemplateIDType := strings.Trim(os.Getenv(nutanixMachineTemplateIDTypeVar), "\"")
	nutanixMachineTemplateIDType = strings.Trim(nutanixMachineTemplateIDType, "'")
	if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierName) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu122Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	} else if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierUUID) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu122Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
	return fillers
}

// GetUbuntu123NutanixFillers returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func GetUbuntu123NutanixFillers() []api.NutanixFiller {
	var fillers []api.NutanixFiller
	nutanixMachineTemplateIDType := strings.Trim(os.Getenv(nutanixMachineTemplateIDTypeVar), "\"")
	nutanixMachineTemplateIDType = strings.Trim(nutanixMachineTemplateIDType, "'")
	if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierName) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu123Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	} else if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierUUID) {
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu123Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
	return fillers
}

// GetUbuntu124NutanixFillers returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func GetUbuntu124NutanixFillers() []api.NutanixFiller {
	fmt.Println("here 1")
	var fillers []api.NutanixFiller
	nutanixMachineTemplateIDType := strings.Trim(os.Getenv(nutanixMachineTemplateIDTypeVar), "\"")
	nutanixMachineTemplateIDType = strings.Trim(nutanixMachineTemplateIDType, "'")
	fmt.Printf("nutanixMachineTemplateIDType %s\n", nutanixMachineTemplateIDType)
	if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierName) {
		fmt.Println("here name")
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateNameUbuntu124Var, api.WithNutanixMachineTemplateImageName),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	} else if nutanixMachineTemplateIDType == string(anywherev1.NutanixIdentifierUUID) {
		fmt.Println("here uuid")
		fillers = append(fillers,
			api.WithNutanixStringFromEnvVar(nutanixTemplateUUIDUbuntu124Var, api.WithNutanixMachineTemplateImageUUID),
			api.WithOsFamilyForAllNutanixMachines(anywherev1.Ubuntu),
		)
	}
	return fillers
}
