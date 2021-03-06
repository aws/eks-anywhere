// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/curatedpackages/bundle.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	gomock "github.com/golang/mock/gomock"
)

// MockReader is a mock of Reader interface.
type MockReader struct {
	ctrl     *gomock.Controller
	recorder *MockReaderMockRecorder
}

// MockReaderMockRecorder is the mock recorder for MockReader.
type MockReaderMockRecorder struct {
	mock *MockReader
}

// NewMockReader creates a new mock instance.
func NewMockReader(ctrl *gomock.Controller) *MockReader {
	mock := &MockReader{ctrl: ctrl}
	mock.recorder = &MockReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReader) EXPECT() *MockReaderMockRecorder {
	return m.recorder
}

// ReadBundlesForVersion mocks base method.
func (m *MockReader) ReadBundlesForVersion(eksaVersion string) (*v1alpha1.Bundles, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadBundlesForVersion", eksaVersion)
	ret0, _ := ret[0].(*v1alpha1.Bundles)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadBundlesForVersion indicates an expected call of ReadBundlesForVersion.
func (mr *MockReaderMockRecorder) ReadBundlesForVersion(eksaVersion interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadBundlesForVersion", reflect.TypeOf((*MockReader)(nil).ReadBundlesForVersion), eksaVersion)
}

// MockBundleRegistry is a mock of BundleRegistry interface.
type MockBundleRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockBundleRegistryMockRecorder
}

// MockBundleRegistryMockRecorder is the mock recorder for MockBundleRegistry.
type MockBundleRegistryMockRecorder struct {
	mock *MockBundleRegistry
}

// NewMockBundleRegistry creates a new mock instance.
func NewMockBundleRegistry(ctrl *gomock.Controller) *MockBundleRegistry {
	mock := &MockBundleRegistry{ctrl: ctrl}
	mock.recorder = &MockBundleRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBundleRegistry) EXPECT() *MockBundleRegistryMockRecorder {
	return m.recorder
}

// GetRegistryBaseRef mocks base method.
func (m *MockBundleRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegistryBaseRef", ctx)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRegistryBaseRef indicates an expected call of GetRegistryBaseRef.
func (mr *MockBundleRegistryMockRecorder) GetRegistryBaseRef(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegistryBaseRef", reflect.TypeOf((*MockBundleRegistry)(nil).GetRegistryBaseRef), ctx)
}
