package gitclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	gitTimeout     = 30 * time.Second
	maxRetries     = 5
	backOffPeriod  = 5 * time.Second
	emptyRepoError = "remote repository is empty"
)

type GitClient struct {
	Auth          transport.AuthMethod
	Client        GoGit
	RepoUrl       string
	RepoDirectory string
	Retrier       *retrier.Retrier
}

type Opt func(*GitClient)

func New(opts ...Opt) *GitClient {
	c := &GitClient{
		Client:  &goGit{},
		Retrier: retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithAuth(auth transport.AuthMethod) Opt {
	return func(c *GitClient) {
		c.Auth = auth
	}
}

func WithRepositoryUrl(repoUrl string) Opt {
	return func(c *GitClient) {
		c.RepoUrl = repoUrl
	}
}

func WithRepositoryDirectory(repoDir string) Opt {
	return func(c *GitClient) {
		c.RepoDirectory = repoDir
	}
}

func (g *GitClient) Clone(ctx context.Context) error {
	_, err := g.Client.Clone(ctx, g.RepoDirectory, g.RepoUrl, g.Auth)
	if err != nil && strings.Contains(err.Error(), emptyRepoError) {
		return &git.RepositoryIsEmptyError{
			Repository: g.RepoDirectory,
		}
	}
	return err
}

func (g *GitClient) Add(filename string) error {
	logger.V(3).Info("Opening directory", "directory", g.RepoDirectory)
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		return err
	}

	logger.V(3).Info("Opening working tree")
	w, err := g.Client.OpenWorktree(r)
	if err != nil {
		return err
	}

	logger.V(3).Info("Tracking specified files", "file", filename)
	err = g.Client.AddGlob(filename, w)
	return err
}

func (g *GitClient) Remove(filename string) error {
	logger.V(3).Info("Opening directory", "directory", g.RepoDirectory)
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		return err
	}

	logger.V(3).Info("Opening working tree")
	w, err := g.Client.OpenWorktree(r)
	if err != nil {
		return err
	}

	logger.V(3).Info("Removing specified files", "file", filename)
	_, err = g.Client.Remove(filename, w)
	return err
}

func (g *GitClient) Commit(message string) error {
	logger.V(3).Info("Opening directory", "directory", g.RepoDirectory)
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		logger.Info("Failed while attempting to open repo")
		return err
	}

	logger.V(3).Info("Opening working tree")
	w, err := g.Client.OpenWorktree(r)
	if err != nil {
		return err
	}

	logger.V(3).Info("Generating Commit object...")
	commitSignature := &object.Signature{
		Name: "EKS-A",
		When: time.Now(),
	}
	commit, err := g.Client.Commit(message, commitSignature, w)
	if err != nil {
		return err
	}

	logger.V(3).Info("Committing Object to local repo", "repo", g.RepoDirectory)
	finalizedCommit, err := g.Client.CommitObject(r, commit)
	logger.Info("Finalized commit and committed to local repository", "hash", finalizedCommit.Hash)
	return err
}

func (g *GitClient) Push(ctx context.Context) error {
	logger.V(3).Info("Pushing to remote", "repo", g.RepoDirectory)
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		return fmt.Errorf("err pushing: %v", err)
	}

	err = g.Client.PushWithContext(ctx, r, g.Auth)
	if err != nil {
		return fmt.Errorf("pushing: %v", err)
	}
	return err
}

func (g *GitClient) Pull(ctx context.Context, branch string) error {
	logger.V(3).Info("Pulling from remote", "repo", g.RepoDirectory, "remote", gogit.DefaultRemoteName)
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		return fmt.Errorf("pulling from remote: %v", err)
	}

	w, err := g.Client.OpenWorktree(r)
	if err != nil {
		return fmt.Errorf("pulling from remote: %v", err)
	}

	branchRef := plumbing.NewBranchReferenceName(branch)

	err = g.Client.PullWithContext(ctx, w, g.Auth, branchRef)

	if errors.Is(err, gogit.NoErrAlreadyUpToDate) {
		logger.V(3).Info("Local repo already up-to-date", "repo", g.RepoDirectory, "remote", gogit.DefaultRemoteName)
		return &git.RepositoryUpToDateError{}
	}

	if err != nil {
		return fmt.Errorf("pulling from remote: %v", err)
	}

	ref, err := g.Client.Head(r)
	if err != nil {
		return fmt.Errorf("pulling from remote: %v", err)
	}

	commit, err := g.Client.CommitObject(r, ref.Hash())
	if err != nil {
		return fmt.Errorf("accessing latest commit after pulling from remote: %v", err)
	}
	logger.V(3).Info("Successfully pulled from remote", "repo", g.RepoDirectory, "remote", gogit.DefaultRemoteName, "latest commit", commit.Hash)
	return nil
}

