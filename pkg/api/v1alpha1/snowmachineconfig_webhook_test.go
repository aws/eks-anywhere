package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	sOld.Spec.InstanceType = v1alpha1.SbeCLarge
	sOld.Spec.Devices = []string{"1.2.3.4"}

	g.Expect(sOld.ValidateCreate()).NotTo(Succeed())
}

func TestSnowMachineConfigValidateCreate(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Spec.AMIID = "testAMI"
	sOld.Spec.InstanceType = v1alpha1.SbeCLarge
	sOld.Spec.Devices = []string{"1.2.3.4"}

	g.Expect(sOld.ValidateCreate()).To(Succeed())
}

func TestSnowMachineConfigValidateUpdateNoDevices(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.InstanceType = v1alpha1.SbeCLarge

	g.Expect(sNew.ValidateUpdate(&sOld)).NotTo(Succeed())
}

func TestSnowMachineConfigValidateUpdate(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sNew := sOld.DeepCopy()
	sNew.Spec.AMIID = "testAMI"
	sNew.Spec.InstanceType = v1alpha1.SbeCLarge
	sNew.Spec.Devices = []string{"1.2.3.4"}

	g.Expect(sNew.ValidateUpdate(&sOld)).To(Succeed())
}

// Unit test to pass the code coverage job
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
