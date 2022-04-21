// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/git/gitclient (interfaces: GoGit)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	git "github.com/go-git/go-git/v5"
	config "github.com/go-git/go-git/v5/config"
	plumbing "github.com/go-git/go-git/v5/plumbing"
	object "github.com/go-git/go-git/v5/plumbing/object"
	transport "github.com/go-git/go-git/v5/plumbing/transport"
	gomock "github.com/golang/mock/gomock"
)

// MockGoGit is a mock of GoGit interface.
type MockGoGit struct {
	ctrl     *gomock.Controller
	recorder *MockGoGitMockRecorder
}

// MockGoGitMockRecorder is the mock recorder for MockGoGit.
type MockGoGitMockRecorder struct {
	mock *MockGoGit
}

// NewMockGoGit creates a new mock instance.
func NewMockGoGit(ctrl *gomock.Controller) *MockGoGit {
	mock := &MockGoGit{ctrl: ctrl}
	mock.recorder = &MockGoGitMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGoGit) EXPECT() *MockGoGitMockRecorder {
	return m.recorder
}

// AddGlob mocks base method.
func (m *MockGoGit) AddGlob(arg0 string, arg1 *git.Worktree) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddGlob", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddGlob indicates an expected call of AddGlob.
func (mr *MockGoGitMockRecorder) AddGlob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddGlob", reflect.TypeOf((*MockGoGit)(nil).AddGlob), arg0, arg1)
}

// Checkout mocks base method.
func (m *MockGoGit) Checkout(arg0 *git.Worktree, arg1 *git.CheckoutOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Checkout", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Checkout indicates an expected call of Checkout.
func (mr *MockGoGitMockRecorder) Checkout(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Checkout", reflect.TypeOf((*MockGoGit)(nil).Checkout), arg0, arg1)
}

// Clone mocks base method.
func (m *MockGoGit) Clone(arg0 context.Context, arg1, arg2 string, arg3 transport.AuthMethod) (*git.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Clone", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*git.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Clone indicates an expected call of Clone.
func (mr *MockGoGitMockRecorder) Clone(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clone", reflect.TypeOf((*MockGoGit)(nil).Clone), arg0, arg1, arg2, arg3)
}

// Commit mocks base method.
func (m *MockGoGit) Commit(arg0 string, arg1 *object.Signature, arg2 *git.Worktree) (plumbing.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", arg0, arg1, arg2)
	ret0, _ := ret[0].(plumbing.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockGoGitMockRecorder) Commit(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockGoGit)(nil).Commit), arg0, arg1, arg2)
}

// CommitObject mocks base method.
func (m *MockGoGit) CommitObject(arg0 *git.Repository, arg1 plumbing.Hash) (*object.Commit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitObject", arg0, arg1)
	ret0, _ := ret[0].(*object.Commit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CommitObject indicates an expected call of CommitObject.
func (mr *MockGoGitMockRecorder) CommitObject(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitObject", reflect.TypeOf((*MockGoGit)(nil).CommitObject), arg0, arg1)
}

// Create mocks base method.
func (m *MockGoGit) Create(arg0 *git.Repository, arg1 string) (*git.Remote, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(*git.Remote)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockGoGitMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockGoGit)(nil).Create), arg0, arg1)
}

// CreateBranch mocks base method.
func (m *MockGoGit) CreateBranch(arg0 *git.Repository, arg1 *config.Branch) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBranch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateBranch indicates an expected call of CreateBranch.
func (mr *MockGoGitMockRecorder) CreateBranch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBranch", reflect.TypeOf((*MockGoGit)(nil).CreateBranch), arg0, arg1)
}

// Head mocks base method.
func (m *MockGoGit) Head(arg0 *git.Repository) (*plumbing.Reference, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Head", arg0)
	ret0, _ := ret[0].(*plumbing.Reference)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Head indicates an expected call of Head.
func (mr *MockGoGitMockRecorder) Head(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Head", reflect.TypeOf((*MockGoGit)(nil).Head), arg0)
}

// Init mocks base method.
func (m *MockGoGit) Init(arg0 string) (*git.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", arg0)
	ret0, _ := ret[0].(*git.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Init indicates an expected call of Init.
func (mr *MockGoGitMockRecorder) Init(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockGoGit)(nil).Init), arg0)
}

// ListRemotes mocks base method.
func (m *MockGoGit) ListRemotes(arg0 *git.Repository, arg1 transport.AuthMethod) ([]*plumbing.Reference, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRemotes", arg0, arg1)
	ret0, _ := ret[0].([]*plumbing.Reference)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRemotes indicates an expected call of ListRemotes.
func (mr *MockGoGitMockRecorder) ListRemotes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRemotes", reflect.TypeOf((*MockGoGit)(nil).ListRemotes), arg0, arg1)
}

// OpenDir mocks base method.
func (m *MockGoGit) OpenDir(arg0 string) (*git.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenDir", arg0)
	ret0, _ := ret[0].(*git.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenDir indicates an expected call of OpenDir.
func (mr *MockGoGitMockRecorder) OpenDir(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenDir", reflect.TypeOf((*MockGoGit)(nil).OpenDir), arg0)
}

// OpenWorktree mocks base method.
func (m *MockGoGit) OpenWorktree(arg0 *git.Repository) (*git.Worktree, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenWorktree", arg0)
	ret0, _ := ret[0].(*git.Worktree)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenWorktree indicates an expected call of OpenWorktree.
func (mr *MockGoGitMockRecorder) OpenWorktree(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenWorktree", reflect.TypeOf((*MockGoGit)(nil).OpenWorktree), arg0)
}

// PullWithContext mocks base method.
func (m *MockGoGit) PullWithContext(arg0 context.Context, arg1 *git.Worktree, arg2 transport.AuthMethod, arg3 plumbing.ReferenceName) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullWithContext", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// PullWithContext indicates an expected call of PullWithContext.
func (mr *MockGoGitMockRecorder) PullWithContext(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullWithContext", reflect.TypeOf((*MockGoGit)(nil).PullWithContext), arg0, arg1, arg2, arg3)
}

// PushWithContext mocks base method.
func (m *MockGoGit) PushWithContext(arg0 context.Context, arg1 *git.Repository, arg2 transport.AuthMethod) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PushWithContext", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PushWithContext indicates an expected call of PushWithContext.
func (mr *MockGoGitMockRecorder) PushWithContext(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushWithContext", reflect.TypeOf((*MockGoGit)(nil).PushWithContext), arg0, arg1, arg2)
}

// Remove mocks base method.
func (m *MockGoGit) Remove(arg0 string, arg1 *git.Worktree) (plumbing.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0, arg1)
	ret0, _ := ret[0].(plumbing.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Remove indicates an expected call of Remove.
func (mr *MockGoGitMockRecorder) Remove(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockGoGit)(nil).Remove), arg0, arg1)
}

// SetRepositoryReference mocks base method.
func (m *MockGoGit) SetRepositoryReference(arg0 *git.Repository, arg1 *plumbing.Reference) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetRepositoryReference", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetRepositoryReference indicates an expected call of SetRepositoryReference.
func (mr *MockGoGitMockRecorder) SetRepositoryReference(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRepositoryReference", reflect.TypeOf((*MockGoGit)(nil).SetRepositoryReference), arg0, arg1)
}
