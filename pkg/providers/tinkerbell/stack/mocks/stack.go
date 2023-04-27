// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/providers/tinkerbell/stack/stack.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	executables "github.com/aws/eks-anywhere/pkg/executables"
	stack "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	v1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	gomock "github.com/golang/mock/gomock"
)

// MockDocker is a mock of Docker interface.
type MockDocker struct {
	ctrl     *gomock.Controller
	recorder *MockDockerMockRecorder
}

// MockDockerMockRecorder is the mock recorder for MockDocker.
type MockDockerMockRecorder struct {
	mock *MockDocker
}

// NewMockDocker creates a new mock instance.
func NewMockDocker(ctrl *gomock.Controller) *MockDocker {
	mock := &MockDocker{ctrl: ctrl}
	mock.recorder = &MockDockerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDocker) EXPECT() *MockDockerMockRecorder {
	return m.recorder
}

// CheckContainerExistence mocks base method.
func (m *MockDocker) CheckContainerExistence(ctx context.Context, name string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckContainerExistence", ctx, name)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckContainerExistence indicates an expected call of CheckContainerExistence.
func (mr *MockDockerMockRecorder) CheckContainerExistence(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckContainerExistence", reflect.TypeOf((*MockDocker)(nil).CheckContainerExistence), ctx, name)
}

// ForceRemove mocks base method.
func (m *MockDocker) ForceRemove(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceRemove", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForceRemove indicates an expected call of ForceRemove.
func (mr *MockDockerMockRecorder) ForceRemove(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceRemove", reflect.TypeOf((*MockDocker)(nil).ForceRemove), ctx, name)
}

// Run mocks base method.
func (m *MockDocker) Run(ctx context.Context, image, name string, cmd []string, flags ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, image, name, cmd}
	for _, a := range flags {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockDockerMockRecorder) Run(ctx, image, name, cmd interface{}, flags ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, image, name, cmd}, flags...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockDocker)(nil).Run), varargs...)
}

// MockHelm is a mock of Helm interface.
type MockHelm struct {
	ctrl     *gomock.Controller
	recorder *MockHelmMockRecorder
}

// MockHelmMockRecorder is the mock recorder for MockHelm.
type MockHelmMockRecorder struct {
	mock *MockHelm
}

// NewMockHelm creates a new mock instance.
func NewMockHelm(ctrl *gomock.Controller) *MockHelm {
	mock := &MockHelm{ctrl: ctrl}
	mock.recorder = &MockHelmMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHelm) EXPECT() *MockHelmMockRecorder {
	return m.recorder
}

// InstallChartWithValuesFile mocks base method.
func (m *MockHelm) InstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstallChartWithValuesFile", ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath)
	ret0, _ := ret[0].(error)
	return ret0
}

// InstallChartWithValuesFile indicates an expected call of InstallChartWithValuesFile.
func (mr *MockHelmMockRecorder) InstallChartWithValuesFile(ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstallChartWithValuesFile", reflect.TypeOf((*MockHelm)(nil).InstallChartWithValuesFile), ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath)
}

// RegistryLogin mocks base method.
func (m *MockHelm) RegistryLogin(ctx context.Context, endpoint, username, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegistryLogin", ctx, endpoint, username, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegistryLogin indicates an expected call of RegistryLogin.
func (mr *MockHelmMockRecorder) RegistryLogin(ctx, endpoint, username, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegistryLogin", reflect.TypeOf((*MockHelm)(nil).RegistryLogin), ctx, endpoint, username, password)
}

// UpgradeChartWithValuesFile mocks base method.
func (m *MockHelm) UpgradeChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string, opts ...executables.HelmOpt) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpgradeChartWithValuesFile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpgradeChartWithValuesFile indicates an expected call of UpgradeChartWithValuesFile.
func (mr *MockHelmMockRecorder) UpgradeChartWithValuesFile(ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, chart, ociURI, version, kubeconfigFilePath, valuesFilePath}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpgradeChartWithValuesFile", reflect.TypeOf((*MockHelm)(nil).UpgradeChartWithValuesFile), varargs...)
}

// MockStackInstaller is a mock of StackInstaller interface.
type MockStackInstaller struct {
	ctrl     *gomock.Controller
	recorder *MockStackInstallerMockRecorder
}

// MockStackInstallerMockRecorder is the mock recorder for MockStackInstaller.
type MockStackInstallerMockRecorder struct {
	mock *MockStackInstaller
}

// NewMockStackInstaller creates a new mock instance.
func NewMockStackInstaller(ctrl *gomock.Controller) *MockStackInstaller {
	mock := &MockStackInstaller{ctrl: ctrl}
	mock.recorder = &MockStackInstallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStackInstaller) EXPECT() *MockStackInstallerMockRecorder {
	return m.recorder
}

// AddNoProxyIP mocks base method.
func (m *MockStackInstaller) AddNoProxyIP(IP string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddNoProxyIP", IP)
}

// AddNoProxyIP indicates an expected call of AddNoProxyIP.
func (mr *MockStackInstallerMockRecorder) AddNoProxyIP(IP interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNoProxyIP", reflect.TypeOf((*MockStackInstaller)(nil).AddNoProxyIP), IP)
}

// CleanupLocalBoots mocks base method.
func (m *MockStackInstaller) CleanupLocalBoots(ctx context.Context, forceCleanup bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanupLocalBoots", ctx, forceCleanup)
	ret0, _ := ret[0].(error)
	return ret0
}

// CleanupLocalBoots indicates an expected call of CleanupLocalBoots.
func (mr *MockStackInstallerMockRecorder) CleanupLocalBoots(ctx, forceCleanup interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupLocalBoots", reflect.TypeOf((*MockStackInstaller)(nil).CleanupLocalBoots), ctx, forceCleanup)
}

// GetNamespace mocks base method.
func (m *MockStackInstaller) GetNamespace() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespace")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetNamespace indicates an expected call of GetNamespace.
func (mr *MockStackInstallerMockRecorder) GetNamespace() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MockStackInstaller)(nil).GetNamespace))
}

// Install mocks base method.
func (m *MockStackInstaller) Install(ctx context.Context, bundle v1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...stack.InstallOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, bundle, tinkerbellIP, kubeconfig, hookOverride}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Install", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install.
func (mr *MockStackInstallerMockRecorder) Install(ctx, bundle, tinkerbellIP, kubeconfig, hookOverride interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, bundle, tinkerbellIP, kubeconfig, hookOverride}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockStackInstaller)(nil).Install), varargs...)
}

// UninstallLocal mocks base method.
func (m *MockStackInstaller) UninstallLocal(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UninstallLocal", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// UninstallLocal indicates an expected call of UninstallLocal.
func (mr *MockStackInstallerMockRecorder) UninstallLocal(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UninstallLocal", reflect.TypeOf((*MockStackInstaller)(nil).UninstallLocal), ctx)
}

// Upgrade mocks base method.
func (m *MockStackInstaller) Upgrade(arg0 context.Context, arg1 v1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upgrade", arg0, arg1, tinkerbellIP, kubeconfig)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upgrade indicates an expected call of Upgrade.
func (mr *MockStackInstallerMockRecorder) Upgrade(arg0, arg1, tinkerbellIP, kubeconfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upgrade", reflect.TypeOf((*MockStackInstaller)(nil).Upgrade), arg0, arg1, tinkerbellIP, kubeconfig)
}
