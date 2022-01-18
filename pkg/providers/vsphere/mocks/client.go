// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers/vsphere (interfaces: ProviderGovcClient,ProviderKubectlClient,ClusterResourceSetManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	executables "github.com/aws/eks-anywhere/pkg/executables"
	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
	v1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	v1 "k8s.io/api/core/v1"
	v1alpha30 "sigs.k8s.io/cluster-api/api/v1alpha3"
	v1alpha31 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
)

// MockProviderGovcClient is a mock of ProviderGovcClient interface.
type MockProviderGovcClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderGovcClientMockRecorder
}

// MockProviderGovcClientMockRecorder is the mock recorder for MockProviderGovcClient.
type MockProviderGovcClientMockRecorder struct {
	mock *MockProviderGovcClient
}

// NewMockProviderGovcClient creates a new mock instance.
func NewMockProviderGovcClient(ctrl *gomock.Controller) *MockProviderGovcClient {
	mock := &MockProviderGovcClient{ctrl: ctrl}
	mock.recorder = &MockProviderGovcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderGovcClient) EXPECT() *MockProviderGovcClientMockRecorder {
	return m.recorder
}

// AddTag mocks base method.
func (m *MockProviderGovcClient) AddTag(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTag", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTag indicates an expected call of AddTag.
func (mr *MockProviderGovcClientMockRecorder) AddTag(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTag", reflect.TypeOf((*MockProviderGovcClient)(nil).AddTag), arg0, arg1, arg2)
}

// ConfigureCertThumbprint mocks base method.
func (m *MockProviderGovcClient) ConfigureCertThumbprint(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigureCertThumbprint", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ConfigureCertThumbprint indicates an expected call of ConfigureCertThumbprint.
func (mr *MockProviderGovcClientMockRecorder) ConfigureCertThumbprint(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigureCertThumbprint", reflect.TypeOf((*MockProviderGovcClient)(nil).ConfigureCertThumbprint), arg0, arg1, arg2)
}

// CreateCategoryForVM mocks base method.
func (m *MockProviderGovcClient) CreateCategoryForVM(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCategoryForVM", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCategoryForVM indicates an expected call of CreateCategoryForVM.
func (mr *MockProviderGovcClientMockRecorder) CreateCategoryForVM(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCategoryForVM", reflect.TypeOf((*MockProviderGovcClient)(nil).CreateCategoryForVM), arg0, arg1)
}

// CreateLibrary mocks base method.
func (m *MockProviderGovcClient) CreateLibrary(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLibrary", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateLibrary indicates an expected call of CreateLibrary.
func (mr *MockProviderGovcClientMockRecorder) CreateLibrary(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLibrary", reflect.TypeOf((*MockProviderGovcClient)(nil).CreateLibrary), arg0, arg1, arg2)
}

// CreateTag mocks base method.
func (m *MockProviderGovcClient) CreateTag(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTag", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateTag indicates an expected call of CreateTag.
func (mr *MockProviderGovcClientMockRecorder) CreateTag(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTag", reflect.TypeOf((*MockProviderGovcClient)(nil).CreateTag), arg0, arg1, arg2)
}

// DatacenterExists mocks base method.
func (m *MockProviderGovcClient) DatacenterExists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DatacenterExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DatacenterExists indicates an expected call of DatacenterExists.
func (mr *MockProviderGovcClientMockRecorder) DatacenterExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DatacenterExists", reflect.TypeOf((*MockProviderGovcClient)(nil).DatacenterExists), arg0, arg1)
}

// DeleteLibraryElement mocks base method.
func (m *MockProviderGovcClient) DeleteLibraryElement(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLibraryElement", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLibraryElement indicates an expected call of DeleteLibraryElement.
func (mr *MockProviderGovcClientMockRecorder) DeleteLibraryElement(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLibraryElement", reflect.TypeOf((*MockProviderGovcClient)(nil).DeleteLibraryElement), arg0, arg1)
}

// DeployTemplateFromLibrary mocks base method.
func (m *MockProviderGovcClient) DeployTemplateFromLibrary(arg0 context.Context, arg1, arg2, arg3, arg4, arg5, arg6 string, arg7 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeployTemplateFromLibrary", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeployTemplateFromLibrary indicates an expected call of DeployTemplateFromLibrary.
func (mr *MockProviderGovcClientMockRecorder) DeployTemplateFromLibrary(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeployTemplateFromLibrary", reflect.TypeOf((*MockProviderGovcClient)(nil).DeployTemplateFromLibrary), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// GetCertThumbprint mocks base method.
func (m *MockProviderGovcClient) GetCertThumbprint(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCertThumbprint", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCertThumbprint indicates an expected call of GetCertThumbprint.
func (mr *MockProviderGovcClientMockRecorder) GetCertThumbprint(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCertThumbprint", reflect.TypeOf((*MockProviderGovcClient)(nil).GetCertThumbprint), arg0)
}

// GetLibraryElementContentVersion mocks base method.
func (m *MockProviderGovcClient) GetLibraryElementContentVersion(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLibraryElementContentVersion", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLibraryElementContentVersion indicates an expected call of GetLibraryElementContentVersion.
func (mr *MockProviderGovcClientMockRecorder) GetLibraryElementContentVersion(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLibraryElementContentVersion", reflect.TypeOf((*MockProviderGovcClient)(nil).GetLibraryElementContentVersion), arg0, arg1)
}

// GetTags mocks base method.
func (m *MockProviderGovcClient) GetTags(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTags", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTags indicates an expected call of GetTags.
func (mr *MockProviderGovcClientMockRecorder) GetTags(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTags", reflect.TypeOf((*MockProviderGovcClient)(nil).GetTags), arg0, arg1)
}

// GetWorkloadAvailableSpace mocks base method.
func (m *MockProviderGovcClient) GetWorkloadAvailableSpace(arg0 context.Context, arg1 string) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkloadAvailableSpace", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWorkloadAvailableSpace indicates an expected call of GetWorkloadAvailableSpace.
func (mr *MockProviderGovcClientMockRecorder) GetWorkloadAvailableSpace(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkloadAvailableSpace", reflect.TypeOf((*MockProviderGovcClient)(nil).GetWorkloadAvailableSpace), arg0, arg1)
}

// ImportTemplate mocks base method.
func (m *MockProviderGovcClient) ImportTemplate(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ImportTemplate", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ImportTemplate indicates an expected call of ImportTemplate.
func (mr *MockProviderGovcClientMockRecorder) ImportTemplate(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImportTemplate", reflect.TypeOf((*MockProviderGovcClient)(nil).ImportTemplate), arg0, arg1, arg2, arg3)
}

// IsCertSelfSigned mocks base method.
func (m *MockProviderGovcClient) IsCertSelfSigned(arg0 context.Context) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsCertSelfSigned", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsCertSelfSigned indicates an expected call of IsCertSelfSigned.
func (mr *MockProviderGovcClientMockRecorder) IsCertSelfSigned(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsCertSelfSigned", reflect.TypeOf((*MockProviderGovcClient)(nil).IsCertSelfSigned), arg0)
}

// LibraryElementExists mocks base method.
func (m *MockProviderGovcClient) LibraryElementExists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LibraryElementExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LibraryElementExists indicates an expected call of LibraryElementExists.
func (mr *MockProviderGovcClientMockRecorder) LibraryElementExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LibraryElementExists", reflect.TypeOf((*MockProviderGovcClient)(nil).LibraryElementExists), arg0, arg1)
}

// ListCategories mocks base method.
func (m *MockProviderGovcClient) ListCategories(arg0 context.Context) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCategories", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListCategories indicates an expected call of ListCategories.
func (mr *MockProviderGovcClientMockRecorder) ListCategories(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCategories", reflect.TypeOf((*MockProviderGovcClient)(nil).ListCategories), arg0)
}

// ListTags mocks base method.
func (m *MockProviderGovcClient) ListTags(arg0 context.Context) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTags", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTags indicates an expected call of ListTags.
func (mr *MockProviderGovcClientMockRecorder) ListTags(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTags", reflect.TypeOf((*MockProviderGovcClient)(nil).ListTags), arg0)
}

// NetworkExists mocks base method.
func (m *MockProviderGovcClient) NetworkExists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkExists indicates an expected call of NetworkExists.
func (mr *MockProviderGovcClientMockRecorder) NetworkExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkExists", reflect.TypeOf((*MockProviderGovcClient)(nil).NetworkExists), arg0, arg1)
}

// SearchTemplate mocks base method.
func (m *MockProviderGovcClient) SearchTemplate(arg0 context.Context, arg1 string, arg2 *v1alpha1.VSphereMachineConfig) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchTemplate", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchTemplate indicates an expected call of SearchTemplate.
func (mr *MockProviderGovcClientMockRecorder) SearchTemplate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchTemplate", reflect.TypeOf((*MockProviderGovcClient)(nil).SearchTemplate), arg0, arg1, arg2)
}

// TemplateHasSnapshot mocks base method.
func (m *MockProviderGovcClient) TemplateHasSnapshot(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TemplateHasSnapshot", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TemplateHasSnapshot indicates an expected call of TemplateHasSnapshot.
func (mr *MockProviderGovcClientMockRecorder) TemplateHasSnapshot(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TemplateHasSnapshot", reflect.TypeOf((*MockProviderGovcClient)(nil).TemplateHasSnapshot), arg0, arg1)
}

// ValidateVCenterAuthentication mocks base method.
func (m *MockProviderGovcClient) ValidateVCenterAuthentication(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateVCenterAuthentication", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateVCenterAuthentication indicates an expected call of ValidateVCenterAuthentication.
func (mr *MockProviderGovcClientMockRecorder) ValidateVCenterAuthentication(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateVCenterAuthentication", reflect.TypeOf((*MockProviderGovcClient)(nil).ValidateVCenterAuthentication), arg0)
}

// ValidateVCenterConnection mocks base method.
func (m *MockProviderGovcClient) ValidateVCenterConnection(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateVCenterConnection", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateVCenterConnection indicates an expected call of ValidateVCenterConnection.
func (mr *MockProviderGovcClientMockRecorder) ValidateVCenterConnection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateVCenterConnection", reflect.TypeOf((*MockProviderGovcClient)(nil).ValidateVCenterConnection), arg0, arg1)
}

// ValidateVCenterSetupMachineConfig mocks base method.
func (m *MockProviderGovcClient) ValidateVCenterSetupMachineConfig(arg0 context.Context, arg1 *v1alpha1.VSphereDatacenterConfig, arg2 *v1alpha1.VSphereMachineConfig, arg3 *bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateVCenterSetupMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateVCenterSetupMachineConfig indicates an expected call of ValidateVCenterSetupMachineConfig.
func (mr *MockProviderGovcClientMockRecorder) ValidateVCenterSetupMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateVCenterSetupMachineConfig", reflect.TypeOf((*MockProviderGovcClient)(nil).ValidateVCenterSetupMachineConfig), arg0, arg1, arg2, arg3)
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

// ApplyTolerationsFromTaintsToDaemonSet mocks base method.
func (m *MockProviderKubectlClient) ApplyTolerationsFromTaintsToDaemonSet(arg0 context.Context, arg1, arg2 []v1.Taint, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyTolerationsFromTaintsToDaemonSet", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyTolerationsFromTaintsToDaemonSet indicates an expected call of ApplyTolerationsFromTaintsToDaemonSet.
func (mr *MockProviderKubectlClientMockRecorder) ApplyTolerationsFromTaintsToDaemonSet(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyTolerationsFromTaintsToDaemonSet", reflect.TypeOf((*MockProviderKubectlClient)(nil).ApplyTolerationsFromTaintsToDaemonSet), arg0, arg1, arg2, arg3, arg4)
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

// DeleteEksaDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) DeleteEksaDatacenterConfig(arg0 context.Context, arg1, arg2, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEksaDatacenterConfig", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEksaDatacenterConfig indicates an expected call of DeleteEksaDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) DeleteEksaDatacenterConfig(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEksaDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).DeleteEksaDatacenterConfig), arg0, arg1, arg2, arg3, arg4)
}

// DeleteEksaMachineConfig mocks base method.
func (m *MockProviderKubectlClient) DeleteEksaMachineConfig(arg0 context.Context, arg1, arg2, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEksaMachineConfig", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEksaMachineConfig indicates an expected call of DeleteEksaMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) DeleteEksaMachineConfig(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEksaMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).DeleteEksaMachineConfig), arg0, arg1, arg2, arg3, arg4)
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

// GetEksaVSphereDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) GetEksaVSphereDatacenterConfig(arg0 context.Context, arg1, arg2, arg3 string) (*v1alpha1.VSphereDatacenterConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksaVSphereDatacenterConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1alpha1.VSphereDatacenterConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksaVSphereDatacenterConfig indicates an expected call of GetEksaVSphereDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) GetEksaVSphereDatacenterConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksaVSphereDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEksaVSphereDatacenterConfig), arg0, arg1, arg2, arg3)
}

// GetEksaVSphereMachineConfig mocks base method.
func (m *MockProviderKubectlClient) GetEksaVSphereMachineConfig(arg0 context.Context, arg1, arg2, arg3 string) (*v1alpha1.VSphereMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksaVSphereMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1alpha1.VSphereMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksaVSphereMachineConfig indicates an expected call of GetEksaVSphereMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) GetEksaVSphereMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksaVSphereMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetEksaVSphereMachineConfig), arg0, arg1, arg2, arg3)
}

// GetEtcdadmCluster mocks base method.
func (m *MockProviderKubectlClient) GetEtcdadmCluster(arg0 context.Context, arg1 *types.Cluster, arg2 string, arg3 ...executables.KubectlOpt) (*v1alpha3.EtcdadmCluster, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEtcdadmCluster", varargs...)
	ret0, _ := ret[0].(*v1alpha3.EtcdadmCluster)
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
func (m *MockProviderKubectlClient) GetKubeadmControlPlane(arg0 context.Context, arg1 *types.Cluster, arg2 string, arg3 ...executables.KubectlOpt) (*v1alpha31.KubeadmControlPlane, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetKubeadmControlPlane", varargs...)
	ret0, _ := ret[0].(*v1alpha31.KubeadmControlPlane)
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
func (m *MockProviderKubectlClient) GetMachineDeployment(arg0 context.Context, arg1 *types.Cluster, arg2 string, arg3 ...executables.KubectlOpt) (*v1alpha30.MachineDeployment, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMachineDeployment", varargs...)
	ret0, _ := ret[0].(*v1alpha30.MachineDeployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachineDeployment indicates an expected call of GetMachineDeployment.
func (mr *MockProviderKubectlClientMockRecorder) GetMachineDeployment(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachineDeployment", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetMachineDeployment), varargs...)
}

// GetNamespace mocks base method.
func (m *MockProviderKubectlClient) GetNamespace(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetNamespace indicates an expected call of GetNamespace.
func (mr *MockProviderKubectlClientMockRecorder) GetNamespace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetNamespace), arg0, arg1, arg2)
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

// SearchVsphereDatacenterConfig mocks base method.
func (m *MockProviderKubectlClient) SearchVsphereDatacenterConfig(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1alpha1.VSphereDatacenterConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchVsphereDatacenterConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1alpha1.VSphereDatacenterConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchVsphereDatacenterConfig indicates an expected call of SearchVsphereDatacenterConfig.
func (mr *MockProviderKubectlClientMockRecorder) SearchVsphereDatacenterConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchVsphereDatacenterConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).SearchVsphereDatacenterConfig), arg0, arg1, arg2, arg3)
}

// SearchVsphereMachineConfig mocks base method.
func (m *MockProviderKubectlClient) SearchVsphereMachineConfig(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1alpha1.VSphereMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchVsphereMachineConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1alpha1.VSphereMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchVsphereMachineConfig indicates an expected call of SearchVsphereMachineConfig.
func (mr *MockProviderKubectlClientMockRecorder) SearchVsphereMachineConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchVsphereMachineConfig", reflect.TypeOf((*MockProviderKubectlClient)(nil).SearchVsphereMachineConfig), arg0, arg1, arg2, arg3)
}

// SetDaemonSetImage mocks base method.
func (m *MockProviderKubectlClient) SetDaemonSetImage(arg0 context.Context, arg1, arg2, arg3, arg4, arg5 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDaemonSetImage", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDaemonSetImage indicates an expected call of SetDaemonSetImage.
func (mr *MockProviderKubectlClientMockRecorder) SetDaemonSetImage(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDaemonSetImage", reflect.TypeOf((*MockProviderKubectlClient)(nil).SetDaemonSetImage), arg0, arg1, arg2, arg3, arg4, arg5)
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

// MockClusterResourceSetManager is a mock of ClusterResourceSetManager interface.
type MockClusterResourceSetManager struct {
	ctrl     *gomock.Controller
	recorder *MockClusterResourceSetManagerMockRecorder
}

// MockClusterResourceSetManagerMockRecorder is the mock recorder for MockClusterResourceSetManager.
type MockClusterResourceSetManagerMockRecorder struct {
	mock *MockClusterResourceSetManager
}

// NewMockClusterResourceSetManager creates a new mock instance.
func NewMockClusterResourceSetManager(ctrl *gomock.Controller) *MockClusterResourceSetManager {
	mock := &MockClusterResourceSetManager{ctrl: ctrl}
	mock.recorder = &MockClusterResourceSetManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClusterResourceSetManager) EXPECT() *MockClusterResourceSetManagerMockRecorder {
	return m.recorder
}

// ForceUpdate mocks base method.
func (m *MockClusterResourceSetManager) ForceUpdate(arg0 context.Context, arg1, arg2 string, arg3, arg4 *types.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceUpdate", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForceUpdate indicates an expected call of ForceUpdate.
func (mr *MockClusterResourceSetManagerMockRecorder) ForceUpdate(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceUpdate", reflect.TypeOf((*MockClusterResourceSetManager)(nil).ForceUpdate), arg0, arg1, arg2, arg3, arg4)
}
