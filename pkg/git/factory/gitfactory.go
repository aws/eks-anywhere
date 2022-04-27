package gitfactory

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gitclient"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

func Build(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter) (*addonclients.GitTools, error) {
	var clientOptions []gitclient.Opt
	var repo string
	var err error

	tools := &addonclients.GitTools{}

	switch {
	case fluxConfig.Spec.Github != nil:
		githubToken, err := github.GetGithubAccessTokenFromEnv()
		if err != nil {
			return nil, err
		}

		repo = fluxConfig.Spec.Github.Repository
		clientOptions = append(clientOptions, gitclient.WithRepositoryUrl(github.RepoUrl(fluxConfig.Spec.Github.Owner, repo)))

		clientAuth := git.TokenAuth{Token: githubToken, Username: fluxConfig.Spec.Github.Owner}
		clientOptions = append(clientOptions, gitclient.WithTokenAuth(clientAuth))

		tools.Provider, err = buildGithubProvider(ctx, clientAuth, fluxConfig.Spec.Github.Owner, repo, fluxConfig.Spec.Github.Personal)
		if err != nil {
			return tools, fmt.Errorf("building github provider: %v", err)
		}
	default:
		return nil, fmt.Errorf("no valid git provider in FluxConfigSpec. Spec: %v", fluxConfig)
	}

	localGitRepoPath := filepath.Join(cluster.Name, "git", repo)
	clientOptions = append(clientOptions, gitclient.WithRepositoryDirectory(localGitRepoPath))
	tools.Client = gitclient.New(clientOptions...)

	repoWriter, err := newRepositoryWriter(writer, repo)
	if err != nil {
		return nil, err
	}
	tools.Writer = repoWriter

	return tools, nil
}

func buildGithubProvider(ctx context.Context, auth git.TokenAuth, owner string, repo string, personal bool) (git.ProviderClient, error) {
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	githubProviderOpts := github.Options{
		Repository: repo,
		Owner:      owner,
		Personal:   personal,
	}
	provider, err := github.New(githubProviderClient, githubProviderOpts, auth)
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
