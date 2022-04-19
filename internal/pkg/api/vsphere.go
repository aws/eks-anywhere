package api

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type VSphereConfig struct {
	datacenterConfig *anywherev1.VSphereDatacenterConfig
	machineConfigs   map[string]*anywherev1.VSphereMachineConfig
}

type VSphereFiller func(config VSphereConfig)

func AutoFillVSphereProvider(filename string, fillers ...VSphereFiller) ([]byte, error) {
	vsphereDatacenterConfig, err := anywherev1.GetVSphereDatacenterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get vsphere datacenter config from file: %v", err)
	}

	vsphereMachineConfigs, err := anywherev1.GetVSphereMachineConfigs(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get vsphere machine config from file: %v", err)
	}

	config := VSphereConfig{
		datacenterConfig: vsphereDatacenterConfig,
		machineConfigs:   vsphereMachineConfigs,
	}

	for _, f := range fillers {
		f(config)
	}

	resources := make([]interface{}, 0, len(config.machineConfigs)+1)
	resources = append(resources, config.datacenterConfig)
	for _, m := range config.machineConfigs {
		resources = append(resources, m)
	}

	yamlResources := make([][]byte, 0, len(resources))
	for _, r := range resources {
		yamlContent, err := yaml.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshalling vsphere resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithOsFamilyForAllMachines(value anywherev1.OSFamily) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.OSFamily = value
		}
	}
}

func WithNumCPUsForAllMachines(value int) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.NumCPUs = value
		}
	}
}

func WithDiskGiBForAllMachines(value int) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.DiskGiB = value
		}
	}
}

func WithMemoryMiBForAllMachines(value int) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.MemoryMiB = value
		}
	}
}

func WithTLSInsecure(value bool) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Insecure = value
	}
}

func WithTLSThumbprint(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Thumbprint = value
	}
}

func WithTemplateForAllMachines(value string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Template = value
		}
	}
}

func WithStoragePolicyNameForAllMachines(value string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.StoragePolicyName = value
		}
	}
}

func WithVSphereConfigNamespaceForAllMachinesAndDatacenter(ns string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Namespace = ns
		for _, m := range config.machineConfigs {
			m.Namespace = ns
		}
	}
}

func WithSSHAuthorizedKeyForAllMachines(key string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			setSSHKeyForFirstUser(m, key)
		}
	}
}

func WithServer(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Server = value
	}
}

func WithResourcePoolForAllMachines(value string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.ResourcePool = value
		}
	}
}

func WithNetwork(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Network = value
	}
}

func WithFolderForAllMachines(value string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Folder = value
		}
	}
}

func WithDatastoreForAllMachines(value string) VSphereFiller {
	return func(config VSphereConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Datastore = value
		}
	}
}

func WithDatacenter(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Datacenter = value
	}
}

func WithVSphereStringFromEnvVar(envVar string, opt func(string) VSphereFiller) VSphereFiller {
	return opt(os.Getenv(envVar))
}

func WithVSphereBoolFromEnvVar(envVar string, opt func(bool) VSphereFiller) VSphereFiller {
	return opt(os.Getenv(envVar) == "true")
}

func WithVSphereMachineConfig(name string, fillers ...VSphereMachineConfigFiller) VSphereFiller {
	return func(config VSphereConfig) {
		m, ok := config.machineConfigs[name]
		if !ok {
			m = &anywherev1.VSphereMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.VSphereMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			config.machineConfigs[name] = m
		}

		FillVSphereMachineConfig(m, fillers...)
	}
}
