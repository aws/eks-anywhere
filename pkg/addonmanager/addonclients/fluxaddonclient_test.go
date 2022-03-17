package addonclients_test

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
	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	addonClientMocks "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients/mocks"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/git"
	gitMocks "github.com/aws/eks-anywhere/pkg/git/mocks"
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

func TestFluxAddonClientInstallGitOpsOnManagementClusterWithPrexistingRepo(t *testing.T) {
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
			f, m, g := newAddonClient(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			m.flux.EXPECT().BootstrapToolkitsComponents(ctx, cluster, clusterSpec.GitOpsConfig)

			m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
			m.git.EXPECT().Clone(ctx).Return(nil)
			m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
			m.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			m.git.EXPECT().Push(ctx).Return(nil)
			m.git.EXPECT().Pull(ctx, clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)

			err := f.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("FluxAddonClient.InstallGitOps() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestFluxAddonClientInstallGitOpsOnWorkloadClusterWithPrexistingRepo(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterName := "workload-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterConfig.SetManagedBy("management-cluster")
	f, m, g := newAddonClient(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	m.flux.EXPECT().BootstrapToolkitsComponents(ctx, cluster, clusterSpec.GitOpsConfig)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	m.git.EXPECT().Add(path.Dir("clusters/management-cluster")).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(ctx).Return(nil)
	m.git.EXPECT().Pull(ctx, clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)

	err := f.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("FluxAddonClient.InstallGitOps() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-workload.yaml")

	expectedKustomizationPath := path.Join(g.Writer.Dir(), "clusters/management-cluster/workload-cluster/eksa-system", defaultKustomizationManifestFileName)
	test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

	expectedFluxPatchesPath := path.Join(g.Writer.Dir(), "clusters/management-cluster/flux-system", defaultFluxPatchesFileName)
	if _, err := os.Stat(expectedFluxPatchesPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxPatchesPath)
	}

	expectedFluxSyncPath := path.Join(g.Writer.Dir(), "clusters/management-cluster/flux-system", defaultFluxSyncFileName)
	if _, err := os.Stat(expectedFluxSyncPath); errors.Is(err, os.ErrExist) {
		t.Errorf("File exists at %s, should not exist", expectedFluxSyncPath)
	}
}

func TestFluxAddonClientInstallGitOpsNoPrexistingRepo(t *testing.T) {
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
			f, m, g := newAddonClient(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			m.flux.EXPECT().BootstrapToolkitsComponents(ctx, cluster, clusterSpec.GitOpsConfig)

			n := clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository
			o := clusterSpec.GitOpsConfig.Spec.Flux.Github.Owner
			p := clusterSpec.GitOpsConfig.Spec.Flux.Github.Personal
			b := clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch
			d := "EKS-A cluster configuration repository"
			createRepoOpts := git.CreateRepoOpts{Name: n, Owner: o, Description: d, Personal: p, Privacy: true}

			returnRepo := git.Repository{
				Name:         clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository,
				Owner:        clusterSpec.GitOpsConfig.Spec.Flux.Github.Owner,
				Organization: "",
				CloneUrl:     fmt.Sprintf("https://github.com/%s/%s.git", o, n),
			}
			m.git.EXPECT().GetRepo(ctx).Return(nil, nil)
			m.git.EXPECT().CreateRepo(ctx, createRepoOpts).Return(&returnRepo, nil)
			m.git.EXPECT().Init().Return(nil)
			m.git.EXPECT().Commit(gomock.Any()).Return(nil)
			m.git.EXPECT().Branch(b).Return(nil)
			m.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			m.git.EXPECT().Push(ctx).Return(nil)
			m.git.EXPECT().Pull(ctx, b).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			err := f.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("FluxAddonClient.InstallGitOps() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestFluxAddonClientInstallGitOpsToolkitsBareRepo(t *testing.T) {
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
			f, m, g := newAddonClient(t)
			clusterSpec := newClusterSpec(t, clusterConfig, tt.fluxpath)

			m.flux.EXPECT().BootstrapToolkitsComponents(ctx, cluster, clusterSpec.GitOpsConfig)

			m.git.EXPECT().GetRepo(ctx).MaxTimes(2).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
			m.git.EXPECT().Clone(ctx).MaxTimes(2).Return(&git.RepositoryIsEmptyError{Repository: "testRepo"})
			m.git.EXPECT().Init().Return(nil)
			m.git.EXPECT().Commit(gomock.Any()).Return(nil)
			m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
			m.git.EXPECT().Add(path.Dir(tt.expectedClusterConfigGitPath)).Return(nil)
			m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
			m.git.EXPECT().Push(ctx).Return(nil)
			m.git.EXPECT().Pull(ctx, clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)

			datacenterConfig := datacenterConfig(tt.clusterName)
			machineConfig := machineConfig(tt.clusterName)
			err := f.InstallGitOps(ctx, cluster, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
			if err != nil {
				t.Errorf("FluxAddonClient.InstallGitOpsToolkits() error = %v, want nil", err)
			}
			expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedEksaConfigFileName)
			test.AssertFilesEquals(t, expectedEksaClusterConfigPath, tt.expectedConfigFileContents)

			expectedKustomizationPath := path.Join(g.Writer.Dir(), tt.expectedEksaSystemDirPath, tt.expectedKustomizationFileName)
			test.AssertFilesEquals(t, expectedKustomizationPath, "./testdata/kustomization.yaml")

			expectedFluxPatchesPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxPatchesFileName)
			test.AssertFilesEquals(t, expectedFluxPatchesPath, "./testdata/gotk-patches.yaml")

			expectedFluxSyncPath := path.Join(g.Writer.Dir(), tt.expectedFluxSystemDirPath, tt.expectedFluxSyncFileName)
			test.AssertFilesEquals(t, expectedFluxSyncPath, "./testdata/gotk-sync.yaml")
		})
	}
}

