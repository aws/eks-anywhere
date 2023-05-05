package v1alpha1

import (
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	VSphereMachineConfigKind = "VSphereMachineConfig"
	DefaultVSphereDiskGiB    = 25
	DefaultVSphereNumCPUs    = 2
	DefaultVSphereMemoryMiB  = 8192
	DefaultVSphereOSFamily   = Bottlerocket
)

// Used for generating yaml for generate clusterconfig command.
func NewVSphereMachineConfigGenerate(name string) *VSphereMachineConfigGenerate {
	return &VSphereMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       VSphereMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: VSphereMachineConfigSpec{
			DiskGiB:   DefaultVSphereDiskGiB,
			NumCPUs:   DefaultVSphereNumCPUs,
			MemoryMiB: DefaultVSphereMemoryMiB,
			OSFamily:  DefaultVSphereOSFamily,
			Users: []UserConfiguration{{
				Name:              "ec2-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
			}},
		},
	}
}

func (c *VSphereMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *VSphereMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *VSphereMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetVSphereMachineConfigs(fileName string) (map[string]*VSphereMachineConfig, error) {
	configs := make(map[string]*VSphereMachineConfig)
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config VSphereMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == VSphereMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == VSphereMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", VSphereMachineConfigKind)
	}
	return configs, nil
}
