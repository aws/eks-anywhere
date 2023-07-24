package clustermarshaller_test

import (
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func TestWriteClusterConfigWithOIDCAndGitOps(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.APIVersion = v1alpha1.GroupVersion.String()
		s.Cluster.TypeMeta.Kind = v1alpha1.ClusterKind
		s.Cluster.CreationTimestamp = v1.Time{Time: time.Now()}
		s.Cluster.Name = "mycluster"
		s.Cluster.Spec.KubernetesVersion = ""
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
			Count: 3,
		}
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "config",
		}
		s.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "config",
			},
		}
		s.OIDCConfig = &v1alpha1.OIDCConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.OIDCConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "config",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.OIDCConfigSpec{
				IssuerUrl: "https://url",
			},
		}

		s.GitOpsConfig = &v1alpha1.GitOpsConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.GitOpsConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "config",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.GitOpsConfigSpec{
				Flux: v1alpha1.Flux{
					Github: v1alpha1.Github{
						Owner:               "me",
						Branch:              "main",
						ClusterConfigPath:   "clusters/mycluster",
						FluxSystemNamespace: "flux-system",
					},
				},
			},
		}
		s.Cluster.SetSelfManaged()
	})

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.VSphereDatacenterKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              "config",
			CreationTimestamp: v1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Server: "https://url",
		},
	}

	machineConfigs := []providers.MachineConfig{
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-1",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-2",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
	}
	g := NewWithT(t)

	folder, writer := test.NewWriter(t)
	gotFile := filepath.Join(folder, "mycluster-eks-a-cluster.yaml")

	g.Expect(clustermarshaller.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs, writer)).To(Succeed())

	test.AssertFilesEquals(t, gotFile, "testdata/expected_marshalled_cluster.yaml")
}

func TestWriteClusterConfigWithFluxAndGitOpsConfigs(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.APIVersion = v1alpha1.GroupVersion.String()
		s.Cluster.TypeMeta.Kind = v1alpha1.ClusterKind
		s.Cluster.CreationTimestamp = v1.Time{Time: time.Now()}
		s.Cluster.Name = "mycluster"
		s.Cluster.Spec.KubernetesVersion = ""
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
			Count: 3,
		}
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "config",
		}

		s.FluxConfig = &v1alpha1.FluxConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.FluxConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "config",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.FluxConfigSpec{
				Github: &v1alpha1.GithubProviderConfig{
					Owner:      "test",
					Repository: "test",
				},
			},
		}

		s.GitOpsConfig = &v1alpha1.GitOpsConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.GitOpsConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "config",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.GitOpsConfigSpec{
				Flux: v1alpha1.Flux{
					Github: v1alpha1.Github{
						Owner:               "me",
						Branch:              "main",
						ClusterConfigPath:   "clusters/mycluster",
						FluxSystemNamespace: "flux-system",
						Repository:          "test",
					},
				},
			},
		}
		s.Cluster.SetSelfManaged()
	})

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.VSphereDatacenterKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              "config",
			CreationTimestamp: v1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Server: "https://url",
		},
	}

	machineConfigs := []providers.MachineConfig{
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-1",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-2",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
	}
	g := NewWithT(t)

	folder, writer := test.NewWriter(t)
	gotFile := filepath.Join(folder, "mycluster-eks-a-cluster.yaml")

	g.Expect(clustermarshaller.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs, writer)).To(Succeed())
	test.AssertFilesEquals(t, gotFile, "testdata/expected_marshalled_cluster_flux_and_gitops.yaml")
}

func TestWriteClusterConfigWithFluxConfig(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.APIVersion = v1alpha1.GroupVersion.String()
		s.Cluster.TypeMeta.Kind = v1alpha1.ClusterKind
		s.Cluster.CreationTimestamp = v1.Time{Time: time.Now()}
		s.Cluster.Name = "mycluster"
		s.Cluster.Spec.KubernetesVersion = ""
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
			Count: 3,
		}
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.FluxConfigKind,
			Name: "config",
		}

		s.FluxConfig = &v1alpha1.FluxConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.FluxConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "config",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.FluxConfigSpec{
				Github: &v1alpha1.GithubProviderConfig{
					Owner:      "test",
					Repository: "test",
				},
			},
		}
		s.Cluster.SetSelfManaged()
	})

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.VSphereDatacenterKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              "config",
			CreationTimestamp: v1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Server: "https://url",
		},
	}

	machineConfigs := []providers.MachineConfig{
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-1",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
		&v1alpha1.VSphereMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.VSphereMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              "machineconf-2",
				CreationTimestamp: v1.Time{Time: time.Now()},
			},
			Spec: v1alpha1.VSphereMachineConfigSpec{
				Folder: "my-folder",
			},
		},
	}
	g := NewWithT(t)

	folder, writer := test.NewWriter(t)
	gotFile := filepath.Join(folder, "mycluster-eks-a-cluster.yaml")

	g.Expect(clustermarshaller.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs, writer)).To(Succeed())

	test.AssertFilesEquals(t, gotFile, "testdata/expected_marshalled_cluster_flux_config.yaml")
}

func TestWriteClusterConfigSnow(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.ClusterKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "testcluster",
			},
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.SnowDatacenterKind,
					Name: "testsnow",
				},
				ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.SnowMachineConfigKind,
						Name: "testsnow",
					},
				},
			},
		}

		s.SnowIPPools = map[string]*v1alpha1.SnowIPPool{
			"ippool": {
				TypeMeta: v1.TypeMeta{
					Kind:       v1alpha1.SnowIPPoolKind,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
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
			},
		}
		s.Cluster.SetSelfManaged()
	})

	datacenterConfig := &v1alpha1.SnowDatacenterConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.SnowDatacenterKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "testsnow",
		},
	}

	machineConfigs := []providers.MachineConfig{
		&v1alpha1.SnowMachineConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       v1alpha1.SnowMachineConfigKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "testsnow",
			},
			Spec: v1alpha1.SnowMachineConfigSpec{
				Network: v1alpha1.SnowNetwork{
					DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
						{
							Index: 1,
							IPPoolRef: &v1alpha1.Ref{
								Kind: v1alpha1.SnowIPPoolKind,
								Name: "ippool",
							},
							Primary: true,
						},
					},
				},
			},
		},
	}
	g := NewWithT(t)
	folder, writer := test.NewWriter(t)
	g.Expect(clustermarshaller.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs, writer)).To(Succeed())
	test.AssertFilesEquals(t, filepath.Join(folder, "testcluster-eks-a-cluster.yaml"), "testdata/expected_marshalled_snow.yaml")
}
