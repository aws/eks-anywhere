// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/providers/tinkerbell/reconciler/reconciler.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	cluster "github.com/aws/eks-anywhere/pkg/cluster"
	controller "github.com/aws/eks-anywhere/pkg/controller"
	logr "github.com/go-logr/logr"
	gomock "github.com/golang/mock/gomock"
)

// MockIPValidator is a mock of IPValidator interface.
type MockIPValidator struct {
	ctrl     *gomock.Controller
	recorder *MockIPValidatorMockRecorder
}

// MockIPValidatorMockRecorder is the mock recorder for MockIPValidator.
type MockIPValidatorMockRecorder struct {
	mock *MockIPValidator
}

// NewMockIPValidator creates a new mock instance.
func NewMockIPValidator(ctrl *gomock.Controller) *MockIPValidator {
	mock := &MockIPValidator{ctrl: ctrl}
	mock.recorder = &MockIPValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIPValidator) EXPECT() *MockIPValidatorMockRecorder {
	return m.recorder
}

// ValidateControlPlaneIP mocks base method.
func (m *MockIPValidator) ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateControlPlaneIP", ctx, log, spec)
	ret0, _ := ret[0].(controller.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateControlPlaneIP indicates an expected call of ValidateControlPlaneIP.
func (mr *MockIPValidatorMockRecorder) ValidateControlPlaneIP(ctx, log, spec interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateControlPlaneIP", reflect.TypeOf((*MockIPValidator)(nil).ValidateControlPlaneIP), ctx, log, spec)
}
