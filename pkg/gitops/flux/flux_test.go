package flux_test

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	gitMocks "github.com/aws/eks-anywhere/pkg/git/mocks"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	"github.com/aws/eks-anywhere/pkg/gitops/flux/mocks"
	fluxMocks "github.com/aws/eks-anywhere/pkg/gitops/flux/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	defaultKustomizationManifestFileName = "kustomization.yaml"
	defaultEksaClusterConfigFileName     = "eksa-cluster.yaml"
	defaultFluxPatchesFileName           = "gotk-patches.yaml"
	defaultFluxSyncFileName              = "gotk-sync.yaml"
)

type fluxTest struct {
	*WithT
	*testing.T
	ctx         context.Context
	flux        *fluxMocks.MockGitOpsFluxClient
	git         *fluxMocks.MockGitClient
	provider    *mocksprovider.MockProvider
	gitOpsFlux  *flux.Flux
	writer      filewriter.FileWriter
	clusterSpec *cluster.Spec
}

func newFluxTest(t *testing.T) fluxTest {
	mockCtrl := gomock.NewController(t)
	mockGitOpsFlux := fluxMocks.NewMockGitOpsFluxClient(mockCtrl)
	mockGit := fluxMocks.NewMockGitClient(mockCtrl)
	mockProvider := mocksprovider.NewMockProvider(gomock.NewController(t))
	_, w := test.NewWriter(t)
	f := flux.NewFluxFromGitOpsFluxClient(mockGitOpsFlux, mockGit, w, nil)

	clusterConfig := NewCluster("management-cluster")
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = clusterConfig
	})

	return fluxTest{
		T:           t,
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		gitOpsFlux:  f,
		flux:        mockGitOpsFlux,
		git:         mockGit,
		provider:    mockProvider,
		writer:      w,
		clusterSpec: clusterSpec,
	}
}

func (t *fluxTest) setupFlux() (owner, repo, path string) {
	t.Helper()
	path = "fluxFolder"
	owner = "aws"
	repo = "eksa-gitops"
	t.clusterSpec.FluxConfig = &v1alpha1.FluxConfig{
		Spec: v1alpha1.FluxConfigSpec{
			ClusterConfigPath: path,
			Branch:            "main",
			Github: &v1alpha1.GithubProviderConfig{
				Owner:      owner,
				Repository: repo,
			},
		},
	}

	if err := cluster.SetConfigDefaults(t.clusterSpec.Config); err != nil {
		t.Fatal(err)
	}

	return owner, repo, path
}

func runValidations(validations []validations.Validation) error {
	for _, v := range validations {
		if err := v().Err; err != nil {
			return err
		}
	}
	return nil
}

func datacenterConfig(clusterName string) *v1alpha1.VSphereDatacenterConfig {
	return &v1alpha1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.VSphereDatacenterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "SDDC-Datacenter",
		},
	}
}

func machineConfig(clusterName string) *v1alpha1.VSphereMachineConfig {
	return &v1alpha1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.VSphereMachineConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Template: "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6",
		},
	}
}

func newClusterSpec(t *testing.T, clusterConfig *v1alpha1.Cluster, fluxPath string) *cluster.Spec {
	t.Helper()

	fluxConfig := v1alpha1.FluxConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.FluxConfigKind,
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitops",
			Namespace: "default",
		},
		Spec: v1alpha1.FluxConfigSpec{
			SystemNamespace:   "flux-system",
			ClusterConfigPath: fluxPath,
			Branch:            "testBranch",
			Github: &v1alpha1.GithubProviderConfig{
				Owner:      "mFolwer",
				Repository: "testRepo",
				Personal:   true,
			},
		},
	}

	clusterConfig.Spec.GitOpsRef = &v1alpha1.Ref{Kind: v1alpha1.FluxConfigKind, Name: "test-gitops"}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = clusterConfig
		s.VersionsBundles["1.19"].Flux = fluxBundle()
		s.Bundles.Spec.VersionsBundles[0].Flux = fluxBundle()
		s.FluxConfig = &fluxConfig
	})
	if err := cluster.SetConfigDefaults(clusterSpec.Config); err != nil {
		t.Fatal(err)
	}
	return clusterSpec
}

