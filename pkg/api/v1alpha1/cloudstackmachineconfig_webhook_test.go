package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCPCloudStackMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateComputeOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ComputeOffering = "oldComputeOffering"
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = "newComputeOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateComputeOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.ComputeOffering = "oldComputeOffering"
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = "newComputeOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = "oldDiskOffering"
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = "newDiskOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateDiskOfferingImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.DiskOffering = "oldDiskOffering"
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = "newDiskOffering"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestManagementCloudStackMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagement("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestManagementCloudStackMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagement("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDetailsImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Details = map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	c := vOld.DeepCopy()

	c.Spec.Details = map[string]string{
		"k1": "v2",
		"k2": "v1",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackMachineValidateUpdateDetailsImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.Details = map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	c := vOld.DeepCopy()

	c.Spec.Details = map[string]string{
		"k1": "v2",
		"k2": "v1",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackMachineValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackMachineConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackMachineConfig() v1alpha1.CloudStackMachineConfig {
	return v1alpha1.CloudStackMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.CloudStackMachineConfigSpec{},
		Status:     v1alpha1.CloudStackMachineConfigStatus{},
	}
}
