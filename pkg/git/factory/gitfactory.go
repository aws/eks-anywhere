package gitfactory

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gitclient"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

type GitTools struct {
	Provider git.ProviderClient
	Client   git.Client
	Writer   filewriter.FileWriter
}

func Build(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter) (*GitTools, error) {
	var provider git.ProviderClient
	var repo string
	var repoUrl string
	var tokenAuth *git.TokenAuth
	var err error

	switch {
	case fluxConfig.Spec.Github != nil:
		githubToken, err := github.GetGithubAccessTokenFromEnv()
		if err != nil {
			return nil, err
		}

		repo = fluxConfig.Spec.Github.Repository
		repoUrl = github.RepoUrl(fluxConfig.Spec.Github.Owner, repo)
		tokenAuth = &git.TokenAuth{Token: githubToken, Username: fluxConfig.Spec.Github.Owner}
		githubProviderOpts := github.Options{
			Repository: repo,
			Owner:      fluxConfig.Spec.Github.Owner,
			Personal:   fluxConfig.Spec.Github.Personal,
		}
		provider, err = buildGithubProvider(ctx, *tokenAuth, githubProviderOpts)
		if err != nil {
			return nil, fmt.Errorf("building github provider: %v", err)
		}
	default:
		return nil, fmt.Errorf("no valid git provider in FluxConfigSpec. Spec: %v", fluxConfig)
	}

	localGitRepoPath := filepath.Join(cluster.Name, "git", repo)
	client := buildGitClient(ctx, tokenAuth, repoUrl, localGitRepoPath)

	repoWriter, err := newRepositoryWriter(writer, repo)
	if err != nil {
		return nil, err
	}

	return &GitTools{
		Writer:   repoWriter,
		Client:   client,
		Provider: provider,
	}, nil
}

func buildGitClient(ctx context.Context, tokenAuth *git.TokenAuth, repoUrl string, repo string) *gitclient.GitClient {
	opts := []gitclient.Opt{
		gitclient.WithRepositoryUrl(repoUrl),
		gitclient.WithRepositoryDirectory(repo),
	}
	// right now, we only support token auth
	// however, the generic git provider will support both token auth and SSH auth
	if tokenAuth != nil {
		opts = append(opts, gitclient.WithTokenAuth(*tokenAuth))
	}
	return gitclient.New(opts...)
}

func buildGithubProvider(ctx context.Context, auth git.TokenAuth, options github.Options) (git.ProviderClient, error) {
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	provider, err := github.New(githubProviderClient, options, auth)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func newRepositoryWriter(writer filewriter.FileWriter, repository string) (filewriter.FileWriter, error) {
	localGitWriterPath := filepath.Join("git", repository)
	gitwriter, err := writer.WithDir(localGitWriterPath)
	if err != nil {
		return nil, fmt.Errorf("creating file writer: %v", err)
	}
	gitwriter.CleanUpTemp()
	return gitwriter, nil
}
