package cilium_test

import (
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

func TestBuildUpgradePlan(t *testing.T) {
	tests := []struct {
		name         string
		installation *cilium.Installation
		clusterSpec  *cluster.Spec
		want         cilium.UpgradePlan
	}{
		{
			name: "no upgrade needed",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.0"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.ComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
			},
		},
		{
			name: "daemon set not installed",
			installation: &cilium.Installation{
				Operator: deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.0"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "DaemonSet doesn't exist",
					NewImage:      "cilium:v1.0.0",
				},
				Operator: cilium.ComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
			},
		},
		{
			name: "daemon container old version",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.1"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.0"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "DaemonSet container agent doesn't match image [cilium:v1.0.0] -> [cilium:v1.0.1]",
					OldImage:      "cilium:v1.0.0",
					NewImage:      "cilium:v1.0.1",
				},
				Operator: cilium.ComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
			},
		},
		{
			name: "daemon init container old version",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.1", func(ds *appsv1.DaemonSet) {
					ds.Spec.Template.Spec.InitContainers = []corev1.Container{
						{
							Name:  "init",
							Image: "cilium:v1.0.0",
						},
					}
				}),
				Operator: deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.1"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.0"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "DaemonSet container init doesn't match image [cilium:v1.0.0] -> [cilium:v1.0.1]",
					OldImage:      "cilium:v1.0.0",
					NewImage:      "cilium:v1.0.1",
				},
				Operator: cilium.ComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
			},
		},
		{
			name: "operator is not present",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.0"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "Operator deployment doesn't exist",
					NewImage:      "cilium-operator:v1.0.0",
				},
			},
		},
		{
			name: "operator 0 containers",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator: deployment("cilium-operator:v1.0.0", func(d *appsv1.Deployment) {
					d.Spec.Template.Spec.Containers = nil
				}),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.1"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "Operator deployment doesn't have any containers",
					NewImage:      "cilium-operator:v1.0.1",
				},
			},
		},
		{
			name: "operator container old version",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundle.Cilium.Operator.URI = "cilium-operator:v1.0.1"
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "Operator container doesn't match the provided image [cilium-operator:v1.0.0] -> [cilium-operator:v1.0.1]",
					OldImage:      "cilium-operator:v1.0.0",
					NewImage:      "cilium-operator:v1.0.1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(
				cilium.BuildUpgradePlan(tt.installation, tt.clusterSpec),
			).To(Equal(tt.want))
		})
	}
}

type deploymentOpt func(*appsv1.Deployment)

func deployment(image string, opts ...deploymentOpt) *appsv1.Deployment {
	d := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: image,
						},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

type dsOpt func(*appsv1.DaemonSet)

func daemonSet(image string, opts ...dsOpt) *appsv1.DaemonSet {
	d := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "agent",
							Image: image,
						},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func TestComponentUpgradePlanNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.ComponentUpgradePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.ComponentUpgradePlan{
				UpgradeReason: "",
			},
			want: false,
		},
		{
			name: "needed",
			info: cilium.ComponentUpgradePlan{
				UpgradeReason: "missing ds",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.Needed()).To(Equal(tt.want))
		})
	}
}

func TestUpgradePlanNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.UpgradePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{},
				Operator:  cilium.ComponentUpgradePlan{},
			},
			want: false,
		},
		{
			name: "ds needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.ComponentUpgradePlan{},
			},
			want: true,
		},
		{
			name: "operator needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
			},
			want: true,
		},
		{
			name: "both needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.Needed()).To(Equal(tt.want))
		})
	}
}

func TestUpgradePlanReason(t *testing.T) {
	tests := []struct {
		name string
		info cilium.UpgradePlan
		want string
	}{
		{
			name: "not needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{},
				Operator:  cilium.ComponentUpgradePlan{},
			},
			want: "",
		},
		{
			name: "ds needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.ComponentUpgradePlan{},
			},
			want: "ds old version",
		},
		{
			name: "operator needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
			},
			want: "operator old version",
		},
		{
			name: "both needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.ComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.ComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
			},
			want: "ds old version - operator old version",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.Reason()).To(Equal(tt.want))
		})
	}
}
