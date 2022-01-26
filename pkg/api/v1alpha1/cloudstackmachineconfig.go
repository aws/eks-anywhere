package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const CloudStackMachineConfigKind = "CloudStackMachineConfig"

// Used for generating yaml for generate clusterconfig command
func NewCloudStackMachineConfigGenerate(name string) *CloudStackMachineConfigGenerate {
	return &CloudStackMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudStackMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: CloudStackMachineConfigSpec{
			ComputeOffering: "",
			OSFamily:        Redhat,
			Template:        "",
			Users: []UserConfiguration{{
				Name:              "capc",
				SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
			}},
		},
	}
}

func (c *CloudStackMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudStackMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudStackMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudStackMachineConfigs(fileName string) (map[string]*CloudStackMachineConfig, error) {
	configs := make(map[string]*CloudStackMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config CloudStackMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == CloudStackMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == CloudStackMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", CloudStackMachineConfigKind)
	}
	return configs, nil
}