func (g *GitClient) Init() error {
	r, err := g.Client.Init(g.RepoDirectory)
	if err != nil {
		return err
	}

	if _, err = g.Client.Create(r, g.RepoUrl); err != nil {
		return fmt.Errorf("initializing repository: %v", err)
	}
	return nil
}

func (g *GitClient) Branch(name string) error {
	r, err := g.Client.OpenDir(g.RepoDirectory)
	if err != nil {
		return fmt.Errorf("creating branch %s: %v", name, err)
	}

	localBranchRef := plumbing.NewBranchReferenceName(name)

	branchOpts := &config.Branch{
		Name:   name,
		Remote: gogit.DefaultRemoteName,
		Merge:  localBranchRef,
		Rebase: "true",
	}

	err = g.Client.CreateBranch(r, branchOpts)
	branchExistsLocally := errors.Is(err, gogit.ErrBranchExists)

	if err != nil && !branchExistsLocally {
		return fmt.Errorf("creating branch %s: %v", name, err)
	}

	if branchExistsLocally {
		logger.V(3).Info("Branch already exists locally", "branch", name)
	}

	if !branchExistsLocally {
		logger.V(3).Info("Branch does not exist locally", "branch", name)
		headref, err := g.Client.Head(r)
		if err != nil {
			return fmt.Errorf("creating branch %s: %v", name, err)
		}
		h := headref.Hash()
		err = g.Client.SetRepositoryReference(r, plumbing.NewHashReference(localBranchRef, h))
		if err != nil {
			return fmt.Errorf("creating branch %s: %v", name, err)
		}
	}

	w, err := g.Client.OpenWorktree(r)
	if err != nil {
		return fmt.Errorf("creating branch %s: %v", name, err)
	}

	err = g.Client.Checkout(w, &gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName(localBranchRef.String()),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("creating branch %s: %v", name, err)
	}

	err = g.pullIfRemoteExists(r, w, name, localBranchRef)
	if err != nil {
		return fmt.Errorf("creating branch %s: %v", name, err)
	}

	return nil
}

func (g *GitClient) ValidateRemoteExists(ctx context.Context) error {
	logger.V(3).Info("Validating git setup", "repoUrl", g.RepoUrl)
	remote := g.Client.NewRemote(g.RepoUrl, gogit.DefaultRemoteName)
	// Check if we are able to make a connection to the remote by attempting to list refs
	_, err := g.Client.ListWithContext(ctx, remote, g.Auth)
	if err != nil {
		return fmt.Errorf("connecting with remote %v for repository: %v", gogit.DefaultRemoteName, err)
	}
	return nil
}

func (g *GitClient) pullIfRemoteExists(r *gogit.Repository, w *gogit.Worktree, branchName string, localBranchRef plumbing.ReferenceName) error {
	err := g.Retrier.Retry(func() error {
		remoteExists, err := g.remoteBranchExists(r, localBranchRef)
		if err != nil {
			return fmt.Errorf("checking if remote branch exists %s: %v", branchName, err)
		}

		if remoteExists {
			err = g.Client.PullWithContext(context.Background(), w, g.Auth, localBranchRef)
			if err != nil && !errors.Is(err, gogit.NoErrAlreadyUpToDate) && !errors.Is(err, gogit.ErrRemoteNotFound) {
				return fmt.Errorf("pulling from remote when checking out existing branch %s: %v", branchName, err)
			}
		}
		return nil
	})
	return err
}

func (g *GitClient) remoteBranchExists(r *gogit.Repository, localBranchRef plumbing.ReferenceName) (bool, error) {
	reflist, err := g.Client.ListRemotes(r, g.Auth)
	if err != nil {
		if strings.Contains(err.Error(), emptyRepoError) {
			return false, nil
		}
		return false, fmt.Errorf("listing remotes: %v", err)
	}
	lb := localBranchRef.String()
	for _, ref := range reflist {
		if ref.Name().String() == lb {
			return true, nil
		}
	}
	return false, nil
}

