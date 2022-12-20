package flux

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/git"
	gitFactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/git/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type gitClientTest struct {
	*WithT
	ctx context.Context
	c   *gitClient
	g   *mocks.MockClient
	p   *mocks.MockProviderClient
}

func newGitClientTest(t *testing.T) *gitClientTest {
	ctrl := gomock.NewController(t)
	g := mocks.NewMockClient(ctrl)
	p := mocks.NewMockProviderClient(ctrl)
	tool := &gitFactory.GitTools{
		Provider: p,
		Client:   g,
	}
	c := newGitClient(tool)
	c.Retrier = retrier.NewWithMaxRetries(maxRetries, 0)
	return &gitClientTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		c:     c,
		g:     g,
		p:     p,
	}
}

func TestGitClientGetRepoSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.p.EXPECT().GetRepo(tt.ctx).Return(nil, errors.New("error in get repo")).Times(4)
	tt.p.EXPECT().GetRepo(tt.ctx).Return(nil, nil).Times(1)

	_, err := tt.c.GetRepo(tt.ctx)
	tt.Expect(err).To(Succeed(), "gitClient.GetRepo() should succeed with 5 tries")
}

func TestGitClientGetRepoSkip(t *testing.T) {
	tt := newGitClientTest(t)

	c := newGitClient(&gitFactory.GitTools{Provider: nil, Client: tt.g})
	_, err := c.GetRepo(tt.ctx)
	tt.Expect(err).To(Succeed())
}

func TestGitClientGetRepoError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.p.EXPECT().GetRepo(tt.ctx).Return(nil, errors.New("error in get repo")).Times(5)
	tt.p.EXPECT().GetRepo(tt.ctx).Return(nil, nil).AnyTimes()

	_, err := tt.c.GetRepo(tt.ctx)
	tt.Expect(err).To(MatchError(ContainSubstring("error in get repo")), "gitClient.GetRepo() should fail after 5 tries")
}

func TestGitClientCreateRepoSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	opts := git.CreateRepoOpts{}
	tt.p.EXPECT().CreateRepo(tt.ctx, opts).Return(nil, errors.New("error in create repo")).Times(4)
	tt.p.EXPECT().CreateRepo(tt.ctx, opts).Return(nil, nil).Times(1)

	tt.Expect(tt.c.CreateRepo(tt.ctx, opts)).To(Succeed(), "gitClient.CreateRepo() should succeed with 5 tries")
}

func TestGitClientCreateRepoSkip(t *testing.T) {
	tt := newGitClientTest(t)
	opts := git.CreateRepoOpts{}
	c := newGitClient(&gitFactory.GitTools{Provider: nil, Client: tt.g})

	tt.Expect(c.CreateRepo(tt.ctx, opts)).To(Succeed())
}

func TestGitClientCreateRepoError(t *testing.T) {
	tt := newGitClientTest(t)
	opts := git.CreateRepoOpts{}
	tt.p.EXPECT().CreateRepo(tt.ctx, opts).Return(nil, errors.New("error in create repo")).Times(5)
	tt.p.EXPECT().CreateRepo(tt.ctx, opts).Return(nil, nil).AnyTimes()

	tt.Expect(tt.c.CreateRepo(tt.ctx, opts)).To(MatchError(ContainSubstring("error in create repo")), "gitClient.CreateRepo() should fail after 5 tries")
}

func TestGitClientCloneSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Clone(tt.ctx).Return(errors.New("error in clone repo")).Times(4)
	tt.g.EXPECT().Clone(tt.ctx).Return(nil).Times(1)

	tt.Expect(tt.c.Clone(tt.ctx)).To(Succeed(), "gitClient.Clone() should succeed with 5 tries")
}

func TestGitClientCloneError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Clone(tt.ctx).Return(errors.New("error in clone repo")).Times(5)
	tt.g.EXPECT().Clone(tt.ctx).Return(nil).AnyTimes()

	tt.Expect(tt.c.Clone(tt.ctx)).To(MatchError(ContainSubstring("error in clone repo")), "gitClient.Clone() should fail after 5 tries")
}

func TestGitClientPushSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Push(tt.ctx).Return(errors.New("error in push repo")).Times(4)
	tt.g.EXPECT().Push(tt.ctx).Return(nil).Times(1)

	tt.Expect(tt.c.Push(tt.ctx)).To(Succeed(), "gitClient.Push() should succeed with 5 tries")
}

