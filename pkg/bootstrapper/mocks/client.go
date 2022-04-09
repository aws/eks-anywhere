// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/bootstrapper (interfaces: ClusterClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	bootstrapper "github.com/aws/eks-anywhere/pkg/bootstrapper"
	cluster "github.com/aws/eks-anywhere/pkg/cluster"
	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
)

// MockClusterClient is a mock of ClusterClient interface.
type MockClusterClient struct {
	ctrl     *gomock.Controller
	recorder *MockClusterClientMockRecorder
}

// MockClusterClientMockRecorder is the mock recorder for MockClusterClient.
type MockClusterClientMockRecorder struct {
	mock *MockClusterClient
}

// NewMockClusterClient creates a new mock instance.
func NewMockClusterClient(ctrl *gomock.Controller) *MockClusterClient {
	mock := &MockClusterClient{ctrl: ctrl}
	mock.recorder = &MockClusterClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClusterClient) EXPECT() *MockClusterClientMockRecorder {
	return m.recorder
}

// ApplyKubeSpecFromBytes mocks base method.
func (m *MockClusterClient) ApplyKubeSpecFromBytes(arg0 context.Context, arg1 *types.Cluster, arg2 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyKubeSpecFromBytes", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyKubeSpecFromBytes indicates an expected call of ApplyKubeSpecFromBytes.
func (mr *MockClusterClientMockRecorder) ApplyKubeSpecFromBytes(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyKubeSpecFromBytes", reflect.TypeOf((*MockClusterClient)(nil).ApplyKubeSpecFromBytes), arg0, arg1, arg2)
}

// ClusterExists mocks base method.
func (m *MockClusterClient) ClusterExists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClusterExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ClusterExists indicates an expected call of ClusterExists.
func (mr *MockClusterClientMockRecorder) ClusterExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterExists", reflect.TypeOf((*MockClusterClient)(nil).ClusterExists), arg0, arg1)
}

// CreateBootstrapCluster mocks base method.
func (m *MockClusterClient) CreateBootstrapCluster(arg0 context.Context, arg1 *cluster.Spec, arg2 ...bootstrapper.BootstrapClusterClientOption) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateBootstrapCluster", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBootstrapCluster indicates an expected call of CreateBootstrapCluster.
func (mr *MockClusterClientMockRecorder) CreateBootstrapCluster(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBootstrapCluster", reflect.TypeOf((*MockClusterClient)(nil).CreateBootstrapCluster), varargs...)
}

// CreateNamespace mocks base method.
func (m *MockClusterClient) CreateNamespace(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNamespace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateNamespace indicates an expected call of CreateNamespace.
func (mr *MockClusterClientMockRecorder) CreateNamespace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNamespace", reflect.TypeOf((*MockClusterClient)(nil).CreateNamespace), arg0, arg1, arg2)
}

// DeleteBootstrapCluster mocks base method.
func (m *MockClusterClient) DeleteBootstrapCluster(arg0 context.Context, arg1 *types.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBootstrapCluster", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBootstrapCluster indicates an expected call of DeleteBootstrapCluster.
func (mr *MockClusterClientMockRecorder) DeleteBootstrapCluster(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBootstrapCluster", reflect.TypeOf((*MockClusterClient)(nil).DeleteBootstrapCluster), arg0, arg1)
}

// GetClusters mocks base method.
func (m *MockClusterClient) GetClusters(arg0 context.Context, arg1 *types.Cluster) ([]types.CAPICluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusters", arg0, arg1)
	ret0, _ := ret[0].([]types.CAPICluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusters indicates an expected call of GetClusters.
func (mr *MockClusterClientMockRecorder) GetClusters(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusters", reflect.TypeOf((*MockClusterClient)(nil).GetClusters), arg0, arg1)
}

// GetKubeconfig mocks base method.
func (m *MockClusterClient) GetKubeconfig(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKubeconfig", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKubeconfig indicates an expected call of GetKubeconfig.
func (mr *MockClusterClientMockRecorder) GetKubeconfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKubeconfig", reflect.TypeOf((*MockClusterClient)(nil).GetKubeconfig), arg0, arg1)
}

// GetNamespace mocks base method.
func (m *MockClusterClient) GetNamespace(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetNamespace indicates an expected call of GetNamespace.
func (mr *MockClusterClientMockRecorder) GetNamespace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MockClusterClient)(nil).GetNamespace), arg0, arg1, arg2)
}

// ValidateClustersCRD mocks base method.
func (m *MockClusterClient) ValidateClustersCRD(arg0 context.Context, arg1 *types.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateClustersCRD", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateClustersCRD indicates an expected call of ValidateClustersCRD.
func (mr *MockClusterClientMockRecorder) ValidateClustersCRD(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateClustersCRD", reflect.TypeOf((*MockClusterClient)(nil).ValidateClustersCRD), arg0, arg1)
}

// WithDefaultCNIDisabled mocks base method.
func (m *MockClusterClient) WithDefaultCNIDisabled() bootstrapper.BootstrapClusterClientOption {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithDefaultCNIDisabled")
	ret0, _ := ret[0].(bootstrapper.BootstrapClusterClientOption)
	return ret0
}

// WithDefaultCNIDisabled indicates an expected call of WithDefaultCNIDisabled.
func (mr *MockClusterClientMockRecorder) WithDefaultCNIDisabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithDefaultCNIDisabled", reflect.TypeOf((*MockClusterClient)(nil).WithDefaultCNIDisabled))
}

// WithEnv mocks base method.
func (m *MockClusterClient) WithEnv(arg0 map[string]string) bootstrapper.BootstrapClusterClientOption {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithEnv", arg0)
	ret0, _ := ret[0].(bootstrapper.BootstrapClusterClientOption)
	return ret0
}

// WithEnv indicates an expected call of WithEnv.
func (mr *MockClusterClientMockRecorder) WithEnv(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithEnv", reflect.TypeOf((*MockClusterClient)(nil).WithEnv), arg0)
}

// WithExtraDockerMounts mocks base method.
func (m *MockClusterClient) WithExtraDockerMounts() bootstrapper.BootstrapClusterClientOption {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithExtraDockerMounts")
	ret0, _ := ret[0].(bootstrapper.BootstrapClusterClientOption)
	return ret0
}

// WithExtraDockerMounts indicates an expected call of WithExtraDockerMounts.
func (mr *MockClusterClientMockRecorder) WithExtraDockerMounts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithExtraDockerMounts", reflect.TypeOf((*MockClusterClient)(nil).WithExtraDockerMounts))
}

// WithExtraPortMappings mocks base method.
func (m *MockClusterClient) WithExtraPortMappings(arg0 []int) bootstrapper.BootstrapClusterClientOption {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithExtraPortMappings", arg0)
	ret0, _ := ret[0].(bootstrapper.BootstrapClusterClientOption)
	return ret0
}

// WithExtraPortMappings indicates an expected call of WithExtraPortMappings.
func (mr *MockClusterClientMockRecorder) WithExtraPortMappings(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithExtraPortMappings", reflect.TypeOf((*MockClusterClient)(nil).WithExtraPortMappings), arg0)
}
