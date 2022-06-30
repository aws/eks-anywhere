// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers/cloudstack (interfaces: ProviderCmkClient,ProviderKubectlClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	executables "github.com/aws/eks-anywhere/pkg/executables"
	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
	v1beta1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	v1beta10 "sigs.k8s.io/cluster-api/api/v1beta1"
	v1beta11 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

// MockProviderCmkClient is a mock of ProviderCmkClient interface.
type MockProviderCmkClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderCmkClientMockRecorder
}

// MockProviderCmkClientMockRecorder is the mock recorder for MockProviderCmkClient.
type MockProviderCmkClientMockRecorder struct {
	mock *MockProviderCmkClient
}

// NewMockProviderCmkClient creates a new mock instance.
func NewMockProviderCmkClient(ctrl *gomock.Controller) *MockProviderCmkClient {
	mock := &MockProviderCmkClient{ctrl: ctrl}
	mock.recorder = &MockProviderCmkClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderCmkClient) EXPECT() *MockProviderCmkClientMockRecorder {
	return m.recorder
}

// GetManagementApiEndpoint mocks base method.
func (m *MockProviderCmkClient) GetManagementApiEndpoint(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManagementApiEndpoint", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManagementApiEndpoint indicates an expected call of GetManagementApiEndpoint.
func (mr *MockProviderCmkClientMockRecorder) GetManagementApiEndpoint(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManagementApiEndpoint", reflect.TypeOf((*MockProviderCmkClient)(nil).GetManagementApiEndpoint), arg0)
}

// ValidateAccountPresent mocks base method.
func (m *MockProviderCmkClient) ValidateAccountPresent(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateAccountPresent", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateAccountPresent indicates an expected call of ValidateAccountPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateAccountPresent(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateAccountPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateAccountPresent), arg0, arg1, arg2, arg3)
}

// ValidateAffinityGroupsPresent mocks base method.
func (m *MockProviderCmkClient) ValidateAffinityGroupsPresent(arg0 context.Context, arg1, arg2, arg3 string, arg4 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateAffinityGroupsPresent", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateAffinityGroupsPresent indicates an expected call of ValidateAffinityGroupsPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateAffinityGroupsPresent(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateAffinityGroupsPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateAffinityGroupsPresent), arg0, arg1, arg2, arg3, arg4)
}

// ValidateCloudStackConnection mocks base method.
func (m *MockProviderCmkClient) ValidateCloudStackConnection(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateCloudStackConnection", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateCloudStackConnection indicates an expected call of ValidateCloudStackConnection.
func (mr *MockProviderCmkClientMockRecorder) ValidateCloudStackConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateCloudStackConnection", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateCloudStackConnection), arg0, arg1)
}

// ValidateDiskOfferingPresent mocks base method.
func (m *MockProviderCmkClient) ValidateDiskOfferingPresent(arg0 context.Context, arg1, arg2 string, arg3 v1alpha1.CloudStackResourceDiskOffering) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateDiskOfferingPresent", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateDiskOfferingPresent indicates an expected call of ValidateDiskOfferingPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateDiskOfferingPresent(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateDiskOfferingPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateDiskOfferingPresent), arg0, arg1, arg2, arg3)
}

// ValidateDomainPresent mocks base method.
func (m *MockProviderCmkClient) ValidateDomainPresent(arg0 context.Context, arg1, arg2 string) (v1alpha1.CloudStackResourceIdentifier, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateDomainPresent", arg0, arg1, arg2)
	ret0, _ := ret[0].(v1alpha1.CloudStackResourceIdentifier)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateDomainPresent indicates an expected call of ValidateDomainPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateDomainPresent(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateDomainPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateDomainPresent), arg0, arg1, arg2)
}

// ValidateNetworkPresent mocks base method.
func (m *MockProviderCmkClient) ValidateNetworkPresent(arg0 context.Context, arg1, arg2 string, arg3 v1alpha1.CloudStackResourceIdentifier, arg4, arg5 string, arg6 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateNetworkPresent", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateNetworkPresent indicates an expected call of ValidateNetworkPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateNetworkPresent(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateNetworkPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateNetworkPresent), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// ValidateServiceOfferingPresent mocks base method.
func (m *MockProviderCmkClient) ValidateServiceOfferingPresent(arg0 context.Context, arg1, arg2 string, arg3 v1alpha1.CloudStackResourceIdentifier) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateServiceOfferingPresent", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateServiceOfferingPresent indicates an expected call of ValidateServiceOfferingPresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateServiceOfferingPresent(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateServiceOfferingPresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateServiceOfferingPresent), arg0, arg1, arg2, arg3)
}

// ValidateTemplatePresent mocks base method.
func (m *MockProviderCmkClient) ValidateTemplatePresent(arg0 context.Context, arg1, arg2, arg3, arg4 string, arg5 v1alpha1.CloudStackResourceIdentifier) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateTemplatePresent", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateTemplatePresent indicates an expected call of ValidateTemplatePresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateTemplatePresent(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateTemplatePresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateTemplatePresent), arg0, arg1, arg2, arg3, arg4, arg5)
}

