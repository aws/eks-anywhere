package executables_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	githubToken                = "GITHUB_TOKEN"
	eksaGithubTokenEnv         = "EKSA_GITHUB_TOKEN"
	validPATValue              = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	githubProvider             = "github"
	gitProvider                = "git"
	validPassword              = "testPassword"
	validPrivateKeyfilePath    = "testdata/nonemptyprivatekey"
	validGitKnownHostsFilePath = "testdata/known_hosts"
)

func setupFluxContext(t *testing.T) {
	t.Setenv(eksaGithubTokenEnv, validPATValue)
	t.Setenv(githubToken, os.Getenv(eksaGithubTokenEnv))
}

func TestFluxInstallGithubToolkitsSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	setupFluxContext(t)

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
				"bootstrap", githubProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--kubeconfig", "f.kubeconfig",
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
				"bootstrap", githubProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--personal",
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
				"bootstrap", githubProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--branch", "main",
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
				"bootstrap", githubProvider, "--repository", repo, "--owner", owner, "--path", path, "--ssh-key-algorithm", "ecdsa", "--namespace", "flux-system",
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
				"bootstrap", githubProvider, "--repository", "", "--owner", "", "--path", "", "--ssh-key-algorithm", "ecdsa",
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
			if err := f.BootstrapGithub(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.BootstrapGithub() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxUninstallGitOpsToolkitsComponents(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	setupFluxContext(t)

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
			if err := f.Uninstall(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.Uninstall() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxPauseKustomization(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	setupFluxContext(t)

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
			if err := f.SuspendKustomization(ctx, tt.cluster, tt.fluxConfig); err != nil {
				t.Errorf("flux.SuspendKustomization() error = %v, want nil", err)
			}
		})
	}
}

func TestFluxResumeKustomization(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	setupFluxContext(t)

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

	setupFluxContext(t)

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

func TestFluxInstallGitToolkitsSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	repoUrl := "ssh://git@example.com/repository.git"
	path := "clusters/cluster-name"
	privateKeyFilePath := validPrivateKeyfilePath
	password := validPassword
	envmap := map[string]string{"SSH_KNOWN_HOSTS": validGitKnownHostsFilePath}

	tests := []struct {
		testName     string
		cluster      *types.Cluster
		fluxConfig   *v1alpha1.FluxConfig
		wantExecArgs []interface{}
		cliConfig    *config.CliConfig
	}{
		{
			testName: "with kubeconfig",
			cluster: &types.Cluster{
				KubeconfigFile: "f.kubeconfig",
			},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Git: &v1alpha1.GitProviderConfig{
						RepositoryUrl: repoUrl,
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--url", repoUrl, "--path", path, "--private-key-file", privateKeyFilePath, "--silent", "--kubeconfig", "f.kubeconfig", "--ssh-key-algorithm", "ecdsa", "--password", password,
			},
			cliConfig: &config.CliConfig{
				GitSshKeyPassphrase: validPassword,
				GitPrivateKeyFile:   validPrivateKeyfilePath,
				GitKnownHostsFile:   validGitKnownHostsFilePath,
			},
		},
		{
			testName: "with branch",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Branch:            "main",
					Git: &v1alpha1.GitProviderConfig{
						RepositoryUrl: repoUrl,
					},
				},
			},

			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--url", repoUrl, "--path", path, "--private-key-file", privateKeyFilePath, "--silent", "--branch", "main",
				"--ssh-key-algorithm", "ecdsa",
			},
			cliConfig: &config.CliConfig{
				GitSshKeyPassphrase: "",
				GitPrivateKeyFile:   validPrivateKeyfilePath,
				GitKnownHostsFile:   validGitKnownHostsFilePath,
			},
		},
		{
			testName: "with namespace",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					SystemNamespace:   "flux-system",
					Git: &v1alpha1.GitProviderConfig{
						RepositoryUrl: repoUrl,
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--url", repoUrl, "--path", path, "--private-key-file", privateKeyFilePath, "--silent", "--namespace", "flux-system",
				"--ssh-key-algorithm", "ecdsa", "--password", password,
			},
			cliConfig: &config.CliConfig{
				GitSshKeyPassphrase: validPassword,
				GitPrivateKeyFile:   validPrivateKeyfilePath,
				GitKnownHostsFile:   validGitKnownHostsFilePath,
			},
		},
		{
			testName: "with ssh key algorithm",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: path,
					Git: &v1alpha1.GitProviderConfig{
						RepositoryUrl:   repoUrl,
						SshKeyAlgorithm: "rsa",
					},
				},
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--url", repoUrl, "--path", path, "--private-key-file", privateKeyFilePath, "--silent",
				"--ssh-key-algorithm", "rsa", "--password", password,
			},
			cliConfig: &config.CliConfig{
				GitSshKeyPassphrase: validPassword,
				GitPrivateKeyFile:   validPrivateKeyfilePath,
				GitKnownHostsFile:   validGitKnownHostsFilePath,
			},
		},
		{
			testName: "minimum args",
			cluster:  &types.Cluster{},
			fluxConfig: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Git: &v1alpha1.GitProviderConfig{},
				},
			},
			cliConfig: &config.CliConfig{
				GitSshKeyPassphrase: validPassword,
				GitPrivateKeyFile:   validPrivateKeyfilePath,
				GitKnownHostsFile:   validGitKnownHostsFilePath,
			},
			wantExecArgs: []interface{}{
				"bootstrap", gitProvider, "--url", "", "--path", "", "--private-key-file", privateKeyFilePath, "--silent", "--ssh-key-algorithm", "ecdsa",
				"--password", password,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().ExecuteWithEnv(
				ctx,
				envmap,
				tt.wantExecArgs...,
			).Return(bytes.Buffer{}, nil)

			f := executables.NewFlux(executable)
			if err := f.BootstrapGit(ctx, tt.cluster, tt.fluxConfig, tt.cliConfig); err != nil {
				t.Errorf("flux.BootstrapGit() error = %v, want nil", err)
			}
		})
	}
}
