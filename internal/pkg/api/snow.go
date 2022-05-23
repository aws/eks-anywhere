package api

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type SnowConfig struct {
	datacenterConfig *anywherev1.SnowDatacenterConfig
	machineConfigs   map[string]*anywherev1.SnowMachineConfig
}

type SnowFiller func(config SnowConfig)

func AutoFillSnowProvider(filename string, fillers ...SnowFiller) ([]byte, error) {
	config, err := cluster.ParseConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

	snowConfig := SnowConfig{
		datacenterConfig: config.SnowDatacenter,
		machineConfigs:   config.SnowMachineConfigs,
	}

	for _, f := range fillers {
		f(snowConfig)
	}

	resources := make([]interface{}, 0, len(snowConfig.machineConfigs)+1)
	resources = append(resources, snowConfig.datacenterConfig)

	for _, m := range snowConfig.machineConfigs {
		resources = append(resources, m)
	}

	yamlResources := make([][]byte, 0, len(resources))
	for _, r := range resources {
		yamlContent, err := yaml.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshalling snow resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithSnowStringFromEnvVar(envVar string, opt func(string) SnowFiller) SnowFiller {
	return opt(os.Getenv(envVar))
}

func WithSnowAMIID(id string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.AMIID = id
		}
	}
}

func WithSnowInstanceType(instanceType anywherev1.SnowInstanceType) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.InstanceType = instanceType
		}
	}
}

func WithSnowPhysicalNetworkConnector(connectorType anywherev1.PhysicalNetworkConnectorType) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.PhysicalNetworkConnector = connectorType
		}
	}
}

func WithSnowSshKeyName(keyName string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.SshKeyName = keyName
		}
	}
}
