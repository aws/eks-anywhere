package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const CloudstackMachineConfigKind = "CloudstackMachineConfig"

// Used for generating yaml for generate clusterconfig command
func NewCloudstackMachineConfigGenerate(name string) *CloudstackMachineConfigGenerate {
	return &CloudstackMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudstackMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: CloudstackMachineConfigSpec{
			Template: "Cloudstack template name",
			ComputeOffering: "Cloudstack compute offering name",
			DiskOffering: "Cloudstack disk offering name",
			KeyPair: "cloudstack keypair name",
		},
	}
}

func (c *CloudstackMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudstackMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudstackMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudstackMachineConfigs(fileName string) (map[string]*CloudstackMachineConfig, error) {
	configs := make(map[string]*CloudstackMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config CloudstackMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == CloudstackMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == CloudstackMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", CloudstackMachineConfigKind)
	}
	return configs, nil
}
