// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/filewriter (interfaces: FileWriter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	filewriter "github.com/aws/eks-anywhere/pkg/filewriter"
	gomock "github.com/golang/mock/gomock"
)

// MockFileWriter is a mock of FileWriter interface.
type MockFileWriter struct {
	ctrl     *gomock.Controller
	recorder *MockFileWriterMockRecorder
}

// MockFileWriterMockRecorder is the mock recorder for MockFileWriter.
type MockFileWriterMockRecorder struct {
	mock *MockFileWriter
}

// NewMockFileWriter creates a new mock instance.
func NewMockFileWriter(ctrl *gomock.Controller) *MockFileWriter {
	mock := &MockFileWriter{ctrl: ctrl}
	mock.recorder = &MockFileWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileWriter) EXPECT() *MockFileWriterMockRecorder {
	return m.recorder
}

// CleanUp mocks base method.
func (m *MockFileWriter) CleanUp() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CleanUp")
}

// CleanUp indicates an expected call of CleanUp.
func (mr *MockFileWriterMockRecorder) CleanUp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanUp", reflect.TypeOf((*MockFileWriter)(nil).CleanUp))
}

// CleanUpTemp mocks base method.
func (m *MockFileWriter) CleanUpTemp() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CleanUpTemp")
}

// CleanUpTemp indicates an expected call of CleanUpTemp.
func (mr *MockFileWriterMockRecorder) CleanUpTemp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanUpTemp", reflect.TypeOf((*MockFileWriter)(nil).CleanUpTemp))
}

// Copy mocks base method.
func (m *MockFileWriter) Copy(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Copy", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Copy indicates an expected call of Copy.
func (mr *MockFileWriterMockRecorder) Copy(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Copy", reflect.TypeOf((*MockFileWriter)(nil).Copy), arg0, arg1)
}

// Dir mocks base method.
func (m *MockFileWriter) Dir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dir")
	ret0, _ := ret[0].(string)
	return ret0
}

// Dir indicates an expected call of Dir.
func (mr *MockFileWriterMockRecorder) Dir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dir", reflect.TypeOf((*MockFileWriter)(nil).Dir))
}

// WithDir mocks base method.
func (m *MockFileWriter) WithDir(arg0 string) (filewriter.FileWriter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithDir", arg0)
	ret0, _ := ret[0].(filewriter.FileWriter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithDir indicates an expected call of WithDir.
func (mr *MockFileWriterMockRecorder) WithDir(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithDir", reflect.TypeOf((*MockFileWriter)(nil).WithDir), arg0)
}

// Write mocks base method.
func (m *MockFileWriter) Write(arg0 string, arg1 []byte, arg2 ...filewriter.FileOptionsFunc) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Write", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockFileWriterMockRecorder) Write(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockFileWriter)(nil).Write), varargs...)
}
