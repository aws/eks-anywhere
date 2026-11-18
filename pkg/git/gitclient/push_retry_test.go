package gitclient_test

import (
	"context"
	"errors"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"go.uber.org/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/git/gitclient"
	mockGitClient "github.com/aws/eks-anywhere/pkg/git/gitclient/mocks"
)

func TestGoGitPushRecoversGitHubReceivePackFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mockGitClient.NewMockGoGit(ctrl)
	server := (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "Internal Server Error"}).Error()
	repository := &gogit.Repository{}
	client.EXPECT().OpenDir(repoDir).Return(repository, nil)
	gomock.InOrder(
		client.EXPECT().PushWithContext(gomock.Any(), repository, nil).Return(server),
		client.EXPECT().PushWithContext(gomock.Any(), repository, nil).Return(gogit.NoErrAlreadyUpToDate),
	)
	g := &gitclient.GitClient{RepoDirectory: repoDir, RepoUrl: "https://github.com/testuser/testrepo.git", Client: client}
	if err := g.Push(context.Background()); err != nil {
		t.Fatalf("push retry failed: %v", err)
	}
}

func TestGoGitPushPreservesPermanentCause(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mockGitClient.NewMockGoGit(ctrl)
	want := errors.New("permission denied")
	client.EXPECT().OpenDir(repoDir).Return(&gogit.Repository{}, nil)
	client.EXPECT().PushWithContext(gomock.Any(), gomock.Any(), nil).Return(want)
	g := &gitclient.GitClient{RepoDirectory: repoDir, RepoUrl: "https://github.com/testuser/testrepo.git", Client: client}
	if err := g.Push(context.Background()); !errors.Is(err, want) {
		t.Fatalf("permanent cause changed: %v", err)
	}
}
