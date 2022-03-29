package v1alpha1_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
)

func TestCloudStackMachineConfigValidateCreateFeatureDisabled(t *testing.T) {
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	err := os.Unsetenv(features.CloudStackProviderEnvVar)
	if err != nil {
		return
	}
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)

	c := cloudstackMachineConfig()
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateTemplateMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldTemplate",
	}
	c := vOld.DeepCopy()

	c.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "newTemplate",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateTemplateMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldTemplate",
	}
	c := vOld.DeepCopy()

	c.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "newTemplate",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCPCloudStackMachineValidateUpdateComputeOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldComputeOffering",
	}
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "newComputeOffering",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateComputeOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldComputeOffering",
	}
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "newComputeOffering",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCloudStackMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagement("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCloudStackMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagement("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
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
