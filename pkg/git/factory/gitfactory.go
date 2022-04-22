package gitfactory

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

type gitProviderFactory struct {
	GitClient github.GitClient
}

type Options struct {
	GithubGitClient github.GitClient
}

func New(opts Options) *gitProviderFactory {
	return &gitProviderFactory{GitClient: opts.GithubGitClient}
}

// BuildProvider will configure and return the proper Github based on the given GitOps configuration.
func (g *gitProviderFactory) BuildProvider(ctx context.Context, fluxConfig *v1alpha1.FluxConfigSpec) (git.Provider, error) {
	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return nil, err
	}
	auth := git.TokenAuth{Token: token, Username: fluxConfig.Github.Owner}
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	githubProviderOpts := github.Options{
		Repository: fluxConfig.Github.Repository,
		Owner:      fluxConfig.Github.Owner,
		Personal:   fluxConfig.Github.Personal,
	}
	provider, err := github.New(g.GitClient, githubProviderClient, githubProviderOpts, auth)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
