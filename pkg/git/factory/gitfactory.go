package gitfactory

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gitclient"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

type gitProviderFactory struct {
	GitClient *git.Client
}

type Options struct {
	GithubGitClient git.Client
}

func New(opts Options) *gitProviderFactory {
	return &gitProviderFactory{GitClient: &opts.GithubGitClient}
}

func NewFactory() *gitProviderFactory {
	return &gitProviderFactory{}
}

func (g *gitProviderFactory) Build(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig) (git.ProviderClient, git.Client, error) {
	var providerClient git.ProviderClient
	var gitClient git.Client
	var err error
	switch {
	case fluxConfig.Spec.Github != nil:
		providerClient, gitClient, err = g.buildGithubProvider(ctx, cluster.Name, fluxConfig.Spec.Github.Owner, fluxConfig.Spec.Github.Repository, fluxConfig.Spec.Github.Personal)
		if err != nil {
			return nil, nil, fmt.Errorf("building github provider: %v", err)
		}
	default:
		return nil, nil, fmt.Errorf("no valid git provider in FluxConfigSpec. Spec: %v", fluxConfig)
	}
	return providerClient, gitClient, nil
}

func (g *gitProviderFactory) buildGithubProvider(ctx context.Context, clusterName string, owner string, repo string, personal bool) (git.ProviderClient, git.Client, error) {
	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return nil, nil, err
	}
	auth := git.TokenAuth{Token: token, Username: owner}
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	githubProviderOpts := github.Options{
		Repository: repo,
		Owner:      owner,
		Personal:   personal,
	}
	provider, err := github.New(githubProviderClient, githubProviderOpts, auth)
	if err != nil {
		return nil, nil, err
	}

	localGitRepoPath := filepath.Join(clusterName, "git", repo)
	client := gitclient.New(auth, localGitRepoPath, github.RepoUrl(owner, repo))

	return provider, client, nil
}
