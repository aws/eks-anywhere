package executables

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	fluxPath                   = "flux"
	eksaGithubTokenEnv         = "EKSA_GITHUB_TOKEN"
	githubTokenEnv             = "GITHUB_TOKEN"
	githubProvider             = "github"
	gitProvider                = "git"
	defaultPrivateKeyAlgorithm = "ecdsa"
)

type Flux struct {
	Executable
}

func NewFlux(executable Executable) *Flux {
	return &Flux{
		Executable: executable,
	}
}

// BootstrapGithub creates the GitHub repository if it doesn’t exist, and commits the toolkit
// components manifests to the main branch. Then it configures the target cluster to synchronize with the repository.
// If the toolkit components are present on the cluster, the bootstrap command will perform an upgrade if needed.
func (f *Flux) BootstrapGithub(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	params := []string{
		"bootstrap",
		githubProvider,
		"--repository", c.Github.Repository,
		"--owner", c.Github.Owner,
		"--path", c.ClusterConfigPath,
		"--ssh-key-algorithm", defaultPrivateKeyAlgorithm,
	}
	params = setUpCommonParamsBootstrap(cluster, fluxConfig, params)

	if c.Github.Personal {
		params = append(params, "--personal")
	}

	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return fmt.Errorf("setting token env: %v", err)
	}

	env := make(map[string]string)
	env[githubTokenEnv] = token

	_, err = f.ExecuteWithEnv(ctx, env, params...)
	if err != nil {
		return fmt.Errorf("executing flux bootstrap github: %v", err)
	}

	return err
}

// BootstrapGit commits the toolkit components manifests to the branch of a Git repository.
// It then configures the target cluster to synchronize with the repository. If the toolkit components are present on the cluster, the
// bootstrap command will perform an upgrade if needed.
func (f *Flux) BootstrapGit(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig, cliConfig *config.CliConfig) error {
	c := fluxConfig.Spec
	params := []string{
		"bootstrap",
		gitProvider,
		"--url", c.Git.RepositoryUrl,
		"--path", c.ClusterConfigPath,
		"--private-key-file", cliConfig.GitPrivateKeyFile,
		"--silent",
	}

	params = setUpCommonParamsBootstrap(cluster, fluxConfig, params)
	if fluxConfig.Spec.Git.SshKeyAlgorithm != "" {
		params = append(params, "--ssh-key-algorithm", fluxConfig.Spec.Git.SshKeyAlgorithm)
	} else {
		params = append(params, "--ssh-key-algorithm", defaultPrivateKeyAlgorithm)
	}

	if cliConfig.GitSshKeyPassphrase != "" {
		params = append(params, "--password", cliConfig.GitSshKeyPassphrase)
	}

	env := make(map[string]string)
	env["SSH_KNOWN_HOSTS"] = cliConfig.GitKnownHostsFile
	_, err := f.ExecuteWithEnv(ctx, env, params...)
	if err != nil {
		return fmt.Errorf("executing flux bootstrap git: %v", err)
	}
	return err
}

func setUpCommonParamsBootstrap(cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig, params []string) []string {
	c := fluxConfig.Spec
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	if c.Branch != "" {
		params = append(params, "--branch", c.Branch)
	}
	if c.SystemNamespace != "" {
		params = append(params, "--namespace", c.SystemNamespace)
	}
	return params
}

func (f *Flux) Uninstall(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	params := []string{
		"uninstall",
		"--silent",
	}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	if c.SystemNamespace != "" {
		params = append(params, "--namespace", c.SystemNamespace)
	}

	_, err := f.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("uninstalling flux: %v", err)
	}
	return err
}

func (f *Flux) SuspendKustomization(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	if c.SystemNamespace == "" {
		return fmt.Errorf("executing flux suspend kustomization: namespace empty")
	}
	params := []string{"suspend", "ks", c.SystemNamespace, "--namespace", c.SystemNamespace}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	_, err := f.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing flux suspend kustomization: %v", err)
	}

	return err
}

func (f *Flux) ResumeKustomization(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	if c.SystemNamespace == "" {
		return fmt.Errorf("executing flux resume kustomization: namespace empty")
	}
	params := []string{"resume", "ks", c.SystemNamespace, "--namespace", c.SystemNamespace}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	_, err := f.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing flux resume kustomization: %v", err)
	}

	return err
}

func (f *Flux) Reconcile(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	params := []string{"reconcile", "source", "git"}

	if c.SystemNamespace != "" {
		params = append(params, c.SystemNamespace, "--namespace", c.SystemNamespace)
	} else {
		params = append(params, "flux-system")
	}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	if _, err := f.Execute(ctx, params...); err != nil {
		return fmt.Errorf("executing flux reconcile: %v", err)
	}

	return nil
}
