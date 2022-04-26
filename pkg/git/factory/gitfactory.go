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

type gitProviderFactory struct{}

func New() *gitProviderFactory {
	return &gitProviderFactory{}
}

func (g *gitProviderFactory) Build(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter) (git.ProviderClient, git.Client, filewriter.FileWriter, error) {
	var providerClient git.ProviderClient
	var gitClient git.Client
	var repoWriter filewriter.FileWriter
	var err error

	switch {
	case fluxConfig.Spec.Github != nil:
		providerClient, gitClient, repoWriter, err = g.buildGithubProvider(ctx, cluster.Name, fluxConfig.Spec.Github.Owner, fluxConfig.Spec.Github.Repository, fluxConfig.Spec.Github.Personal, writer)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("building github provider: %v", err)
		}
	default:
		return nil, nil, nil, fmt.Errorf("no valid git provider in FluxConfigSpec. Spec: %v", fluxConfig)
	}
	return providerClient, gitClient, repoWriter, nil
}

func (g *gitProviderFactory) buildGithubProvider(ctx context.Context, clusterName string, owner string, repo string, personal bool, writer filewriter.FileWriter) (git.ProviderClient, git.Client, filewriter.FileWriter, error) {
	token, err := github.GetGithubAccessTokenFromEnv()
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	localGitRepoPath := filepath.Join(clusterName, "git", repo)
	client := gitclient.New(auth, localGitRepoPath, github.RepoUrl(owner, repo))

	w, err := newRepositoryWriter(writer, repo)
	if err != nil {
		return nil, nil, nil, err
	}

	return provider, client, w, nil
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
