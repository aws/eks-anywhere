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
			gitProviderConfig := v1alpha1.Github{Owner: "Jeff"}
			fluxConfig := v1alpha1.Flux{Github: gitProviderConfig}
			gitopsConfig := &v1alpha1.GitOpsConfigSpec{Flux: fluxConfig}

			githubProviderClient := githubMocks.NewMockGitProviderClient(mockCtrl)
			githubProviderClient.EXPECT().SetTokenAuth(gomock.Any(), fluxConfig.Github.Owner)
			opts := gitFactory.Options{GithubGitClient: githubProviderClient}
			factory := gitFactory.New(opts)

			_, err := factory.BuildProvider(context.Background(), gitopsConfig)
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
