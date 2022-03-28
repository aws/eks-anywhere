package api

import (
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type TinkerbellConfig struct {
	clusterConfig    *anywherev1.Cluster
	datacenterConfig *anywherev1.TinkerbellDatacenterConfig
	machineConfigs   map[string]*anywherev1.TinkerbellMachineConfig
	templateConfigs  map[string]*anywherev1.TinkerbellTemplateConfig
}

type TinkerbellFiller func(config TinkerbellConfig) error

func AutoFillTinkerbellProvider(filename string, fillers ...TinkerbellFiller) ([]byte, error) {
	tinkerbellDatacenterConfig, err := anywherev1.GetTinkerbellDatacenterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get tinkerbell datacenter config from file: %v", err)
	}

	tinkerbellMachineConfigs, err := anywherev1.GetTinkerbellMachineConfigs(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get tinkerbell machine config from file: %v", err)
	}

	tinkerbellTemplateConfigs, err := anywherev1.GetTinkerbellTemplateConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get tinkerbell template configs from file: %v", err)
	}

	clusterConfig, err := anywherev1.GetClusterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get tinkerbell cluster config from file: %v", err)
	}

	config := TinkerbellConfig{
		clusterConfig:    clusterConfig,
		datacenterConfig: tinkerbellDatacenterConfig,
		machineConfigs:   tinkerbellMachineConfigs,
		templateConfigs:  tinkerbellTemplateConfigs,
	}

	for _, f := range fillers {
		err := f(config)
		if err != nil {
			return nil, fmt.Errorf("failed to apply tinkerbell config filler: %v", err)
		}
	}

	resources := make([]interface{}, 0, len(config.machineConfigs)+len(config.templateConfigs)+1)
	resources = append(resources, config.datacenterConfig)

	for _, m := range config.machineConfigs {
		resources = append(resources, m)
	}

	for _, m := range config.templateConfigs {
		resources = append(resources, m)
	}

	yamlResources := make([][]byte, 0, len(resources))
	for _, r := range resources {
		yamlContent, err := yaml.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshalling tinkerbell resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithTinkerbellServer(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		config.datacenterConfig.Spec.TinkerbellIP = value
		return nil
	}
}

func WithTinkerbellCertURL(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		config.datacenterConfig.Spec.TinkerbellCertURL = value
		return nil
	}
}

func WithTinkerbellGRPCAuthEndpoint(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		config.datacenterConfig.Spec.TinkerbellGRPCAuth = value
		return nil
	}
}

func WithTinkerbellPBnJGRPCAuthEndpoint(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		config.datacenterConfig.Spec.TinkerbellPBnJGRPCAuth = value
		return nil
	}
}

func WithStringFromEnvVarTinkerbell(envVar string, opt func(string) TinkerbellFiller) TinkerbellFiller {
	return opt(os.Getenv(envVar))
}

func WithOsFamilyForAllTinkerbellMachines(value anywherev1.OSFamily) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		for _, m := range config.machineConfigs {
			m.Spec.OSFamily = value
		}
		return nil
	}
}

func WithTinkerbellHegelURL(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		for _, t := range config.templateConfigs {
			for _, task := range t.Spec.Template.Tasks {
				for _, action := range task.Actions {
					if action.Name == "add-tink-cloud-init-config" {
						contents := action.Environment["CONTENTS"]
						action.Environment["CONTENTS"] = strings.ReplaceAll(contents, "http://<REPLACE WITH TINKERBELL IP>:50061", value)
					}
				}
			}
		}
		return nil
	}
}

func WithImageUrlForAllTinkerbellMachines(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		for _, t := range config.templateConfigs {
			for _, task := range t.Spec.Template.Tasks {
				for _, action := range task.Actions {
					if action.Name == "stream-image" {
						action.Environment["IMG_URL"] = value
					}
				}
			}
		}
		return nil
	}
}

func WithSSHAuthorizedKeyForAllTinkerbellMachines(key string) TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		for _, m := range config.machineConfigs {
			if len(m.Spec.Users) == 0 {
				m.Spec.Users = []anywherev1.UserConfiguration{{}}
			}

			m.Spec.Users[0].Name = "ec2-user"
			m.Spec.Users[0].SshAuthorizedKeys = []string{key}
		}
		return nil
	}
}

func WithTinkerbellEtcdMachineConfig() TinkerbellFiller {
	return func(config TinkerbellConfig) error {
		clusterName := config.clusterConfig.Name
		name := providers.GetEtcdNodeName(clusterName)

		_, ok := config.machineConfigs[name]
		if !ok {
			m := &anywherev1.TinkerbellMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.TinkerbellMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: anywherev1.TinkerbellMachineConfigSpec{
					TemplateRef: anywherev1.Ref{
						Name: clusterName,
						Kind: anywherev1.TinkerbellTemplateConfigKind,
					},
				},
			}
			config.machineConfigs[name] = m
		}
		return nil
	}
}
