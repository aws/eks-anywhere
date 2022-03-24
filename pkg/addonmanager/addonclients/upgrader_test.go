package addonclients_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
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
		s.Cluster = &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "management-cluster",
			},
			Spec: v1alpha1.ClusterSpec{
				GitOpsRef: &v1alpha1.Ref{
					Name: "testGitOpsRef",
				},
			},
		}
	})

	return &upgraderTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "management-cluster",
			KubeconfigFile: "k.kubeconfig",
		},
		fluxConfig: v1alpha1.Flux{
			Github: v1alpha1.Github{
				Owner:               "mFowler",
				Repository:          "testRepo",
				FluxSystemNamespace: "flux-system",
				Branch:              "testBranch",
				ClusterConfigPath:   "clusters/management-cluster",
				Personal:            true,
			},
		},
	}
}

func TestFluxUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestFluxUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.1.0"

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestFluxUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.2.0"

	tt.newSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: tt.fluxConfig,
		},
	}
	f, m, g := newAddonClient(t)

	if err := setupTestFiles(t, g); err != nil {
		t.Errorf("setting up files: %v", err)
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "Flux",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
	}

	m.git.EXPECT().GetRepo(tt.ctx).Return(&git.Repository{Name: tt.fluxConfig.Github.Repository}, nil)
	m.git.EXPECT().Clone(tt.ctx).Return(nil)
	m.git.EXPECT().Branch(tt.fluxConfig.Github.Branch).Return(nil)
	m.git.EXPECT().Add(tt.fluxConfig.Github.ClusterConfigPath).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(tt.ctx).Return(nil)

	m.flux.EXPECT().DeleteFluxSystemSecret(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace)
	m.flux.EXPECT().BootstrapToolkitsComponents(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig)
	m.flux.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig)

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestFluxUpgradeError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.Flux.Version = "v0.2.0"

	tt.newSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: tt.fluxConfig,
		},
	}
	f, m, g := newAddonClient(t)

	if err := setupTestFiles(t, g); err != nil {
		t.Errorf("setting up files: %v", err)
	}

	m.git.EXPECT().GetRepo(tt.ctx).Return(&git.Repository{Name: tt.fluxConfig.Github.Repository}, nil)
	m.git.EXPECT().Clone(tt.ctx).Return(nil)
	m.git.EXPECT().Branch(tt.fluxConfig.Github.Branch).Return(nil)
	m.git.EXPECT().Add(tt.fluxConfig.Github.ClusterConfigPath).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(tt.ctx).Return(nil)

	m.flux.EXPECT().DeleteFluxSystemSecret(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace)
	m.flux.EXPECT().BootstrapToolkitsComponents(tt.ctx, tt.cluster, tt.newSpec.GitOpsConfig).Return(errors.New("error from client"))

	_, err := f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestFluxUpgradeNoGitOpsConfig(t *testing.T) {
	tt := newUpgraderTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.GitOpsConfig = nil

	tt.Expect(f.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func setupTestFiles(t *testing.T, g *addonclients.GitOptions) error {
	w, err := g.Writer.WithDir("clusters/management-cluster/management-cluster/eksa-system")
	if err != nil {
		return fmt.Errorf("failed to create test eksa-system directory: %v", err)
	}
	eksaContent, err := ioutil.ReadFile("./testdata/cluster-config-default-path-management.yaml")
	if err != nil {
		return fmt.Errorf("File [%s] reading error in test: %v", "cluster-config-default-path-management.yaml", err)
	}
	_, err = w.Write(defaultEksaClusterConfigFileName, eksaContent, filewriter.PersistentFile)
	if err != nil {
		return fmt.Errorf("failed to write eksa-cluster.yaml in test: %v", err)
	}
	return nil
}
