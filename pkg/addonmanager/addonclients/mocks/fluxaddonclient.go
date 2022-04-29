// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/pkg/addonmanager/addonclients (interfaces: Flux)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	types "github.com/aws/eks-anywhere/pkg/types"
	gomock "github.com/golang/mock/gomock"
)

// MockFlux is a mock of Flux interface.
type MockFlux struct {
	ctrl     *gomock.Controller
	recorder *MockFluxMockRecorder
}

// MockFluxMockRecorder is the mock recorder for MockFlux.
type MockFluxMockRecorder struct {
	mock *MockFlux
}

// NewMockFlux creates a new mock instance.
func NewMockFlux(ctrl *gomock.Controller) *MockFlux {
	mock := &MockFlux{ctrl: ctrl}
	mock.recorder = &MockFluxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFlux) EXPECT() *MockFluxMockRecorder {
	return m.recorder
}

// BootstrapToolkitsComponentsGithub mocks base method.
func (m *MockFlux) BootstrapToolkitsComponentsGithub(arg0 context.Context, arg1 *types.Cluster, arg2 *v1alpha1.FluxConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BootstrapToolkitsComponentsGithub", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BootstrapToolkitsComponentsGithub indicates an expected call of BootstrapToolkitsComponentsGithub.
func (mr *MockFluxMockRecorder) BootstrapToolkitsComponentsGithub(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BootstrapToolkitsComponentsGithub", reflect.TypeOf((*MockFlux)(nil).BootstrapToolkitsComponentsGithub), arg0, arg1, arg2)
}

// DeleteFluxSystemSecret mocks base method.
func (m *MockFlux) DeleteFluxSystemSecret(arg0 context.Context, arg1 *types.Cluster, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFluxSystemSecret", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFluxSystemSecret indicates an expected call of DeleteFluxSystemSecret.
func (mr *MockFluxMockRecorder) DeleteFluxSystemSecret(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFluxSystemSecret", reflect.TypeOf((*MockFlux)(nil).DeleteFluxSystemSecret), arg0, arg1, arg2)
}

// ForceReconcileGitRepo mocks base method.
func (m *MockFlux) ForceReconcileGitRepo(arg0 context.Context, arg1 *types.Cluster, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceReconcileGitRepo", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForceReconcileGitRepo indicates an expected call of ForceReconcileGitRepo.
func (mr *MockFluxMockRecorder) ForceReconcileGitRepo(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceReconcileGitRepo", reflect.TypeOf((*MockFlux)(nil).ForceReconcileGitRepo), arg0, arg1, arg2)
}

// PauseKustomization mocks base method.
func (m *MockFlux) PauseKustomization(arg0 context.Context, arg1 *types.Cluster, arg2 *v1alpha1.FluxConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PauseKustomization", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PauseKustomization indicates an expected call of PauseKustomization.
func (mr *MockFluxMockRecorder) PauseKustomization(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PauseKustomization", reflect.TypeOf((*MockFlux)(nil).PauseKustomization), arg0, arg1, arg2)
}

// Reconcile mocks base method.
func (m *MockFlux) Reconcile(arg0 context.Context, arg1 *types.Cluster, arg2 *v1alpha1.FluxConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile.
func (mr *MockFluxMockRecorder) Reconcile(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockFlux)(nil).Reconcile), arg0, arg1, arg2)
}

// ResumeKustomization mocks base method.
func (m *MockFlux) ResumeKustomization(arg0 context.Context, arg1 *types.Cluster, arg2 *v1alpha1.FluxConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResumeKustomization", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResumeKustomization indicates an expected call of ResumeKustomization.
func (mr *MockFluxMockRecorder) ResumeKustomization(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResumeKustomization", reflect.TypeOf((*MockFlux)(nil).ResumeKustomization), arg0, arg1, arg2)
}

// UninstallToolkitsComponents mocks base method.
func (m *MockFlux) UninstallToolkitsComponents(arg0 context.Context, arg1 *types.Cluster, arg2 *v1alpha1.FluxConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UninstallToolkitsComponents", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UninstallToolkitsComponents indicates an expected call of UninstallToolkitsComponents.
func (mr *MockFluxMockRecorder) UninstallToolkitsComponents(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UninstallToolkitsComponents", reflect.TypeOf((*MockFlux)(nil).UninstallToolkitsComponents), arg0, arg1, arg2)
}