// ValidateZonePresent mocks base method.
func (m *MockProviderCmkClient) ValidateZonePresent(arg0 context.Context, arg1 string, arg2 v1alpha1.CloudStackZone) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateZonePresent", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateZonePresent indicates an expected call of ValidateZonePresent.
func (mr *MockProviderCmkClientMockRecorder) ValidateZonePresent(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateZonePresent", reflect.TypeOf((*MockProviderCmkClient)(nil).ValidateZonePresent), arg0, arg1, arg2)
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

// ApplyKubeSpecFromBytes mocks base method.
func (m *MockProviderKubectlClient) ApplyKubeSpecFromBytes(arg0 context.Context, arg1 *types.Cluster, arg2 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyKubeSpecFromBytes", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyKubeSpecFromBytes indicates an expected call of ApplyKubeSpecFromBytes.
func (mr *MockProviderKubectlClientMockRecorder) ApplyKubeSpecFromBytes(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyKubeSpecFromBytes", reflect.TypeOf((*MockProviderKubectlClient)(nil).ApplyKubeSpecFromBytes), arg0, arg1, arg2)
}

// CreateNamespace mocks base method.
func (m *MockProviderKubectlClient) CreateNamespace(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNamespace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateNamespace indicates an expected call of CreateNamespace.
func (mr *MockProviderKubectlClientMockRecorder) CreateNamespace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNamespace", reflect.TypeOf((*MockProviderKubectlClient)(nil).CreateNamespace), arg0, arg1, arg2)
}

// DeleteEksaCloudStackDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) DeleteEksaCloudStackDatacenterConfig(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEksaCloudStackDatacenterConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEksaCloudStackDatacenterConfig indicates an expected call of DeleteEksaCloudStackDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) DeleteEksaCloudStackDatacenterConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEksaCloudStackDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).DeleteEksaCloudStackDatacenterConfig), arg0, arg1, arg2, arg3)
}

// DeleteEksaCloudStackMachineConfig mocks base method.
func (m *MockProviderKubectlClient) DeleteEksaCloudStackMachineConfig(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEksaCloudStackMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEksaCloudStackMachineConfig indicates an expected call of DeleteEksaCloudStackMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) DeleteEksaCloudStackMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEksaCloudStackMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).DeleteEksaCloudStackMachineConfig), arg0, arg1, arg2, arg3)
}

// GetEksaCloudStackDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) GetEksaCloudStackDatacenterConfig(arg0 context.Context, arg1, arg2, arg3 string) (*v1alpha1.CloudStackDatacenterConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksaCloudStackDatacenterConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1alpha1.CloudStackDatacenterConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksaCloudStackDatacenterConfig indicates an expected call of GetEksaCloudStackDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) GetEksaCloudStackDatacenterConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksaCloudStackDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEksaCloudStackDatacenterConfig), arg0, arg1, arg2, arg3)
}

// GetEksaCloudStackMachineConfig mocks base method.
func (m *MockProviderKubectlClient) GetEksaCloudStackMachineConfig(arg0 context.Context, arg1, arg2, arg3 string) (*v1alpha1.CloudStackMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksaCloudStackMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1alpha1.CloudStackMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksaCloudStackMachineConfig indicates an expected call of GetEksaCloudStackMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) GetEksaCloudStackMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksaCloudStackMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEksaCloudStackMachineConfig), arg0, arg1, arg2, arg3)
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

// GetSecret mocks base method.
func (m *MockProviderKubectlClient) GetSecret(arg0 context.Context, arg1 string, arg2 ...executables.KubectlOpt) (*v1.Secret, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetSecret", varargs...)
	ret0, _ := ret[0].(*v1.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockProviderKubectlClientMockRecorder) GetSecret(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetSecret), varargs...)
}

// LoadSecret mocks base method.
func (m *MockProviderKubectlClient) LoadSecret(arg0 context.Context, arg1, arg2, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadSecret", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadSecret indicates an expected call of LoadSecret.
func (mr *MockProviderKubectlClientMockRecorder) LoadSecret(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadSecret", reflect.TypeOf((*MockProviderKubectlClient)(nil).LoadSecret), arg0, arg1, arg2, arg3, arg4)
}

// SearchCloudStackDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) SearchCloudStackDatacenterConfig(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1alpha1.CloudStackDatacenterConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchCloudStackDatacenterConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1alpha1.CloudStackDatacenterConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchCloudStackDatacenterConfig indicates an expected call of SearchCloudStackDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) SearchCloudStackDatacenterConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchCloudStackDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).SearchCloudStackDatacenterConfig), arg0, arg1, arg2, arg3)
}

// SearchCloudStackMachineConfig mocks base method.
func (m *MockProviderKubectlClient) SearchCloudStackMachineConfig(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1alpha1.CloudStackMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchCloudStackMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1alpha1.CloudStackMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchCloudStackMachineConfig indicates an expected call of SearchCloudStackMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) SearchCloudStackMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchCloudStackMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).SearchCloudStackMachineConfig), arg0, arg1, arg2, arg3)
}

// SetEksaControllerEnvVar mocks base method.
func (m *MockProviderKubectlClient) SetEksaControllerEnvVar(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetEksaControllerEnvVar", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetEksaControllerEnvVar indicates an expected call of SetEksaControllerEnvVar.
func (mr *MockProviderKubectlClientMockRecorder) SetEksaControllerEnvVar(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetEksaControllerEnvVar", reflect.TypeOf((*MockProviderKubectlClient)(nil).SetEksaControllerEnvVar), arg0, arg1, arg2, arg3)
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
