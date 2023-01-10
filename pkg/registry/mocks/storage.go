// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/registry/storage.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	registry "github.com/aws/eks-anywhere/pkg/registry"
	gomock "github.com/golang/mock/gomock"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	registry0 "oras.land/oras-go/v2/registry"
)

// MockStorageClient is a mock of StorageClient interface.
type MockStorageClient struct {
	ctrl     *gomock.Controller
	recorder *MockStorageClientMockRecorder
}

// MockStorageClientMockRecorder is the mock recorder for MockStorageClient.
type MockStorageClientMockRecorder struct {
	mock *MockStorageClient
}

// NewMockStorageClient creates a new mock instance.
func NewMockStorageClient(ctrl *gomock.Controller) *MockStorageClient {
	mock := &MockStorageClient{ctrl: ctrl}
	mock.recorder = &MockStorageClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageClient) EXPECT() *MockStorageClientMockRecorder {
	return m.recorder
}

// CopyGraph mocks base method.
func (m *MockStorageClient) CopyGraph(ctx context.Context, srcStorage, dstStorage registry0.Repository, desc v1.Descriptor) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CopyGraph", ctx, srcStorage, dstStorage, desc)
	ret0, _ := ret[0].(error)
	return ret0
}

// CopyGraph indicates an expected call of CopyGraph.
func (mr *MockStorageClientMockRecorder) CopyGraph(ctx, srcStorage, dstStorage, desc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CopyGraph", reflect.TypeOf((*MockStorageClient)(nil).CopyGraph), ctx, srcStorage, dstStorage, desc)
}

// Destination mocks base method.
func (m *MockStorageClient) Destination(image registry.Artifact) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destination", image)
	ret0, _ := ret[0].(string)
	return ret0
}

// Destination indicates an expected call of Destination.
func (mr *MockStorageClientMockRecorder) Destination(image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destination", reflect.TypeOf((*MockStorageClient)(nil).Destination), image)
}

// GetStorage mocks base method.
func (m *MockStorageClient) GetStorage(ctx context.Context, image registry.Artifact) (registry0.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStorage", ctx, image)
	ret0, _ := ret[0].(registry0.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStorage indicates an expected call of GetStorage.
func (mr *MockStorageClientMockRecorder) GetStorage(ctx, image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorage", reflect.TypeOf((*MockStorageClient)(nil).GetStorage), ctx, image)
}

// Init mocks base method.
func (m *MockStorageClient) Init() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init")
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockStorageClientMockRecorder) Init() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockStorageClient)(nil).Init))
}

// Resolve mocks base method.
func (m *MockStorageClient) Resolve(ctx context.Context, srcStorage registry0.Repository, versionedImage string) (v1.Descriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resolve", ctx, srcStorage, versionedImage)
	ret0, _ := ret[0].(v1.Descriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Resolve indicates an expected call of Resolve.
func (mr *MockStorageClientMockRecorder) Resolve(ctx, srcStorage, versionedImage interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resolve", reflect.TypeOf((*MockStorageClient)(nil).Resolve), ctx, srcStorage, versionedImage)
}

// SetProject mocks base method.
func (m *MockStorageClient) SetProject(project string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetProject", project)
}

// SetProject indicates an expected call of SetProject.
func (mr *MockStorageClientMockRecorder) SetProject(project interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetProject", reflect.TypeOf((*MockStorageClient)(nil).SetProject), project)
}
