package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultCloudStackUser is the default CloudStackMachingConfig username.
const DefaultCloudStackUser = "capc"

// CloudStackMachineConfigKind is the kind value for a CloudStackMachineConfig.
const CloudStackMachineConfigKind = "CloudStackMachineConfig"

// Taken from https://github.com/shapeblue/cloudstack/blob/08bb4ad9fea7e422c3d3ac6d52f4670b1e89eed7/api/src/main/java/com/cloud/vm/VmDetailConstants.java
// These fields should be modeled separately in eks-a and not used by the additionalDetails cloudstack VM field.
var restrictedUserCustomDetails = [...]string{
	"keyboard", "cpu.corespersocket", "rootdisksize", "boot.mode", "nameonhypervisor",
	"nicAdapter", "rootDiskController", "dataDiskController", "svga.vramSize", "nestedVirtualizationFlag", "ramReservation",
	"hypervisortoolsversion", "platform", "timeoffset", "kvm.vnc.port", "kvm.vnc.address", "video.hardware", "video.ram",
	"smc.present", "firmware", "cpuNumber", "cpuSpeed", "memory", "cpuOvercommitRatio", "memoryOvercommitRatio",
	"Message.ReservedCapacityFreed.Flag", "deployvm", "SSH.PublicKey", "SSH.KeyPairNames", "password", "Encrypted.Password",
	"configDriveLocation", "nic", "network", "ip4Address", "ip6Address", "disk", "diskOffering", "configurationId",
	"keypairnames", "controlNodeLoginUser",
}

// Used for generating yaml for generate clusterconfig command.
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
			ComputeOffering: CloudStackResourceIdentifier{
				Id: "",
			},
			Template: CloudStackResourceIdentifier{
				Id: "",
			},
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

func validateCloudStackMachineConfig(machineConfig *CloudStackMachineConfig) error {
	if len(machineConfig.Spec.ComputeOffering.Id) == 0 && len(machineConfig.Spec.ComputeOffering.Name) == 0 {
		return fmt.Errorf("computeOffering is not set for CloudStackMachineConfig %s. Default computeOffering is not supported in CloudStack, please provide a computeOffering name or ID", machineConfig.Name)
	}
	if len(machineConfig.Spec.Template.Id) == 0 && len(machineConfig.Spec.Template.Name) == 0 {
		return fmt.Errorf("template is not set for CloudStackMachineConfig %s. Default template is not supported in CloudStack, please provide a template name or ID", machineConfig.Name)
	}
	if err, fieldName, fieldValue := machineConfig.Spec.DiskOffering.Validate(); err != nil {
		return fmt.Errorf("machine config %s validation failed: %s: %s invalid, %v", machineConfig.Name, fieldName, fieldValue, err)
	}
	for _, restrictedKey := range restrictedUserCustomDetails {
		if _, found := machineConfig.Spec.UserCustomDetails[restrictedKey]; found {
			return fmt.Errorf("restricted key %s found in custom user details", restrictedKey)
		}
	}
	if err := validateAffinityConfig(machineConfig); err != nil {
		return err
	}
	return nil
}

func validateAffinityConfig(machineConfig *CloudStackMachineConfig) error {
	if len(machineConfig.Spec.Affinity) > 0 && len(machineConfig.Spec.AffinityGroupIds) > 0 {
		return fmt.Errorf("affinity and affinityGroupIds cannot be set at the same time for CloudStackMachineConfig %s. Please provide either one of them or none", machineConfig.Name)
	}
	if len(machineConfig.Spec.Affinity) > 0 {
		if machineConfig.Spec.Affinity != "pro" && machineConfig.Spec.Affinity != "anti" && machineConfig.Spec.Affinity != "no" {
			return fmt.Errorf("invalid affinity type %s for CloudStackMachineConfig %s. Please provide \"pro\", \"anti\" or \"no\"", machineConfig.Spec.Affinity, machineConfig.Name)
		}
	}
	return nil
}
