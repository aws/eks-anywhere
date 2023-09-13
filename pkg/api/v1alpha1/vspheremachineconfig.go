package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	if machineConfig.Spec.OSFamily == "" {
		logger.Info("Warning: OS family not specified in machine config specification. Defaulting to Bottlerocket.")
		machineConfig.Spec.OSFamily = Bottlerocket
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
	if err := validateVSphereMachineConfigOSFamilyUser(config); err != nil {
		return err
	}
	if err := validateHostOSConfig(config.Spec.HostOSConfiguration, config.Spec.OSFamily); err != nil {
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
