package api

import (
	"encoding/base64"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

// NutanixConfig is a wrapper for the Nutanix provider spec.
type NutanixConfig struct {
	datacenterConfig *anywherev1.NutanixDatacenterConfig
	machineConfigs   map[string]*anywherev1.NutanixMachineConfig
}

type NutanixFiller func(config *NutanixConfig)

// NutanixToConfigFiller transforms a set of NutanixFiller's in a single ClusterConfigFiller.
func NutanixToConfigFiller(fillers ...NutanixFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		updateNutanix(c, fillers...)
	}
}

func updateNutanix(config *cluster.Config, fillers ...NutanixFiller) {
	nc := &NutanixConfig{
		datacenterConfig: config.NutanixDatacenter,
		machineConfigs:   config.NutanixMachineConfigs,
	}

	for _, f := range fillers {
		f(nc)
	}
}

// WithNutanixStringFromEnvVar returns a NutanixFiller that sets the given string value to the given environment variable.
func WithNutanixStringFromEnvVar(envVar string, opt func(string) NutanixFiller) NutanixFiller {
	return opt(os.Getenv(envVar))
}

// WithNutanixIntFromEnvVar returns a NutanixFiller that sets the given integer value to the given environment variable.
func WithNutanixIntFromEnvVar(envVar string, opt func(int) NutanixFiller) NutanixFiller {
	intVar, _ := strconv.Atoi(os.Getenv(envVar))
	return opt(intVar)
}

// WithNutanixInt32FromEnvVar returns a NutanixFiller that sets the given int32 value to the given environment variable.
func WithNutanixInt32FromEnvVar(envVar string, opt func(int32) NutanixFiller) NutanixFiller {
	intVar, _ := strconv.ParseInt(os.Getenv(envVar), 10, 32)
	return opt(int32(intVar))
}

// WithNutanixBoolFromEnvVar returns a NutanixFiller that sets the given int32 value to the given environment variable.
func WithNutanixBoolFromEnvVar(envVar string, opt func(bool) NutanixFiller) NutanixFiller {
	return opt(os.Getenv(envVar) == "true")
}

// WithNutanixEndpoint returns a NutanixFiller that sets the endpoint for the Nutanix provider.
func WithNutanixEndpoint(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		config.datacenterConfig.Spec.Endpoint = value
	}
}

// WithNutanixPort returns a NutanixFiller that sets the port for the Nutanix provider.
func WithNutanixPort(value int) NutanixFiller {
	return func(config *NutanixConfig) {
		config.datacenterConfig.Spec.Port = value
	}
}

// WithNutanixAdditionalTrustBundle returns a NutanixFiller that sets the additional trust bundle for the Nutanix provider.
func WithNutanixAdditionalTrustBundle(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		certificate, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			logger.Info("Warning: Failed to decode AdditionalTrustBundle. AdditionalTrustBundle won't be added")
		} else {
			config.datacenterConfig.Spec.AdditionalTrustBundle = string(certificate)
		}
	}
}

// WithNutanixInsecure returns a NutanixFiller that sets the insecure for the Nutanix provider.
func WithNutanixInsecure(value bool) NutanixFiller {
	return func(config *NutanixConfig) {
		config.datacenterConfig.Spec.Insecure = value
	}
}

// WithNutanixMachineMemorySize returns a NutanixFiller that sets the memory size for the Nutanix machine.
func WithNutanixMachineMemorySize(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.MemorySize = resource.MustParse(value)
		}
	}
}

// WithNutanixMachineSystemDiskSize returns a NutanixFiller that sets the system disk size for the Nutanix machine.
func WithNutanixMachineSystemDiskSize(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.SystemDiskSize = resource.MustParse(value)
		}
	}
}

// WithNutanixMachineVCPUsPerSocket returns a NutanixFiller that sets the vCPUs per socket for the Nutanix machine.
func WithNutanixMachineVCPUsPerSocket(value int32) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.VCPUsPerSocket = value
		}
	}
}

// WithNutanixMachineVCPUSocket returns a NutanixFiller that sets the vCPU sockets for the Nutanix machine.
func WithNutanixMachineVCPUSocket(value int32) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.VCPUSockets = value
		}
	}
}

// WithNutanixMachineTemplateImageName returns a NutanixFiller that sets the image name for the Nutanix machine template.
func WithNutanixMachineTemplateImageName(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Image = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

// WithOsFamilyForAllNutanixMachines sets the osFamily for all Nutanix machines to value.
func WithOsFamilyForAllNutanixMachines(value anywherev1.OSFamily) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.OSFamily = value
		}
	}
}

// WithNutanixSubnetName returns a NutanixFiller that sets the subnet name for the Nutanix machine.
func WithNutanixSubnetName(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Subnet = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

// WithNutanixPrismElementClusterName returns a NutanixFiller that sets the cluster name for the Nutanix machine.
func WithNutanixPrismElementClusterName(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Cluster = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

// WithNutanixMachineTemplateImageUUID returns a NutanixFiller that sets the image UUID for the Nutanix machine.
func WithNutanixMachineTemplateImageUUID(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Image = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierUUID, UUID: &value}
		}
	}
}

// WithNutanixSubnetUUID returns a NutanixFiller that sets the subnet UUID for the Nutanix machine.
func WithNutanixSubnetUUID(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Subnet = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierUUID, UUID: &value}
		}
	}
}

// WithNutanixPrismElementClusterUUID returns a NutanixFiller that sets the cluster UUID for the Nutanix machine.
func WithNutanixPrismElementClusterUUID(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Cluster = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierUUID, UUID: &value}
		}
	}
}

// WithNutanixSSHAuthorizedKey returns a NutanixFiller that sets the SSH authorized key for the Nutanix machine.
func WithNutanixSSHAuthorizedKey(value string) NutanixFiller {
	return func(config *NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Users = []anywherev1.UserConfiguration{
				{
					Name:              anywherev1.DefaultNutanixMachineConfigUser,
					SshAuthorizedKeys: []string{value},
				},
			}
		}
	}
}
