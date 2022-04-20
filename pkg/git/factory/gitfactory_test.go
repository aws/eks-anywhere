package gitfactory_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
	githubMocks "github.com/aws/eks-anywhere/pkg/git/providers/github/mocks"
)

const (
	validPATValue = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func TestGitFactoryHappyPath(t *testing.T) {
	tests := []struct {
		testName     string
		authTokenEnv string
	}{
		{
			testName:     "valid token var",
			authTokenEnv: validPATValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var tctx testContext
			tctx.SaveContext(validPATValue)
			defer tctx.RestoreContext()

			mockCtrl := gomock.NewController(t)

			gitProviderConfig := v1alpha1.GithubProviderConfig{
				Owner:      "Jeff",
				Repository: "testRepo",
				Personal:   true,
			}

			fluxConfig := v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &gitProviderConfig,
				},
			}

			githubProviderClient := githubMocks.NewMockGitProviderClient(mockCtrl)
			githubProviderClient.EXPECT().SetTokenAuth(gomock.Any(), fluxConfig.Spec.Github.Owner)
			opts := gitFactory.Options{GithubGitClient: githubProviderClient}
			factory := gitFactory.New(opts)

			_, err := factory.BuildProvider(context.Background(), &fluxConfig.Spec)
			if err != nil {
				t.Errorf("gitfactory.BuldProvider returned err, wanted nil. err: %v", err)
			}
		})
	}
}

type testContext struct {
	oldGithubToken   string
	isGithubTokenSet bool
}

func (tctx *testContext) SaveContext(token string) {
	tctx.oldGithubToken, tctx.isGithubTokenSet = os.LookupEnv(github.EksaGithubTokenEnv)
	os.Setenv(github.EksaGithubTokenEnv, validPATValue)
	os.Setenv(github.GithubTokenEnv, validPATValue)
}

func (tctx *testContext) RestoreContext() {
	if tctx.isGithubTokenSet {
		os.Setenv(github.EksaGithubTokenEnv, tctx.oldGithubToken)
	} else {
		os.Unsetenv(github.EksaGithubTokenEnv)
	}
}
