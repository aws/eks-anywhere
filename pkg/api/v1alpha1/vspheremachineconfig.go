package v1alpha1

import (
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
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

func setVSphereMachineConfigDefaults(machineConfig *VSphereMachineConfig) {
	if len(machineConfig.Spec.Folder) <= 0 {
		logger.Info("VSphereMachineConfig Folder is not set or is empty. Defaulting to root vSphere folder.")
	}

	if machineConfig.Spec.MemoryMiB <= 0 {
		logger.V(1).Info("VSphereMachineConfig MemoryMiB is not set or is empty. Defaulting to 8192.", "machineConfig", machineConfig.Name)
		machineConfig.Spec.MemoryMiB = 8192
	}

	if machineConfig.Spec.MemoryMiB < 2048 {
		logger.Info("Warning: VSphereMachineConfig MemoryMiB should not be less than 2048. Defaulting to 2048. Recommended memory is 8192.", "machineConfig", machineConfig.Name)
		machineConfig.Spec.MemoryMiB = 2048
	}

	if machineConfig.Spec.NumCPUs <= 0 {
		logger.V(1).Info("VSphereMachineConfig NumCPUs is not set or is empty. Defaulting to 2.", "machineConfig", machineConfig.Name)
		machineConfig.Spec.NumCPUs = 2
	}

	if len(machineConfig.Spec.Users) <= 0 {
		machineConfig.Spec.Users = []UserConfiguration{{}}
	}

	if len(machineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	if machineConfig.Spec.OSFamily == "" {
		logger.Info("Warning: OS family not specified in machine config specification. Defaulting to Bottlerocket.")
		machineConfig.Spec.OSFamily = Bottlerocket
	}

	if len(machineConfig.Spec.Users) == 0 || machineConfig.Spec.Users[0].Name == "" {
		if machineConfig.Spec.OSFamily == Bottlerocket {
			machineConfig.Spec.Users[0].Name = constants.BottlerocketDefaultUser
		} else {
			machineConfig.Spec.Users[0].Name = constants.UbuntuDefaultUser
		}
		logger.V(1).Info("SSHUsername is not set or is empty for VSphereMachineConfig, using default", "machineConfig", machineConfig.Name, "user", machineConfig.Spec.Users[0].Name)
	}
}

func validateVSphereMachineConfig(config *VSphereMachineConfig) error {
	if len(config.Spec.Datastore) <= 0 {
		return fmt.Errorf("VSphereMachineConfig %s datastore is not set or is empty", config.Name)
	}
	if len(config.Spec.ResourcePool) <= 0 {
		return fmt.Errorf("VSphereMachineConfig %s VM resourcePool is not set or is empty", config.Name)
	}
	if config.Spec.OSFamily != Bottlerocket && config.Spec.OSFamily != Ubuntu && config.Spec.OSFamily != RedHat {
		return fmt.Errorf("VSphereMachineConfig %s osFamily: %s is not supported, please use one of the following: %s, %s, %s", config.Name, config.Spec.OSFamily, Bottlerocket, Ubuntu, RedHat)
	}
	if config.Spec.OSFamily == Bottlerocket && config.Spec.Users[0].Name != constants.BottlerocketDefaultUser {
		return fmt.Errorf("SSHUsername %s is invalid. Please use 'ec2-user' for Bottlerocket", config.Spec.Users[0].Name)
	}
	if err := validateHostOSConfig(config.Spec.HostOSConfiguration); err != nil {
		return fmt.Errorf("HostOSConfiguration is invalid for VSphereMachineConfig %s: %v", config.Name, err)
	}

	return nil
}

func validateVSphereMachineConfigHasTemplate(config *VSphereMachineConfig) error {
	if config.Spec.Template == "" {
		return fmt.Errorf("template field is required")
	}

	return nil
}
