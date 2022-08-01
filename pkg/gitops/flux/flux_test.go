package flux_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/git"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	gitMocks "github.com/aws/eks-anywhere/pkg/git/mocks"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	fluxMocks "github.com/aws/eks-anywhere/pkg/gitops/flux/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
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
			ctx := context.Background()
			cluster := &types.Cluster{}
			clusterConfig := v1alpha1.NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			g.flux.EXPECT().BootstrapGithub(ctx, cluster, clusterSpec.FluxConfig)

			g.gitProvider.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.gitClient.EXPECT().Clone(ctx).Return(nil)
			g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.gitClient.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.gitClient.EXPECT().Push(ctx).Return(nil)
			g.gitClient.EXPECT().Pull(ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)

			err := g.gitOpsFlux.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("gitOpsFlux.InstallGitOps() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
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
			ctx := context.Background()
			cluster := &types.Cluster{}
			clusterConfig := v1alpha1.NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			g.flux.EXPECT().BootstrapGithub(ctx, cluster, clusterSpec.FluxConfig)

			g.gitProvider.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.gitClient.EXPECT().Clone(ctx).Return(nil)
			g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.gitClient.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.gitClient.EXPECT().Push(ctx).Return(nil)
			g.gitClient.EXPECT().Pull(ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			err := g.gitOpsFlux.InstallGitOps(ctx, cluster, clusterSpec, nil, nil)
			if err != nil {
				t.Errorf("gitOpsFlux.InstallGitOps() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			g.Expect(validations.FileExists(expectedEksaClusterConfigPath)).To(Equal(false))

			expectedKustomizationPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			g.Expect(validations.FileExists(expectedKustomizationPath)).To(Equal(false))

			expectedFluxPatchesPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestInstallGitOpsOnWorkloadClusterWithPrexistingRepo(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterName := "workload-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterConfig.SetManagedBy("management-cluster")
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g.flux.EXPECT().BootstrapGithub(ctx, cluster, clusterSpec.FluxConfig)

	g.gitProvider.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.gitClient.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(nil)
	g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.gitClient.EXPECT().Push(ctx).Return(nil)
	g.gitClient.EXPECT().Pull(ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	err := g.gitOpsFlux.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("gitOpsFlux.InstallGitOps() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-workload.yaml")

	expectedKustomizationPath := path.Join(g.gitTools.Writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultKustomizationManifestFileName)
	test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

	expectedFluxPatchesPath := path.Join(g.gitTools.Writer.Dir(), "clusters/management-cluster/flux-system", defaultFluxPatchesFileName)
	if _, err := os.Stat(expectedFluxPatchesPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxPatchesPath)
	}

	expectedFluxSyncPath := path.Join(g.gitTools.Writer.Dir(), "clusters/management-cluster/flux-system", defaultFluxSyncFileName)
	if _, err := os.Stat(expectedFluxSyncPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxSyncPath)
	}
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
			ctx := context.Background()
			cluster := &types.Cluster{}
			clusterConfig := v1alpha1.NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			g.flux.EXPECT().BootstrapGithub(ctx, cluster, clusterSpec.FluxConfig)

			n := clusterSpec.FluxConfig.Spec.Github.Repository
			o := clusterSpec.FluxConfig.Spec.Github.Owner
			p := clusterSpec.FluxConfig.Spec.Github.Personal
			b := clusterSpec.FluxConfig.Spec.Branch
			d := "EKS-A cluster configuration repository"
			createRepoOpts := git.CreateRepoOpts{Name: n, Owner: o, Description: d, Personal: p, Privacy: true}

			returnRepo := git.Repository{
				Name:         clusterSpec.FluxConfig.Spec.Github.Repository,
				Owner:        clusterSpec.FluxConfig.Spec.Github.Owner,
				Organization: "",
				CloneUrl:     fmt.Sprintf("https://github.com/%s/%s.git", o, n),
			}

			g.gitProvider.EXPECT().GetRepo(ctx).Return(nil, nil)
			g.gitProvider.EXPECT().CreateRepo(ctx, createRepoOpts).Return(&returnRepo, nil)

			g.gitClient.EXPECT().Init().Return(nil)
			g.gitClient.EXPECT().Commit(gomock.Any()).Return(nil)
			g.gitClient.EXPECT().Branch(b).Return(nil)
			g.gitClient.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.gitClient.EXPECT().Push(ctx).Return(nil)
			g.gitClient.EXPECT().Pull(ctx, b).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			err := g.gitOpsFlux.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("gitOpsFlux.InstallGitOps() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
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
			ctx := context.Background()
			cluster := &types.Cluster{}
			clusterConfig := v1alpha1.NewCluster(tt.clusterName)
			g := newFluxTest(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			g.flux.EXPECT().BootstrapGithub(ctx, cluster, clusterSpec.FluxConfig)

			g.gitProvider.EXPECT().GetRepo(ctx).MaxTimes(2).Return(&git.Repository{Name: clusterSpec.FluxConfig.Spec.Github.Repository}, nil)

			g.gitClient.EXPECT().Clone(ctx).MaxTimes(2).Return(&git.RepositoryIsEmptyError{Repository: "testRepo"})
			g.gitClient.EXPECT().Init().Return(nil)
			g.gitClient.EXPECT().Commit(gomock.Any()).Return(nil)
			g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
			g.gitClient.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
			g.gitClient.EXPECT().Push(ctx).Return(nil)
			g.gitClient.EXPECT().Pull(ctx, clusterSpec.FluxConfig.Spec.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			err := g.gitOpsFlux.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("gitOpsFlux.InstallGitOpsToolkits() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.gitTools.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestPauseKustomization(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.flux.EXPECT().SuspendKustomization(ctx, cluster, clusterSpec.FluxConfig)

	err := g.gitOpsFlux.PauseGitOpsKustomization(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.PauseGitOpsKustomization() error = %v, want nil", err)
	}
}

func TestResumeKustomization(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("management-cluster")

	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.flux.EXPECT().ResumeKustomization(ctx, cluster, clusterSpec.FluxConfig)

	err := g.gitOpsFlux.ResumeGitOpsKustomization(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.ResumeGitOpsKustomization() error = %v, want nil", err)
	}
}

func TestUpdateGitRepoEksaSpecLocalRepoNotExists(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	g := newFluxTest(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.gitClient.EXPECT().Add(eksaSystemDirPath).Return(nil)
	g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.gitClient.EXPECT().Push(ctx).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(g.gitTools.Writer.Dir(), eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestUpdateGitRepoEksaSpecLocalRepoExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	mocks := fluxMocks.NewMockClient(mockCtrl)

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Add(eksaSystemDirPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(ctx).Return(nil)

	writePath, w := test.NewWriter(t)
	if _, err := w.WithDir(".git"); err != nil {
		t.Errorf("failed to add .git dir: %v", err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(mocks, fGitOptions, nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(writePath, eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestUpdateGitRepoEksaSpecErrorCloneRepo(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).MaxTimes(2).Return(errors.New("failed to cloneIfExists repo"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = nil, want failed to cloneIfExists repo")
	}
}

func TestUpdateGitRepoEksaSpecErrorSwitchBranch(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(errors.New("failed to switch branch"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = nil, want failed to switch branch")
	}
}

func TestUpdateGitRepoEksaSpecErrorAddFile(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.gitClient.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(errors.New("failed to add file"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = nil, want failed to add file")
	}
}

func TestUpdateGitRepoEksaSpecErrorCommit(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.gitClient.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	g.gitClient.EXPECT().Commit(test.OfType("string")).Return(errors.New("failed to commit"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = nil, want failed to commit code")
	}
}

func TestUpdateGitRepoEksaSpecErrorPushAfterRetry(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	g.gitClient.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	g.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	g.gitClient.EXPECT().Push(ctx).MaxTimes(2).Return(errors.New("failed to push code"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := g.gitOpsFlux.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = nil, want failed to push code")
	}
}

func TestUpdateGitRepoEksaSpecSkip(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f := flux.NewFlux(nil, nil, nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("gitOpsFlux.UpdateGitEksaSpec() error = %v, want nil", err)
	}
}

func TestForceReconcileGitRepo(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.flux.EXPECT().ForceReconcileGitRepo(ctx, cluster, "flux-system")

	err := g.gitOpsFlux.ForceReconcileGitRepo(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.ForceReconcileGitRepo() error = %v, want nil", err)
	}
}

func TestCleanupGitRepo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	expectedClusterPath := "clusters/management-cluster"
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Clone(ctx).Return(nil)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(nil, fGitOptions, nil)

	err := f.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestCleanupGitRepoWorkloadCluster(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("workload-cluster")
	clusterConfig.SetManagedBy("management-cluster")
	expectedClusterPath := "clusters/management-cluster/workload-cluster/" + constants.EksaSystemNamespace
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)

	gitClient := gitMocks.NewMockClient(mockCtrl)
	gitClient.EXPECT().Clone(ctx).Return(nil)
	gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)
	gitClient.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitClient.EXPECT().Push(ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(nil, fGitOptions, nil)

	err := f.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestCleanupGitRepoSkip(t *testing.T) {
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	g := newFluxTest(t)

	g.gitClient.EXPECT().Clone(ctx).Return(nil)
	g.gitClient.EXPECT().Branch(clusterSpec.FluxConfig.Spec.Branch).Return(nil)

	err := g.gitOpsFlux.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("gitOpsFlux.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestValidationsSkipFLux(t *testing.T) {
	tt := newFluxTest(t)
	tt.gitTools = nil
	tt.gitOpsFlux = flux.NewFlux(tt.flux, tt.gitTools, nil)

	tt.Expect(tt.gitOpsFlux.Validations(tt.ctx, tt.clusterSpec)).To(BeEmpty())
}

func TestValidationsErrorFromPathExists(t *testing.T) {
	tt := newFluxTest(t)
	owner, repo, path := tt.setupFlux()
	tt.gitProvider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(false, errors.New("error from git"))

	tt.Expect(runValidations(tt.gitOpsFlux.Validations(tt.ctx, tt.clusterSpec))).NotTo(Succeed())
}

func TestValidationsPath(t *testing.T) {
	tt := newFluxTest(t)
	owner, repo, path := tt.setupFlux()
	tt.gitProvider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(true, nil)

	tt.Expect(runValidations(tt.gitOpsFlux.Validations(tt.ctx, tt.clusterSpec))).NotTo(Succeed())
}

func TestValidationsSuccess(t *testing.T) {
	tt := newFluxTest(t)
	owner, repo, path := tt.setupFlux()
	tt.gitProvider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(false, nil)

	tt.Expect(runValidations(tt.gitOpsFlux.Validations(tt.ctx, tt.clusterSpec))).To(Succeed())
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
		s.VersionsBundle.Flux = fluxBundle()
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

type fluxTest struct {
	*WithT
	*testing.T
	ctx         context.Context
	flux        *fluxMocks.MockClient
	gitProvider *gitMocks.MockProviderClient
	gitClient   *gitMocks.MockClient
	gitTools    *gitFactory.GitTools
	gitOpsFlux  *flux.Flux
	clusterSpec *cluster.Spec
}

func newFluxTest(t *testing.T) fluxTest {
	mockCtrl := gomock.NewController(t)
	mocks := fluxMocks.NewMockClient(mockCtrl)
	gitProvider := gitMocks.NewMockProviderClient(mockCtrl)
	gitClient := gitMocks.NewMockClient(mockCtrl)
	_, w := test.NewWriter(t)
	gitTools := &gitFactory.GitTools{
		Provider: gitProvider,
		Client:   gitClient,
		Writer:   w,
	}
	f := flux.NewFlux(mocks, gitTools, nil)
	retrier := retrier.NewWithMaxRetries(2, 1)
	f.SetRetier(retrier)

	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = clusterConfig
	})

	return fluxTest{
		T:           t,
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		gitOpsFlux:  f,
		flux:        mocks,
		gitProvider: gitProvider,
		gitClient:   gitClient,
		gitTools:    gitTools,
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
