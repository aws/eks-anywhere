package executables_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	githubToken        = "GITHUB_TOKEN"
	eksaGithubTokenEnv = "EKSA_GITHUB_TOKEN"
	validPATValue      = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	gitProvider        = "github"
)

type testFluxContext struct {
	oldGithubToken   string
	isGithubTokenSet bool
}

func (tctx *testFluxContext) SaveContext() {
	tctx.oldGithubToken, tctx.isGithubTokenSet = os.LookupEnv(eksaGithubTokenEnv)
	os.Setenv(eksaGithubTokenEnv, validPATValue)
	os.Setenv(githubToken, os.Getenv(eksaGithubTokenEnv))
}

func (tctx *testFluxContext) RestoreContext() {
	if tctx.isGithubTokenSet {
		os.Setenv(eksaGithubTokenEnv, tctx.oldGithubToken)
	} else {
		os.Unsetenv(eksaGithubTokenEnv)
	}
}

func TestFluxInstallGitOpsToolkitsSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var tctx testFluxContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	owner := "janedoe"
	repo := "gitops-fleet"
	path := "clusters/cluster-name"

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
	}{
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Github: &v1alpha1.GithubProviderConfig{
						Owner:      owner,
						Repository: repo,
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--kubeconfig", "f.kubeconfig",
			},
		},
		{
			testName: "with personal",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Github: &v1alpha1.GithubProviderConfig{
						Owner:      owner,
						Repository: repo,
						Personal:   true,
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--personal",
			},
		},
		{
			testName: "with branch",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Branch:            "main",
					Github: &v1alpha1.GithubProviderConfig{
						Owner:      owner,
						Repository: repo,
					},
				},
			},

			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--branch", "main",
			},
		},
		{
			testName: "with namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					SystemNamespace:   "flux-system",
					Github: &v1alpha1.GithubProviderConfig{
						Owner:      owner,
						Repository: repo,
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--namespace", "flux-system",
			},
		},
		{
			testName: "minimum args",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--repository", "", "--owner", "", "--path", "", "--ssh-key-algorithm", "ecdsa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			env := map[string]string{githubToken: validPATValue}
			executable.EXPECT().ExecuteWithEnv(
				ctx,
				env,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.BootstrapToolkitsComponentsGithub(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.BootstrapToolkitsComponents() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxUninstallGitOpsToolkitsComponents(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var tctx testFluxContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
	}{
		{
			testName:   "minimum args",
			cluster:    &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{},
			wantExecArgs: []interface{}{
				"uninstall", "--silent",
			},
		},
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{},
			wantExecArgs: []interface{}{
				"uninstall", "--silent", "--kubeconfig", "f.kubeconfig",
			},
		},
		{
			testName: "with namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "flux-system",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"uninstall", "--silent", "--namespace", "flux-system",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(
				ctx,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.UninstallToolkitsComponents(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.UninstallToolkitsComponents() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxPauseKustomization(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var tctx testFluxContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
	}{
		{
			testName: "minimum args",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "flux-system",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"suspend", "ks", "flux-system", "--namespace", "flux-system",
			},
		},
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "flux-system",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"suspend", "ks", "flux-system", "--namespace", "flux-system", "--kubeconfig", "f.kubeconfig",
			},
		},
		{
			testName: "with namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "custom-ns",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"suspend", "ks", "custom-ns", "--namespace", "custom-ns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(
				ctx,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.PauseKustomization(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.PauseKustomization() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxResumeKustomization(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var tctx testFluxContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
	}{
		{
			testName: "minimum args",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "flux-system",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"resume", "ks", "flux-system", "--namespace", "flux-system",
			},
		},
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "flux-system",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"resume", "ks", "flux-system", "--namespace", "flux-system", "--kubeconfig", "f.kubeconfig",
			},
		},
		{
			testName: "with namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "custom-ns",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"resume", "ks", "custom-ns", "--namespace", "custom-ns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)

			executable.EXPECT().Execute(
				ctx,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.ResumeKustomization(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.ResumeKustomization() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxReconcile(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var tctx testFluxContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
	}{
		{
			testName:   "minimum args",
			cluster:    &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{},
			wantExecArgs: []interface{}{
				"reconcile", "source", "git", "flux-system",
			},
		},
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{},
			wantExecArgs: []interface{}{
				"reconcile", "source", "git", "flux-system", "--kubeconfig", "f.kubeconfig",
			},
		},
		{
			testName: "with custom namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "custom-ns",
					Github:          &v1alpha1.GithubProviderConfig{},
				},
			},
			wantExecArgs: []interface{}{
				"reconcile", "source", "git", "custom-ns", "--namespace", "custom-ns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)

			executable.EXPECT().Execute(
				ctx,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.Reconcile(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.Reconcile() error = %v, want nil", err)
			}
		})
	}
}
