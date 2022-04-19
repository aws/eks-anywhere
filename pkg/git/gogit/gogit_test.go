package gogit_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogit"
	mockGoGit "github.com/aws/eks-anywhere/pkg/git/gogit/mocks"
)

func TestGoGitClone(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    bool
		throwError error
		matchError error
	}{
		{
			name:    "clone repo success",
			wantErr: false,
		},
		{
			name:       "empty repository error",
			wantErr:    true,
			throwError: fmt.Errorf("remote repository is empty"),
			matchError: &git.RepositoryIsEmptyError{
				Repository: "testrepo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, client, opts := newGoGit(t)
			repoUrl := "testurl"

			g := &gogit.GoGit{
				Opts:   opts,
				Client: client,
			}

			client.EXPECT().Clone(ctx, opts.RepositoryDirectory, repoUrl, opts.Auth).Return(&goGit.Repository{}, tt.throwError)

			err := g.Clone(ctx, repoUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !reflect.DeepEqual(err, tt.matchError) {
					t.Errorf("Clone() error = %v, matchError %v", err, tt.matchError)
				}
			}
		})
	}
}

func TestGoGitAdd(t *testing.T) {
	_, client, opts := newGoGit(t)
	filename := "testfile"

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
	client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(&goGit.Worktree{}, nil)
	client.EXPECT().AddGlob(gomock.Any(), gomock.Any()).Do(func(arg0 string, arg1 *goGit.Worktree) {}).Return(nil)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Add(filename)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}
}

func TestGoGitRemove(t *testing.T) {
	_, client, opts := newGoGit(t)
	filename := "testfile"

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
	client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(&goGit.Worktree{}, nil)
	client.EXPECT().Remove(gomock.Any(), gomock.Any()).Do(func(arg0 string, arg1 *goGit.Worktree) {}).Return(plumbing.Hash{}, nil)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Remove(filename)
	if err != nil {
		t.Errorf("Remove() error = %v", err)
		return
	}
}

func TestGoGitCommit(t *testing.T) {
	_, client, opts := newGoGit(t)
	message := "message"

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
	client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(&goGit.Worktree{}, nil)
	client.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(arg0 string, arg1 *object.Signature, arg2 *goGit.Worktree) {}).Return(plumbing.Hash{}, nil)
	client.EXPECT().CommitObject(gomock.Any(), gomock.Any()).Do(func(arg0 *goGit.Repository, arg1 plumbing.Hash) {}).Return(&object.Commit{}, nil)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Commit(message)
	if err != nil {
		t.Errorf("Commit() error = %v", err)
		return
	}
}

func TestGoGitPush(t *testing.T) {
	ctx, client, opts := newGoGit(t)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
	client.EXPECT().PushWithContext(ctx, gomock.Any(), gomock.Any()).Do(func(arg0 context.Context, arg1 *goGit.Repository, arg2 transport.AuthMethod) {}).Return(nil)

	err := g.Push(ctx)
	if err != nil {
		t.Errorf("Push() error = %v", err)
		return
	}
}

func TestGoGitPull(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    bool
		throwError error
		matchError error
	}{
		{
			name:    "pull success",
			wantErr: false,
		},
		{
			name:       "repo already up-to-date",
			wantErr:    true,
			throwError: fmt.Errorf("already up-to-date"),
			matchError: fmt.Errorf("pulling from remote: %v", goGit.NoErrAlreadyUpToDate),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, client, opts := newGoGit(t)
			branch := "testbranch"

			g := &gogit.GoGit{
				Opts:   opts,
				Client: client,
			}

			client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
			client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(&goGit.Worktree{}, nil)
			client.EXPECT().PullWithContext(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(arg0 context.Context, arg1 *goGit.Worktree, arg2 transport.AuthMethod, name plumbing.ReferenceName) {
			}).Return(tt.throwError)
			if !tt.wantErr {
				client.EXPECT().Head(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(&plumbing.Reference{}, nil)
				client.EXPECT().CommitObject(gomock.Any(), gomock.Any()).Do(func(arg0 *goGit.Repository, arg1 plumbing.Hash) {}).Return(&object.Commit{}, nil)
			}

			err := g.Pull(ctx, branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !reflect.DeepEqual(err, tt.matchError) {
					t.Errorf("Pull() error = %v, matchError %v", err, tt.matchError)
				}
			}
		})
	}
}

