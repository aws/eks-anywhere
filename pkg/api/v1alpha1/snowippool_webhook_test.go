package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestSnowIPPoolValidateCreate(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	g.Expect(new.ValidateCreate()).To(Succeed())
}

func TestSnowIPPoolValidateUpdate(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	g.Expect(new.ValidateUpdate(old)).To(Succeed())
}

func TestSnowIPPoolValidateUpdateImmutableFields(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	old.Spec.Pools = []snowv1.IPPool{
		{
			IPStart: ptr.String("start"),
		},
	}
	g.Expect(new.ValidateUpdate(old)).To(MatchError(ContainSubstring("spec.pools: Forbidden: field is immutable")))
}

func TestSnowIPPoolValidateDelete(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	g.Expect(new.ValidateDelete()).To(Succeed())
}

func snowIPPool() v1alpha1.SnowIPPool {
	return v1alpha1.SnowIPPool{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.SnowIPPoolSpec{},
		Status:     v1alpha1.SnowIPPoolStatus{},
	}
}