func TestFluxAddonClientPauseKustomization(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.flux.EXPECT().PauseKustomization(ctx, cluster, clusterSpec.GitOpsConfig)

	err := f.PauseGitOpsKustomization(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.PauseGitOpsKustomization() error = %v, want nil", err)
	}
}

func TestFluxAddonClientResumeKustomization(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("management-cluster")

	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.flux.EXPECT().ResumeKustomization(ctx, cluster, clusterSpec.GitOpsConfig)

	err := f.ResumeGitOpsKustomization(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.ResumeGitOpsKustomization() error = %v, want nil", err)
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecLocalRepoNotExists(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	f, m, g := newAddonClient(t)
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	m.git.EXPECT().Add(eksaSystemDirPath).Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(ctx).Return(nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestFluxAddonClientUpdateGitRepoEksaSpecLocalRepoExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	eksaSystemDirPath := "clusters/management-cluster/management-cluster/eksa-system"
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	flux := addonClientMocks.NewMockFlux(mockCtrl)

	gitProvider := gitMocks.NewMockProvider(mockCtrl)
	gitProvider.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	gitProvider.EXPECT().Add(eksaSystemDirPath).Return(nil)
	gitProvider.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitProvider.EXPECT().Push(ctx).Return(nil)

	writePath, w := test.NewWriter(t)
	if _, err := w.WithDir(".git"); err != nil {
		t.Errorf("failed to add .git dir: %v", err)
	}
	fGitOptions := &addonclients.GitOptions{Git: gitProvider, Writer: w}
	f := addonclients.NewFluxAddonClient(flux, fGitOptions)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = %v, want nil", err)
	}
	expectedEksaClusterConfigPath := path.Join(writePath, eksaSystemDirPath, defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorGetRepo(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).MaxTimes(2).Return(nil, errors.New("fail to get repo"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to describe repo")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorCloneRepo(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).MaxTimes(2).Return(errors.New("failed to clone repo"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to clone repo")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorSwitchBranch(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(errors.New("failed to switch branch"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to switch branch")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorAddFile(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	m.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(errors.New("failed to add file"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to add file")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorCommit(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	m.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(errors.New("failed to commit"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to commit code")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecErrorPushAfterRetry(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	m.git.EXPECT().Add("clusters/management-cluster/management-cluster/eksa-system").Return(nil)
	m.git.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.git.EXPECT().Push(ctx).MaxTimes(2).Return(errors.New("failed to push code"))

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err == nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = nil, want failed to push code")
	}
}

func TestFluxAddonClientUpdateGitRepoEksaSpecSkip(t *testing.T) {
	ctx := context.Background()
	clusterName := "management-cluster"
	clusterConfig := v1alpha1.NewCluster(clusterName)
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f := addonclients.NewFluxAddonClient(nil, nil)

	datacenterConfig := datacenterConfig(clusterName)
	machineConfig := machineConfig(clusterName)
	err := f.UpdateGitEksaSpec(ctx, clusterSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	if err != nil {
		t.Errorf("FluxAddonClient.UpdateGitEksaSpec() error = %v, want nil", err)
	}
}

func TestFluxAddonClientForceReconcileGitRepo(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterConfig := v1alpha1.NewCluster("")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.flux.EXPECT().ForceReconcileGitRepo(ctx, cluster, "flux-system")

	err := f.ForceReconcileGitRepo(ctx, cluster, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.ForceReconcileGitRepo() error = %v, want nil", err)
	}
}

func TestFluxAddonClientCleanupGitRepo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	expectedClusterPath := "clusters/management-cluster"
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProvider(mockCtrl)
	gitProvider.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	gitProvider.EXPECT().Clone(ctx).Return(nil)
	gitProvider.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	gitProvider.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitProvider.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitProvider.EXPECT().Push(ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &addonclients.GitOptions{Git: gitProvider, Writer: w}
	f := addonclients.NewFluxAddonClient(nil, fGitOptions)

	err := f.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestFluxAddonClientCleanupGitRepoWorkloadCluster(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("workload-cluster")
	clusterConfig.SetManagedBy("management-cluster")
	expectedClusterPath := "clusters/management-cluster/workload-cluster/" + constants.EksaSystemNamespace
	clusterSpec := newClusterSpec(t, clusterConfig, "")

	gitProvider := gitMocks.NewMockProvider(mockCtrl)
	gitProvider.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	gitProvider.EXPECT().Clone(ctx).Return(nil)
	gitProvider.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)
	gitProvider.EXPECT().Remove(expectedClusterPath).Return(nil)
	gitProvider.EXPECT().Commit(test.OfType("string")).Return(nil)
	gitProvider.EXPECT().Push(ctx).Return(nil)

	_, w := test.NewWriter(t)
	if _, err := w.WithDir(expectedClusterPath); err != nil {
		t.Errorf("failed to add %s dir: %v", expectedClusterPath, err)
	}
	fGitOptions := &addonclients.GitOptions{Git: gitProvider, Writer: w}
	f := addonclients.NewFluxAddonClient(nil, fGitOptions)

	err := f.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestFluxAddonClientCleanupGitRepoSkip(t *testing.T) {
	ctx := context.Background()
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := newClusterSpec(t, clusterConfig, "")
	f, m, _ := newAddonClient(t)

	m.git.EXPECT().GetRepo(ctx).Return(&git.Repository{Name: clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository}, nil)
	m.git.EXPECT().Clone(ctx).Return(nil)
	m.git.EXPECT().Branch(clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch).Return(nil)

	err := f.CleanupGitRepo(ctx, clusterSpec)
	if err != nil {
		t.Errorf("FluxAddonClient.CleanupGitRepo() error = %v, want nil", err)
	}
}

func TestFluxAddonClientValidationsSkipFLux(t *testing.T) {
	tt := newTest(t)
	tt.gitOptions = nil
	tt.f = addonclients.NewFluxAddonClient(tt.flux, tt.gitOptions)

	tt.Expect(tt.f.Validations(tt.ctx, tt.clusterSpec)).To(BeEmpty())
}

func TestFluxAddonClientValidationsErrorFromPathExists(t *testing.T) {
	tt := newTest(t)
	owner, repo, path := tt.setupFlux()
	tt.provider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(false, errors.New("error from git"))

	tt.Expect(runValidations(tt.f.Validations(tt.ctx, tt.clusterSpec))).NotTo(Succeed())
}

func TestFluxAddonClientValidationsPath(t *testing.T) {
	tt := newTest(t)
	owner, repo, path := tt.setupFlux()
	tt.provider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(true, nil)

	tt.Expect(runValidations(tt.f.Validations(tt.ctx, tt.clusterSpec))).NotTo(Succeed())
}

func TestFluxAddonClientValidationsSuccess(t *testing.T) {
	tt := newTest(t)
	owner, repo, path := tt.setupFlux()
	tt.provider.EXPECT().PathExists(tt.ctx, owner, repo, "main", path).Return(false, nil)

	tt.Expect(runValidations(tt.f.Validations(tt.ctx, tt.clusterSpec))).To(Succeed())
}

func runValidations(validations []validations.Validation) error {
	for _, v := range validations {
		if err := v().Err; err != nil {
			return err
		}
	}
	return nil
}

type fluxTest struct {
	*WithT
	*testing.T
	f           *addonclients.FluxAddonClient
	flux        *addonClientMocks.MockFlux
	provider    *gitMocks.MockProvider
	clusterSpec *c.Spec
	gitOptions  *addonclients.GitOptions
	ctx         context.Context
}

func newTest(t *testing.T) *fluxTest {
	ctrl := gomock.NewController(t)
	gitProvider := gitMocks.NewMockProvider(ctrl)
	flux := addonClientMocks.NewMockFlux(ctrl)
	_, w := test.NewWriter(t)
	gitOptions := &addonclients.GitOptions{Git: gitProvider, Writer: w}
	f := addonclients.NewFluxAddonClient(flux, gitOptions)
	clusterConfig := v1alpha1.NewCluster("management-cluster")
	clusterSpec := test.NewClusterSpec(func(s *c.Spec) {
		s.Cluster = clusterConfig
	})
	return &fluxTest{
		T:           t,
		f:           f,
		flux:        flux,
		provider:    gitProvider,
		clusterSpec: clusterSpec,
		gitOptions:  gitOptions,
		ctx:         context.Background(),
		WithT:       NewGomegaWithT(t),
	}
}

func (tt *fluxTest) setupFlux() (owner, repo, path string) {
	tt.Helper()
	path = "fluxFolder"
	owner = "aws"
	repo = "eksa-gitops"
	tt.clusterSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: v1alpha1.Flux{
				Github: v1alpha1.Github{
					ClusterConfigPath: path,
					Owner:             owner,
					Repository:        repo,
				},
			},
		},
	}
	if err := c.SetConfigDefaults(tt.clusterSpec.Config); err != nil {
		tt.Fatal(err)
	}

	return owner, repo, path
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

func newClusterSpec(t *testing.T, clusterConfig *v1alpha1.Cluster, fluxPath string) *c.Spec {
	t.Helper()
	fluxConfig := v1alpha1.Flux{
		Github: v1alpha1.Github{
			Owner:               "mFowler",
			Repository:          "testRepo",
			FluxSystemNamespace: "flux-system",
			Branch:              "testBranch",
			ClusterConfigPath:   fluxPath,
			Personal:            true,
		},
	}

	gitOpsConfig := v1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.GitOpsConfigKind,
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitops",
			Namespace: "default",
		},
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: fluxConfig,
		},
	}

	clusterConfig.Spec.GitOpsRef = &v1alpha1.Ref{Kind: v1alpha1.GitOpsConfigKind, Name: "test-gitops"}

	clusterSpec := test.NewClusterSpec(func(s *c.Spec) {
		s.Cluster = clusterConfig
		s.VersionsBundle.Flux = fluxBundle()
		s.GitOpsConfig = &gitOpsConfig
	})
	if err := c.SetConfigDefaults(clusterSpec.Config); err != nil {
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

type mocks struct {
	flux *addonClientMocks.MockFlux
	git  *gitMocks.MockProvider
}

func newAddonClient(t *testing.T) (*addonclients.FluxAddonClient, *mocks, *addonclients.GitOptions) {
	mockCtrl := gomock.NewController(t)
	m := &mocks{
		flux: addonClientMocks.NewMockFlux(mockCtrl),
		git:  gitMocks.NewMockProvider(mockCtrl),
	}
	_, w := test.NewWriter(t)
	gitOpts := &addonclients.GitOptions{Git: m.git, Writer: w}
	f := addonclients.NewFluxAddonClient(m.flux, gitOpts)
	retrier := retrier.NewWithMaxRetries(2, 1)
	f.SetRetier(retrier)
	return f, m, gitOpts
}
