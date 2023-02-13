package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func TestSnowMachineConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Default()

	g.Expect(sOld.Spec.InstanceType).To(Equal(v1alpha1.DefaultSnowInstanceType))
	g.Expect(sOld.Spec.PhysicalNetworkConnector).To(Equal(v1alpha1.DefaultSnowPhysicalNetworkConnectorType))
}

func TestSnowMachineConfigValidateCreateNoAMI(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Spec.SshKeyName = "testKey"
	sOld.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sOld.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sOld.Spec.Devices = []string{"1.2.3.4"}
	sOld.Spec.OSFamily = v1alpha1.Bottlerocket
	sOld.Spec.ContainersVolume = &snowv1.Volume{
		Size: 25,
	}
	sOld.Spec.Network = v1alpha1.SnowNetwork{
		DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
			{
				Index:   1,
				DHCP:    true,
				Primary: true,
			},
		},
	}

	g.Expect(sOld.ValidateCreate()).To(Succeed())
}

func TestSnowMachineConfigValidateCreateInvalidInstanceType(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Spec.SshKeyName = "testKey"
	sOld.Spec.InstanceType = "invalid-instance-type"

	g.Expect(sOld.ValidateCreate()).To(MatchError(ContainSubstring("SnowMachineConfig InstanceType invalid-instance-type is not supported")))
}

func TestSnowMachineConfigValidateCreateEmptySSHKeyName(t *testing.T) {
	g := NewWithT(t)
	s := snowMachineConfig()
	s.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	s.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	s.Spec.Devices = []string{"1.2.3.4"}
	s.Spec.OSFamily = v1alpha1.Ubuntu
	s.Spec.Network = v1alpha1.SnowNetwork{
		DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
			{
				Index:   1,
				DHCP:    true,
				Primary: true,
			},
		},
	}
	g.Expect(s.ValidateCreate()).To(MatchError(ContainSubstring("SnowMachineConfig SshKeyName must not be empty")))
}

func TestSnowMachineConfigValidateCreate(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Spec.AMIID = "testAMI"
	sOld.Spec.SshKeyName = "testKey"
	sOld.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sOld.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sOld.Spec.Devices = []string{"1.2.3.4"}
	sOld.Spec.OSFamily = v1alpha1.Bottlerocket
	sOld.Spec.ContainersVolume = &snowv1.Volume{
		Size: 25,
	}
	sOld.Spec.Network = v1alpha1.SnowNetwork{
		DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
			{
				Index:   1,
				DHCP:    true,
				Primary: true,
			},
		},
	}

	g.Expect(sOld.ValidateCreate()).To(Succeed())
}

func TestSnowMachineConfigValidateUpdate(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.SshKeyName = "testKey"
	sNew.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sNew.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sNew.Spec.Devices = []string{"1.2.3.4"}
	sNew.Spec.OSFamily = v1alpha1.Bottlerocket
	sNew.Spec.ContainersVolume = &snowv1.Volume{
		Size: 25,
	}
	sNew.Spec.Network = v1alpha1.SnowNetwork{
		DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
			{
				Index:   1,
				DHCP:    true,
				Primary: true,
			},
		},
	}

	g.Expect(sNew.ValidateUpdate(&sOld)).To(Succeed())
}

func TestSnowMachineConfigValidateUpdateNoDevices(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.SshKeyName = "testKey"
	sNew.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sNew.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sNew.Spec.OSFamily = v1alpha1.Bottlerocket

	g.Expect(sNew.ValidateUpdate(&sOld)).To(MatchError(ContainSubstring("Devices must contain at least one device IP")))
}

func TestSnowMachineConfigValidateUpdateEmptySSHKeyName(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sNew.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sNew.Spec.OSFamily = v1alpha1.Bottlerocket

	g.Expect(sNew.ValidateUpdate(&sOld)).To(MatchError(ContainSubstring("SnowMachineConfig SshKeyName must not be empty")))
}

// Unit test to pass the code coverage job.
func TestSnowMachineConfigValidateDelete(t *testing.T) {
	g := NewWithT(t)
	sOld := snowMachineConfig()
	g.Expect(sOld.ValidateDelete()).To(Succeed())
}

func snowMachineConfig() v1alpha1.SnowMachineConfig {
	return v1alpha1.SnowMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.SnowMachineConfigSpec{},
		Status:     v1alpha1.SnowMachineConfigStatus{},
	}
}
