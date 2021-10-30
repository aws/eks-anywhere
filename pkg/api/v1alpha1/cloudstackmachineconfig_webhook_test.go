package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCPCloudstackMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudstackMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudstackMachineValidateUpdateComputeOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ComputeOffering = "oldComputeOffering"
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = "newComputeOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudstackMachineValidateUpdateComputeOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.ComputeOffering = "oldComputeOffering"
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = "newComputeOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudstackMachineValidateUpdateDiskOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = "oldDiskOffering"
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = "newDiskOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudstackMachineValidateUpdateDiskOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.DiskOffering = "oldDiskOffering"
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = "newDiskOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudstackMachineValidateUpdateKeypairImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.KeyPair = "oldKeypair"
	c := vOld.DeepCopy()

	c.Spec.KeyPair = "newKeypair"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackMachineValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudstackMachineConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackMachineConfig() v1alpha1.CloudstackMachineConfig {
	return v1alpha1.CloudstackMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.CloudstackMachineConfigSpec{},
		Status:     v1alpha1.CloudstackMachineConfigStatus{},
	}
}
