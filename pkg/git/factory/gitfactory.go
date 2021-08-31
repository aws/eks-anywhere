package gitfactory

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

type gitProviderFactory struct {
	GithubGitClient github.GitProviderClient
}

type Options struct {
	GithubGitClient github.GitProviderClient
}

func New(opts Options) *gitProviderFactory {
	return &gitProviderFactory{GithubGitClient: opts.GithubGitClient}
}

// BuildProvider will configure and return the proper Github based on the given GitOps configuration.
func (g *gitProviderFactory) BuildProvider(ctx context.Context, gitOpsConfig *v1alpha1.GitOpsConfigSpec) (git.Provider, error) {
	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return nil, err
	}
	auth := git.TokenAuth{Token: token, Username: gitOpsConfig.Flux.Github.Owner}
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	githubProviderOpts := github.Options{
		Repository: gitOpsConfig.Flux.Github.Repository,
		Owner:      gitOpsConfig.Flux.Github.Owner,
		Personal:   gitOpsConfig.Flux.Github.Personal,
	}
	provider, err := github.New(g.GithubGitClient, githubProviderClient, githubProviderOpts, auth)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
