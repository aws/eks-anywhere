package cluster_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestSetDefaultFluxConfigPath(t *testing.T) {
	tests := []struct {
		name           string
		config         *cluster.Config
		wantConfigPath string
	}{
		{
			name: "self-managed cluster",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c-1",
					},
					Spec: anywherev1.ClusterSpec{
						ManagementCluster: anywherev1.ManagementCluster{
							Name: "c-1",
						},
					},
				},
				FluxConfig: &anywherev1.FluxConfig{},
			},
			wantConfigPath: "clusters/c-1",
		},
		{
			name: "managed cluster",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c-1",
					},
					Spec: anywherev1.ClusterSpec{
						ManagementCluster: anywherev1.ManagementCluster{
							Name: "c-m",
						},
					},
				},
				FluxConfig: &anywherev1.FluxConfig{},
			},
			wantConfigPath: "clusters/c-m",
		},
		{
			name: "config path is already set",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c-1",
					},
					Spec: anywherev1.ClusterSpec{
						ManagementCluster: anywherev1.ManagementCluster{
							Name: "c-m",
						},
					},
				},
				FluxConfig: &anywherev1.FluxConfig{
					Spec: anywherev1.FluxConfigSpec{
						ClusterConfigPath: "my-path",
					},
				},
			},
			wantConfigPath: "my-path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(cluster.SetDefaultFluxConfigPath(tt.config)).To(Succeed())
			g.Expect(tt.config.FluxConfig.Spec.ClusterConfigPath).To(Equal(tt.wantConfigPath))
		})
	}
}

func TestSetConfigDefaults(t *testing.T) {
	g := NewWithT(t)
	c := clusterConfigFromFile(t, "testdata/cluster_1_19.yaml")
	originalC := clusterConfigFromFile(t, "testdata/cluster_1_19.yaml")
	g.Expect(cluster.SetConfigDefaults(c)).To(Succeed())
	g.Expect(c).NotTo(Equal(originalC))
}

func TestNewControlPlaneIPCheckAnnotationDefaulterNoAnnotation(t *testing.T) {
	g := NewWithT(t)

	newControlPlaneIPCheckAnnotationDefaulter := cluster.NewControlPlaneIPCheckAnnotationDefaulter(false)

	c := baseCluster()

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: c,
		},
	}

	updatedClusterSpec, err := newControlPlaneIPCheckAnnotationDefaulter.ControlPlaneIPCheckDefault(context.Background(), clusterSpec)

	g.Expect(err).To(BeNil())
	g.Expect(clusterSpec).To(Equal(updatedClusterSpec))
}

func TestNewControlPlaneIPCheckAnnotationDefaulterAddAnnotation(t *testing.T) {
	g := NewWithT(t)

	newControlPlaneIPCheckAnnotationDefaulter := cluster.NewControlPlaneIPCheckAnnotationDefaulter(true)

	c := baseCluster()

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: c,
		},
	}

	oldCluster := clusterSpec.Config.Cluster.DeepCopy()
	oldCluster.DisableControlPlaneIPCheck()

	_, err := newControlPlaneIPCheckAnnotationDefaulter.ControlPlaneIPCheckDefault(context.Background(), clusterSpec)

	g.Expect(err).To(BeNil())
	g.Expect(oldCluster).To(Equal(clusterSpec.Config.Cluster))
}

type clusterOpt func(c *anywherev1.Cluster)

func baseCluster(opts ...clusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mgmt",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube121,
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:       "bundles-1",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}},
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
			},
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
