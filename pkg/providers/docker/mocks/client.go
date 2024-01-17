// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers/docker (interfaces: ProviderClient,ProviderKubectlClient,KubeconfigReader)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	executables "github.com/aws/eks-anywhere/pkg/executables"
	types "github.com/aws/eks-anywhere/pkg/types"
	v1beta1 "github.com/aws/etcdadm-controller/api/v1beta1"
	gomock "github.com/golang/mock/gomock"
	v1beta10 "sigs.k8s.io/cluster-api/api/v1beta1"
	v1beta11 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

// MockProviderClient is a mock of ProviderClient interface.
type MockProviderClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderClientMockRecorder
}

// MockProviderClientMockRecorder is the mock recorder for MockProviderClient.
type MockProviderClientMockRecorder struct {
	mock *MockProviderClient
}

// NewMockProviderClient creates a new mock instance.
func NewMockProviderClient(ctrl *gomock.Controller) *MockProviderClient {
	mock := &MockProviderClient{ctrl: ctrl}
	mock.recorder = &MockProviderClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderClient) EXPECT() *MockProviderClientMockRecorder {
	return m.recorder
}

// GetDockerLBPort mocks base method.
func (m *MockProviderClient) GetDockerLBPort(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDockerLBPort", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDockerLBPort indicates an expected call of GetDockerLBPort.
func (mr *MockProviderClientMockRecorder) GetDockerLBPort(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDockerLBPort", reflect.TypeOf((*MockProviderClient)(nil).GetDockerLBPort), arg0, arg1)
}

// MockProviderKubectlClient is a mock of ProviderKubectlClient interface.
type MockProviderKubectlClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderKubectlClientMockRecorder
}

// MockProviderKubectlClientMockRecorder is the mock recorder for MockProviderKubectlClient.
type MockProviderKubectlClientMockRecorder struct {
	mock *MockProviderKubectlClient
}

// NewMockProviderKubectlClient creates a new mock instance.
func NewMockProviderKubectlClient(ctrl *gomock.Controller) *MockProviderKubectlClient {
	mock := &MockProviderKubectlClient{ctrl: ctrl}
	mock.recorder = &MockProviderKubectlClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderKubectlClient) EXPECT() *MockProviderKubectlClientMockRecorder {
	return m.recorder
}

// GetEksaCluster mocks base method.
func (m *MockProviderKubectlClient) GetEksaCluster(arg0 context.Context, arg1 *types.Cluster, arg2 string) (*v1alpha1.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksaCluster", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1alpha1.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksaCluster indicates an expected call of GetEksaCluster.
func (mr *MockProviderKubectlClientMockRecorder) GetEksaCluster(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksaCluster", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEksaCluster), arg0, arg1, arg2)
}

// GetEtcdadmCluster mocks base method.
func (m *MockProviderKubectlClient) GetEtcdadmCluster(arg0 context.Context, arg1 *types.Cluster, arg2 string, arg3 ...executables.KubectlOpt) (*v1beta1.EtcdadmCluster, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEtcdadmCluster", varargs...)
	ret0, _ := ret[0].(*v1beta1.EtcdadmCluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEtcdadmCluster indicates an expected call of GetEtcdadmCluster.
func (mr *MockProviderKubectlClientMockRecorder) GetEtcdadmCluster(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEtcdadmCluster", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEtcdadmCluster), varargs...)
}

// GetKubeadmControlPlane mocks base method.
func (m *MockProviderKubectlClient) GetKubeadmControlPlane(arg0 context.Context, arg1 *types.Cluster, arg2 string, arg3 ...executables.KubectlOpt) (*v1beta11.KubeadmControlPlane, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetKubeadmControlPlane", varargs...)
	ret0, _ := ret[0].(*v1beta11.KubeadmControlPlane)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKubeadmControlPlane indicates an expected call of GetKubeadmControlPlane.
func (mr *MockProviderKubectlClientMockRecorder) GetKubeadmControlPlane(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKubeadmControlPlane", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetKubeadmControlPlane), varargs...)
}

// GetMachineDeployment mocks base method.
func (m *MockProviderKubectlClient) GetMachineDeployment(arg0 context.Context, arg1 string, arg2 ...executables.KubectlOpt) (*v1beta10.MachineDeployment, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMachineDeployment", varargs...)
	ret0, _ := ret[0].(*v1beta10.MachineDeployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachineDeployment indicates an expected call of GetMachineDeployment.
func (mr *MockProviderKubectlClientMockRecorder) GetMachineDeployment(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachineDeployment", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetMachineDeployment), varargs...)
}

// UpdateAnnotation mocks base method.
func (m *MockProviderKubectlClient) UpdateAnnotation(arg0 context.Context, arg1, arg2 string, arg3 map[string]string, arg4 ...executables.KubectlOpt) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateAnnotation", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAnnotation indicates an expected call of UpdateAnnotation.
func (mr *MockProviderKubectlClientMockRecorder) UpdateAnnotation(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAnnotation", reflect.TypeOf((*MockProviderKubectlClient)(nil).UpdateAnnotation), varargs...)
}

// MockKubeconfigReader is a mock of KubeconfigReader interface.
type MockKubeconfigReader struct {
	ctrl     *gomock.Controller
	recorder *MockKubeconfigReaderMockRecorder
}

// MockKubeconfigReaderMockRecorder is the mock recorder for MockKubeconfigReader.
type MockKubeconfigReaderMockRecorder struct {
	mock *MockKubeconfigReader
}

// NewMockKubeconfigReader creates a new mock instance.
func NewMockKubeconfigReader(ctrl *gomock.Controller) *MockKubeconfigReader {
	mock := &MockKubeconfigReader{ctrl: ctrl}
	mock.recorder = &MockKubeconfigReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKubeconfigReader) EXPECT() *MockKubeconfigReaderMockRecorder {
	return m.recorder
}

// GetClusterKubeconfig mocks base method.
func (m *MockKubeconfigReader) GetClusterKubeconfig(arg0 context.Context, arg1, arg2 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterKubeconfig", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusterKubeconfig indicates an expected call of GetClusterKubeconfig.
func (mr *MockKubeconfigReaderMockRecorder) GetClusterKubeconfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterKubeconfig", reflect.TypeOf((*MockKubeconfigReader)(nil).GetClusterKubeconfig), arg0, arg1, arg2)
}
