package cilium_test

import (
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "daemon set not installed",
			installation: &cilium.Installation{
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "DaemonSet doesn't exist",
					NewImage:      "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "daemon container old version",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.1"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "DaemonSet container agent doesn't match image [cilium:v1.0.0] -> [cilium:v1.0.1]",
					OldImage:      "cilium:v1.0.0",
					NewImage:      "cilium:v1.0.1",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
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
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.1"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "DaemonSet container init doesn't match image [cilium:v1.0.0] -> [cilium:v1.0.1]",
					OldImage:      "cilium:v1.0.0",
					NewImage:      "cilium:v1.0.1",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "operator is not present",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "Operator deployment doesn't exist",
					NewImage:      "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
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
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.1"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "Operator deployment doesn't have any containers",
					NewImage:      "cilium-operator:v1.0.1",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "operator container old version",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.1"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "Operator container doesn't match the provided image [cilium-operator:v1.0.0] -> [cilium-operator:v1.0.1]",
					OldImage:      "cilium-operator:v1.0.0",
					NewImage:      "cilium-operator:v1.0.1",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "config map doesn't exist",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "Cilium config doesn't exist",
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							NewValue: "default",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "PolicyEnforcementMode has changed",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", ""),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{
						PolicyEnforcementMode: anywherev1.CiliumPolicyModeAlways,
					},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "Cilium enable-policy changed: [default] -> [always]",
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:         cilium.PolicyEnforcementComponentName,
							OldValue:     "default",
							NewValue:     "always",
							UpdateReason: "Cilium enable-policy changed: [default] -> [always]",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "PolicyEnforcementMode not present in config",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", "", func(cm *corev1.ConfigMap) {
					cm.Data = nil
				}),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{
						PolicyEnforcementMode: anywherev1.CiliumPolicyModeAlways,
					},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "Cilium enable-policy field is not present in config",
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:         cilium.PolicyEnforcementComponentName,
							OldValue:     "",
							NewValue:     "always",
							UpdateReason: "Cilium enable-policy field is not present in config",
						},
						{
							Name: "EgressMasqueradeInterfaces",
						},
					},
				},
			},
		},
		{
			name: "EgressMasqueradeInterfaces has changed",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", "old"),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{
						EgressMasqueradeInterfaces: "new",
					},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "Egress masquerade interfaces changed: [old] -> [new]",
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:     cilium.PolicyEnforcementComponentName,
							OldValue: "default",
							NewValue: "default",
						},
						{
							Name:         cilium.EgressMasqueradeInterfacesComponentName,
							OldValue:     "old",
							NewValue:     "new",
							UpdateReason: "Egress masquerade interfaces changed: [old] -> [new]",
						},
					},
				},
			},
		},
		{
			name: "EgressMasqueradeInterfaces not present in config",
			installation: &cilium.Installation{
				DaemonSet: daemonSet("cilium:v1.0.0"),
				Operator:  deployment("cilium-operator:v1.0.0"),
				ConfigMap: ciliumConfigMap("default", "", func(cm *corev1.ConfigMap) {
					cm.Data = nil
				}),
			},
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:v1.0.0"
				s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:v1.0.0"
				s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{
						EgressMasqueradeInterfaces: "new",
					},
				}
			}),
			want: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium:v1.0.0",
					NewImage: "cilium:v1.0.0",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					OldImage: "cilium-operator:v1.0.0",
					NewImage: "cilium-operator:v1.0.0",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "Cilium enable-policy field is not present in config - Egress masquerade interfaces field is not present in config but is configured in cluster spec",
					Components: []cilium.ConfigComponentUpdatePlan{
						{
							Name:         cilium.PolicyEnforcementComponentName,
							OldValue:     "",
							NewValue:     "default",
							UpdateReason: "Cilium enable-policy field is not present in config",
						},
						{
							Name:         cilium.EgressMasqueradeInterfacesComponentName,
							OldValue:     "",
							NewValue:     "new",
							UpdateReason: "Egress masquerade interfaces field is not present in config but is configured in cluster spec",
						},
					},
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

type cmOpt func(*corev1.ConfigMap)

func ciliumConfigMap(enforcementMode string, egressMasqueradeInterface string, opts ...cmOpt) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cilium.ConfigMapName,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			cilium.PolicyEnforcementConfigMapKey:    enforcementMode,
			cilium.EgressMasqueradeInterfacesMapKey: egressMasqueradeInterface,
		},
	}

	for _, o := range opts {
		o(cm)
	}

	return cm
}

func TestConfigUpdatePlanNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.ConfigUpdatePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.ConfigUpdatePlan{
				UpdateReason: "",
			},
			want: false,
		},
		{
			name: "needed",
			info: cilium.ConfigUpdatePlan{
				UpdateReason: "missing ds",
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

func TestVersionedComponentUpgradePlanNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.VersionedComponentUpgradePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.VersionedComponentUpgradePlan{
				UpgradeReason: "",
			},
			want: false,
		},
		{
			name: "needed",
			info: cilium.VersionedComponentUpgradePlan{
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
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: false,
		},
		{
			name: "ds needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: true,
		},
		{
			name: "operator needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: true,
		},
		{
			name: "config needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "config has changed",
				},
			},
			want: true,
		},
		{
			name: "all needed needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "config has changed",
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

func TestUpgradePlanVersionUpgradeNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.UpgradePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "config has changed",
				},
			},
			want: false,
		},
		{
			name: "ds needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: true,
		},
		{
			name: "operator needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: true,
		},
		{
			name: "both needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.VersionUpgradeNeeded()).To(Equal(tt.want))
		})
	}
}

func TestUpgradePlanConfigUpdateNeeded(t *testing.T) {
	tests := []struct {
		name string
		info cilium.UpgradePlan
		want bool
	}{
		{
			name: "not needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{},
			},
			want: false,
		},
		{
			name: "config needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator:  cilium.VersionedComponentUpgradePlan{},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "config has changed",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.ConfigUpdateNeeded()).To(Equal(tt.want))
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
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator:  cilium.VersionedComponentUpgradePlan{},
			},
			want: "",
		},
		{
			name: "ds needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.VersionedComponentUpgradePlan{},
			},
			want: "ds old version",
		},
		{
			name: "operator needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
			},
			want: "operator old version",
		},
		{
			name: "all needed",
			info: cilium.UpgradePlan{
				DaemonSet: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "ds old version",
				},
				Operator: cilium.VersionedComponentUpgradePlan{
					UpgradeReason: "operator old version",
				},
				ConfigMap: cilium.ConfigUpdatePlan{
					UpdateReason: "config has changed",
				},
			},
			want: "ds old version - operator old version - config has changed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.info.Reason()).To(Equal(tt.want))
		})
	}
}
