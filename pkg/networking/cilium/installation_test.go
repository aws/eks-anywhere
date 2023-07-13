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

func TestInstallationVersion(t *testing.T) {
	tests := []struct {
		name         string
		installation cilium.Installation
		want         string
		wantErr      string
	}{
		{
			name: "no cilium daemon set, no version",
			installation: cilium.Installation{
				DaemonSet: nil,
				Operator:  &appsv1.Deployment{},
			},
			want:    "",
			wantErr: "",
		},
		{
			name: "repository, cilium and version",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "public.ecr.aws/isovalent/cilium:v1.11.15-eksa.1"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want:    "v1.11.15-eksa.1",
			wantErr: "",
		},
		{
			name: "cilium and version",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "cilium:v1.11.15-eksa.1"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want:    "v1.11.15-eksa.1",
			wantErr: "",
		},
		{
			name: "cilium and version only, no pre-release",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "cilium:v1.11.15"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want:    "v1.11.15",
			wantErr: "",
		},
		{
			name: "invalid semver",
			installation: cilium.Installation{
				DaemonSet: &appsv1.DaemonSet{
					Spec: appsv1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Image: "v1.15-eksa.1"},
								},
							},
						},
					},
				},
				Operator: &appsv1.Deployment{},
			},
			want:    "",
			wantErr: "invalid major version in semver",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			version, err := tt.installation.Version()
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring("")))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(version).To(Equal(tt.want))
			}
		})
	}
}