func TestGoGitInit(t *testing.T) {
	_, client, opts := newGoGit(t)
	url := "testurl"

	client.EXPECT().Init(opts.RepositoryDirectory).Return(&goGit.Repository{}, nil)
	client.EXPECT().Create(gomock.Any(), url).Do(func(arg0 *goGit.Repository, arg1 string) {}).Return(&goGit.Remote{}, nil)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Init(url)
	if err != nil {
		t.Errorf("Init() error = %v", err)
		return
	}
}

func TestGoGitBranch(t *testing.T) {
	_, client, opts := newGoGit(t)

	repo := &goGit.Repository{}
	headRef := &plumbing.Reference{}
	worktree := &goGit.Worktree{}
	bOpts := &config.Branch{
		Name:   "testBranch",
		Remote: "origin",
		Merge:  "refs/heads/testBranch",
	}
	cOpts := &goGit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("testBranch"),
		Force:  true,
	}

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(repo, nil)
	client.EXPECT().CreateBranch(repo, bOpts).Return(nil)
	client.EXPECT().Head(repo).Return(headRef, nil)
	client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(worktree, nil)
	client.EXPECT().SetRepositoryReference(repo, gomock.Any()).Return(nil)
	client.EXPECT().Checkout(worktree, cOpts).Return(nil)
	client.EXPECT().ListRemotes(repo, gomock.Any()).Return(nil, nil)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Branch("testBranch")
	if err != nil {
		t.Errorf("Branch() error = %v", err)
		return
	}
}

func TestGoGitBranchRemoteExists(t *testing.T) {
	_, client, opts := newGoGit(t)

	repo := &goGit.Repository{}
	headRef := &plumbing.Reference{}
	worktree := &goGit.Worktree{}
	bOpts := &config.Branch{
		Name:   "testBranch",
		Remote: "origin",
		Merge:  "refs/heads/testBranch",
	}
	localBranchRef := plumbing.NewBranchReferenceName("testBranch")
	cOpts := &goGit.CheckoutOptions{
		Branch: localBranchRef,
		Force:  true,
	}

	returnReferences := []*plumbing.Reference{
		plumbing.NewHashReference("refs/heads/testBranch", headRef.Hash()),
	}

	client.EXPECT().OpenDir(opts.RepositoryDirectory).Return(repo, nil)
	client.EXPECT().CreateBranch(repo, bOpts).Return(nil)
	client.EXPECT().Head(repo).Return(headRef, nil)
	client.EXPECT().OpenWorktree(gomock.Any()).Do(func(arg0 *goGit.Repository) {}).Return(worktree, nil)
	client.EXPECT().SetRepositoryReference(repo, gomock.Any()).Return(nil)
	client.EXPECT().Checkout(worktree, cOpts).Return(nil)
	client.EXPECT().ListRemotes(repo, gomock.Any()).Return(returnReferences, nil)
	client.EXPECT().PullWithContext(gomock.Any(), worktree, gomock.Any(), localBranchRef)

	g := &gogit.GoGit{
		Opts:   opts,
		Client: client,
	}

	err := g.Branch("testBranch")
	if err != nil {
		t.Errorf("Branch() error = %v", err)
		return
	}
}

func newGoGit(t *testing.T) (context.Context, *mockGoGit.MockGoGitClient, gogit.Options) {
	opts := gogit.Options{
		RepositoryDirectory: "testrepo",
	}
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mockGoGit.NewMockGoGitClient(ctrl)

	return ctx, client, opts
}
