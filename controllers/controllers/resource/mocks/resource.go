// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/eks-anywhere/controllers/controllers/resource (interfaces: ResourceFetcher,ResourceUpdater)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "k8s.io/apimachinery/pkg/types"
	v1alpha30 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	v1alpha31 "sigs.k8s.io/cluster-api/api/v1alpha3"
	v1alpha32 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	cluster "github.com/aws/eks-anywhere/pkg/cluster"
)

// MockResourceFetcher is a mock of ResourceFetcher interface.
type MockResourceFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockResourceFetcherMockRecorder
}

// MockResourceFetcherMockRecorder is the mock recorder for MockResourceFetcher.
type MockResourceFetcherMockRecorder struct {
	mock *MockResourceFetcher
}

// NewMockResourceFetcher creates a new mock instance.
func NewMockResourceFetcher(ctrl *gomock.Controller) *MockResourceFetcher {
	mock := &MockResourceFetcher{ctrl: ctrl}
	mock.recorder = &MockResourceFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceFetcher) EXPECT() *MockResourceFetcherMockRecorder {
	return m.recorder
}

// AWSIamConfig mocks base method.
func (m *MockResourceFetcher) AWSIamConfig(arg0 context.Context, arg1 *v1alpha1.Ref, arg2 string) (*v1alpha1.AWSIamConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AWSIamConfig", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1alpha1.AWSIamConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AWSIamConfig indicates an expected call of AWSIamConfig.
func (mr *MockResourceFetcherMockRecorder) AWSIamConfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AWSIamConfig", reflect.TypeOf((*MockResourceFetcher)(nil).AWSIamConfig), arg0, arg1, arg2)
}

