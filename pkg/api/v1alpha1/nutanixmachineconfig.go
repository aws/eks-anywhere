package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	NutanixMachineConfigKind        = "NutanixMachineConfig"
	DefaultNutanixOSFamily          = Ubuntu
	DefaultNutanixSystemDiskSizeGi  = "20Gi"
	DefaultNutanixMemorySizeGi      = "2Gi"
	DefaultNutanixVCPUsPerSocket    = 1
	DefaultNutanixVCPUSockets       = 1
	DefaultNutanixMachineConfigUser = "nutanix-user"
)

// +kubebuilder:object:generate=false
type NutanixMachineConfigGenerateOpt func(config *NutanixMachineConfigGenerate)

// Used for generating yaml for generate clusterconfig command
func NewNutanixMachineConfigGenerate(name string, opts ...NutanixMachineConfigGenerateOpt) *NutanixMachineConfigGenerate {
	machineConfig := &NutanixMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       NutanixMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: NutanixMachineConfigSpec{
			OSFamily: DefaultNutanixOSFamily,
			Users: []UserConfiguration{
				{
					Name:              DefaultNutanixMachineConfigUser,
					SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
				},
			},
			VCPUsPerSocket: DefaultNutanixVCPUsPerSocket,
			VCPUSockets:    DefaultNutanixVCPUSockets,
			MemorySize:     resource.MustParse(DefaultNutanixMemorySizeGi),
			// Image:          NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: nil},
			// Cluster:        NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: nil},
			// Subnet:         NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: nil},
			SystemDiskSize: resource.MustParse(DefaultNutanixSystemDiskSizeGi),
		},
	}

	for _, opt := range opts {
		opt(machineConfig)
	}

	return machineConfig
}

func (c *NutanixMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *NutanixMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *NutanixMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetNutanixMachineConfigs(fileName string) (map[string]*NutanixMachineConfig, error) {
	configs := make(map[string]*NutanixMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config NutanixMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == NutanixMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == NutanixMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", NutanixMachineConfigKind)
	}
	return configs, nil
}
