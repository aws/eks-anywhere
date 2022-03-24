package gogithub

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	goGithub "github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type GoGithub struct {
	Opts   Options
	Client Client
}

type Options struct {
	Auth git.TokenAuth
}

func New(ctx context.Context, opts Options) *GoGithub {
	return &GoGithub{
		Opts:   opts,
		Client: newClient(ctx, opts),
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client interface {
	CreateRepo(ctx context.Context, org string, repo *goGithub.Repository) (*goGithub.Repository, *goGithub.Response, error)
	Repo(ctx context.Context, owner, repo string) (*goGithub.Repository, *goGithub.Response, error)
	User(ctx context.Context, user string) (*goGithub.User, *goGithub.Response, error)
	Organization(ctx context.Context, org string) (*goGithub.Organization, *goGithub.Response, error)
	GetContents(ctx context.Context, owner, repo, path string, opt *goGithub.RepositoryContentGetOptions) (
		fileContent *goGithub.RepositoryContent, directoryContent []*goGithub.RepositoryContent, resp *goGithub.Response, err error,
	)
	DeleteRepo(ctx context.Context, owner, repo string) (*goGithub.Response, error)
}

type githubClient struct {
	client *goGithub.Client
}

var HttpClient HTTPClient

func init() {
	HttpClient = &http.Client{}
}

func (ggc *githubClient) CreateRepo(ctx context.Context, org string, repo *goGithub.Repository) (*goGithub.Repository, *goGithub.Response, error) {
	return ggc.client.Repositories.Create(ctx, org, repo)
}

func (ggc *githubClient) Repo(ctx context.Context, owner, repo string) (*goGithub.Repository, *goGithub.Response, error) {
	return ggc.client.Repositories.Get(ctx, owner, repo)
}

func (ggc *githubClient) User(ctx context.Context, user string) (*goGithub.User, *goGithub.Response, error) {
	return ggc.client.Users.Get(ctx, user)
}

func (ggc *githubClient) Organization(ctx context.Context, org string) (*goGithub.Organization, *goGithub.Response, error) {
	return ggc.client.Organizations.Get(ctx, org)
}

func (ggc *githubClient) GetContents(ctx context.Context, owner, repo, path string, opt *goGithub.RepositoryContentGetOptions) (fileContent *goGithub.RepositoryContent, directoryContent []*goGithub.RepositoryContent, resp *goGithub.Response, err error) {
	return ggc.client.Repositories.GetContents(ctx, owner, repo, path, opt)
}

func (ggc *githubClient) DeleteRepo(ctx context.Context, owner, repo string) (*goGithub.Response, error) {
	return ggc.client.Repositories.Delete(ctx, owner, repo)
}

// CreateRepo creates an empty Github Repository. The repository must be initialized locally or
// file must be added to it via the github api before it can be successfully cloned.
func (g *GoGithub) CreateRepo(ctx context.Context, opts git.CreateRepoOpts) (repository *git.Repository, err error) {
	logger.V(3).Info("Attempting to create new Github repo", "repo", opts.Name, "owner", opts.Owner)
	r := &goGithub.Repository{Name: &opts.Name, Private: &opts.Privacy, Description: &opts.Description}

	org := ""
	if !opts.Personal {
		org = opts.Owner
		logger.V(4).Info("Not a personal repository; using repository Owner as Org", "org", org, "owner", opts.Owner)
	}

	repo, _, err := g.Client.CreateRepo(ctx, org, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Github repo %s: %v", opts.Name, err)
	}
	logger.V(3).Info("Successfully created new Github repo", "repo", repo.GetName(), "owner", opts.Owner)
	return &git.Repository{
		Name:         repo.GetName(),
		CloneUrl:     repo.GetCloneURL(),
		Owner:        repo.GetOwner().GetName(),
		Organization: repo.GetOrganization().GetName(),
	}, err
}

func (g *GoGithub) GetAccessTokenPermissions(accessToken string) (string, error) {
	req, err := http.NewRequest("HEAD", "https://api.github.com/users/codertocat", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)

	var resp *http.Response
	r := retrier.New(3 * time.Minute)
	err = r.Retry(func() error {
		resp, err = HttpClient.Do(req)
		if err != nil {
			return fmt.Errorf("getting Github Personal Access Token permissions %v", err)
		}
		return nil
	})

	permissionsScopes := resp.Header.Get("X-Oauth-Scopes")
	defer resp.Body.Close()

	return permissionsScopes, nil
}

func (g *GoGithub) CheckAccessTokenPermissions(checkPATPermission string, allPermissionScopes string) error {
	logger.Info("Checking Github Access Token permissions")

	allPermissions := strings.Split(allPermissionScopes, ", ")
	for _, permission := range allPermissions {
		if permission == checkPATPermission {
			return nil
		}
	}
	return errors.New("github access token does not have repo permissions")
}

// GetRepo describes a remote repository, return the repo name if it exists.
// If the repo does not exist, resulting in a 404 exception, it returns a `RepoDoesNotExist` error.
func (g *GoGithub) GetRepo(ctx context.Context, opts git.GetRepoOpts) (*git.Repository, error) {
	r := opts.Repository
	o := opts.Owner
	logger.V(3).Info("Describing Github repository", "name", r, "owner", o)
	repo, _, err := g.Client.Repo(ctx, o, r)
	if err != nil {
		if isNotFound(err) {
			return nil, &git.RepositoryDoesNotExistError{Err: err}
		}
		return nil, fmt.Errorf("unexpected error when describing repository %s: %w", r, err)
	}
	return &git.Repository{
		Name:         repo.GetName(),
		CloneUrl:     repo.GetCloneURL(),
		Owner:        repo.GetOwner().GetName(),
		Organization: repo.GetOrganization().GetName(),
	}, err
}

func (g *GoGithub) AuthenticatedUser(ctx context.Context) (*goGithub.User, error) {
	githubUser, _, err := g.Client.User(ctx, "") // passing the empty string will fetch the authenticated
	if err != nil {
		return nil, fmt.Errorf("failed while getting the authenticated github user %v", err)
	}
	return githubUser, nil
}

func (g *GoGithub) Organization(ctx context.Context, org string) (*goGithub.Organization, error) {
	organization, _, err := g.Client.Organization(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed while getting github organization %s details %v", org, err)
	}
	return organization, nil
}

// PathExists checks if a path exists in the remote repository. If the owner, repository or branch doesn't exist,
// it returns false and no error
func (g *GoGithub) PathExists(ctx context.Context, owner, repo, branch, path string) (bool, error) {
	_, _, _, err := g.Client.GetContents(
		ctx,
		owner,
		repo,
		path,
		&goGithub.RepositoryContentGetOptions{Ref: branch},
	)

	if isNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed checking if path %s exists in remote github repository: %v", path, err)
	}

	return true, nil
}

// DeleteRepo deletes a Github repository.
func (g *GoGithub) DeleteRepo(ctx context.Context, opts git.DeleteRepoOpts) error {
	r := opts.Repository
	o := opts.Owner
	logger.V(3).Info("Deleting Github repository", "name", r, "owner", o)
	_, err := g.Client.DeleteRepo(ctx, o, r)
	if err != nil {
		return fmt.Errorf("deleting repository %s: %v", r, err)
	}
	return nil
}

func newClient(ctx context.Context, opts Options) Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Auth.Token})
	tc := oauth2.NewClient(ctx, ts)
	return &githubClient{goGithub.NewClient(tc)}
}

func isNotFound(err error) bool {
	var e *goGithub.ErrorResponse
	return errors.As(err, &e) && e.Response.StatusCode == http.StatusNotFound
}