func TestGitClientPushError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Push(tt.ctx).Return(errors.New("error in push repo")).Times(5)
	tt.g.EXPECT().Push(tt.ctx).Return(nil).AnyTimes()

	tt.Expect(tt.c.Push(tt.ctx)).To(MatchError(ContainSubstring("error in push repo")), "gitClient.Push() should fail after 5 tries")
}

func TestGitClientPullSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Pull(tt.ctx, "").Return(errors.New("error in pull repo")).Times(4)
	tt.g.EXPECT().Pull(tt.ctx, "").Return(nil).Times(1)

	tt.Expect(tt.c.Pull(tt.ctx, "")).To(Succeed(), "gitClient.Pull() should succeed with 5 tries")
}

func TestGitClientPullError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Pull(tt.ctx, "").Return(errors.New("error in pull repo")).Times(5)
	tt.g.EXPECT().Pull(tt.ctx, "").Return(nil).AnyTimes()

	tt.Expect(tt.c.Pull(tt.ctx, "")).To(MatchError(ContainSubstring("error in pull repo")), "gitClient.Pull() should fail after 5 tries")
}

func TestGitClientPathExistsSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.p.EXPECT().PathExists(tt.ctx, "", "", "", "").Return(false, errors.New("error in get repo")).Times(4)
	tt.p.EXPECT().PathExists(tt.ctx, "", "", "", "").Return(true, nil).Times(1)

	exists, err := tt.c.PathExists(tt.ctx, "", "", "", "")
	tt.Expect(exists).To(Equal(true))
	tt.Expect(err).To(Succeed(), "gitClient.PathExists() should succeed with 5 tries")
}

func TestGitClientPathExistsSkip(t *testing.T) {
	tt := newGitClientTest(t)

	c := newGitClient(&gitFactory.GitTools{Provider: nil, Client: tt.g})
	exists, err := c.PathExists(tt.ctx, "", "", "", "")
	tt.Expect(exists).To(Equal(false))
	tt.Expect(err).To(Succeed())
}

func TestGitClientPathExistsError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.p.EXPECT().PathExists(tt.ctx, "", "", "", "").Return(false, errors.New("error in get repo")).Times(5)
	tt.p.EXPECT().PathExists(tt.ctx, "", "", "", "").Return(true, nil).AnyTimes()

	exists, err := tt.c.PathExists(tt.ctx, "", "", "", "")
	tt.Expect(exists).To(Equal(false))
	tt.Expect(err).To(MatchError(ContainSubstring("error in get repo")), "gitClient.PathExists() should fail after 5 tries")
}

func TestGitClientAddSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Add("").Return(nil)

	tt.Expect(tt.c.Add("")).To(Succeed(), "gitClient.Add() should succeed with 1 try")
}

func TestGitClientAddError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Add("").Return(errors.New("error in add"))

	tt.Expect(tt.c.Add("")).To(MatchError(ContainSubstring("error in add")), "gitClient.Add() should fail after 1 try")
}

func TestGitClientRemoveSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Remove("").Return(nil)

	tt.Expect(tt.c.Remove("")).To(Succeed(), "gitClient.Remove() should succeed with 1 try")
}

func TestGitClientRemoveError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Remove("").Return(errors.New("error in remove"))

	tt.Expect(tt.c.Remove("")).To(MatchError(ContainSubstring("error in remove")), "gitClient.Remove() should fail after 1 try")
}

func TestGitClientCommitSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Commit("").Return(nil)

	tt.Expect(tt.c.Commit("")).To(Succeed(), "gitClient.Commit() should succeed with 1 try")
}

func TestGitClientCommitError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Commit("").Return(errors.New("error in commit"))

	tt.Expect(tt.c.Commit("")).To(MatchError(ContainSubstring("error in commit")), "gitClient.Commit() should fail after 1 try")
}

func TestGitClientBranchSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Branch("").Return(nil)

	tt.Expect(tt.c.Branch("")).To(Succeed(), "gitClient.Branch() should succeed with 1 try")
}

func TestGitClientBranchError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Branch("").Return(errors.New("error in branch"))

	tt.Expect(tt.c.Branch("")).To(MatchError(ContainSubstring("error in branch")), "gitClient.Branch() should fail after 1 try")
}

func TestGitClientInitSuccess(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Init().Return(nil)

	tt.Expect(tt.c.Init()).To(Succeed(), "gitClient.Init() should succeed with 1 try")
}

func TestGitClientInitError(t *testing.T) {
	tt := newGitClientTest(t)
	tt.g.EXPECT().Init().Return(errors.New("error in init"))

	tt.Expect(tt.c.Init()).To(MatchError(ContainSubstring("error in init")), "gitClient.Init() should fail after 1 try")
}
