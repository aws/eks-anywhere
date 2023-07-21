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
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("spec.server: Forbidden: field is immutable")))
}

func TestVSphereDatacenterValidateUpdateDataCenterImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Datacenter = "oldCruftyDatacenter"
	c := vOld.DeepCopy()

	c.Spec.Datacenter = "/shinyNewDatacenter"
	c.Spec.Network = "/shinyNewDatacenter/network"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("spec.datacenter: Forbidden: field is immutable")))
}

func TestVSphereDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Network = "OldNet"
	c := vOld.DeepCopy()

	c.Spec.Network = "/datacenter/network"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("spec.network: Forbidden: field is immutable")))
}

func TestVSphereDatacenterValidateUpdateTLSInsecureMutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Insecure = true
	c := vOld.DeepCopy()

	c.Spec.Insecure = false
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereDatacenterValidateUpdateTlsThumbprintMutable(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Thumbprint = "5334E1D85B267B78F99BAF553FEB2F94E72EFDFD"
	c := vOld.DeepCopy()

	c.Spec.Thumbprint = "B3D1C464976E725E599D3548180CB56311818F224E701F9D56F22E8079A7B396"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	vOld.Spec.Network = "/datacenter/oldNetwork"
	c := vOld.DeepCopy()

	c.Spec.Network = "/datacenter/network/newNetwork"

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereDatacenterValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.VSphereDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).To(MatchError(ContainSubstring("expected a VSphereDataCenterConfig but got a *v1alpha1.Cluster")))
}

func TestVSphereDatacenterValidateUpdateInvalidServer(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Server = ""
	c.Spec.Server = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("VSphereDatacenterConfig server is not set or is empty")))
}

func TestVSphereDatacenterValidateUpdateInvalidDatacenter(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Datacenter = ""
	c.Spec.Datacenter = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("VSphereDatacenterConfig datacenter is not set or is empty")))
}

func TestVSphereDatacenterValidateUpdateInvalidNetwork(t *testing.T) {
	vOld := vsphereDatacenterConfig()
	c := vOld.DeepCopy()
	vOld.Spec.Network = ""
	c.Spec.Network = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(MatchError(ContainSubstring("VSphereDatacenterConfig VM network is not set or is empty")))
}

func TestVSphereDatacenterConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := vsphereDatacenterConfig()
	sOld.Spec.Network = "network-1"
	sOld.Default()

	g.Expect(sOld.Spec.Network).To(Equal("/datacenter/network/network-1"))
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

func TestVSphereDatacenterValidateCreateFullManagementCycleOn(t *testing.T) {
	dataCenterConfig := vsphereDatacenterConfig()

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).To(Succeed())
}

func TestVSphereDatacenterValidateCreate(t *testing.T) {
	dataCenterConfig := vsphereDatacenterConfig()

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).To(Succeed())
}

func TestVSphereDatacenterValidateCreateFail(t *testing.T) {
	dataCenterConfig := vsphereDatacenterConfig()
	dataCenterConfig.Spec.Datacenter = ""

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).To(MatchError(ContainSubstring("VSphereDatacenterConfig datacenter is not set or is empty")))
}
