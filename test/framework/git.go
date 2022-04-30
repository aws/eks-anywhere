package framework

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
)

func (e *ClusterE2ETest) NewGitTools(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter, repoPath string) (*gitFactory.GitTools, error) {
	if fluxConfig == nil {
		return nil, nil
	}

	var localGitWriterPath string
	var localGitRepoPath string
	if repoPath == "" {
		localGitWriterPath = filepath.Join("git", fluxConfig.Spec.Github.Repository)
		localGitRepoPath = filepath.Join(cluster.Name, "git", fluxConfig.Spec.Github.Repository)
	} else {
		localGitWriterPath = repoPath
		localGitRepoPath = repoPath
	}

	tools, err := gitFactory.Build(ctx, cluster, fluxConfig, writer, gitFactory.WithRepositoryDirectory(localGitRepoPath))
	if err != nil {
		return nil, fmt.Errorf("creating Git provider: %v", err)
	}
	err = tools.Provider.Validate(ctx)
	if err != nil {
		return nil, err
	}
	gitwriter, err := writer.WithDir(localGitWriterPath)
	if err != nil {
		return nil, fmt.Errorf("creating file writer: %v", err)
	}
	gitwriter.CleanUpTemp()
	tools.Writer = gitwriter
	return tools, nil
}
