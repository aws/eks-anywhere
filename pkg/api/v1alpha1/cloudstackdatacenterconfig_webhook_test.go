package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

//func TestCloudstackDatacenterValidateUpdateServerImmutable(t *testing.T) {
//	vOld := cloudstackDatacenterConfig()
//	vOld.Spec.Server = "https://realOldServer.realOldDatacenter.com"
//	c := vOld.DeepCopy()
//
//	c.Spec.Server = "https://newFancyServer.newFancyCloud.io"
//	g := NewWithT(t)
//	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
//}

func TestCloudstackDatacenterValidateUpdateDomainImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Domain = "oldCruftyDomain"
	c := vOld.DeepCopy()

	c.Spec.Domain = "shinyNewDomain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateZoneImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Zone = "oldCruftyZone"
	c := vOld.DeepCopy()

	c.Spec.Zone = "shinyNewZone"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}



func TestCloudstackDatacenterValidateUpdateProjectImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Project = "oldCruftyDatacenter"
	c := vOld.DeepCopy()

	c.Spec.Project = "shinyNewDatacenter"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateAccountImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Account = "oldCruftyAccount"
	c := vOld.DeepCopy()

	c.Spec.Account = "shinyNewAccount"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = "OldNet"
	c := vOld.DeepCopy()

	c.Spec.Network = "NewNet"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateControlPlaneEndpointImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.ControlPlaneEndpoint = "OldControlPlaneEndpoint"
	c := vOld.DeepCopy()

	c.Spec.ControlPlaneEndpoint = "NewControlPlaneEndpoint"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateTLSInsecureImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Insecure = true
	c := vOld.DeepCopy()

	c.Spec.Insecure = false
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudstackDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = "oldNetwork"
	c := vOld.DeepCopy()

	c.Spec.Network = "newNetwork"

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCloudstackDatacenterValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudstackDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackDatacenterConfig() v1alpha1.CloudstackDatacenterConfig {
	return v1alpha1.CloudstackDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.CloudstackDatacenterConfigSpec{},
		Status:     v1alpha1.CloudstackDatacenterConfigStatus{},
	}
}
