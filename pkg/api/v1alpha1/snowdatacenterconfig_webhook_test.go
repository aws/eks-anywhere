package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowDatacenterConfigValidateCreateValid(t *testing.T) {
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"
	snowDC.Spec.IdentityRef.Kind = v1alpha1.SnowIdentityKind

	g.Expect(snowDC.ValidateCreate()).To(Succeed())
}

func TestSnowDatacenterConfigValidateCreateEmptyIdentityRef(t *testing.T) {
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()

	g.Expect(snowDC.ValidateCreate()).To(MatchError(ContainSubstring("IdentityRef name must not be empty")))
}

func TestSnowDatacenterConfigValidateCreateEmptyIdentityKind(t *testing.T) {
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"

	g.Expect(snowDC.ValidateCreate()).To(MatchError(ContainSubstring("IdentityRef kind must not be empty")))
}

func TestSnowDatacenterConfigValidateCreateIdentityKindNotSnow(t *testing.T) {
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"
	snowDC.Spec.IdentityRef.Kind = v1alpha1.OIDCConfigKind

	g.Expect(snowDC.ValidateCreate()).To(MatchError(ContainSubstring("is invalid, the only supported kind is Secret")))
}

func TestSnowDatacenterConfigValidateValidateEmptyIdentityRef(t *testing.T) {
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	g.Expect(snowDCNew.ValidateUpdate(&snowDCOld)).To(MatchError(ContainSubstring("IdentityRef name must not be empty")))
}

func TestSnowDatacenterConfigValidateValidateEmptyIdentityKind(t *testing.T) {
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	snowDCNew.Spec.IdentityRef.Name = "refName"

	g.Expect(snowDCNew.ValidateUpdate(&snowDCOld)).To(MatchError(ContainSubstring("IdentityRef kind must not be empty")))
}

func TestSnowDatacenterConfigValidateValidateIdentityKindNotSnow(t *testing.T) {
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	snowDCNew.Spec.IdentityRef.Name = "refName"
	snowDCNew.Spec.IdentityRef.Kind = v1alpha1.OIDCConfigKind

	g.Expect(snowDCNew.ValidateUpdate(&snowDCOld)).To(MatchError(ContainSubstring("is invalid, the only supported kind is Secret")))
}

func snowDatacenterConfig() v1alpha1.SnowDatacenterConfig {
	return v1alpha1.SnowDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.SnowDatacenterConfigSpec{},
		Status:     v1alpha1.SnowDatacenterConfigStatus{},
	}
}
