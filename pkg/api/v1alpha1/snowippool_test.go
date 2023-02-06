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

func TestSnowIPPoolValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *v1alpha1.SnowIPPool
		wantErr string
	}{
		{
			name: "valid ip pool",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "ip start empty",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPStart can not be empty",
		},
		{
			name: "ip start invalid",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "invalid",
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPStart is invalid",
		},
		{
			name: "ip end empty",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPEnd can not be empty",
		},
		{
			name: "ip end invalid",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "invalid",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPEnd is invalid",
		},
		{
			name: "ip gateway empty",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.3.5",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].Gateway can not be empty",
		},
		{
			name: "ip gateway invalid",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.3.5",
							Gateway: "invalid",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].Gateway is invalid",
		},
		{
			name: "ip end smaller than ip start",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.5",
							IPEnd:   "1.2.3.2",
							Gateway: "1.2.3.2",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPStart should be smaller than IPEnd",
		},
		{
			name: "subnet empty",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].Subnet can not be empty",
		},
		{
			name: "subnet invalid",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
							Subnet:  "invalid",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].Subnet is invalid",
		},
		{
			name: "ip start fell out of subnet range",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.2.2",
							IPEnd:   "1.2.3.5",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPStart should be within the subnet range 1.2.3.4/24",
		},
		{
			name: "ip end fell out of subnet range",
			obj: &v1alpha1.SnowIPPool{
				Spec: v1alpha1.SnowIPPoolSpec{
					Pools: []v1alpha1.IPPool{
						{
							IPStart: "1.2.3.2",
							IPEnd:   "1.2.4.4",
							Gateway: "1.2.3.1",
							Subnet:  "1.2.3.4/24",
						},
					},
				},
			},
			wantErr: "SnowIPPool Pools[0].IPEnd should be within the subnet range 1.2.3.4/24",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := tt.obj.Validate()
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
