package framework

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/git/gogit"
)

type GitOptions struct {
	Git    git.Provider
	Writer filewriter.FileWriter
}

func (e *ClusterE2ETest) NewGitOptions(ctx context.Context, cluster *v1alpha1.Cluster, gitOpsConfig *v1alpha1.GitOpsConfig, writer filewriter.FileWriter, repoPath string) (*GitOptions, error) {
	if gitOpsConfig == nil {
		return nil, nil
	}

	var localGitRepoPath string
	var localGitWriterPath string
	if repoPath == "" {
		localGitRepoPath = filepath.Join(cluster.Name, "git", gitOpsConfig.Spec.Flux.Github.Repository)
		localGitWriterPath = filepath.Join("git", gitOpsConfig.Spec.Flux.Github.Repository)
	} else {
		localGitRepoPath = repoPath
		localGitWriterPath = repoPath
	}

	gogitOptions := gogit.Options{
		RepositoryDirectory: localGitRepoPath,
	}
	goGit := gogit.New(gogitOptions)

	gitProviderFactoryOptions := gitFactory.Options{GithubGitClient: goGit}
	gitProviderFactory := gitFactory.New(gitProviderFactoryOptions)
	gitProvider, err := gitProviderFactory.BuildProvider(ctx, &gitOpsConfig.Spec)
	if err != nil {
		return nil, fmt.Errorf("creating Git provider: %v", err)
	}
	err = gitProvider.Validate(ctx)
	if err != nil {
		return nil, err
	}
	gitwriter, err := writer.WithDir(localGitWriterPath)
	if err != nil {
		return nil, fmt.Errorf("creating file writer: %v", err)
	}
	gitwriter.CleanUpTemp()
	return &GitOptions{
		Git:    gitProvider,
		Writer: gitwriter,
	}, nil
}
