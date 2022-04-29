package gitfactory_test

import (
	"context"
	"os"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
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

			gitProviderConfig := v1alpha1.GithubProviderConfig{
				Owner:      "Jeff",
				Repository: "testRepo",
				Personal:   true,
			}

			cluster := &v1alpha1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCluster",
				},
			}

			fluxConfig := &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &gitProviderConfig,
				},
			}

			_, w := test.NewWriter(t)

			_, err := gitFactory.Build(context.Background(), cluster, fluxConfig, w)
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