// ControlPlane mocks base method.
func (m *MockResourceFetcher) ControlPlane(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha32.KubeadmControlPlane, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ControlPlane", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha32.KubeadmControlPlane)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ControlPlane indicates an expected call of ControlPlane.
func (mr *MockResourceFetcherMockRecorder) ControlPlane(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ControlPlane", reflect.TypeOf((*MockResourceFetcher)(nil).ControlPlane), arg0, arg1)
}

// Etcd mocks base method.
func (m *MockResourceFetcher) Etcd(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha3.EtcdadmCluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Etcd", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha3.EtcdadmCluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Etcd indicates an expected call of Etcd.
func (mr *MockResourceFetcherMockRecorder) Etcd(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Etcd", reflect.TypeOf((*MockResourceFetcher)(nil).Etcd), arg0, arg1)
}

// ExistingVSphereControlPlaneMachineConfig mocks base method.
func (m *MockResourceFetcher) ExistingVSphereControlPlaneMachineConfig(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha1.VSphereMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistingVSphereControlPlaneMachineConfig", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha1.VSphereMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExistingVSphereControlPlaneMachineConfig indicates an expected call of ExistingVSphereControlPlaneMachineConfig.
func (mr *MockResourceFetcherMockRecorder) ExistingVSphereControlPlaneMachineConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistingVSphereControlPlaneMachineConfig", reflect.TypeOf((*MockResourceFetcher)(nil).ExistingVSphereControlPlaneMachineConfig), arg0, arg1)
}

// ExistingVSphereDatacenterConfig mocks base method.
func (m *MockResourceFetcher) ExistingVSphereDatacenterConfig(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha1.VSphereDatacenterConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistingVSphereDatacenterConfig", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha1.VSphereDatacenterConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExistingVSphereDatacenterConfig indicates an expected call of ExistingVSphereDatacenterConfig.
func (mr *MockResourceFetcherMockRecorder) ExistingVSphereDatacenterConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistingVSphereDatacenterConfig", reflect.TypeOf((*MockResourceFetcher)(nil).ExistingVSphereDatacenterConfig), arg0, arg1)
}

// ExistingVSphereEtcdMachineConfig mocks base method.
func (m *MockResourceFetcher) ExistingVSphereEtcdMachineConfig(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha1.VSphereMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistingVSphereEtcdMachineConfig", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha1.VSphereMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExistingVSphereEtcdMachineConfig indicates an expected call of ExistingVSphereEtcdMachineConfig.
func (mr *MockResourceFetcherMockRecorder) ExistingVSphereEtcdMachineConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistingVSphereEtcdMachineConfig", reflect.TypeOf((*MockResourceFetcher)(nil).ExistingVSphereEtcdMachineConfig), arg0, arg1)
}

// ExistingVSphereWorkerMachineConfig mocks base method.
func (m *MockResourceFetcher) ExistingVSphereWorkerMachineConfig(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha1.VSphereMachineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistingVSphereWorkerMachineConfig", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha1.VSphereMachineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExistingVSphereWorkerMachineConfig indicates an expected call of ExistingVSphereWorkerMachineConfig.
func (mr *MockResourceFetcherMockRecorder) ExistingVSphereWorkerMachineConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistingVSphereWorkerMachineConfig", reflect.TypeOf((*MockResourceFetcher)(nil).ExistingVSphereWorkerMachineConfig), arg0, arg1)
}

// Fetch mocks base method.
func (m *MockResourceFetcher) Fetch(arg0 context.Context, arg1, arg2, arg3, arg4 string) (*unstructured.Unstructured, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fetch", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*unstructured.Unstructured)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Fetch indicates an expected call of Fetch.
func (mr *MockResourceFetcherMockRecorder) Fetch(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fetch", reflect.TypeOf((*MockResourceFetcher)(nil).Fetch), arg0, arg1, arg2, arg3, arg4)
}

// FetchAppliedSpec mocks base method.
func (m *MockResourceFetcher) FetchAppliedSpec(arg0 context.Context, arg1 *v1alpha1.Cluster) (*cluster.Spec, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchAppliedSpec", arg0, arg1)
	ret0, _ := ret[0].(*cluster.Spec)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchAppliedSpec indicates an expected call of FetchAppliedSpec.
func (mr *MockResourceFetcherMockRecorder) FetchAppliedSpec(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchAppliedSpec", reflect.TypeOf((*MockResourceFetcher)(nil).FetchAppliedSpec), arg0, arg1)
}

// FetchCluster mocks base method.
func (m *MockResourceFetcher) FetchCluster(arg0 context.Context, arg1 types.NamespacedName) (*v1alpha1.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchCluster", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha1.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCluster indicates an expected call of FetchCluster.
func (mr *MockResourceFetcherMockRecorder) FetchCluster(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCluster", reflect.TypeOf((*MockResourceFetcher)(nil).FetchCluster), arg0, arg1)
}

// FetchObject mocks base method.
func (m *MockResourceFetcher) FetchObject(arg0 context.Context, arg1 types.NamespacedName, arg2 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchObject", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// FetchObject indicates an expected call of FetchObject.
func (mr *MockResourceFetcherMockRecorder) FetchObject(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchObject", reflect.TypeOf((*MockResourceFetcher)(nil).FetchObject), arg0, arg1, arg2)
}

// FetchObjectByName mocks base method.
func (m *MockResourceFetcher) FetchObjectByName(arg0 context.Context, arg1, arg2 string, arg3 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchObjectByName", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// FetchObjectByName indicates an expected call of FetchObjectByName.
func (mr *MockResourceFetcherMockRecorder) FetchObjectByName(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchObjectByName", reflect.TypeOf((*MockResourceFetcher)(nil).FetchObjectByName), arg0, arg1, arg2, arg3)
}

// MachineDeployment mocks base method.
func (m *MockResourceFetcher) MachineDeployment(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha31.MachineDeployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineDeployment", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha31.MachineDeployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MachineDeployment indicates an expected call of MachineDeployment.
func (mr *MockResourceFetcherMockRecorder) MachineDeployment(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineDeployment", reflect.TypeOf((*MockResourceFetcher)(nil).MachineDeployment), arg0, arg1)
}

// OIDCConfig mocks base method.
func (m *MockResourceFetcher) OIDCConfig(arg0 context.Context, arg1 *v1alpha1.Ref, arg2 string) (*v1alpha1.OIDCConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OIDCConfig", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1alpha1.OIDCConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OIDCConfig indicates an expected call of OIDCConfig.
func (mr *MockResourceFetcherMockRecorder) OIDCConfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OIDCConfig", reflect.TypeOf((*MockResourceFetcher)(nil).OIDCConfig), arg0, arg1, arg2)
}

// VSphereWorkerMachineTemplate mocks base method.
func (m *MockResourceFetcher) VSphereWorkerMachineTemplate(arg0 context.Context, arg1 *v1alpha1.Cluster) (*v1alpha30.VSphereMachineTemplate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VSphereWorkerMachineTemplate", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha30.VSphereMachineTemplate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VSphereWorkerMachineTemplate indicates an expected call of VSphereWorkerMachineTemplate.
func (mr *MockResourceFetcherMockRecorder) VSphereWorkerMachineTemplate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VSphereWorkerMachineTemplate", reflect.TypeOf((*MockResourceFetcher)(nil).VSphereWorkerMachineTemplate), arg0, arg1)
}

// MockResourceUpdater is a mock of ResourceUpdater interface.
type MockResourceUpdater struct {
	ctrl     *gomock.Controller
	recorder *MockResourceUpdaterMockRecorder
}

// MockResourceUpdaterMockRecorder is the mock recorder for MockResourceUpdater.
type MockResourceUpdaterMockRecorder struct {
	mock *MockResourceUpdater
}

// NewMockResourceUpdater creates a new mock instance.
func NewMockResourceUpdater(ctrl *gomock.Controller) *MockResourceUpdater {
	mock := &MockResourceUpdater{ctrl: ctrl}
	mock.recorder = &MockResourceUpdaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceUpdater) EXPECT() *MockResourceUpdaterMockRecorder {
	return m.recorder
}

// ApplyPatch mocks base method.
func (m *MockResourceUpdater) ApplyPatch(arg0 context.Context, arg1 client.Object, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyPatch", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyPatch indicates an expected call of ApplyPatch.
func (mr *MockResourceUpdaterMockRecorder) ApplyPatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyPatch", reflect.TypeOf((*MockResourceUpdater)(nil).ApplyPatch), arg0, arg1, arg2)
}

// ApplyTemplate mocks base method.
func (m *MockResourceUpdater) ApplyTemplate(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 map[string]interface{}, arg3 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyTemplate", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyTemplate indicates an expected call of ApplyTemplate.
func (mr *MockResourceUpdaterMockRecorder) ApplyTemplate(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyTemplate", reflect.TypeOf((*MockResourceUpdater)(nil).ApplyTemplate), arg0, arg1, arg2, arg3)
}

// ApplyUpdatedTemplate mocks base method.
func (m *MockResourceUpdater) ApplyUpdatedTemplate(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyUpdatedTemplate", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyUpdatedTemplate indicates an expected call of ApplyUpdatedTemplate.
func (mr *MockResourceUpdaterMockRecorder) ApplyUpdatedTemplate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyUpdatedTemplate", reflect.TypeOf((*MockResourceUpdater)(nil).ApplyUpdatedTemplate), arg0, arg1, arg2)
}

// CreateResource mocks base method.
func (m *MockResourceUpdater) CreateResource(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateResource", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateResource indicates an expected call of CreateResource.
func (mr *MockResourceUpdaterMockRecorder) CreateResource(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResource", reflect.TypeOf((*MockResourceUpdater)(nil).CreateResource), arg0, arg1, arg2)
}

// ForceApplyTemplate mocks base method.
func (m *MockResourceUpdater) ForceApplyTemplate(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceApplyTemplate", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForceApplyTemplate indicates an expected call of ForceApplyTemplate.
func (mr *MockResourceUpdaterMockRecorder) ForceApplyTemplate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceApplyTemplate", reflect.TypeOf((*MockResourceUpdater)(nil).ForceApplyTemplate), arg0, arg1, arg2)
}

// UpdateTemplate mocks base method.
func (m *MockResourceUpdater) UpdateTemplate(arg0 *unstructured.Unstructured, arg1 map[string]interface{}) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateTemplate", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateTemplate indicates an expected call of UpdateTemplate.
func (mr *MockResourceUpdaterMockRecorder) UpdateTemplate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateTemplate", reflect.TypeOf((*MockResourceUpdater)(nil).UpdateTemplate), arg0, arg1)
}
