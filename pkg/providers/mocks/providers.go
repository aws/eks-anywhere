// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers (interfaces: Provider,DatacenterConfig,MachineConfig)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	bootstrapper "github.com/aws/eks-anywhere/pkg/bootstrapper"
	cluster "github.com/aws/eks-anywhere/pkg/cluster"
	providers "github.com/aws/eks-anywhere/pkg/providers"
	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
)

// MockProvider is a mock of Provider interface.
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
}

// MockProviderMockRecorder is the mock recorder for MockProvider.
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance.
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// BootstrapClusterOpts mocks base method.
func (m *MockProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BootstrapClusterOpts")
	ret0, _ := ret[0].([]bootstrapper.BootstrapClusterOption)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BootstrapClusterOpts indicates an expected call of BootstrapClusterOpts.
func (mr *MockProviderMockRecorder) BootstrapClusterOpts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BootstrapClusterOpts", reflect.TypeOf((*MockProvider)(nil).BootstrapClusterOpts))
}

// BootstrapSetup mocks base method.
func (m *MockProvider) BootstrapSetup(arg0 context.Context, arg1 *v1alpha1.Cluster, arg2 *types.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BootstrapSetup", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BootstrapSetup indicates an expected call of BootstrapSetup.
func (mr *MockProviderMockRecorder) BootstrapSetup(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BootstrapSetup", reflect.TypeOf((*MockProvider)(nil).BootstrapSetup), arg0, arg1, arg2)
}

// CleanupProviderInfrastructure mocks base method.
func (m *MockProvider) CleanupProviderInfrastructure(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanupProviderInfrastructure", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CleanupProviderInfrastructure indicates an expected call of CleanupProviderInfrastructure.
func (mr *MockProviderMockRecorder) CleanupProviderInfrastructure(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupProviderInfrastructure", reflect.TypeOf((*MockProvider)(nil).CleanupProviderInfrastructure), arg0)
}

// DatacenterConfig mocks base method.
func (m *MockProvider) DatacenterConfig() providers.DatacenterConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DatacenterConfig")
	ret0, _ := ret[0].(providers.DatacenterConfig)
	return ret0
}

// DatacenterConfig indicates an expected call of DatacenterConfig.
func (mr *MockProviderMockRecorder) DatacenterConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DatacenterConfig", reflect.TypeOf((*MockProvider)(nil).DatacenterConfig))
}

// DatacenterResourceType mocks base method.
func (m *MockProvider) DatacenterResourceType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DatacenterResourceType")
	ret0, _ := ret[0].(string)
	return ret0
}

// DatacenterResourceType indicates an expected call of DatacenterResourceType.
func (mr *MockProviderMockRecorder) DatacenterResourceType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DatacenterResourceType", reflect.TypeOf((*MockProvider)(nil).DatacenterResourceType))
}

// EnvMap mocks base method.
func (m *MockProvider) EnvMap() (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnvMap")
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnvMap indicates an expected call of EnvMap.
func (mr *MockProviderMockRecorder) EnvMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnvMap", reflect.TypeOf((*MockProvider)(nil).EnvMap))
}

// GenerateClusterApiSpecForCreate mocks base method.
func (m *MockProvider) GenerateClusterApiSpecForCreate(arg0 context.Context, arg1 *types.Cluster, arg2 *cluster.Spec) ([]byte, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateClusterApiSpecForCreate", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GenerateClusterApiSpecForCreate indicates an expected call of GenerateClusterApiSpecForCreate.
func (mr *MockProviderMockRecorder) GenerateClusterApiSpecForCreate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateClusterApiSpecForCreate", reflect.TypeOf((*MockProvider)(nil).GenerateClusterApiSpecForCreate), arg0, arg1, arg2)
}

