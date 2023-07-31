package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestTinkerbellDatacenterValidateCreate(t *testing.T) {
	dataCenterConfig := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).To(Succeed())
}

func TestTinkerbellDatacenterValidateCreateFail(t *testing.T) {
	dataCenterConfig := tinkerbellDatacenterConfig()
	dataCenterConfig.Spec.TinkerbellIP = ""

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).NotTo(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateSucceed(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()
	tOld.Spec.TinkerbellIP = "1.1.1.1"
	tNew := tOld.DeepCopy()

	tNew.Spec.TinkerbellIP = "1.1.1.1"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(&tOld)).To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateSucceedOSImageURL(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()
	tNew := tOld.DeepCopy()

	tNew.Spec.OSImageURL = "https://os-image-url"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(&tOld)).To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateFailBadReq(t *testing.T) {
	cOld := &v1alpha1.Cluster{}
	c := &v1alpha1.TinkerbellDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("expected a TinkerbellDatacenterConfig but got a *v1alpha1.Cluster")))
}

func TestTinkerbellDatacenterValidateUpdateImmutableTinkIP(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()
	tOld.Spec.TinkerbellIP = "1.1.1.1"
	tNew := tOld.DeepCopy()

	tNew.Spec.TinkerbellIP = "1.1.1.2"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(&tOld)).To(MatchError(ContainSubstring("spec.tinkerbellIP: Forbidden: field is immutable")))
}

func TestTinkerbellDatacenterValidateDelete(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(tOld.ValidateDelete()).To(Succeed())
}

func tinkerbellDatacenterConfig() v1alpha1.TinkerbellDatacenterConfig {
	return v1alpha1.TinkerbellDatacenterConfig{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string, 1),
			Name:        "tinkerbelldatacenterconfig",
		},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "1.1.1.1",
		},
		Status: v1alpha1.TinkerbellDatacenterConfigStatus{},
	}
}
