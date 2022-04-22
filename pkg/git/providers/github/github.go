package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	goGithub "github.com/google/go-github/v35/github"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	GitProviderName    = "github"
	EksaGithubTokenEnv = "EKSA_GITHUB_TOKEN"
	GithubTokenEnv     = "GITHUB_TOKEN"
	githubUrlTemplate  = "https://github.com/%v/%v.git"
	patRegex           = "^[A-Za-z0-9_]{40}$"
	repoPermissions    = "repo"
)

type githubProvider struct {
	gitProviderClient    GitClient
	githubProviderClient GithubClient
	options              Options
	auth                 git.TokenAuth
}

type Options struct {
	Repository string
	Owner      string
	Personal   bool
}

// GitClient represents the attributes that the GitHub provider requires of a low-level git implementation (e.g. gogit) in order to function.
// Any basic git implementation (gogit, an executable wrapper, etc) which supports these methods can be used by the GitHub specific provider.
type GitClient interface {
	Add(filename string) error
	Remove(filename string) error
	Clone(ctx context.Context, repourl string) error
	Commit(message string) error
	Push(ctx context.Context) error
	Pull(ctx context.Context, branch string) error
	Init(url string) error
	Branch(name string) error
	SetTokenAuth(token string, username string)
}

// GithubClient represents the attributes that the Github provider requires of a library to directly connect to and interact with the Github API.
type GithubClient interface {
	GetRepo(ctx context.Context, opts git.GetRepoOpts) (repo *git.Repository, err error)
	CreateRepo(ctx context.Context, opts git.CreateRepoOpts) (repo *git.Repository, err error)
	AuthenticatedUser(ctx context.Context) (*goGithub.User, error)
	Organization(ctx context.Context, org string) (*goGithub.Organization, error)
	GetAccessTokenPermissions(accessToken string) (string, error)
	CheckAccessTokenPermissions(checkPATPermission string, allPermissionScopes string) error
	PathExists(ctx context.Context, owner, repo, branch, path string) (bool, error)
	DeleteRepo(ctx context.Context, opts git.DeleteRepoOpts) error
}

func New(gitProviderClient GitClient, githubProviderClient GithubClient, opts Options, auth git.TokenAuth) (git.Provider, error) {
	gitProviderClient.SetTokenAuth(auth.Token, auth.Username)
	return &githubProvider{
		gitProviderClient:    gitProviderClient,
		githubProviderClient: githubProviderClient,
		options:              opts,
		auth:                 auth,
	}, nil
}

func (g *githubProvider) Add(filename string) error {
	return g.gitProviderClient.Add(filename)
}

func (g *githubProvider) Remove(filename string) error {
	return g.gitProviderClient.Remove(filename)
}

func (g *githubProvider) Clone(ctx context.Context) error {
	return g.gitProviderClient.Clone(ctx, g.RepoUrl())
}

func (g *githubProvider) Commit(message string) error {
	return g.gitProviderClient.Commit(message)
}

func (g *githubProvider) Push(ctx context.Context) error {
	err := g.gitProviderClient.Push(ctx)
	return err
}

func (g *githubProvider) Pull(ctx context.Context, branch string) error {
	err := g.gitProviderClient.Pull(ctx, branch)
	return err
}

func (g *githubProvider) Init() error {
	return g.gitProviderClient.Init(g.RepoUrl())
}

func (g *githubProvider) Branch(name string) error {
	return g.gitProviderClient.Branch(name)
}

// CreateRepo creates an empty Github Repository. The repository must be initialized locally or
// file must be added to it via the github api before it can be successfully cloned.
func (g *githubProvider) CreateRepo(ctx context.Context, opts git.CreateRepoOpts) (repository *git.Repository, err error) {
	return g.githubProviderClient.CreateRepo(ctx, opts)
}

// GetRepo describes a remote repository, return the repo name if it exists.
// If the repo does not exist, a nil repo is returned.
func (g *githubProvider) GetRepo(ctx context.Context) (*git.Repository, error) {
	r := g.options.Repository
	o := g.options.Owner
	logger.V(3).Info("Describing Github repository", "name", r, "owner", o)
	opts := git.GetRepoOpts{Owner: o, Repository: r}
	repo, err := g.githubProviderClient.GetRepo(ctx, opts)
	if err != nil {
		var e *git.RepositoryDoesNotExistError
		if errors.As(err, &e) {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected error when describing repository %s: %w", r, err)
	}
	return repo, err
}

// validates the github setup and access
func (g *githubProvider) Validate(ctx context.Context) error {
	user, err := g.githubProviderClient.AuthenticatedUser(ctx)
	if err != nil {
		return err
	}
	accessToken := g.auth.Token
	allPermissions, err := g.githubProviderClient.GetAccessTokenPermissions(accessToken)
	if err != nil {
		return err
	}
	err = g.githubProviderClient.CheckAccessTokenPermissions(repoPermissions, allPermissions)
	if err != nil {
		return err
	}
	logger.MarkPass("Github personal access token has the required repo permissions")
	if g.options.Personal {
		if !strings.EqualFold(g.options.Owner, *user.Login) {
			return fmt.Errorf("the authenticated Github.com user and owner %s specified in the EKS-A gitops spec don't match; confirm access token owner is %s", g.options.Owner, g.options.Owner)
		}
		return nil
	}
	org, err := g.githubProviderClient.Organization(ctx, g.options.Owner)
	if err != nil {
		return fmt.Errorf("the authenticated github user doesn't have proper access to github organization %s, %v", g.options.Owner, err)
	}
	if org == nil { // for now only checks if user belongs to the org
		return fmt.Errorf("the authenticated github user doesn't have proper access to github organization %s", g.options.Owner)
	}
	return nil
}

func validateGithubAccessToken() error {
	r := regexp.MustCompile(patRegex)
	logger.V(4).Info("Checking validity of Github Access Token environment variable", "env var", EksaGithubTokenEnv)
	val, ok := os.LookupEnv(EksaGithubTokenEnv)
	if !ok {
		return fmt.Errorf("github access token environment variable %s is invalid; could not get var from environment", EksaGithubTokenEnv)
	}
	if !r.MatchString(val) {
		return fmt.Errorf("github access token environment variable %s is invalid; must match format %s", EksaGithubTokenEnv, patRegex)
	}
	return nil
}

func GetGithubAccessTokenFromEnv() (string, error) {
	err := validateGithubAccessToken()
	if err != nil {
		return "", err
	}

	env := make(map[string]string)
	if val, ok := os.LookupEnv(EksaGithubTokenEnv); ok && len(val) > 0 {
		env[GithubTokenEnv] = val
		if err := os.Setenv(GithubTokenEnv, val); err != nil {
			return "", fmt.Errorf("unable to set %s: %v", GithubTokenEnv, err)
		}
	}
	return env[GithubTokenEnv], nil
}

func (g *githubProvider) RepoUrl() string {
	return fmt.Sprintf(githubUrlTemplate, g.options.Owner, g.options.Repository)
}

func (g *githubProvider) PathExists(ctx context.Context, owner, repo, branch, path string) (bool, error) {
	return g.githubProviderClient.PathExists(ctx, owner, repo, branch, path)
}

func (g *githubProvider) DeleteRepo(ctx context.Context, opts git.DeleteRepoOpts) error {
	return g.githubProviderClient.DeleteRepo(ctx, opts)
}

type GitProviderNotFoundError struct {
	Provider string
}

func (e *GitProviderNotFoundError) Error() string {
	return fmt.Sprintf("git provider %s not found", e.Provider)
}