// GenerateClusterApiSpecForUpgrade mocks base method.
func (m *MockProvider) GenerateClusterApiSpecForUpgrade(arg0 context.Context, arg1, arg2 *types.Cluster, arg3 *cluster.Spec) ([]byte, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateClusterApiSpecForUpgrade", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GenerateClusterApiSpecForUpgrade indicates an expected call of GenerateClusterApiSpecForUpgrade.
func (mr *MockProviderMockRecorder) GenerateClusterApiSpecForUpgrade(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateClusterApiSpecForUpgrade", reflect.TypeOf((*MockProvider)(nil).GenerateClusterApiSpecForUpgrade), arg0, arg1, arg2, arg3)
}

// GenerateMHC mocks base method.
func (m *MockProvider) GenerateMHC() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateMHC")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateMHC indicates an expected call of GenerateMHC.
func (mr *MockProviderMockRecorder) GenerateMHC() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateMHC", reflect.TypeOf((*MockProvider)(nil).GenerateMHC))
}

// GenerateStorageClass mocks base method.
func (m *MockProvider) GenerateStorageClass() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateStorageClass")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GenerateStorageClass indicates an expected call of GenerateStorageClass.
func (mr *MockProviderMockRecorder) GenerateStorageClass() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateStorageClass", reflect.TypeOf((*MockProvider)(nil).GenerateStorageClass))
}

// GetDeployments mocks base method.
func (m *MockProvider) GetDeployments() map[string][]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeployments")
	ret0, _ := ret[0].(map[string][]string)
	return ret0
}

// GetDeployments indicates an expected call of GetDeployments.
func (mr *MockProviderMockRecorder) GetDeployments() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeployments", reflect.TypeOf((*MockProvider)(nil).GetDeployments))
}

// GetInfrastructureBundle mocks base method.
func (m *MockProvider) GetInfrastructureBundle(arg0 *cluster.Spec) *types.InfrastructureBundle {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfrastructureBundle", arg0)
	ret0, _ := ret[0].(*types.InfrastructureBundle)
	return ret0
}

// GetInfrastructureBundle indicates an expected call of GetInfrastructureBundle.
func (mr *MockProviderMockRecorder) GetInfrastructureBundle(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfrastructureBundle", reflect.TypeOf((*MockProvider)(nil).GetInfrastructureBundle), arg0)
}

// MachineConfigs mocks base method.
func (m *MockProvider) MachineConfigs() []providers.MachineConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineConfigs")
	ret0, _ := ret[0].([]providers.MachineConfig)
	return ret0
}

// MachineConfigs indicates an expected call of MachineConfigs.
func (mr *MockProviderMockRecorder) MachineConfigs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineConfigs", reflect.TypeOf((*MockProvider)(nil).MachineConfigs))
}

// MachineResourceType mocks base method.
func (m *MockProvider) MachineResourceType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineResourceType")
	ret0, _ := ret[0].(string)
	return ret0
}

// MachineResourceType indicates an expected call of MachineResourceType.
func (mr *MockProviderMockRecorder) MachineResourceType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineResourceType", reflect.TypeOf((*MockProvider)(nil).MachineResourceType))
}

// Name mocks base method.
func (m *MockProvider) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockProviderMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockProvider)(nil).Name))
}

