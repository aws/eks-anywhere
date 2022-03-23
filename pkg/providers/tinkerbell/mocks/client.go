// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers/tinkerbell (interfaces: ProviderKubectlClient,ProviderTinkClient,ProviderPbnjClient,SSHAuthKeyGenerator)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	executables "github.com/aws/eks-anywhere/pkg/executables"
	filewriter "github.com/aws/eks-anywhere/pkg/filewriter"
	pbnj "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	gomock "github.com/golang/mock/gomock"
	hardware "github.com/tinkerbell/tink/protos/hardware"
	workflow "github.com/tinkerbell/tink/protos/workflow"
	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

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

// ApplyHardware mocks base method.
func (m *MockProviderKubectlClient) ApplyHardware(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyHardware", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyHardware indicates an expected call of ApplyHardware.
func (mr *MockProviderKubectlClientMockRecorder) ApplyHardware(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyHardware", reflect.TypeOf((*MockProviderKubectlClient)(nil).ApplyHardware), arg0, arg1, arg2)
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

// GetMachineDeployment mocks base method.
func (m *MockProviderKubectlClient) GetMachineDeployment(arg0 context.Context, arg1 string, arg2 ...executables.KubectlOpt) (*v1beta1.MachineDeployment, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMachineDeployment", varargs...)
	ret0, _ := ret[0].(*v1beta1.MachineDeployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachineDeployment indicates an expected call of GetMachineDeployment.
func (mr *MockProviderKubectlClientMockRecorder) GetMachineDeployment(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachineDeployment", reflect.TypeOf((*MockProviderKubectlClient)(nil).GetMachineDeployment), varargs...)
}

// MockProviderTinkClient is a mock of ProviderTinkClient interface.
type MockProviderTinkClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderTinkClientMockRecorder
}

// MockProviderTinkClientMockRecorder is the mock recorder for MockProviderTinkClient.
type MockProviderTinkClientMockRecorder struct {
	mock *MockProviderTinkClient
}

// NewMockProviderTinkClient creates a new mock instance.
func NewMockProviderTinkClient(ctrl *gomock.Controller) *MockProviderTinkClient {
	mock := &MockProviderTinkClient{ctrl: ctrl}
	mock.recorder = &MockProviderTinkClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderTinkClient) EXPECT() *MockProviderTinkClientMockRecorder {
	return m.recorder
}

// GetHardware mocks base method.
func (m *MockProviderTinkClient) GetHardware(arg0 context.Context) ([]*hardware.Hardware, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHardware", arg0)
	ret0, _ := ret[0].([]*hardware.Hardware)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHardware indicates an expected call of GetHardware.
func (mr *MockProviderTinkClientMockRecorder) GetHardware(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHardware", reflect.TypeOf((*MockProviderTinkClient)(nil).GetHardware), arg0)
}

// GetWorkflow mocks base method.
func (m *MockProviderTinkClient) GetWorkflow(arg0 context.Context) ([]*workflow.Workflow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkflow", arg0)
	ret0, _ := ret[0].([]*workflow.Workflow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWorkflow indicates an expected call of GetWorkflow.
func (mr *MockProviderTinkClientMockRecorder) GetWorkflow(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkflow", reflect.TypeOf((*MockProviderTinkClient)(nil).GetWorkflow), arg0)
}

// MockProviderPbnjClient is a mock of ProviderPbnjClient interface.
type MockProviderPbnjClient struct {
	ctrl     *gomock.Controller
	recorder *MockProviderPbnjClientMockRecorder
}

// MockProviderPbnjClientMockRecorder is the mock recorder for MockProviderPbnjClient.
type MockProviderPbnjClientMockRecorder struct {
	mock *MockProviderPbnjClient
}

// NewMockProviderPbnjClient creates a new mock instance.
func NewMockProviderPbnjClient(ctrl *gomock.Controller) *MockProviderPbnjClient {
	mock := &MockProviderPbnjClient{ctrl: ctrl}
	mock.recorder = &MockProviderPbnjClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderPbnjClient) EXPECT() *MockProviderPbnjClientMockRecorder {
	return m.recorder
}

// GetPowerState mocks base method.
func (m *MockProviderPbnjClient) GetPowerState(arg0 context.Context, arg1 pbnj.BmcSecretConfig) (pbnj.PowerState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPowerState", arg0, arg1)
	ret0, _ := ret[0].(pbnj.PowerState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPowerState indicates an expected call of GetPowerState.
func (mr *MockProviderPbnjClientMockRecorder) GetPowerState(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPowerState", reflect.TypeOf((*MockProviderPbnjClient)(nil).GetPowerState), arg0, arg1)
}

// MockSSHAuthKeyGenerator is a mock of SSHAuthKeyGenerator interface.
type MockSSHAuthKeyGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockSSHAuthKeyGeneratorMockRecorder
}

// MockSSHAuthKeyGeneratorMockRecorder is the mock recorder for MockSSHAuthKeyGenerator.
type MockSSHAuthKeyGeneratorMockRecorder struct {
	mock *MockSSHAuthKeyGenerator
}

// NewMockSSHAuthKeyGenerator creates a new mock instance.
func NewMockSSHAuthKeyGenerator(ctrl *gomock.Controller) *MockSSHAuthKeyGenerator {
	mock := &MockSSHAuthKeyGenerator{ctrl: ctrl}
	mock.recorder = &MockSSHAuthKeyGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSSHAuthKeyGenerator) EXPECT() *MockSSHAuthKeyGeneratorMockRecorder {
	return m.recorder
}

// GenerateSSHAuthKey mocks base method.
func (m *MockSSHAuthKeyGenerator) GenerateSSHAuthKey(arg0 filewriter.FileWriter) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateSSHAuthKey", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateSSHAuthKey indicates an expected call of GenerateSSHAuthKey.
func (mr *MockSSHAuthKeyGeneratorMockRecorder) GenerateSSHAuthKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateSSHAuthKey", reflect.TypeOf((*MockSSHAuthKeyGenerator)(nil).GenerateSSHAuthKey), arg0)
}
