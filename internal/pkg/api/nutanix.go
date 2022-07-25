package api

import (
	"fmt"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type NutanixConfig struct {
	datacenterConfig *anywherev1.NutanixDatacenterConfig
	machineConfigs   map[string]*anywherev1.NutanixMachineConfig
}

type NutanixFiller func(config NutanixConfig)

func AutoFillNutanixProvider(filename string, fillers ...NutanixFiller) ([]byte, error) {
	config, err := cluster.ParseConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

	nutanixConfig := NutanixConfig{
		datacenterConfig: config.NutanixDatacenter,
		machineConfigs:   config.NutanixMachineConfigs,
	}

	for _, f := range fillers {
		f(nutanixConfig)
	}

	resources := make([]interface{}, 0, len(nutanixConfig.machineConfigs)+1)
	resources = append(resources, nutanixConfig.datacenterConfig)

	for _, m := range nutanixConfig.machineConfigs {
		resources = append(resources, m)
	}

	yamlResources := make([][]byte, 0, len(resources))
	for _, r := range resources {
		yamlContent, err := yaml.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshalling nutanix resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithNutanixStringFromEnvVar(envVar string, opt func(string) NutanixFiller) NutanixFiller {
	return opt(os.Getenv(envVar))
}

func WithNutanixIntFromEnvVar(envVar string, opt func(int) NutanixFiller) NutanixFiller {
	intVar, _ := strconv.Atoi(os.Getenv(envVar))
	return opt(intVar)
}

func WithNutanixInt32FromEnvVar(envVar string, opt func(int32) NutanixFiller) NutanixFiller {
	intVar, _ := strconv.Atoi(os.Getenv(envVar))
	return opt(int32(intVar))
}

func WithNutanixBoolFromEnvVar(envVar string, opt func(bool) NutanixFiller) NutanixFiller {
	return opt(os.Getenv(envVar) == "true")
}

func WithNutanixEndpoint(value string) NutanixFiller {
	return func(config NutanixConfig) {
		config.datacenterConfig.Spec.NutanixEndpoint = value
	}
}
func WithNutanixPort(value int) NutanixFiller {
	return func(config NutanixConfig) {
		config.datacenterConfig.Spec.NutanixPort = value
	}
}

func WithNutanixUser(value string) NutanixFiller {
	return func(config NutanixConfig) {
		config.datacenterConfig.Spec.NutanixUser = value
	}
}
func WithNutanixPwd(value string) NutanixFiller {
	return func(config NutanixConfig) {
		config.datacenterConfig.Spec.NutanixPassword = value
	}
}

func WithNutanixInsure(value bool) NutanixFiller {
	return func(config NutanixConfig) {
		config.datacenterConfig.Spec.NutanixInsecure = value
	}
}

func WithNutanixMachineBootType(value int32) NutanixFiller {
	return func(config NutanixConfig) {
		// for _, m := range config.machineConfigs {
		// 	m.Spec. = value
		// }
	}
}

func WithNutanixMachineMemorySize(value string) NutanixFiller {
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.MemorySize = resource.MustParse(value)
		}
	}
}

func WithNutanixMachineSystemDiskSize(value string) NutanixFiller {
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.SystemDiskSize = resource.MustParse(value)
		}
	}
}

func WithNutanixMachineVCPUsPerSocket(value int32) NutanixFiller {
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.VCPUsPerSocket = value
		}
	}
}

func WithNutanixMachineVCPUSocket(value int32) NutanixFiller {
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.VCPUSockets = value
		}
	}
}

func WithNutanixMachineTemplateImageName(value string) NutanixFiller {
	// TODO Handle Type uuid as well
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Image = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

func WithNutanixSubnetName(value string) NutanixFiller {
	// TODO Handle Type uuid as well
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Subnet = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

func WithNutanixPrismElementClusterName(value string) NutanixFiller {
	return func(config NutanixConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Cluster = anywherev1.NutanixResourceIdentifier{Type: anywherev1.NutanixIdentifierName, Name: &value}
		}
	}
}

func WithNutanixSSHAuthorizedKey(value string) NutanixFiller {
	return func(config NutanixConfig) {
	}
}
