package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	SnowMachineConfigKind                   = "SnowMachineConfig"
	DefaultSnowSshKeyName                   = "default"
	DefaultSnowInstanceType                 = SbeCLarge
	DefaultSnowPhysicalNetworkConnectorType = SFPPlus
)

// Used for generating yaml for generate clusterconfig command
func NewSnowMachineConfigGenerate(name string) *SnowMachineConfigGenerate {
	return &SnowMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       SnowMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: SnowMachineConfigSpec{
			AMIID:        "",
			InstanceType: DefaultSnowInstanceType,
			SshKeyName:   DefaultSnowSshKeyName,
		},
	}
}

func (s *SnowMachineConfigGenerate) APIVersion() string {
	return s.TypeMeta.APIVersion
}

func (s *SnowMachineConfigGenerate) Kind() string {
	return s.TypeMeta.Kind
}

func (s *SnowMachineConfigGenerate) Name() string {
	return s.ObjectMeta.Name
}

func GetSnowMachineConfigs(fileName string) (map[string]*SnowMachineConfig, error) {
	configs := make(map[string]*SnowMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config SnowMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == SnowMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == SnowMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", SnowMachineConfigKind)
	}
	return configs, nil
}

func validateSnowMachineConfig(config *SnowMachineConfig) error {
	if config.Spec.AMIID == "" {
		return fmt.Errorf("SnowMachineConfig AMIID is a required field")
	}

	if config.Spec.InstanceType != SbeCLarge && config.Spec.InstanceType != SbeCXLarge && config.Spec.InstanceType != SbeC2XLarge && config.Spec.InstanceType != SbeC4XLarge {
		return fmt.Errorf("SnowMachineConfig InstanceType %s is not supported, please use one of the following: %s, %s, %s, %s ", config.Spec.InstanceType, SbeCLarge, SbeCXLarge, SbeC2XLarge, SbeC4XLarge)
	}
	return nil
}

func setSnowMachineConfigDefaults(config *SnowMachineConfig) {
	if config.Spec.InstanceType == "" {
		config.Spec.InstanceType = DefaultSnowInstanceType
		logger.V(1).Info("SnowMachineConfig InstanceType is empty. Using default", "default instance type", DefaultSnowInstanceType)
	}

	if config.Spec.SshKeyName == "" {
		config.Spec.SshKeyName = DefaultSnowSshKeyName
		logger.V(1).Info("SnowMachineConfig SshKeyName is empty. Using default", "default SSH key name", DefaultSnowSshKeyName)
	}

	if config.Spec.PhysicalNetworkConnector == "" {
		config.Spec.PhysicalNetworkConnector = DefaultSnowPhysicalNetworkConnectorType
		logger.V(1).Info("SnowMachineConfig PhysicalNetworkConnector is empty. Using default", "default physical network connector", DefaultSnowPhysicalNetworkConnectorType)
	}
}
