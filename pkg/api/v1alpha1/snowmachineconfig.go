package v1alpha1

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/logger"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

const (
	SnowMachineConfigKind                   = "SnowMachineConfig"
	DefaultSnowSSHKeyName                   = ""
	DefaultSnowInstanceType                 = "sbe-c.large"
	DefaultSnowPhysicalNetworkConnectorType = SFPPlus
	DefaultOSFamily                         = Ubuntu
	MinimumContainerVolumeSizeUbuntu        = 8
	MinimumContainerVolumeSizeBottlerocket  = 25
	MinimumNonRootVolumeSize                = 8
)

var snowInstanceTypesRegex = regexp.MustCompile(`^sbe-[cg]\.\d*x?large$`)

// NewSnowMachineConfigGenerate generates snowMachineConfig example for generate clusterconfig command.
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
			Devices:                  []string{""},
			InstanceType:             DefaultSnowInstanceType,
			SshKeyName:               DefaultSnowSSHKeyName,
			PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
			OSFamily:                 DefaultOSFamily,
			Network: SnowNetwork{
				DirectNetworkInterfaces: []SnowDirectNetworkInterface{
					{
						Index:   1,
						DHCP:    true,
						Primary: true,
					},
				},
			},
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
	if err := validateSnowMachineConfigInstanceType(string(config.Spec.InstanceType)); err != nil {
		return err
	}

	if config.Spec.PhysicalNetworkConnector != SFPPlus && config.Spec.PhysicalNetworkConnector != QSFP && config.Spec.PhysicalNetworkConnector != RJ45 {
		return fmt.Errorf("SnowMachineConfig PhysicalNetworkConnector %s is not supported, please use one of the following: %s, %s, %s ", config.Spec.PhysicalNetworkConnector, SFPPlus, QSFP, RJ45)
	}

	if len(config.Spec.Devices) == 0 {
		return errors.New("SnowMachineConfig Devices must contain at least one device IP")
	}

	if len(config.Spec.OSFamily) <= 0 {
		return errors.New("SnowMachineConfig OSFamily must be specified")
	}

	if config.Spec.OSFamily != Bottlerocket && config.Spec.OSFamily != Ubuntu {
		return fmt.Errorf("SnowMachineConfig OSFamily %s is not supported, please use one of the following: %s, %s", config.Spec.OSFamily, Bottlerocket, Ubuntu)
	}

	if err := validateSnowMachineConfigNetwork(config.Spec.Network); err != nil {
		return err
	}

	if err := validateSnowMachineConfigContainerVolume(config); err != nil {
		return err
	}

	if err := validateHostOSConfig(config.Spec.HostOSConfiguration, config.Spec.OSFamily); err != nil {
		return fmt.Errorf("SnowMachineConfig HostOSConfiguration is invalid: %v", err)
	}

	return validateSnowMachineConfigNonRootVolumes(config.Spec.NonRootVolumes)
}

func validateSnowMachineConfigInstanceType(instanceType string) error {
	match := snowInstanceTypesRegex.FindStringSubmatch(instanceType)
	if match == nil {
		return fmt.Errorf("SnowMachineConfig InstanceType %s is not supported", instanceType)
	}
	return nil
}

func validateSnowMachineConfigContainerVolume(config *SnowMachineConfig) error {
	// The Bottlerocket AWS Variant AMI only has 2 Gi of data volume, which is insufficient to store EKS-A and user container volumes.
	// Thus the ContainersVolume is required and its size must be no smaller than 25 Gi.
	if config.Spec.OSFamily == Bottlerocket {
		if config.Spec.ContainersVolume == nil {
			return errors.New("SnowMachineConfig ContainersVolume must be specified for Bottlerocket OS")
		}
		if config.Spec.ContainersVolume.Size < MinimumContainerVolumeSizeBottlerocket {
			return fmt.Errorf("SnowMachineConfig ContainersVolume.Size must be no smaller than %d Gi for Bottlerocket OS", MinimumContainerVolumeSizeBottlerocket)
		}
	}

	if config.Spec.OSFamily == Ubuntu && config.Spec.ContainersVolume != nil && config.Spec.ContainersVolume.Size < MinimumContainerVolumeSizeUbuntu {
		return fmt.Errorf("SnowMachineConfig ContainersVolume.Size must be no smaller than %d Gi for Ubuntu OS", MinimumContainerVolumeSizeUbuntu)
	}

	return nil
}

func validateSnowMachineConfigNonRootVolumes(volumes []*snowv1.Volume) error {
	for i, v := range volumes {
		if v == nil {
			continue
		}

		if len(v.DeviceName) <= 0 {
			return fmt.Errorf("SnowMachineConfig NonRootVolumes[%d].DeviceName must be specified", i)
		}

		if strings.HasPrefix(v.DeviceName, "/dev/sda") {
			return fmt.Errorf("SnowMachineConfig NonRootVolumes[%d].DeviceName [%s] is invalid. Device name with prefix /dev/sda* is reserved for root volume and containers volume, please use another name", i, v.DeviceName)
		}

		if v.Size < MinimumNonRootVolumeSize {
			return fmt.Errorf("SnowMachineConfig NonRootVolumes[%d].Size must be no smaller than %d Gi", i, MinimumNonRootVolumeSize)
		}
	}

	return nil
}

func validateSnowMachineConfigNetwork(network SnowNetwork) error {
	if len(network.DirectNetworkInterfaces) <= 0 {
		return errors.New("SnowMachineConfig Network.DirectNetworkInterfaces length must be no smaller than 1")
	}

	primaryDNICount := 0
	for _, dni := range network.DirectNetworkInterfaces {
		if dni.Primary {
			primaryDNICount++
		}
	}
	if primaryDNICount != 1 {
		return errors.New("SnowMachineConfig Network.DirectNetworkInterfaces list must contain one and only one primary DNI")
	}

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
