package v1alpha1

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			AMIID:                    "",
			InstanceType:             DefaultSnowInstanceType,
			SshKeyName:               DefaultSnowSshKeyName,
			PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
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

func validateSnowMachineConfig(config *SnowMachineConfig) error {
	if config.Spec.AMIID == "" {
		return errors.New("SnowMachineConfig AMIID is a required field")
	}

	if config.Spec.InstanceType != SbeCLarge && config.Spec.InstanceType != SbeCXLarge && config.Spec.InstanceType != SbeC2XLarge && config.Spec.InstanceType != SbeC4XLarge {
		return fmt.Errorf("SnowMachineConfig InstanceType %s is not supported, please use one of the following: %s, %s, %s, %s ", config.Spec.InstanceType, SbeCLarge, SbeCXLarge, SbeC2XLarge, SbeC4XLarge)
	}

	// TODO: temporarily remove this validation since `devices` is a newly added, required field.
	// This validation runs in snowmachineconfig webhook and ValidateUpdate fails when upgrading from older eks-a version
	// without the `devices` field. We will add this validation back once users update their clusters to latest version.
	// if len(config.Spec.Devices) == 0 {
	// 	return errors.New("SnowMachineConfig Devices must contain at least one device IP")
	// }
	return nil
}

func setSnowMachineConfigDefaults(config *SnowMachineConfig) {
	if config.Spec.InstanceType == "" {
		config.Spec.InstanceType = DefaultSnowInstanceType
		logger.V(1).Info("SnowMachineConfig InstanceType is empty. Using default", "default instance type", DefaultSnowInstanceType)
	}

	if config.Spec.PhysicalNetworkConnector == "" {
		config.Spec.PhysicalNetworkConnector = DefaultSnowPhysicalNetworkConnectorType
		logger.V(1).Info("SnowMachineConfig PhysicalNetworkConnector is empty. Using default", "default physical network connector", DefaultSnowPhysicalNetworkConnectorType)
	}
}
