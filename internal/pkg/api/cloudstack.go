package api

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type CloudStackConfig struct {
	datacenterConfig *anywherev1.CloudStackDatacenterConfig
	machineConfigs   map[string]*anywherev1.CloudStackMachineConfig
}

type CloudStackFiller func(config CloudStackConfig)

func AutoFillCloudStackProvider(filename string, fillers ...CloudStackFiller) ([]byte, error) {
	cloudstackDatacenterConfig, err := anywherev1.GetCloudStackDatacenterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cloudstack datacenter config from file: %v", err)
	}

	cloudstackMachineConfigs, err := anywherev1.GetCloudStackMachineConfigs(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cloudstack machine config from file: %v", err)
	}

	config := CloudStackConfig{
		datacenterConfig: cloudstackDatacenterConfig,
		machineConfigs:   cloudstackMachineConfigs,
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
			return nil, fmt.Errorf("marshalling cloudstack resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithCloudStackComputeOfferingForAllMachines(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.ComputeOffering.Name = value
		}
	}
}

func WithCloudStackManagementServer(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.ManagementApiEndpoint = value
	}
}

func WithCloudStackAffinityGroupIds(value []string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.AffinityGroupIds = value
		}
	}
}

func WithUserCustomDetails(value map[string]string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.UserCustomDetails = value
		}
	}
}

func WithCloudStackTemplateForAllMachines(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Template.Name = value
		}
	}
}

func WithCloudStackConfigNamespace(ns string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Namespace = ns
		for _, m := range config.machineConfigs {
			m.Namespace = ns
		}
	}
}

//func WithCloudStackSSHUsernameAndAuthorizedKey(username string, key string) CloudStackFiller {
//	return func(config CloudStackConfig) {
//		if len(config.cpMachineConfig.Spec.Users) == 0 {
//			config.cpMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
//		}
//		if len(config.workerMachineConfig.Spec.Users) == 0 {
//			config.workerMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
//		}
//		config.cpMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
//			Name:              username,
//			SshAuthorizedKeys: []string{key},
//		}
//		config.workerMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
//			Name:              username,
//			SshAuthorizedKeys: []string{key},
//		}
//		if config.etcdMachineConfig != nil {
//			if len(config.etcdMachineConfig.Spec.Users) == 0 {
//				config.etcdMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
//			}
//			config.etcdMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
//				Name:              username,
//				SshAuthorizedKeys: []string{key},
//			}
//		}
//	}
//}

func WithCloudStackSSHAuthorizedKey(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			if len(m.Spec.Users) == 0 {
				m.Spec.Users = []anywherev1.UserConfiguration{{Name: "capc"}}
			}
			m.Spec.Users[0].SshAuthorizedKeys[0] = value
		}
	}
}

func WithCloudStackDomain(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Domain = value
	}
}

func WithCloudStackAccount(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Account = value
	}
}

func WithCloudStackZone(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Zones[0].Name = value
	}
}

func WithCloudStackNetwork(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Zones[0].Network.Name = value
	}
}

func WithCloudStackStringFromEnvVar(envVar string, opt func(string) CloudStackFiller) CloudStackFiller {
	return opt(os.Getenv(envVar))
}

func WithCloudStackMachineConfig(name string, fillers ...CloudStackMachineConfigFiller) CloudStackFiller {
	return func(config CloudStackConfig) {
		m, ok := config.machineConfigs[name]
		if !ok {
			m = &anywherev1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.CloudStackMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			config.machineConfigs[name] = m
		}

		FillCloudStackMachineConfig(m, fillers...)
	}
}