func fluxBundle() releasev1alpha1.FluxBundle {
	return releasev1alpha1.FluxBundle{
		SourceController: releasev1alpha1.Image{
			URI: "public.ecr.aws/l0g8r8j6/fluxcd/source-controller:v0.12.1-8539f509df046a4f567d2182dde824b957136599",
		},
		KustomizeController: releasev1alpha1.Image{
			URI: "public.ecr.aws/l0g8r8j6/fluxcd/kustomize-controller:v0.11.1-d82011942ec8a447ba89a70ff9a84bf7b9579492",
		},
		HelmController: releasev1alpha1.Image{
			URI: "public.ecr.aws/l0g8r8j6/fluxcd/helm-controller:v0.10.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
		},
		NotificationController: releasev1alpha1.Image{
			URI: "public.ecr.aws/l0g8r8j6/fluxcd/notification-controller:v0.13.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
		},
	}
}

func TestInstallGitOpsOnManagementClusterWithPrexistingRepo(t *testing.T) {
	tests := []struct {
		testName                      string
		clusterName                   string
		managedbyClusterName          string
		selfManaged                   bool
		fluxpath                      string
		expectedClusterConfigGitPath  string
		expectedEksaSystemDirPath     string
		expectedEksaConfigFileName    string
		expectedKustomizationFileName string
		expectedConfigFileContents    string
		expectedFluxSystemDirPath     string
		expectedFluxPatchesFileName   string
		expectedFluxSyncFileName      string
	}{
		{
			testName:                      "with default config path",
			clusterName:                   "management-cluster",
			selfManaged:                   true,
			fluxpath:                      "",
			expectedClusterConfigGitPath:  "clusters/management-cluster",
			expectedEksaSystemDirPath:     "clusters/management-cluster/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedConfigFileContents:    "./testdata/cluster-config-default-path-management.yaml",
			expectedFluxSystemDirPath:     "clusters/management-cluster/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
		{
			testName:                      "with user provided config path",
			clusterName:                   "management-cluster",
			selfManaged:                   true,
			fluxpath:                      "user/provided/path",
			expectedClusterConfigGitPath:  "user/provided/path",
			expectedEksaSystemDirPath:     "user/provided/path/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedConfigFileContents:    "./testdata/cluster-config-user-provided-path.yaml",
			expectedFluxSystemDirPath:     "user/provided/path/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			clusterConfig := NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)
			managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

			cluster := &types.Cluster{}
			g.flux.EXPECT().BootstrapGithub(g.ctx, cluster, clusterSpec.FluxConfig)

			g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.git.EXPECT().Clone(g.ctx).Return(nil)
			g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.git.EXPECT().Push(g.ctx).Return(nil)
			g.git.EXPECT().Pull(g.ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)

			g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())

			expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedEksaKustomizationPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedEksaKustomizationPath, "./testdata/eksa-kustomization.yaml")

			expectedFluxKustomizationPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedFluxKustomizationPath, "./testdata/flux-kustomization.yaml")

			expectedFluxSyncPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestInstallGitOpsOnManagementClusterWithoutClusterSpec(t *testing.T) {
	tests := []struct {
		testName                      string
		clusterName                   string
		managedbyClusterName          string
		fluxpath                      string
		expectedClusterConfigGitPath  string
		expectedEksaSystemDirPath     string
		expectedEksaConfigFileName    string
		expectedKustomizationFileName string
		expectedFluxSystemDirPath     string
		expectedFluxPatchesFileName   string
		expectedFluxSyncFileName      string
	}{
		{
			testName:                      "with default config path",
			clusterName:                   "management-cluster",
			fluxpath:                      "",
			expectedClusterConfigGitPath:  "clusters/management-cluster",
			expectedEksaSystemDirPath:     "clusters/management-cluster/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedFluxSystemDirPath:     "clusters/management-cluster/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			clusterConfig := NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)
			managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

			cluster := &types.Cluster{}
			g.flux.EXPECT().BootstrapGithub(g.ctx, cluster, clusterSpec.FluxConfig)

			g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.git.EXPECT().Clone(g.ctx).Return(nil)
			g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.git.EXPECT().Push(g.ctx).Return(nil)
			g.git.EXPECT().Pull(g.ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, nil, nil)).To(Succeed())

			expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			g.Expect(validations.FileExists(expectedEksaClusterConfigPath)).To(Equal(false))

			expectedEksaKustomizationPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			g.Expect(validations.FileExists(expectedEksaKustomizationPath)).To(Equal(false))

			expectedFluxKustomizationPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedFluxKustomizationPath, "./testdata/flux-kustomization.yaml")

			expectedFluxSyncPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestInstallGitOpsOnWorkloadClusterWithPrexistingRepo(t *testing.T) {
	clusterName := "workload-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterConfig.SetManagedBy("management-cluster")
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	cluster := &types.Cluster{}
	g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(g.ctx).Return(nil)
	g.git.EXPECT().Pull(g.ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())

	expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-workload.yaml")

	expectedKustomizationPath := path.Join(g.writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultKustomizationManifestFileName)
	test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/eksa-kustomization.yaml")

	expectedFluxKustomizationPath := path.Join(g.writer.Dir(), "clusters/management-cluster/flux-system", defaultKustomizationManifestFileName)
	if _, err := os.Stat(expectedFluxKustomizationPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxKustomizationPath)
	}

	expectedFluxSyncPath := path.Join(g.writer.Dir(), "clusters/management-cluster/flux-system", defaultFluxSyncFileName)
	if _, err := os.Stat(expectedFluxSyncPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxSyncPath)
	}
}

