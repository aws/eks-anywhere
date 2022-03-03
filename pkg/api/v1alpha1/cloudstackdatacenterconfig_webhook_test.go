package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackDatacenterValidateUpdateDomainImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Domain = "oldCruftyDomain"
	c := vOld.DeepCopy()

	c.Spec.Domain = "shinyNewDomain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateZoneImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Zone = v1alpha1.CloudStackResourceRef{
		Type:  v1alpha1.Name,
		Value: "oldCruftyZone",
	}
	c := vOld.DeepCopy()

	c.Spec.Zone = v1alpha1.CloudStackResourceRef{
		Type:  v1alpha1.Name,
		Value: "shinyNewZone",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateAccountImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Account = "oldCruftyAccount"
	c := vOld.DeepCopy()

	c.Spec.Account = "shinyNewAccount"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "OldNet",
		Type:  v1alpha1.Name,
	}
	c := vOld.DeepCopy()

	c.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "NewNet",
		Type:  v1alpha1.Name,
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "oldNetwork",
		Type:  v1alpha1.Name,
	}
	c := vOld.DeepCopy()

	c.Spec.Network = v1alpha1.CloudStackResourceRef{
		Value: "newNetwork",
		Type:  v1alpha1.Name,
	}

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCloudStackDatacenterValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackDatacenterConfig() v1alpha1.CloudStackDatacenterConfig {
	return v1alpha1.CloudStackDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.CloudStackDatacenterConfigSpec{},
		Status:     v1alpha1.CloudStackDatacenterConfigStatus{},
	}
}
