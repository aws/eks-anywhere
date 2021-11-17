package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackDeploymentValidateUpdateDomainImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Domain = "oldCruftyDomain"
	c := vOld.DeepCopy()

	c.Spec.Domain = "shinyNewDomain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateZoneImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Zone = "oldCruftyZone"
	c := vOld.DeepCopy()

	c.Spec.Zone = "shinyNewZone"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateAccountImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Account = "oldCruftyAccount"
	c := vOld.DeepCopy()

	c.Spec.Account = "shinyNewAccount"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = "OldNet"
	c := vOld.DeepCopy()

	c.Spec.Network = "NewNet"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateControlPlaneEndpointImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.ManagementApiEndpoint = "OldManagementApiEndpoint"
	c := vOld.DeepCopy()

	c.Spec.ManagementApiEndpoint = "NewManagementApiEndpoint"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateTLSInsecureImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Insecure = true
	c := vOld.DeepCopy()

	c.Spec.Insecure = false
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDeploymentValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = "oldNetwork"
	c := vOld.DeepCopy()

	c.Spec.Network = "newNetwork"

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCloudStackDeploymentValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackDeploymentConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackDatacenterConfig() v1alpha1.CloudStackDeploymentConfig {
	return v1alpha1.CloudStackDeploymentConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.CloudStackDeploymentConfigSpec{},
		Status:     v1alpha1.CloudStackDeploymentConfigStatus{},
	}
}
