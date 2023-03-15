package cilium_test

import (
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

func TestInstallationInstalled(t *testing.T) {
	tests := []struct {
		name         string
		installation cilium.Installation
		want         bool
	}{
		{
			name: "installed",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "cilium-eksa"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want: true,
		},
		{
			name: "ds not installed",
			installation: cilium.Installation{
				Operator: &appsv1.Deployment{},
			},
			want: false,
		},
		{
			name: "ds not installed with eksa cilium",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "cilium"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want: false,
		},
		{
			name: "operator not installed",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{},
			},
			want: false,
		},
		{
			name:         "none installed",
			installation: cilium.Installation{},
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.installation.Installed()).To(Equal(tt.want))
		})
	}
}
