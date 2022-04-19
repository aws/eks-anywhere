package gogithub_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v35/github"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	mockGoGithub "github.com/aws/eks-anywhere/pkg/git/gogithub/mocks"
)

const repoPermissions = "repo"

func TestGoGithubCreateRepo(t *testing.T) {
	type fields struct {
		opts gogithub.Options
	}
	type args struct {
		opts git.CreateRepoOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "create public repo with organizational owner",
			fields: fields{
				opts: gogithub.Options{},
			},
			args: args{
				opts: git.CreateRepoOpts{
					Name:        "testrepo",
					Description: "unit test repo",
					Owner:       "testorganization",
					Personal:    false,
				},
			},
		},
		{
			name: "create personal repo",
			fields: fields{
				opts: gogithub.Options{},
			},
			args: args{
				opts: git.CreateRepoOpts{
					Name:        "testrepo",
					Description: "unit test repo",
					Owner:       "testuser",
					Personal:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockCtrl := gomock.NewController(t)
			client := mockGoGithub.NewMockClient(mockCtrl)

			returnRepo := &github.Repository{
				Name:     &tt.args.opts.Name,
				CloneURL: &tt.args.opts.Description,
				Owner: &github.User{
					Name: &tt.args.opts.Owner,
				},
				Organization: &github.Organization{
					Name: &tt.args.opts.Name,
				},
			}
			client.EXPECT().CreateRepo(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(arg0 context.Context, org string, arg2 *github.Repository) {
				returnRepo.Organization.Name = &org
			}).Return(
				returnRepo, nil, nil)

			g := &gogithub.GoGithub{
				Opts:   tt.fields.opts,
				Client: client,
			}
			gotRepository, err := g.CreateRepo(ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantRepository := &git.Repository{
				Name:         returnRepo.GetName(),
				Organization: returnRepo.GetOrganization().GetName(),
				CloneUrl:     returnRepo.GetCloneURL(),
				Owner:        returnRepo.GetOwner().GetName(),
			}
			if tt.args.opts.Personal && gotRepository.Organization != "" {
				t.Errorf("for personal account org should be empty")
			}

			if !reflect.DeepEqual(gotRepository, wantRepository) {
				t.Errorf("CreateRepo() gotRepository = %v, want %v", gotRepository, wantRepository)
			}
		})
	}
}

func TestGoGithubGetRepo(t *testing.T) {
	type fields struct {
		opts gogithub.Options
	}
	type args struct {
		opts     git.GetRepoOpts
		name     string
		cloneURL string
		orgName  string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		throwError error
		matchError error
	}{
		{
			name: "Repo no error",
			args: args{
				opts: git.GetRepoOpts{
					Owner:      "owner1",
					Repository: "repo1",
				},
				cloneURL: "url1",
				name:     "repo1",
				orgName:  "org1",
			},
			wantErr: false,
		},
		{
			name: "github client threw generic error",
			args: args{
				opts: git.GetRepoOpts{
					Owner:      "owner1",
					Repository: "repo1",
				},
				cloneURL: "url1",
				name:     "repo1",
				orgName:  "org1",
			},
			wantErr:    true,
			throwError: fmt.Errorf("github client threw error"),
			matchError: fmt.Errorf("unexpected error when describing repository %s: %w", "repo1", fmt.Errorf("github client threw error")),
		},
		{
			name: "github threw 404 error",
			args: args{
				opts: git.GetRepoOpts{
					Owner:      "owner1",
					Repository: "repo1",
				},
				cloneURL: "url1",
				name:     "repo1",
				orgName:  "org1",
			},
			wantErr:    true,
			throwError: notFoundError(),
			matchError: &git.RepositoryDoesNotExistError{Err: fmt.Errorf("GET : 404 Not found []")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockCtrl := gomock.NewController(t)
			client := mockGoGithub.NewMockClient(mockCtrl)

			returnRepo := &github.Repository{
				Name:     &tt.args.name,
				CloneURL: &tt.args.cloneURL,
				Owner: &github.User{
					Name: &tt.args.opts.Owner,
				},
				Organization: &github.Organization{
					Name: &tt.args.orgName,
				},
			}

			client.EXPECT().Repo(ctx, tt.args.opts.Owner, tt.args.opts.Repository).
				Return(returnRepo, nil, tt.throwError)

			g := &gogithub.GoGithub{
				Opts:   tt.fields.opts,
				Client: client,
			}
			got, err := g.GetRepo(ctx, tt.args.opts)

			wantRepository := &git.Repository{
				Name:         tt.args.name,
				Organization: tt.args.orgName,
				CloneUrl:     tt.args.cloneURL,
				Owner:        tt.args.opts.Owner,
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Repo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				switch tt.matchError.(type) {
				case *git.RepositoryDoesNotExistError:
					_, typeMatches := err.(*git.RepositoryDoesNotExistError)
					if !typeMatches || err.Error() != tt.matchError.Error() {
						t.Errorf("Repo() error = %v, matchError %v", err, tt.matchError)
					}
				default:
					if !reflect.DeepEqual(err, tt.matchError) {
						t.Errorf("Repo() error = %v, matchError %v", err, tt.matchError)
					}
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(got, wantRepository) {
				t.Errorf("Repo() got = %v, want %v", got, wantRepository)
			}
		})
	}
}

func TestGoGithubDeleteRepoSuccess(t *testing.T) {
	type fields struct {
		opts gogithub.Options
	}

	tests := []struct {
		name    string
		fields  fields
		args    git.DeleteRepoOpts
		wantErr error
	}{
		{
			name: "github repo deleted successfully",
			args: git.DeleteRepoOpts{
				Owner:      "owner1",
				Repository: "repo1",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			client := mockGoGithub.NewMockClient(mockCtrl)

			client.EXPECT().DeleteRepo(ctx, tt.args.Owner, tt.args.Repository).Return(nil, tt.wantErr)

			g := &gogithub.GoGithub{
				Opts:   tt.fields.opts,
				Client: client,
			}
			err := g.DeleteRepo(ctx, tt.args)
			if err != tt.wantErr {
				t.Errorf("DeleteRepo() got error: %v want error: %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoGithubDeleteRepoFail(t *testing.T) {
	type fields struct {
		opts gogithub.Options
	}

	tests := []struct {
		name     string
		fields   fields
		args     git.DeleteRepoOpts
		wantErr  error
		throwErr error
	}{
		{
			name: "github repo delete fail",
			args: git.DeleteRepoOpts{
				Owner:      "owner1",
				Repository: "repo1",
			},
			wantErr:  fmt.Errorf("deleting repository repo1: github client threw error"),
			throwErr: fmt.Errorf("github client threw error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			client := mockGoGithub.NewMockClient(mockCtrl)

			client.EXPECT().DeleteRepo(ctx, tt.args.Owner, tt.args.Repository).Return(nil, tt.throwErr)

			g := &gogithub.GoGithub{
				Opts:   tt.fields.opts,
				Client: client,
			}
			err := g.DeleteRepo(ctx, tt.args)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("DeleteRepo() got error: %v want error: %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoGithub_CheckAccessTokenPermissions(t *testing.T) {
	type fields struct {
		opts gogithub.Options
	}
	tests := []struct {
		name           string
		allPermissions string
		fields         fields
		wantErr        error
	}{
		{
			name:           "token with repo permissions",
			allPermissions: "admin, repo",
			fields: fields{
				opts: gogithub.Options{},
			},
			wantErr: nil,
		},
		{
			name:           "token without repo permissions",
			allPermissions: "admin, workflow",
			fields: fields{
				opts: gogithub.Options{},
			},
			wantErr: errors.New("github access token does not have repo permissions"),
		},
		{
			name:           "token with repo permissions",
			allPermissions: "",
			fields: fields{
				opts: gogithub.Options{},
			},
			wantErr: errors.New("github access token does not have repo permissions"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			client := mockGoGithub.NewMockClient(mockCtrl)

			g := &gogithub.GoGithub{
				Opts:   tt.fields.opts,
				Client: client,
			}

			gotError := g.CheckAccessTokenPermissions(repoPermissions, tt.allPermissions)
			if !reflect.DeepEqual(gotError, tt.wantErr) {
				t.Errorf("Test %v\n got %v\n want %v", tt.name, gotError, tt.wantErr)
			}
		})
	}
}

func TestPathExistsError(t *testing.T) {
	tt := newTest(t)
	owner, repo, branch, path := pathArgs()
	tt.client.EXPECT().GetContents(
		tt.ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch},
	).Return(nil, nil, nil, errors.New("can't get content"))

	_, err := tt.g.PathExists(tt.ctx, owner, repo, branch, path)
	tt.Expect(err).To(HaveOccurred())
}

func TestPathExistsItDoes(t *testing.T) {
	tt := newTest(t)
	owner, repo, branch, path := pathArgs()
	tt.client.EXPECT().GetContents(
		tt.ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch},
	).Return(nil, nil, nil, notFoundError())

	tt.Expect(tt.g.PathExists(tt.ctx, owner, repo, branch, path)).To(BeFalse())
}

func TestPathExistsItDoesNot(t *testing.T) {
	tt := newTest(t)
	owner, repo, branch, path := pathArgs()
	tt.client.EXPECT().GetContents(
		tt.ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch},
	).Return(nil, nil, nil, nil)

	tt.Expect(tt.g.PathExists(tt.ctx, owner, repo, branch, path)).To(BeTrue())
}

type gogithubTest struct {
	*WithT
	g      *gogithub.GoGithub
	opts   gogithub.Options
	client *mockGoGithub.MockClient
	ctx    context.Context
}

func newTest(t *testing.T) *gogithubTest {
	withT := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mockGoGithub.NewMockClient(ctrl)
	opts := gogithub.Options{
		Auth: git.TokenAuth{
			Username: "user",
			Token:    "token",
		},
	}

	g := &gogithub.GoGithub{
		Opts:   opts,
		Client: client,
	}

	return &gogithubTest{
		WithT:  withT,
		g:      g,
		opts:   opts,
		client: client,
		ctx:    ctx,
	}
}

func pathArgs() (owner, repo, branch, path string) {
	path = "fluxFolder"
	branch = "main"
	owner = "aws"
	repo = "eksa-gitops"

	return owner, repo, path, branch
}

func notFoundError() error {
	return &github.ErrorResponse{
		Message: "Not found",
		Response: &http.Response{
			StatusCode: http.StatusNotFound,
			Request: &http.Request{
				Method: "GET",
				URL:    &url.URL{},
			},
		},
	}
}
