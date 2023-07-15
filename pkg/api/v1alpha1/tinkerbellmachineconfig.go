package v1alpha1

import (
	"fmt"

	yamlutilpkg "github.com/aws/eks-anywhere/pkg/yamlutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const TinkerbellMachineConfigKind = "TinkerbellMachineConfig"

// +kubebuilder:object:generate=false
type TinkerbellMachineConfigGenerateOpt func(config *TinkerbellMachineConfigGenerate)

// Used for generating yaml for generate clusterconfig command.
func NewTinkerbellMachineConfigGenerate(name string, opts ...TinkerbellMachineConfigGenerateOpt) *TinkerbellMachineConfigGenerate {
	machineConfig := &TinkerbellMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: TinkerbellMachineConfigSpec{
			HardwareSelector: HardwareSelector{},
			OSFamily:         Bottlerocket,
			Users: []UserConfiguration{
				{
					Name:              "ec2-user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(machineConfig)
	}

	return machineConfig
}

func (c *TinkerbellMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetTinkerbellMachineConfigs(fileName string) (map[string]*TinkerbellMachineConfig, error) {
	configs := make(map[string]*TinkerbellMachineConfig)
	resources, err := yamlutilpkg.ParseMultiYamlFile(fileName)
	if err != nil {
		return nil, err
	}

	for _, d := range resources {
		var config TinkerbellMachineConfig
		if err := yamlutil.UnmarshalStrict(d, &config); err == nil {
			if config.Kind == TinkerbellMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}

		err := yaml.Unmarshal(d, &config)
		if config.Kind == TinkerbellMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", TinkerbellMachineConfigKind)
	}
	return configs, nil
}

func WithTemplateRef(ref ProviderRefAccessor) TinkerbellMachineConfigGenerateOpt {
	return func(c *TinkerbellMachineConfigGenerate) {
		c.Spec.TemplateRef = Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

func validateTinkerbellMachineConfig(config *TinkerbellMachineConfig) error {
	if err := validateObjectMeta(config.ObjectMeta); err != nil {
		return fmt.Errorf("TinkerbellMachineConfig: %v", err)
	}

	if len(config.Spec.HardwareSelector) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.hardwareSelector: %s", config.Name)
	}

	if len(config.Spec.HardwareSelector) != 1 {
		return fmt.Errorf(
			"TinkerbellMachineConfig: spec.hardwareSelector must contain only 1 key-value pair: %s",
			config.Name,
		)
	}

	if config.Spec.OSFamily == "" {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.osFamily: %s", config.Name)
	}

	if config.Spec.OSFamily != Ubuntu && config.Spec.OSFamily != Bottlerocket && config.Spec.OSFamily != RedHat {
		return fmt.Errorf(
			"TinkerbellMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s, %s",
			config.Spec.OSFamily,
			Ubuntu,
			RedHat,
			Bottlerocket,
		)
	}

	if len(config.Spec.Users) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.Users: %s", config.Name)
	}

	if err := validateHostOSConfig(config.Spec.HostOSConfiguration, config.Spec.OSFamily); err != nil {
		return fmt.Errorf("HostOSConfiguration is invalid for TinkerbellMachineConfig %s: %v", config.Name, err)
	}

	return nil
}

func setTinkerbellMachineConfigDefaults(machineConfig *TinkerbellMachineConfig) {
	if machineConfig.Spec.OSFamily == "" {
		machineConfig.Spec.OSFamily = Bottlerocket
	}
}
