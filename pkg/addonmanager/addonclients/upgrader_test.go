package addonclients_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/types"
)

type upgraderTest struct {
	*WithT
	ctx         context.Context
	currentSpec *cluster.Spec
	newSpec     *cluster.Spec
	cluster     *types.Cluster
	fluxConfig  v1alpha1.Flux
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles.Spec.Number = 1
		s.VersionsBundle.Flux.Version = "v0.1.0"
	})

	return &upgraderTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
		fluxConfig: v1alpha1.Flux{
			Github: v1alpha1.Github{
				Owner:               "mFowler",
				Repository:          "testRepo",
				FluxSystemNamespace: "flux-system",
				Branch:              "testBranch",
				ClusterConfigPath:   "clusters/fluxAddonTestCluster/flux-system",
				Personal:            true,
			},
		},
	}
}

func TestFluxUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestFluxUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.1.0"

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestFluxUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.2.0"
	fluxSystemDirPath := "clusters/fluxAddonTestCluster"

	tt.newSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: tt.fluxConfig,
		},
	}
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(tt.ctx).Return(&git.Repository{Name: tt.fluxConfig.Github.Repository}, nil)
	m.git.EXPECT().Clone(tt.ctx).Return(nil)
	m.git.EXPECT().Branch(tt.fluxConfig.Github.Branch).Return(nil)
	m.git.EXPECT().Add(fluxSystemDirPath).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(tt.ctx).Return(nil)

	m.flux.EXPECT().BootstrapToolkitsComponents(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig)
	m.flux.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig)

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestFluxUpgradeError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.2.0"
	fluxSystemDirPath := "clusters/fluxAddonTestCluster"

	tt.newSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: tt.fluxConfig,
		},
	}
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(tt.ctx).Return(&git.Repository{Name: tt.fluxConfig.Github.Repository}, nil)
	m.git.EXPECT().Clone(tt.ctx).Return(nil)
	m.git.EXPECT().Branch(tt.fluxConfig.Github.Branch).Return(nil)
	m.git.EXPECT().Add(fluxSystemDirPath).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(tt.ctx).Return(nil)

	m.flux.EXPECT().BootstrapToolkitsComponents(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig).Return(errors.New("error from client"))

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).NotTo(Succeed())
}
