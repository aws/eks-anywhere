// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/networkutils/netclient.go

// Package mocks is a generated GoMock package.
package mocks

import (
	net "net"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockNetClient is a mock of NetClient interface.
type MockNetClient struct {
	ctrl     *gomock.Controller
	recorder *MockNetClientMockRecorder
}

// MockNetClientMockRecorder is the mock recorder for MockNetClient.
type MockNetClientMockRecorder struct {
	mock *MockNetClient
}

// NewMockNetClient creates a new mock instance.
func NewMockNetClient(ctrl *gomock.Controller) *MockNetClient {
	mock := &MockNetClient{ctrl: ctrl}
	mock.recorder = &MockNetClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNetClient) EXPECT() *MockNetClientMockRecorder {
	return m.recorder
}

// DialTimeout mocks base method.
func (m *MockNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DialTimeout", network, address, timeout)
	ret0, _ := ret[0].(net.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DialTimeout indicates an expected call of DialTimeout.
func (mr *MockNetClientMockRecorder) DialTimeout(network, address, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DialTimeout", reflect.TypeOf((*MockNetClient)(nil).DialTimeout), network, address, timeout)
}