func TestInstallGitOpsSetupRepoError(t *testing.T) {
	clusterName := "test-cluster"
	clusterConfig := NewCluster(clusterName)
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	cluster := &types.Cluster{}

	g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)
	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(errors.New("error in add"))

	g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, nil, nil)).To(MatchError(ContainSubstring("error in add")))
}

func TestInstallGitOpsBootstrapError(t *testing.T) {
	clusterName := "test-cluster"
	clusterConfig := NewCluster(clusterName)
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)
	cluster := &types.Cluster{}

	g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)
	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(g.ctx).Return(nil)
	g.flux.EXPECT().BootstrapGithub(g.ctx, cluster, clusterSpec.FluxConfig).Return(errors.New("error in bootstrap"))
	g.flux.EXPECT().Uninstall(g.ctx, cluster, clusterSpec.FluxConfig).Return(nil)

	g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, nil, nil)).To(MatchError(ContainSubstring("error in bootstrap")))
}

func TestInstallGitOpsGitProviderSuccess(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	clusterSpec.FluxConfig.Spec.Git = &v1alpha1.GitProviderConfig{RepositoryUrl: "git.xyz"}
	clusterSpec.FluxConfig.Spec.Github = nil

	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	cluster := &types.Cluster{}

	g.flux.EXPECT().BootstrapGit(g.ctx, cluster, clusterSpec.FluxConfig, nil)
	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(g.ctx).Return(nil)
	g.git.EXPECT().Pull(g.ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())
}

func TestInstallGitOpsCommitFilesError(t *testing.T) {
	clusterName := "test-cluster"
	clusterConfig := NewCluster(clusterName)
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	cluster := &types.Cluster{}

	g.git.EXPECT().GetRepo(g.ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)
	g.git.EXPECT().Clone(g.ctx).Return(errors.New("error in clone"))

	g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, nil, nil)).To(MatchError(ContainSubstring("error in clone")))
}

