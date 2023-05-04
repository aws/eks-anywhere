package github_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	goGithub "github.com/google/go-github/v35/github"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
	"github.com/aws/eks-anywhere/pkg/git/providers/github/mocks"
)

var validPATValues = []string{"ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "github_pat_abcdefghijklmnopqrstuv_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456"}

func TestValidate(t *testing.T) {
	tests := []struct {
		testName          string
		owner             string
		repository        string
		personal          bool
		authenticatedUser string
		allPATPermissions string
		wantErr           error
	}{
		{
			testName:          "good personal repo",
			owner:             "Jeff",
			repository:        "testRepo",
			personal:          true,
			authenticatedUser: "Jeff",
			allPATPermissions: "repo, notrepo, admin",
		},
		{
			testName:          "good organization repo",
			owner:             "orgA",
			repository:        "testRepo",
			personal:          false,
			authenticatedUser: "Jeff",
			allPATPermissions: "repo, notrepo, admin",
		},
		{
			testName:          "user specified wrong owner in spec for a personal repo",
			owner:             "nobody",
			repository:        "testRepo",
			personal:          true,
			authenticatedUser: "Jeff",
			allPATPermissions: "repo, notrepo, admin",
			wantErr:           fmt.Errorf("the authenticated Github.com user and owner %s specified in the EKS-A gitops spec don't match; confirm access token owner is %s", "nobody", "nobody"),
		},
		{
			testName:          "user doesn't belong to the organization or wrong organization",
			owner:             "hiddenOrg",
			repository:        "testRepo",
			personal:          false,
			authenticatedUser: "Jeff",
			allPATPermissions: "repo, notrepo, admin",
			wantErr:           fmt.Errorf("the authenticated github user doesn't have proper access to github organization %s", "hiddenOrg"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			validPATValue := validPATValues[rand.Intn(len(validPATValues))]

			ctx := context.Background()
			githubproviderclient := mocks.NewMockGithubClient(mockCtrl)
			authenticatedUser := &goGithub.User{Login: &tt.authenticatedUser}
			githubproviderclient.EXPECT().AuthenticatedUser(ctx).Return(authenticatedUser, nil)
			githubproviderclient.EXPECT().GetAccessTokenPermissions(validPATValue).Return(tt.allPATPermissions, nil)
			githubproviderclient.EXPECT().CheckAccessTokenPermissions("repo", tt.allPATPermissions).Return(nil)

			auth := git.TokenAuth{Token: validPATValue, Username: tt.owner}

			config := &v1alpha1.GithubProviderConfig{
				Owner:      tt.owner,
				Repository: tt.repository,
				Personal:   tt.personal,
			}

			githubProvider, err := github.New(githubproviderclient, config, auth)
			if err != nil {
				t.Errorf("instantiating github provider: %v, wanted nil", err)
			}

			if !tt.personal {
				if tt.wantErr == nil {
					githubproviderclient.EXPECT().Organization(ctx, tt.owner).Return(&goGithub.Organization{Login: &tt.owner}, nil)
				} else {
					githubproviderclient.EXPECT().Organization(ctx, tt.owner).Return(nil, nil)
				}
			}

			err = githubProvider.Validate(ctx)

			if !reflect.DeepEqual(tt.wantErr, err) {
				t.Errorf("%v got = %v, want %v", tt.testName, err, tt.wantErr)
			}
		})
	}
}

func setupContext(t *testing.T) {
	validPATValue := validPATValues[rand.Intn(len(validPATValues))]
	t.Setenv(github.EksaGithubTokenEnv, validPATValue)
	t.Setenv(github.GithubTokenEnv, validPATValue)
}

func TestIsGithubAccessTokenValidWithEnv(t *testing.T) {
	setupContext(t)

	tests := []struct {
		testName string
	}{
		{
			testName: "no token path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := github.GetGithubAccessTokenFromEnv()
			if err != nil {
				t.Errorf("github.GetGithubAccessTokenFromEnv returned an error, wanted none; %s", err)
			}
		})
	}
}

func TestGetRepoSucceeds(t *testing.T) {
	tests := []struct {
		testName    string
		owner       string
		repository  string
		gitProvider string
		personal    bool
	}{
		{
			testName:    "personal repo succeeds",
			owner:       "Jeff",
			repository:  "testRepo",
			gitProvider: github.GitProviderName,
			personal:    true,
		},
		{
			testName:    "organizational repo succeeds",
			owner:       "Jeff",
			repository:  "testRepo",
			gitProvider: github.GitProviderName,
			personal:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			githubproviderclient := mocks.NewMockGithubClient(mockCtrl)
			getRepoOpts := git.GetRepoOpts{Owner: tt.owner, Repository: tt.repository}
			testRepo := &git.Repository{Name: tt.repository, Owner: tt.owner, Organization: "", CloneUrl: "https://github.com/user/repo"}
			githubproviderclient.EXPECT().GetRepo(context.Background(), getRepoOpts).Return(testRepo, nil)

			config := &v1alpha1.GithubProviderConfig{
				Owner:      tt.owner,
				Repository: tt.repository,
				Personal:   tt.personal,
			}

			validPATValue := validPATValues[rand.Intn(len(validPATValues))]
			auth := git.TokenAuth{Token: validPATValue, Username: tt.owner}
			githubProvider, err := github.New(githubproviderclient, config, auth)
			if err != nil {
				t.Errorf("instantiating github provider: %v, wanted nil", err)
			}
			repo, err := githubProvider.GetRepo(context.Background())
			if err != nil {
				t.Errorf("calling Repo %v, wanted nil", err)
			}
			assert.Equal(t, testRepo, repo)
		})
	}
}

func TestGetNonExistantRepoSucceeds(t *testing.T) {
	tests := []struct {
		testName      string
		owner         string
		repository    string
		authTokenPath string
		gitProvider   string
		personal      bool
	}{
		{
			testName:      "personal repo succeeds",
			owner:         "Jeff",
			repository:    "testRepo",
			authTokenPath: "",
			gitProvider:   github.GitProviderName,
			personal:      true,
		},
		{
			testName:      "organizational repo succeeds",
			owner:         "Jeff",
			repository:    "testRepo",
			authTokenPath: "",
			gitProvider:   github.GitProviderName,
			personal:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			githubproviderclient := mocks.NewMockGithubClient(mockCtrl)
			getRepoOpts := git.GetRepoOpts{Owner: tt.owner, Repository: tt.repository}
			githubproviderclient.EXPECT().GetRepo(context.Background(), getRepoOpts).Return(nil, &git.RepositoryDoesNotExistError{})

			config := &v1alpha1.GithubProviderConfig{
				Owner:      tt.owner,
				Repository: tt.repository,
				Personal:   tt.personal,
			}

			validPATValue := validPATValues[rand.Intn(len(validPATValues))]
			auth := git.TokenAuth{Token: validPATValue, Username: tt.owner}
			githubProvider, err := github.New(githubproviderclient, config, auth)
			if err != nil {
				t.Errorf("instantiating github provider: %v, wanted nil", err)
			}
			repo, err := githubProvider.GetRepo(context.Background())
			if err != nil {
				t.Errorf("calling Repo %v, wanted nil", err)
			}
			var nilRepo *git.Repository
			assert.Equal(t, nilRepo, repo)
		})
	}
}
