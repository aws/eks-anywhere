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

func TestSnowIPPoolValidateUpdateInvalidObjectType(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := &v1alpha1.SnowDatacenterConfig{}
	g.Expect(new.ValidateUpdate(old)).To(MatchError(ContainSubstring("expected a SnowIPPool but got a *v1alpha1.SnowDatacenterConfig")))
}

func TestSnowIPPoolValidateUpdateIPPoolsSame(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	new.Spec.Pools = []snowv1.IPPool{
		{
			IPStart: nil,
			IPEnd:   ptr.String("end-2"),
			Subnet:  ptr.String("subnet-2"),
			Gateway: ptr.String("gateway-2"),
		},
		{
			IPStart: ptr.String("start-1"),
			IPEnd:   nil,
			Subnet:  ptr.String("subnet-1"),
			Gateway: nil,
		},
	}
	old.Spec.Pools = []snowv1.IPPool{
		{
			IPStart: ptr.String("start-1"),
			IPEnd:   nil,
			Subnet:  ptr.String("subnet-1"),
			Gateway: nil,
		},
		{
			IPStart: nil,
			IPEnd:   ptr.String("end-2"),
			Subnet:  ptr.String("subnet-2"),
			Gateway: ptr.String("gateway-2"),
		},
	}
	g.Expect(new.ValidateUpdate(old)).To(Succeed())
}

func TestSnowIPPoolValidateUpdateIPPoolsLengthDiff(t *testing.T) {
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

func TestSnowIPPoolValidateUpdateIPPoolsDiff(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	new.Spec.Pools = []snowv1.IPPool{
		{
			IPStart: ptr.String("start-1"),
			IPEnd:   nil,
			Subnet:  ptr.String("subnet-1"),
			Gateway: nil,
		},
	}
	old.Spec.Pools = []snowv1.IPPool{
		{
			IPStart: nil,
			IPEnd:   ptr.String("end-2"),
			Subnet:  ptr.String("subnet-2"),
			Gateway: ptr.String("gateway-2"),
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
