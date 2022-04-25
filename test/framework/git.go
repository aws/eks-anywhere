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
)

type GitOptions struct {
	GitProvider git.ProviderClient
	GitClient   git.Client
	Writer      filewriter.FileWriter
}

func (e *ClusterE2ETest) NewGitOptions(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter, repoPath string) (*GitOptions, error) {
	if fluxConfig == nil {
		return nil, nil
	}

	var localGitWriterPath string
	if repoPath == "" {
		localGitWriterPath = filepath.Join("git", fluxConfig.Spec.Github.Repository)
	} else {
		localGitWriterPath = repoPath
	}

	gitProviderFactory := gitFactory.New()
	gitProvider, gitClient, err := gitProviderFactory.Build(ctx, cluster, fluxConfig)
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
		GitProvider: gitProvider,
		GitClient:   gitClient,
		Writer:      gitwriter,
	}, nil
}
