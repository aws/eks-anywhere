package flux

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/git"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type gitClient struct {
	git         git.Client
	gitProvider git.ProviderClient
	*retrier.Retrier
}

func newGitClient(gitTools *gitFactory.GitTools) *gitClient {
	if gitTools == nil {
		return nil
	}

	return &gitClient{
		git:         gitTools.Client,
		gitProvider: gitTools.Provider,
		Retrier:     retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
}

func (c *gitClient) GetRepo(ctx context.Context) (repo *git.Repository, err error) {
	if c.gitProvider == nil {
		return nil, nil
	}

	err = c.Retry(
		func() error {
			repo, err = c.gitProvider.GetRepo(ctx)
			return err
		},
	)
	return repo, err
}

func (c *gitClient) CreateRepo(ctx context.Context, opts git.CreateRepoOpts) error {
	if c.gitProvider == nil {
		return nil
	}

	return c.Retry(
		func() error {
			_, err := c.gitProvider.CreateRepo(ctx, opts)
			return err
		},
	)
}

func (c *gitClient) Clone(ctx context.Context) error {
	return c.Retry(
		func() error {
			return c.git.Clone(ctx)
		},
	)
}

func (c *gitClient) Push(ctx context.Context) error {
	return c.Retry(
		func() error {
			return c.git.Push(ctx)
		},
	)
}

func (c *gitClient) Pull(ctx context.Context, branch string) error {
	return c.Retry(
		func() error {
			return c.git.Pull(ctx, branch)
		},
	)
}

func (c *gitClient) PathExists(ctx context.Context, owner, repo, branch, path string) (exists bool, err error) {
	if c.gitProvider == nil {
		return false, nil
	}

	err = c.Retry(
		func() error {
			exists, err = c.gitProvider.PathExists(ctx, owner, repo, branch, path)
			return err
		},
	)
	return exists, err
}

func (c *gitClient) Add(filename string) error {
	return c.git.Add(filename)
}

func (c *gitClient) Remove(filename string) error {
	return c.git.Remove(filename)
}

func (c *gitClient) Commit(message string) error {
	return c.git.Commit(message)
}

func (c *gitClient) Branch(name string) error {
	return c.git.Branch(name)
}

func (c *gitClient) Init() error {
	return c.git.Init()
}
