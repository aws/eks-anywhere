package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestVSphereDatacenterValidateUpdateServerImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Server = "https://realOldServer.realOldDatacenter.com"
	c := vOld.DeepCopy()

	c.Spec.Server = "https://newFancyServer.newFancyCloud.io"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateDataCenterImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Datacenter = "oldCruftyDatacenter"
	c := vOld.DeepCopy()

	c.Spec.Datacenter = "shinyNewDatacenter"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Network = "OldNet"
	c := vOld.DeepCopy()

	c.Spec.Network = "NewNet"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateTLSInsecureImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Insecure = true
	c := vOld.DeepCopy()

	c.Spec.Insecure = false
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateTlsThumbprintImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Thumbprint = "5334E1D85B267B78F99BAF553FEB2F94E72EFDFD"
	c := vOld.DeepCopy()

	c.Spec.Thumbprint = "B3D1C464976E725E599D3548180CB56311818F224E701F9D56F22E8079A7B396"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Network = "oldNetwork"
	c := vOld.DeepCopy()

	c.Spec.Network = "newNetwork"

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereDatacenterValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.VSphereDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateInvalidServer(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Server = ""
	c.Spec.Server = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateInvalidDatacenter(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Datacenter = ""
	c.Spec.Datacenter = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereDatacenterValidateUpdateInvalidNetwork(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Network = ""
	c.Spec.Network = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func vsphereDatacenterConfig() v1alpha1.VSphereDatacenterConfig {
	return v1alpha1.VSphereDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "datacenter",
			Network:    "/datacenter/network-1",
			Server:     "vcenter.com",
			Insecure:   false,
			Thumbprint: "abc",
		},
		Status: v1alpha1.VSphereDatacenterConfigStatus{},
	}
}