func TestInstallGitOpsNoPrexistingRepo(t *testing.T) {
	tests := []struct {
		testName                      string
		clusterName                   string
		fluxpath                      string
		expectedClusterConfigGitPath  string
		expectedEksaSystemDirPath     string
		expectedEksaConfigFileName    string
		expectedKustomizationFileName string
		expectedConfigFileContents    string
		expectedFluxSystemDirPath     string
		expectedFluxPatchesFileName   string
		expectedFluxSyncFileName      string
		expectedRepoUrl               string
	}{
		{
			testName:                      "with default config path",
			clusterName:                   "management-cluster",
			fluxpath:                      "",
			expectedClusterConfigGitPath:  "clusters/management-cluster",
			expectedEksaSystemDirPath:     "clusters/management-cluster/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedConfigFileContents:    "./testdata/cluster-config-default-path-management.yaml",
			expectedFluxSystemDirPath:     "clusters/management-cluster/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
		{
			testName:                      "with user provided config path",
			clusterName:                   "management-cluster",
			fluxpath:                      "user/provided/path",
			expectedClusterConfigGitPath:  "user/provided/path",
			expectedEksaSystemDirPath:     "user/provided/path/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedConfigFileContents:    "./testdata/cluster-config-user-provided-path.yaml",
			expectedFluxSystemDirPath:     "user/provided/path/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			clusterConfig := NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)
			managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

			cluster := &types.Cluster{}

			g.flux.EXPECT().BootstrapGithub(g.ctx, cluster, clusterSpec.FluxConfig)

			n := clusterSpec.FluxConfig.Spec.Github.Repository
			o := clusterSpec.FluxConfig.Spec.Github.Owner
			p := clusterSpec.FluxConfig.Spec.Github.Personal
			b := clusterSpec.FluxConfig.Spec.Branch
			d := "EKS-A cluster configuration repository"
			createRepoOpts := git.CreateRepoOpts{Name: n, Owner: o, Description: d, Personal: p, Privacy: true}

			g.git.EXPECT().GetRepo(g.ctx).Return(nil, nil)
			g.git.EXPECT().CreateRepo(g.ctx, createRepoOpts).Return(nil)

			g.git.EXPECT().Init().Return(nil)
			g.git.EXPECT().Commit(gomock.Any()).Return(nil)
			g.git.EXPECT().Branch(b).Return(nil)
			g.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.git.EXPECT().Push(g.ctx).Return(nil)
			g.git.EXPECT().Pull(g.ctx, b).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())

			expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedEksaKustomizationPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedEksaKustomizationPath, "./testdata/eksa-kustomization.yaml")

			expectedFluxKustomizationPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedFluxKustomizationPath, "./testdata/flux-kustomization.yaml")

			expectedFluxSyncPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestInstallGitOpsToolkitsBareRepo(t *testing.T) {
	tests := []struct {
		testName                      string
		clusterName                   string
		fluxpath                      string
		expectedClusterConfigGitPath  string
		expectedEksaSystemDirPath     string
		expectedEksaConfigFileName    string
		expectedKustomizationFileName string
		expectedConfigFileContents    string
		expectedFluxSystemDirPath     string
		expectedFluxPatchesFileName   string
		expectedFluxSyncFileName      string
	}{
		{
			testName:                      "with default config path",
			clusterName:                   "management-cluster",
			fluxpath:                      "",
			expectedClusterConfigGitPath:  "clusters/management-cluster",
			expectedEksaSystemDirPath:     "clusters/management-cluster/management-cluster/eksa-system",
			expectedEksaConfigFileName:    defaultEksaClusterConfigFileName,
			expectedKustomizationFileName: defaultKustomizationManifestFileName,
			expectedConfigFileContents:    "./testdata/cluster-config-default-path-management.yaml",
			expectedFluxSystemDirPath:     "clusters/management-cluster/flux-system",
			expectedFluxPatchesFileName:   defaultFluxPatchesFileName,
			expectedFluxSyncFileName:      defaultFluxSyncFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			clusterConfig := NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)
			managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

			cluster := &types.Cluster{}

			g.flux.EXPECT().BootstrapGithub(g.ctx, cluster, clusterSpec.FluxConfig)

			g.git.EXPECT().GetRepo(g.ctx).MaxTimes(2).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.git.EXPECT().Clone(g.ctx).MaxTimes(2).Return(&git.RepositoryIsEmptyError{Repository: "testRepo"})
			g.git.EXPECT().Init().Return(nil)
			g.git.EXPECT().Commit(gomock.Any()).Return(nil)
			g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.git.EXPECT().Push(g.ctx).Return(nil)
			g.git.EXPECT().Pull(g.ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			g.Expect(g.gitOpsFlux.InstallGitOps(g.ctx, cluster, managementComponents, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())

			expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedEksaKustomizationPath := path.Join(g.writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedEksaKustomizationPath, "./testdata/eksa-kustomization.yaml")

			expectedFluxKustomizationPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedFluxKustomizationPath, "./testdata/flux-kustomization.yaml")

			expectedFluxSyncPath := path.Join(g.writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestResumeClusterResourcesReconcile(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "cp-machine"}
	clusterConfig.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
		{
			MachineGroupRef: &v1alpha1.Ref{Name: "worker-machine"},
		},
	}
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "cp-machine", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "worker-machine", "")
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter")
	g.provider.EXPECT().MachineResourceType().Return("providerMachineConfig").Times(3)

	g.Expect(g.gitOpsFlux.ResumeClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(Succeed())
}

func TestResumeClusterResourcesReconcileEnableClusterReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "").Return(errors.New("error in enable cluster reconcile"))

	g.Expect(g.gitOpsFlux.ResumeClusterResourcesReconcile(g.ctx, cluster, clusterSpec, nil)).To(MatchError(ContainSubstring("error in enable cluster reconcile")))
}

func TestResumeClusterResourcesReconcileEnableDatacenterReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "").Return(errors.New("error in enable datacenter reconcile"))
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter").Times(2)

	g.Expect(g.gitOpsFlux.ResumeClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(MatchError(ContainSubstring("error in enable datacenter reconcile")))
}

