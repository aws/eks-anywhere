// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/providers/tinkerbell (interfaces: ProviderKubectlClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
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
