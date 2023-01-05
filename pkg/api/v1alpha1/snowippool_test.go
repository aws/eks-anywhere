package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowIPPoolConvertConfigToConfigGenerateStruct(t *testing.T) {
	g := NewWithT(t)

	s := &v1alpha1.SnowIPPool{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.SnowIPPoolKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ippool",
		},
		Spec: v1alpha1.SnowIPPoolSpec{
			Pools: []v1alpha1.IPPool{
				{
					IPStart: "start",
					IPEnd:   "end",
					Gateway: "gateway",
					Subnet:  "subnet",
				},
			},
		},
	}

	want := &v1alpha1.SnowIPPoolGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.SnowIPPoolKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1alpha1.ObjectMeta{
			Name:      "ippool",
			Namespace: "default",
		},
		Spec: v1alpha1.SnowIPPoolSpec{
			Pools: []v1alpha1.IPPool{
				{
					IPStart: "start",
					IPEnd:   "end",
					Gateway: "gateway",
					Subnet:  "subnet",
				},
			},
		},
	}

	g.Expect(s.ConvertConfigToConfigGenerateStruct()).To(Equal(want))
}