func TestResumeClusterResourcesReconcileEnableMachineReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "cp-machine"}
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "")
	g.flux.EXPECT().EnableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "cp-machine", "").Return(errors.New("error in enable machine reconcile"))
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter")
	g.provider.EXPECT().MachineResourceType().Return("providerMachineConfig").Times(3)

	g.Expect(g.gitOpsFlux.ResumeClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(MatchError(ContainSubstring("error in enable machine reconcile")))
}

func TestPauseClusterResourcesReconcile(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "cp-machine"}
	clusterConfig.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
		{
			MachineGroupRef: &v1alpha1.Ref{Name: "worker-machine"},
		},
	}

	g := newFluxTest(t)

	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "cp-machine", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "worker-machine", "")
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter")
	g.provider.EXPECT().MachineResourceType().Return("providerMachineConfig").Times(3)

	g.Expect(g.gitOpsFlux.PauseClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(Succeed())
}

func TestPauseClusterResourcesReconcileEnableClusterReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "").Return(errors.New("error in enable cluster reconcile"))

	g.Expect(g.gitOpsFlux.PauseClusterResourcesReconcile(g.ctx, cluster, clusterSpec, nil)).To(MatchError(ContainSubstring("error in enable cluster reconcile")))
}

func TestPauseClusterResourcesReconcileEnableDatacenterReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "").Return(errors.New("error in enable datacenter reconcile"))
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter").Times(2)

	g.Expect(g.gitOpsFlux.PauseClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(MatchError(ContainSubstring("error in enable datacenter reconcile")))
}

