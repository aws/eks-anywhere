package flux_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	"github.com/aws/eks-anywhere/pkg/types"
)

type upgraderTest struct {
	*WithT
	ctx                         context.Context
	currentManagementComponents *cluster.ManagementComponents
	newManagementComponents     *cluster.ManagementComponents
	currentSpec                 *cluster.Spec
	newSpec                     *cluster.Spec
	cluster                     *types.Cluster
	fluxConfig                  v1alpha1.FluxConfig
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles.Spec.Number = 1
		s.VersionsBundles["1.19"].Flux.Version = "v0.1.0"
		s.Bundles.Spec.VersionsBundles[0].Flux.Version = "v0.1.0"

		s.Cluster.ObjectMeta.Name = "management-cluster"
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Name: "testGitOpsRef",
		}
		s.Cluster.Spec.KubernetesVersion = "1.19"
	})

	currentManagementComponents := cluster.ManagementComponentsFromBundles(currentSpec.Bundles)

	newSpec := currentSpec.DeepCopy()
	newManagementComponents := cluster.ManagementComponentsFromBundles(newSpec.Bundles)

	return &upgraderTest{
		WithT:                       NewWithT(t),
		ctx:                         context.Background(),
		currentManagementComponents: currentManagementComponents,
		newManagementComponents:     newManagementComponents,
		currentSpec:                 currentSpec,
		newSpec:                     newSpec,
		cluster: &types.Cluster{
			Name:           "management-cluster",
			KubeconfigFile: "k.kubeconfig",
		},
		fluxConfig: v1alpha1.FluxConfig{
			Spec: v1alpha1.FluxConfigSpec{
				SystemNamespace:   "flux-system",
				ClusterConfigPath: "clusters/management-cluster",
				Branch:            "testBranch",
				Github: &v1alpha1.GithubProviderConfig{
					Owner:      "mFowler",
					Repository: "testRepo",
					Personal:   true,
				},
				Git: &v1alpha1.GitProviderConfig{
					RepositoryUrl: "",
				},
			},
		},
	}
}

func TestFluxUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	g := newFluxTest(t)
	tt.newSpec.Cluster.Name = "workload-cluster"
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestFluxUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	g := newFluxTest(t)
	tt.newManagementComponents.Flux.Version = "v0.1.0"

	tt.Expect(g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestFluxUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newManagementComponents.Flux.Version = "v0.2.0"

	tt.newSpec.FluxConfig = &tt.fluxConfig

	g := newFluxTest(t)

	if err := setupTestFiles(t, g.writer); err != nil {
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

	g.git.EXPECT().Clone(tt.ctx).Return(nil)
	g.git.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(tt.fluxConfig.Spec.ClusterConfigPath).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(tt.ctx).Return(nil)

	g.flux.EXPECT().DeleteSystemSecret(tt.ctx, tt.cluster, tt.newSpec.FluxConfig.Spec.SystemNamespace)
	g.flux.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.newSpec.FluxConfig)
	g.flux.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.newSpec.FluxConfig, nil)
	g.flux.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.newSpec.FluxConfig)

	tt.Expect(g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestFluxUpgradeBootstrapGithubError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newManagementComponents.Flux.Version = "v0.2.0"

	tt.newSpec.FluxConfig = &tt.fluxConfig
	g := newFluxTest(t)

	if err := setupTestFiles(t, g.writer); err != nil {
		t.Errorf("setting up files: %v", err)
	}

	g.git.EXPECT().Clone(tt.ctx).Return(nil)
	g.git.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(tt.fluxConfig.Spec.ClusterConfigPath).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(tt.ctx).Return(nil)

	g.flux.EXPECT().DeleteSystemSecret(tt.ctx, tt.cluster, tt.newSpec.FluxConfig.Spec.SystemNamespace)
	g.flux.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.newSpec.FluxConfig).Return(errors.New("error from client"))

	_, err := g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestFluxUpgradeBootstrapGitError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newManagementComponents.Flux.Version = "v0.2.0"

	tt.newSpec.FluxConfig = &tt.fluxConfig
	g := newFluxTest(t)

	if err := setupTestFiles(t, g.writer); err != nil {
		t.Errorf("setting up files: %v", err)
	}

	g.git.EXPECT().Clone(tt.ctx).Return(nil)
	g.git.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(tt.fluxConfig.Spec.ClusterConfigPath).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(tt.ctx).Return(nil)

	g.flux.EXPECT().DeleteSystemSecret(tt.ctx, tt.cluster, tt.newSpec.FluxConfig.Spec.SystemNamespace)
	g.flux.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.newSpec.FluxConfig)
	g.flux.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.newSpec.FluxConfig, nil).Return(errors.New("error in bootstrap git"))

	_, err := g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("error in bootstrap git")))
}

func TestFluxUpgradeAddError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newManagementComponents.Flux.Version = "v0.2.0"

	tt.newSpec.FluxConfig = &tt.fluxConfig

	g := newFluxTest(t)

	if err := setupTestFiles(t, g.writer); err != nil {
		t.Errorf("setting up files: %v", err)
	}

	g.git.EXPECT().Clone(tt.ctx).Return(nil)
	g.git.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(tt.fluxConfig.Spec.ClusterConfigPath).Return(errors.New("error in add"))

	_, err := g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)
	tt.Expect(err).To(MatchError(ContainSubstring("error in add")))
}

func TestFluxUpgradeNoGitOpsConfig(t *testing.T) {
	tt := newUpgraderTest(t)
	g := newFluxTest(t)
	tt.newSpec.FluxConfig = nil

	tt.Expect(g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestFluxUpgradeNewGitOpsConfig(t *testing.T) {
	tt := newUpgraderTest(t)
	g := newFluxTest(t)
	tt.currentSpec.Cluster.Spec.GitOpsRef = nil
	tt.Expect(g.gitOpsFlux.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func setupTestFiles(t *testing.T, writer filewriter.FileWriter) error {
	w, err := writer.WithDir("clusters/management-cluster/management-cluster/eksa-system")
	if err != nil {
		return fmt.Errorf("failed to create test eksa-system directory: %v", err)
	}
	eksaContent, err := os.ReadFile("./testdata/cluster-config-default-path-management.yaml")
	if err != nil {
		return fmt.Errorf("File [%s] reading error in test: %v", "cluster-config-default-path-management.yaml", err)
	}
	_, err = w.Write(defaultEksaClusterConfigFileName, eksaContent, filewriter.PersistentFile)
	if err != nil {
		return fmt.Errorf("failed to write eksa-cluster.yaml in test: %v", err)
	}
	return nil
}

func TestInstallSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	c := flux.NewFlux(nil, nil, nil, nil)
	tt.currentSpec.Cluster.Spec.GitOpsRef = nil
	tt.Expect(c.Install(tt.ctx, tt.cluster, tt.newManagementComponents, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestInstallSkip(t *testing.T) {
	tests := []struct {
		name     string
		new, old *v1alpha1.Ref
	}{
		{
			name: "gitops ref removed",
			new:  nil,
			old:  &v1alpha1.Ref{Name: "name"},
		},
		{
			name: "gitops ref not exists",
			new:  nil,
			old:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := newUpgraderTest(t)
			g := newFluxTest(t)
			test.currentSpec.Cluster.Spec.GitOpsRef = tt.old
			test.newSpec.Cluster.Spec.GitOpsRef = tt.new
			test.Expect(g.gitOpsFlux.Install(test.ctx, test.cluster, test.newManagementComponents, test.currentSpec, test.newSpec)).To(BeNil())
		})
	}
}
