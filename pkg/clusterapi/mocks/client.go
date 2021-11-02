// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/clusterapi/resourceset_manager.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	v1alpha3 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// ApplyKubeSpecFromBytes mocks base method.
func (m *MockClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyKubeSpecFromBytes", ctx, cluster, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyKubeSpecFromBytes indicates an expected call of ApplyKubeSpecFromBytes.
func (mr *MockClientMockRecorder) ApplyKubeSpecFromBytes(ctx, cluster, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyKubeSpecFromBytes", reflect.TypeOf((*MockClient)(nil).ApplyKubeSpecFromBytes), ctx, cluster, data)
}

// GetClusterResourceSet mocks base method.
func (m *MockClient) GetClusterResourceSet(ctx context.Context, kubeconfigFile, name, namespace string) (*v1alpha3.ClusterResourceSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterResourceSet", ctx, kubeconfigFile, name, namespace)
	ret0, _ := ret[0].(*v1alpha3.ClusterResourceSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusterResourceSet indicates an expected call of GetClusterResourceSet.
func (mr *MockClientMockRecorder) GetClusterResourceSet(ctx, kubeconfigFile, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterResourceSet", reflect.TypeOf((*MockClient)(nil).GetClusterResourceSet), ctx, kubeconfigFile, name, namespace)
}

// GetConfigMap mocks base method.
func (m *MockClient) GetConfigMap(ctx context.Context, kubeconfigFile, name, namespace string) (*v1.ConfigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigMap", ctx, kubeconfigFile, name, namespace)
	ret0, _ := ret[0].(*v1.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfigMap indicates an expected call of GetConfigMap.
func (mr *MockClientMockRecorder) GetConfigMap(ctx, kubeconfigFile, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigMap", reflect.TypeOf((*MockClient)(nil).GetConfigMap), ctx, kubeconfigFile, name, namespace)
}

// GetSecretFromNamespace mocks base method.
func (m *MockClient) GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*v1.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecretFromNamespace", ctx, kubeconfigFile, name, namespace)
	ret0, _ := ret[0].(*v1.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecretFromNamespace indicates an expected call of GetSecretFromNamespace.
func (mr *MockClientMockRecorder) GetSecretFromNamespace(ctx, kubeconfigFile, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecretFromNamespace", reflect.TypeOf((*MockClient)(nil).GetSecretFromNamespace), ctx, kubeconfigFile, name, namespace)
}