// SetupAndValidateCreateCluster mocks base method.
func (m *MockProvider) SetupAndValidateCreateCluster(arg0 context.Context, arg1 *cluster.Spec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupAndValidateCreateCluster", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupAndValidateCreateCluster indicates an expected call of SetupAndValidateCreateCluster.
func (mr *MockProviderMockRecorder) SetupAndValidateCreateCluster(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupAndValidateCreateCluster", reflect.TypeOf((*MockProvider)(nil).SetupAndValidateCreateCluster), arg0, arg1)
}

// SetupAndValidateDeleteCluster mocks base method.
func (m *MockProvider) SetupAndValidateDeleteCluster(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupAndValidateDeleteCluster", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupAndValidateDeleteCluster indicates an expected call of SetupAndValidateDeleteCluster.
func (mr *MockProviderMockRecorder) SetupAndValidateDeleteCluster(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupAndValidateDeleteCluster", reflect.TypeOf((*MockProvider)(nil).SetupAndValidateDeleteCluster), arg0)
}

// SetupAndValidateUpgradeCluster mocks base method.
func (m *MockProvider) SetupAndValidateUpgradeCluster(arg0 context.Context, arg1 *cluster.Spec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupAndValidateUpgradeCluster", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupAndValidateUpgradeCluster indicates an expected call of SetupAndValidateUpgradeCluster.
func (mr *MockProviderMockRecorder) SetupAndValidateUpgradeCluster(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupAndValidateUpgradeCluster", reflect.TypeOf((*MockProvider)(nil).SetupAndValidateUpgradeCluster), arg0, arg1)
}

// UpdateKubeConfig mocks base method.
func (m *MockProvider) UpdateKubeConfig(arg0 *[]byte, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateKubeConfig", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateKubeConfig indicates an expected call of UpdateKubeConfig.
func (mr *MockProviderMockRecorder) UpdateKubeConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateKubeConfig", reflect.TypeOf((*MockProvider)(nil).UpdateKubeConfig), arg0, arg1)
}

// UpdateSecrets mocks base method.
func (m *MockProvider) UpdateSecrets(arg0 context.Context, arg1 *types.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSecrets", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSecrets indicates an expected call of UpdateSecrets.
func (mr *MockProviderMockRecorder) UpdateSecrets(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSecrets", reflect.TypeOf((*MockProvider)(nil).UpdateSecrets), arg0, arg1)
}

// ValidateNewSpec mocks base method.
func (m *MockProvider) ValidateNewSpec(arg0 context.Context, arg1 *types.Cluster, arg2 *cluster.Spec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateNewSpec", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateNewSpec indicates an expected call of ValidateNewSpec.
func (mr *MockProviderMockRecorder) ValidateNewSpec(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateNewSpec", reflect.TypeOf((*MockProvider)(nil).ValidateNewSpec), arg0, arg1, arg2)
}

// Version mocks base method.
func (m *MockProvider) Version(arg0 *cluster.Spec) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Version", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// Version indicates an expected call of Version.
func (mr *MockProviderMockRecorder) Version(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Version", reflect.TypeOf((*MockProvider)(nil).Version), arg0)
}

// MockDatacenterConfig is a mock of DatacenterConfig interface.
type MockDatacenterConfig struct {
	ctrl     *gomock.Controller
	recorder *MockDatacenterConfigMockRecorder
}

// MockDatacenterConfigMockRecorder is the mock recorder for MockDatacenterConfig.
type MockDatacenterConfigMockRecorder struct {
	mock *MockDatacenterConfig
}

// NewMockDatacenterConfig creates a new mock instance.
func NewMockDatacenterConfig(ctrl *gomock.Controller) *MockDatacenterConfig {
	mock := &MockDatacenterConfig{ctrl: ctrl}
	mock.recorder = &MockDatacenterConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatacenterConfig) EXPECT() *MockDatacenterConfigMockRecorder {
	return m.recorder
}

// ClearPauseAnnotation mocks base method.
func (m *MockDatacenterConfig) ClearPauseAnnotation() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClearPauseAnnotation")
}

// ClearPauseAnnotation indicates an expected call of ClearPauseAnnotation.
func (mr *MockDatacenterConfigMockRecorder) ClearPauseAnnotation() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearPauseAnnotation", reflect.TypeOf((*MockDatacenterConfig)(nil).ClearPauseAnnotation))
}

// Kind mocks base method.
func (m *MockDatacenterConfig) Kind() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind.
func (mr *MockDatacenterConfigMockRecorder) Kind() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockDatacenterConfig)(nil).Kind))
}

// PauseReconcile mocks base method.
func (m *MockDatacenterConfig) PauseReconcile() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PauseReconcile")
}

// PauseReconcile indicates an expected call of PauseReconcile.
func (mr *MockDatacenterConfigMockRecorder) PauseReconcile() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PauseReconcile", reflect.TypeOf((*MockDatacenterConfig)(nil).PauseReconcile))
}

// MockMachineConfig is a mock of MachineConfig interface.
type MockMachineConfig struct {
	ctrl     *gomock.Controller
	recorder *MockMachineConfigMockRecorder
}

// MockMachineConfigMockRecorder is the mock recorder for MockMachineConfig.
type MockMachineConfigMockRecorder struct {
	mock *MockMachineConfig
}

// NewMockMachineConfig creates a new mock instance.
func NewMockMachineConfig(ctrl *gomock.Controller) *MockMachineConfig {
	mock := &MockMachineConfig{ctrl: ctrl}
	mock.recorder = &MockMachineConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMachineConfig) EXPECT() *MockMachineConfigMockRecorder {
	return m.recorder
}
