package git

import (
	"context"
	"fmt"
)

type Client interface {
	Add(filename string) error
	Remove(filename string) error
	Clone(ctx context.Context) error
	Commit(message string) error
	Push(ctx context.Context) error
	Pull(ctx context.Context, branch string) error
	Init() error
	Branch(name string) error
}

type ProviderClient interface {
	GetRepo(ctx context.Context) (repo *Repository, err error)
	CreateRepo(ctx context.Context, opts CreateRepoOpts) (repo *Repository, err error)
	DeleteRepo(ctx context.Context, opts DeleteRepoOpts) error
	Validate(ctx context.Context) error
	PathExists(ctx context.Context, owner, repo, branch, path string) (bool, error)
}

type CreateRepoOpts struct {
	Name        string
	Owner       string
	Description string
	Personal    bool
	Privacy     bool
}

type GetRepoOpts struct {
	Owner      string
	Repository string
}

type DeleteRepoOpts struct {
	Owner      string
	Repository string
}

type Repository struct {
	Name         string
	Owner        string
	Organization string
	CloneUrl     string
}

type TokenAuth struct {
	Username string
	Token    string
}

type RepositoryDoesNotExistError struct {
	repository string
	owner      string
	Err        error
}

func (e *RepositoryDoesNotExistError) Error() string {
	return fmt.Sprintf("repository %s with owner %s not found: %s", e.repository, e.owner, e.Err)
}

type RepositoryIsEmptyError struct {
	Repository string
}

func (e *RepositoryIsEmptyError) Error() string {
	return fmt.Sprintf("repository %s is empty can cannot be cloned", e.Repository)
}

type RepositoryUpToDateError struct {
	Repository string
}

func (e *RepositoryUpToDateError) Error() string {
	return fmt.Sprintf("error pulling from repository %s: already up-to-date", e.Repository)
}

type RemoteBranchDoesNotExistError struct {
	Repository string
	Branch     string
}

func (e *RemoteBranchDoesNotExistError) Error() string {
	return fmt.Sprintf("error pulling from repository %s: remote branch %s does not exist", e.Repository, e.Branch)
}