type GoGit interface {
	AddGlob(f string, w *gogit.Worktree) error
	Checkout(w *gogit.Worktree, opts *gogit.CheckoutOptions) error
	Clone(ctx context.Context, dir string, repoUrl string, auth transport.AuthMethod) (*gogit.Repository, error)
	Commit(m string, sig *object.Signature, w *gogit.Worktree) (plumbing.Hash, error)
	CommitObject(r *gogit.Repository, h plumbing.Hash) (*object.Commit, error)
	Create(r *gogit.Repository, url string) (*gogit.Remote, error)
	CreateBranch(r *gogit.Repository, config *config.Branch) error
	Head(r *gogit.Repository) (*plumbing.Reference, error)
	NewRemote(url, remoteName string) *gogit.Remote
	Init(dir string) (*gogit.Repository, error)
	OpenDir(dir string) (*gogit.Repository, error)
	OpenWorktree(r *gogit.Repository) (*gogit.Worktree, error)
	PushWithContext(ctx context.Context, r *gogit.Repository, auth transport.AuthMethod) error
	PullWithContext(ctx context.Context, w *gogit.Worktree, auth transport.AuthMethod, ref plumbing.ReferenceName) error
	ListRemotes(r *gogit.Repository, auth transport.AuthMethod) ([]*plumbing.Reference, error)
	ListWithContext(ctx context.Context, r *gogit.Remote, auth transport.AuthMethod) ([]*plumbing.Reference, error)
	Remove(f string, w *gogit.Worktree) (plumbing.Hash, error)
	SetRepositoryReference(r *gogit.Repository, p *plumbing.Reference) error
}

type goGit struct{}

func (gg *goGit) Clone(ctx context.Context, dir string, repourl string, auth transport.AuthMethod) (*gogit.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	return gogit.PlainCloneContext(ctx, dir, false, &gogit.CloneOptions{
		Auth:     auth,
		URL:      repourl,
		Progress: os.Stdout,
	})
}

func (gg *goGit) OpenDir(dir string) (*gogit.Repository, error) {
	return gogit.PlainOpen(dir)
}

func (gg *goGit) OpenWorktree(r *gogit.Repository) (*gogit.Worktree, error) {
	return r.Worktree()
}

func (gg *goGit) AddGlob(f string, w *gogit.Worktree) error {
	return w.AddGlob(f)
}

func (gg *goGit) Commit(m string, sig *object.Signature, w *gogit.Worktree) (plumbing.Hash, error) {
	return w.Commit(m, &gogit.CommitOptions{
		Author:            sig,
		AllowEmptyCommits: true,
	})
}

func (gg *goGit) CommitObject(r *gogit.Repository, h plumbing.Hash) (*object.Commit, error) {
	return r.CommitObject(h)
}

func (gg *goGit) PushWithContext(ctx context.Context, r *gogit.Repository, auth transport.AuthMethod) error {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	return r.PushContext(ctx, &gogit.PushOptions{
		Auth: auth,
	})
}

func (gg *goGit) PullWithContext(ctx context.Context, w *gogit.Worktree, auth transport.AuthMethod, ref plumbing.ReferenceName) error {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	return w.PullContext(ctx, &gogit.PullOptions{RemoteName: gogit.DefaultRemoteName, Auth: auth, ReferenceName: ref})
}

func (gg *goGit) Head(r *gogit.Repository) (*plumbing.Reference, error) {
	return r.Head()
}

func (gg *goGit) Init(dir string) (*gogit.Repository, error) {
	return gogit.PlainInit(dir, false)
}

func (ggc *goGit) NewRemote(url, remoteName string) *gogit.Remote {
	return gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: remoteName,
		URLs: []string{url},
	})
}

func (gg *goGit) Checkout(worktree *gogit.Worktree, opts *gogit.CheckoutOptions) error {
	return worktree.Checkout(opts)
}

func (gg *goGit) Create(r *gogit.Repository, url string) (*gogit.Remote, error) {
	return r.CreateRemote(&config.RemoteConfig{
		Name: gogit.DefaultRemoteName,
		URLs: []string{url},
	})
}

func (gg *goGit) CreateBranch(repo *gogit.Repository, config *config.Branch) error {
	return repo.CreateBranch(config)
}

func (gg *goGit) ListRemotes(r *gogit.Repository, auth transport.AuthMethod) ([]*plumbing.Reference, error) {
	remote, err := r.Remote(gogit.DefaultRemoteName)
	if err != nil {
		if errors.Is(err, gogit.ErrRemoteNotFound) {
			return []*plumbing.Reference{}, nil
		}
		return nil, err
	}
	refList, err := remote.List(&gogit.ListOptions{Auth: auth})
	if err != nil {
		return nil, err
	}
	return refList, nil
}

func (gg *goGit) Remove(f string, w *gogit.Worktree) (plumbing.Hash, error) {
	return w.Remove(f)
}

func (ggc *goGit) ListWithContext(ctx context.Context, r *gogit.Remote, auth transport.AuthMethod) ([]*plumbing.Reference, error) {
	refList, err := r.ListContext(ctx, &gogit.ListOptions{Auth: auth})
	if err != nil {
		return nil, err
	}
	return refList, nil
}

func (gg *goGit) SetRepositoryReference(r *gogit.Repository, p *plumbing.Reference) error {
	return r.Storer.SetReference(p)
}
