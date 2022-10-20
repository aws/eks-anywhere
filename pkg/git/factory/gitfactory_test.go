package gitfactory_test

import (
	"context"
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
		opt          gitFactory.GitToolsOpt
	}{
		{
			testName:     "valid token var",
			authTokenEnv: validPATValue,
		},
		{
			testName:     "valid token var with opt",
			authTokenEnv: validPATValue,
			opt:          gitFactory.WithRepositoryDirectory("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			setupContext(t)

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

			_, err := gitFactory.Build(context.Background(), cluster, fluxConfig, w, tt.opt)
			if err != nil {
				t.Errorf("gitfactory.BuldProvider returned err, wanted nil. err: %v", err)
			}
		})
	}
}

func setupContext(t *testing.T) {
	t.Setenv(github.EksaGithubTokenEnv, validPATValue)
	t.Setenv(github.GithubTokenEnv, validPATValue)
}
