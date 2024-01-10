package e2e

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

func (e *E2ESession) TestGithubClient(ctx context.Context, githubToken, owner, repository string, personal bool) (git.ProviderClient, error) {
	auth := git.TokenAuth{Token: githubToken, Username: owner}
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)

	config := &v1alpha1.GithubProviderConfig{
		Owner:      owner,
		Repository: repository,
		Personal:   personal,
	}
	provider, err := github.New(githubProviderClient, config, auth)
	if err != nil {
		return nil, fmt.Errorf("creating test git provider: %v", err)
	}

	return provider, nil
}