func TestPauseClusterResourcesReconcileEnableMachineReconcileError(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterConfig.Spec.DatacenterRef = v1alpha1.Ref{Name: "datacenter"}
	clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "cp-machine"}
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g := newFluxTest(t)

	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "clusters.anywhere.eks.amazonaws.com", "management-cluster", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerDatacenter", "datacenter", "")
	g.flux.EXPECT().DisableResourceReconcile(g.ctx, cluster, "providerMachineConfig", "cp-machine", "").Return(errors.New("error in enable machine reconcile"))
	g.provider.EXPECT().DatacenterResourceType().Return("providerDatacenter")
	g.provider.EXPECT().MachineResourceType().Return("providerMachineConfig").Times(3)

	g.Expect(g.gitOpsFlux.PauseClusterResourcesReconcile(g.ctx, cluster, clusterSpec, g.provider)).To(MatchError(ContainSubstring("error in enable machine reconcile")))
}

func TestUpdateGitRepoEksaSpecLocalRepoNotExists(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add(eksaSystemDirPath).Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(g.ctx).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())
	expectedEksaClusterConfigPath := path.Join(g.writer.Dir(), eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestUpdateGitRepoEksaSpecLocalRepoExists(t *testing.T) {
	g := newFluxTest(t)
	mockCtrl := gomock.NewController(t)
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	mocks := fluxMocks.NewMockFluxClient(mockCtrl)

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Add(eksaSystemDirPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(g.ctx).Return(nil)

	writePath, w := test.NewWriter(t)
	if _, err := w.WithDir(".git"); err != nil {
		t.Errorf("failed to add .git dir: %v", err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(mocks, nil, fGitOptions, nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	g.Expect(f.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())

	expectedEksaClusterConfigPath := path.Join(writePath, eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestUpdateGitRepoEksaSpecErrorCloneRepo(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).MaxTimes(2).Return(errors.New("error in cloneIfExists repo"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(MatchError(ContainSubstring("error in cloneIfExists repo")))
}

func TestUpdateGitRepoEksaSpecErrorSwitchBranch(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(errors.New("failed to switch branch"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(MatchError(ContainSubstring("failed to switch branch")))
}

func TestUpdateGitRepoEksaSpecErrorAddFile(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(errors.New("failed to add file"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(MatchError(ContainSubstring("failed to add file")))
}

func TestUpdateGitRepoEksaSpecErrorCommit(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(errors.New("failed to commit"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(MatchError(ContainSubstring("failed to commit")))
}

func TestUpdateGitRepoEksaSpecErrorPushAfterRetry(t *testing.T) {
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	g.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.git.EXPECT().Push(g.ctx).MaxTimes(2).Return(errors.New("failed to push code"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(g.gitOpsFlux.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(MatchError(ContainSubstring("failed to push code")))
}

func TestUpdateGitRepoEksaSpecSkip(t *testing.T) {
	g := newFluxTest(t)
	clusterName := "management-cluster"
	clusterConfig := NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f := flux.NewFlux(nil, nil, nil, nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	g.Expect(f.UpdateGitEksaSpec(g.ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})).To(Succeed())
}

func TestForceReconcileGitRepo(t *testing.T) {
	cluster := &types.Cluster{}
	clusterConfig := NewCluster("")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.flux.EXPECT().ForceReconcile(g.ctx, cluster, "flux-system")

	g.Expect(g.gitOpsFlux.ForceReconcileGitRepo(g.ctx, cluster, clusterSpec)).To(Succeed())
}

func TestForceReconcileGitRepoSkip(t *testing.T) {
	cluster := &types.Cluster{}
	g := newFluxTest(t)
	f := flux.NewFlux(nil, nil, nil, nil)

	g.Expect(f.ForceReconcileGitRepo(g.ctx, cluster, g.clusterSpec)).To(Succeed())
}

func TestCleanupGitRepo(t *testing.T) {
	g := newFluxTest(t)
	mockCtrl := gomock.NewController(t)
	clusterConfig := NewCluster("management-cluster")
	expectedClusterPath := "clusters/management-cluster"
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Clone(g.ctx).Return(nil)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(g.ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(nil, nil, fGitOptions, nil)

	g.Expect(f.CleanupGitRepo(g.ctx, clusterSpec)).To(Succeed())
}

func TestCleanupGitRepoWorkloadCluster(t *testing.T) {
	g := newFluxTest(t)
	mockCtrl := gomock.NewController(t)
	clusterConfig := NewCluster("workload-cluster")
	clusterConfig.SetManagedBy("management-cluster")
	expectedClusterPath := "clusters/management-cluster/workload-cluster/" + constants.EksaSystemNamespace
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Clone(g.ctx).Return(nil)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(g.ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(nil, nil, fGitOptions, nil)

	g.Expect(f.CleanupGitRepo(g.ctx, clusterSpec)).To(Succeed())
}

func TestCleanupGitRepoSkip(t *testing.T) {
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.git.EXPECT().Clone(g.ctx).Return(nil)
	g.git.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)

	g.Expect(g.gitOpsFlux.CleanupGitRepo(g.ctx, clusterSpec)).To(Succeed())
}

func TestCleanupGitRepoRemoveError(t *testing.T) {
	g := newFluxTest(t)
	mockCtrl := gomock.NewController(t)
	clusterConfig := NewCluster("management-cluster")
	expectedClusterPath := "clusters/management-cluster"
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Clone(g.ctx).Return(nil)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Remove(expectedClusterPath).Return(errors.New("error in remove"))

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(nil, nil, fGitOptions, nil)

	g.Expect(f.CleanupGitRepo(g.ctx, clusterSpec)).To(MatchError(ContainSubstring("error in remove")))
}

func TestValidationsSkipFLux(t *testing.T) {
	g := newFluxTest(t)
	g.gitOpsFlux = flux.NewFlux(g.flux, nil, nil, nil)

	g.Expect(g.gitOpsFlux.Validations(g.ctx, g.clusterSpec)).To(BeEmpty())
}

func TestValidationsErrorFromPathExists(t *testing.T) {
	g := newFluxTest(t)
	owner, repo, path := g.setupFlux()
	g.git.EXPECT().PathExists(g.ctx, owner, repo, "main", path).Return(false, errors.New("error from git"))

	g.Expect(runValidations(g.gitOpsFlux.Validations(g.ctx, g.clusterSpec))).NotTo(Succeed())
}

func TestValidationsPath(t *testing.T) {
	g := newFluxTest(t)
	owner, repo, path := g.setupFlux()
	g.git.EXPECT().PathExists(g.ctx, owner, repo, "main", path).Return(true, nil)

	g.Expect(runValidations(g.gitOpsFlux.Validations(g.ctx, g.clusterSpec))).NotTo(Succeed())
}

func TestValidationsSuccess(t *testing.T) {
	g := newFluxTest(t)
	owner, repo, path := g.setupFlux()
	g.git.EXPECT().PathExists(g.ctx, owner, repo, "main", path).Return(false, nil)

	g.Expect(runValidations(g.gitOpsFlux.Validations(g.ctx, g.clusterSpec))).To(Succeed())
}

func TestBootstrapGithubSkip(t *testing.T) {
	g := newFluxTest(t)
	c := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	clusterSpec.FluxConfig.Spec.Github = nil

	g.Expect(g.gitOpsFlux.Bootstrap(g.ctx, c, clusterSpec)).To(Succeed())
}

func TestBootstrapGithubError(t *testing.T) {
	g := newFluxTest(t)
	c := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g.flux.EXPECT().BootstrapGithub(g.ctx, c, clusterSpec.FluxConfig).Return(errors.New("error in bootstrap github"))
	g.flux.EXPECT().Uninstall(g.ctx, c, clusterSpec.FluxConfig).Return(nil)

	g.Expect(g.gitOpsFlux.Bootstrap(g.ctx, c, clusterSpec)).To(MatchError(ContainSubstring("error in bootstrap github")))
}

func TestBootstrapGitError(t *testing.T) {
	g := newFluxTest(t)
	c := &types.Cluster{}
	clusterConfig := NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	clusterSpec.FluxConfig.Spec.Git = &v1alpha1.GitProviderConfig{RepositoryUrl: "abc"}

	g.flux.EXPECT().BootstrapGithub(g.ctx, c, clusterSpec.FluxConfig).Return(nil)
	g.flux.EXPECT().BootstrapGit(g.ctx, c, clusterSpec.FluxConfig, nil).Return(errors.New("error in bootstrap git"))
	g.flux.EXPECT().Uninstall(g.ctx, c, clusterSpec.FluxConfig).Return(nil)

	g.Expect(g.gitOpsFlux.Bootstrap(g.ctx, c, clusterSpec)).To(MatchError(ContainSubstring("error in bootstrap git")))
}

func TestUninstallError(t *testing.T) {
	g := newFluxTest(t)
	c := &types.Cluster{}

	g.flux.EXPECT().Uninstall(g.ctx, c, g.clusterSpec.FluxConfig).Return(errors.New("error in uninstall"))

	g.Expect(g.gitOpsFlux.Uninstall(g.ctx, c, g.clusterSpec)).To(MatchError(ContainSubstring("error in uninstall")))
}

func TestFluxBootstrapGithub(t *testing.T) {
	c := &types.Cluster{}
	ctx := context.Background()
	testCases := []struct {
		name                string
		spec                *cluster.Spec
		needBootstrapGithub bool
	}{
		{
			name: "management cluster",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "management-cluster"
				s.Cluster.SetSelfManaged()
				s.FluxConfig = &anywherev1.FluxConfig{
					Spec: anywherev1.FluxConfigSpec{
						Github: &anywherev1.GithubProviderConfig{},
					},
				}
			}),
			needBootstrapGithub: true,
		},
		{
			name: "workload cluster",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "workload-cluster"
				s.Cluster.SetManagedBy("management-cluster")
			}),
			needBootstrapGithub: false,
		},
		{
			name: "management cluster not github configured",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "management-cluster"
				s.Cluster.SetSelfManaged()
				s.FluxConfig = &anywherev1.FluxConfig{}
			}),
			needBootstrapGithub: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctrl := gomock.NewController(t)
			mockFluxClient := mocks.NewMockFluxClient(ctrl)
			if tc.needBootstrapGithub {
				mockFluxClient.EXPECT().BootstrapGithub(ctx, c, tc.spec.FluxConfig).Return(nil)
			}

			f := flux.NewFlux(mockFluxClient, nil, nil, nil)
			g.Expect(f.BootstrapGithub(ctx, c, tc.spec)).NotTo(HaveOccurred())
		})
	}
}

func TestFluxBootstrapGit(t *testing.T) {
	c := &types.Cluster{}
	ctx := context.Background()
	testCases := []struct {
		name             string
		spec             *cluster.Spec
		needBootstrapGit bool
	}{
		{
			name: "management cluster",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "management-cluster"
				s.Cluster.SetSelfManaged()
				s.FluxConfig = &anywherev1.FluxConfig{
					Spec: anywherev1.FluxConfigSpec{
						Git: &anywherev1.GitProviderConfig{},
					},
				}
			}),
			needBootstrapGit: true,
		},
		{
			name: "workload cluster",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "workload-cluster"
				s.Cluster.SetManagedBy("management-cluster")
			}),
			needBootstrapGit: false,
		},
		{
			name: "management cluster not git configured",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "management-cluster"
				s.Cluster.SetSelfManaged()
				s.FluxConfig = &anywherev1.FluxConfig{}
			}),
			needBootstrapGit: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctrl := gomock.NewController(t)
			mockFluxClient := mocks.NewMockFluxClient(ctrl)
			if tc.needBootstrapGit {
				mockFluxClient.EXPECT().BootstrapGit(ctx, c, tc.spec.FluxConfig, nil).Return(nil)
			}

			f := flux.NewFlux(mockFluxClient, nil, nil, nil)
			g.Expect(f.BootstrapGit(ctx, c, tc.spec)).NotTo(HaveOccurred())
		})
	}
}
