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
