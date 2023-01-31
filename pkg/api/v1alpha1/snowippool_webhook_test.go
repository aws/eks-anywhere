package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowIPPoolValidateCreate(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	g.Expect(new.ValidateCreate()).To(Succeed())
}

func TestSnowIPPoolValidateCreateInvalidIPPool(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	new.Spec.Pools[0].IPStart = "invalid"
	g.Expect(new.ValidateCreate()).To(MatchError(ContainSubstring("SnowIPPool Pools[0].IPStart is invalid")))
}

func TestSnowIPPoolValidateUpdate(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	g.Expect(new.ValidateUpdate(old)).To(Succeed())
}

func TestSnowIPPoolValidateUpdateInvalidIPPool(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	new.Spec.Pools[0].IPStart = "invalid"
	old := new.DeepCopy()
	g.Expect(new.ValidateUpdate(old)).To(MatchError(ContainSubstring("SnowIPPool Pools[0].IPStart is invalid")))
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
	new.Spec.Pools = []v1alpha1.IPPool{
		{
			IPStart: "192.168.1.20",
			IPEnd:   "192.168.1.30",
			Gateway: "192.168.1.1",
			Subnet:  "192.168.1.0/24",
		},
		{
			IPStart: "192.168.1.2",
			IPEnd:   "192.168.1.14",
			Gateway: "192.168.1.1",
			Subnet:  "192.168.1.0/24",
		},
	}
	g.Expect(new.ValidateUpdate(old)).To(Succeed())
}

func TestSnowIPPoolValidateUpdateIPPoolsLengthDiff(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	old.Spec.Pools = []v1alpha1.IPPool{
		{
			IPStart: "start",
		},
	}
	g.Expect(new.ValidateUpdate(old)).To(MatchError(ContainSubstring("spec.pools: Forbidden: field is immutable")))
}

func TestSnowIPPoolValidateUpdateIPPoolsDiff(t *testing.T) {
	g := NewWithT(t)
	new := snowIPPool()
	old := new.DeepCopy()
	new.Spec.Pools = []v1alpha1.IPPool{
		{
			IPStart: "192.168.1.21",
			IPEnd:   "192.168.1.30",
			Gateway: "192.168.1.1",
			Subnet:  "192.168.1.0/24",
		},
		{
			IPStart: "192.168.1.2",
			IPEnd:   "192.168.1.14",
			Gateway: "192.168.1.1",
			Subnet:  "192.168.1.0/24",
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
		Spec: v1alpha1.SnowIPPoolSpec{
			Pools: []v1alpha1.IPPool{
				{
					IPStart: "192.168.1.2",
					IPEnd:   "192.168.1.14",
					Gateway: "192.168.1.1",
					Subnet:  "192.168.1.0/24",
				},
				{
					IPStart: "192.168.1.20",
					IPEnd:   "192.168.1.30",
					Gateway: "192.168.1.1",
					Subnet:  "192.168.1.0/24",
				},
			},
		},
		Status: v1alpha1.SnowIPPoolStatus{},
	}
}
