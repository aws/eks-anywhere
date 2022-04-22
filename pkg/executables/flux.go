package executables

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	fluxPath            = "flux"
	githubTokenEnv      = "GITHUB_TOKEN"
	gitProvider         = "github"
	privateKeyAlgorithm = "ecdsa"
)

type Flux struct {
	Executable
}

func NewFlux(executable Executable) *Flux {
	return &Flux{
		Executable: executable,
	}
}

// BootstrapToolkitsComponentsGithub creates the GitHub repository if it doesn’t exist, and commits the toolkit
// components manifests to the main branch. Then it configures the target cluster to synchronize with the repository.
// If the toolkit components are present on the cluster, the bootstrap command will perform an upgrade if needed.
func (f *Flux) BootstrapToolkitsComponentsGithub(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	c := fluxConfig.Spec
	params := []string{
		"bootstrap",
		gitProvider,
		"--repository", c.Github.Repository,
		"--owner", c.Github.Owner,
		"--path", c.ClusterConfigPath,
		"--ssh-key-algorithm", privateKeyAlgorithm,
	}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	if c.Github.Personal {
		params = append(params, "--personal")
	}
	if c.Branch != "" {
		params = append(params, "--branch", c.Branch)
	}
	if c.SystemNamespace != "" {
		params = append(params, "--namespace", c.SystemNamespace)
	}

	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return fmt.Errorf("setting token env: %v", err)
	}

	env := make(map[string]string)
	env[githubTokenEnv] = token

	_, err = f.ExecuteWithEnv(ctx, env, params...)
	if err != nil {
		return fmt.Errorf("executing flux bootstrap: %v", err)
	}

	return err
}

// BootstrapToolkitsComponentsGit creates the Git repository if it does not exist, and commits the toolkit
// components manifests to the main branch. Then it configures the target cluster to synchronize with the repository.
// If the toolkit components are present on the cluster, the bootstrap command will perform an upgrade if needed.
func (f *Flux) BootstrapToolkitsComponentsGit(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
	return nil
}

func (f *Flux) UninstallToolkitsComponents(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
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

func (f *Flux) PauseKustomization(ctx context.Context, cluster *types.Cluster, fluxConfig *v1alpha1.FluxConfig) error {
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
