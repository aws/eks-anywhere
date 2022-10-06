package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const TinkerbellMachineConfigKind = "TinkerbellMachineConfig"

// +kubebuilder:object:generate=false
type TinkerbellMachineConfigGenerateOpt func(config *TinkerbellMachineConfigGenerate)

// Used for generating yaml for generate clusterconfig command
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
			UpgradeRolloutStrategy:	UpgradeRolloutStrategy{
				Type:	"RollingUpdate",
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
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config TinkerbellMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == TinkerbellMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
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
