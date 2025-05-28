package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func TestSnowMachineConfigSetDefaults(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	sOld := snowMachineConfig()
	err := sOld.Default(ctx, &sOld)
	g.Expect(err).To(BeNil())

	g.Expect(sOld.Spec.InstanceType).To(Equal(v1alpha1.DefaultSnowInstanceType))
	g.Expect(sOld.Spec.PhysicalNetworkConnector).To(Equal(v1alpha1.DefaultSnowPhysicalNetworkConnectorType))
}

func TestSnowMachineConfigValidateCreateNoAMI(t *testing.T) {
	ctx := context.Background()
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

	g.Expect(sOld.ValidateCreate(ctx, &sOld)).Error().To(Succeed())
}

func TestSnowMachineConfigValidateCreateInvalidInstanceType(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Spec.SshKeyName = "testKey"
	sOld.Spec.InstanceType = "invalid-instance-type"

	g.Expect(sOld.ValidateCreate(ctx, &sOld)).Error().To(MatchError(ContainSubstring("SnowMachineConfig InstanceType invalid-instance-type is not supported")))
}

func TestSnowMachineConfigValidateCreateEmptySSHKeyName(t *testing.T) {
	ctx := context.Background()
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
	g.Expect(s.ValidateCreate(ctx, &s)).Error().To(MatchError(ContainSubstring("SnowMachineConfig SshKeyName must not be empty")))
}

func TestSnowMachineConfigValidateCreate(t *testing.T) {
	ctx := context.Background()
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

	g.Expect(sOld.ValidateCreate(ctx, &sOld)).Error().To(Succeed())
}

func TestSnowMachineConfigValidateUpdate(t *testing.T) {
	ctx := context.Background()
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

	g.Expect(sNew.ValidateUpdate(ctx, &sOld, sNew)).Error().To(Succeed())
}

func TestSnowMachineConfigValidateUpdateNoDevices(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.SshKeyName = "testKey"
	sNew.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sNew.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sNew.Spec.OSFamily = v1alpha1.Bottlerocket

	g.Expect(sNew.ValidateUpdate(ctx, &sOld, sNew)).Error().To(MatchError(ContainSubstring("Devices must contain at least one device IP")))
}

func TestSnowMachineConfigValidateUpdateEmptySSHKeyName(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.InstanceType = v1alpha1.DefaultSnowInstanceType
	sNew.Spec.PhysicalNetworkConnector = v1alpha1.SFPPlus
	sNew.Spec.OSFamily = v1alpha1.Bottlerocket

	g.Expect(sNew.ValidateUpdate(ctx, sNew, &sOld)).Error().To(MatchError(ContainSubstring("SnowMachineConfig SshKeyName must not be empty")))
}

// Unit test to pass the code coverage job.
func TestSnowMachineConfigValidateDelete(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	sOld := snowMachineConfig()
	g.Expect(sOld.ValidateDelete(ctx, &sOld)).Error().To(Succeed())
}

func snowMachineConfig() v1alpha1.SnowMachineConfig {
	return v1alpha1.SnowMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.SnowMachineConfigSpec{},
		Status:     v1alpha1.SnowMachineConfigStatus{},
	}
}

func TestSnowMachineConfigDefaultCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomDefaulter
	config := &v1alpha1.SnowMachineConfig{}

	// Call Default with the wrong type
	err := config.Default(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowMachineConfig"))
}

func TestSnowMachineConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowMachineConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowMachineConfig"))
}

func TestSnowMachineConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowMachineConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), &v1alpha1.SnowMachineConfig{}, wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowMachineConfig"))
}

func TestSnowMachineConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowMachineConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowMachineConfig"))
}
