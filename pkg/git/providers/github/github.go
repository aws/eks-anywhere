package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	goGithub "github.com/google/go-github/v35/github"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	GitProviderName    = "github"
	EksaGithubTokenEnv = "EKSA_GITHUB_TOKEN"
	GithubTokenEnv     = "GITHUB_TOKEN"
	githubUrlTemplate  = "https://github.com/%v/%v.git"
	patRegex           = "^ghp_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}$"
	repoPermissions    = "repo"
)

type githubProvider struct {
	githubProviderClient GithubClient
	config               *v1alpha1.GithubProviderConfig
	auth                 git.TokenAuth
}

type Options struct {
	Repository string
	Owner      string
	Personal   bool
}

// GithubClient represents the attributes that the Github provider requires of a library to directly connect to and interact with the Github API.
type GithubClient interface {
	GetRepo(ctx context.Context, opts git.GetRepoOpts) (repo *git.Repository, err error)
	CreateRepo(ctx context.Context, opts git.CreateRepoOpts) (repo *git.Repository, err error)
	AddDeployKeyToRepo(ctx context.Context, opts git.AddDeployKeyOpts) error
	AuthenticatedUser(ctx context.Context) (*goGithub.User, error)
	Organization(ctx context.Context, org string) (*goGithub.Organization, error)
	GetAccessTokenPermissions(accessToken string) (string, error)
	CheckAccessTokenPermissions(checkPATPermission string, allPermissionScopes string) error
	PathExists(ctx context.Context, owner, repo, branch, path string) (bool, error)
	DeleteRepo(ctx context.Context, opts git.DeleteRepoOpts) error
}

func New(githubProviderClient GithubClient, config *v1alpha1.GithubProviderConfig, auth git.TokenAuth) (*githubProvider, error) {
	return &githubProvider{
		githubProviderClient: githubProviderClient,
		config:               config,
		auth:                 auth,
	}, nil
}

// CreateRepo creates an empty Github Repository. The repository must be initialized locally or
// file must be added to it via the github api before it can be successfully cloned.
func (g *githubProvider) CreateRepo(ctx context.Context, opts git.CreateRepoOpts) (repository *git.Repository, err error) {
	return g.githubProviderClient.CreateRepo(ctx, opts)
}

// GetRepo describes a remote repository, return the repo name if it exists.
// If the repo does not exist, a nil repo is returned.
func (g *githubProvider) GetRepo(ctx context.Context) (*git.Repository, error) {
	r := g.config.Repository
	o := g.config.Owner
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

func (g *githubProvider) AddDeployKeyToRepo(ctx context.Context, opts git.AddDeployKeyOpts) error {
	return g.githubProviderClient.AddDeployKeyToRepo(ctx, opts)
}

// validates the github setup and access.
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
	if g.config.Personal {
		if !strings.EqualFold(g.config.Owner, *user.Login) {
			return fmt.Errorf("the authenticated Github.com user and owner %s specified in the EKS-A gitops spec don't match; confirm access token owner is %s", g.config.Owner, g.config.Owner)
		}
		return nil
	}
	org, err := g.githubProviderClient.Organization(ctx, g.config.Owner)
	if err != nil {
		return fmt.Errorf("the authenticated github user doesn't have proper access to github organization %s, %v", g.config.Owner, err)
	}
	if org == nil { // for now only checks if user belongs to the org
		return fmt.Errorf("the authenticated github user doesn't have proper access to github organization %s", g.config.Owner)
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

func RepoUrl(owner string, repo string) string {
	return fmt.Sprintf(githubUrlTemplate, owner, repo)
}
