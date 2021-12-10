package resource_test

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	"github.com/aws/eks-anywhere/controllers/controllers/resource/mocks"
	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

//go:embed testdata/kubeadmcontrolplane.yaml
var kubeadmcontrolplaneFile string

//go:embed testdata/etcdadmcluster.yaml
var etcdadmclusterFile string

//go:embed testdata/vsphereMachineTemplate.yaml
var vsphereMachineTemplateFile string

//go:embed testdata/machineDeployment.yaml
var machineDeploymentFile string

//go:embed testdata/expectedMachineDeployment.yaml
var expectedMachineDeploymentFile string

//go:embed testdata/expectedMachineDeploymentOnlyReplica.yaml
var expectedMachineDeploymentOnlyReplica string

//go:embed testdata/vsphereDatacenterConfigSpec.yaml
var vsphereDatacenterConfigSpecPath string

//go:embed testdata/vsphereMachineConfigSpec.yaml
var vsphereMachineConfigSpecPath string

func TestClusterReconcilerReconcile(t *testing.T) {
	type args struct {
		objectKey types.NamespacedName
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    controllerruntime.Result
		wantErr bool
		prepare func(context.Context, *mocks.MockResourceFetcher, *mocks.MockResourceUpdater, string, string)
	}{
		{
			name: "worker node reconcile (Vsphere provider) - worker nodes has changes",
			args: args{
				namespace: "namespaceA",
				name:      "nameA",
				objectKey: types.NamespacedName{
					Name:      "nameA",
					Namespace: "namespaceA",
				},
			},
			want: controllerruntime.Result{},
			prepare: func(ctx context.Context, fetcher *mocks.MockResourceFetcher, resourceUpdater *mocks.MockResourceUpdater, name string, namespace string) {
				replicasInput := 3
				cluster := &anywherev1.Cluster{}
				cluster.SetName(name)
				cluster.SetNamespace(namespace)
				cluster.Spec.DatacenterRef.Name = "testDataRef"
				cluster.Spec.DatacenterRef.Kind = anywherev1.VSphereDatacenterKind
				cluster.Spec.ControlPlaneConfiguration = anywherev1.ControlPlaneConfiguration{Count: replicasInput, MachineGroupRef: &anywherev1.Ref{Name: "testMachineGroupRef-cp"}}
				cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{Count: replicasInput, MachineGroupRef: &anywherev1.Ref{Name: "testMachineGroupRef"}}}
				cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{Count: replicasInput, MachineGroupRef: &anywherev1.Ref{Name: "testMachineGroupRef-etcd"}}
				cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.96.0.0/12"}

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "testDataRef", "expected Name to be testDataRef")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "testMachineGroupRef-cp", "expected Name to be testMachineGroupRef-cp")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "testMachineGroupRef", "expected Name to be testMachineGroupRef")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "testMachineGroupRef-etcd", "expected Name to be testMachineGroupRef-etcd")
				}).Return(nil)

				kubeAdmControlPlane := &bootstrapv1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(kubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(etcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().Etcd(ctx, gomock.Any()).Return(etcdadmCluster, nil)
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereEtcdMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ApplyPatch(ctx, gomock.Any(), false).Return(nil)
				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "VSphereMachineTemplate":
						if strings.Contains(template.GetName(), "worker-node") {
							expectedMachineTemplate := &unstructured.Unstructured{}
							if err := yaml.Unmarshal([]byte(vsphereMachineTemplateFile), expectedMachineTemplate); err != nil {
								t.Errorf("unmarshal failed: %v", err)
							}
							assert.Equal(t, expectedMachineTemplate, template, "values", expectedMachineTemplate, template)
						}
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedMachineDeploymentFile), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Vsphere provider) - worker nodes has NO machine-template changes",
			args: args{
				namespace: "namespaceA",
				name:      "nameA",
				objectKey: types.NamespacedName{
					Name:      "nameA",
					Namespace: "namespaceA",
				},
			},
			want: controllerruntime.Result{},
			prepare: func(ctx context.Context, fetcher *mocks.MockResourceFetcher, resourceUpdater *mocks.MockResourceUpdater, name string, namespace string) {
				cluster := &anywherev1.Cluster{}
				cluster.SetName(name)
				cluster.SetNamespace(namespace)
				cluster.Spec.DatacenterRef.Name = "testDataRef"
				cluster.Spec.DatacenterRef.Kind = anywherev1.VSphereDatacenterKind
				cluster.Spec.ControlPlaneConfiguration = anywherev1.ControlPlaneConfiguration{Count: 1, MachineGroupRef: &anywherev1.Ref{Name: "testMachineGroupRef-cp"}}
				cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{Count: 1, MachineGroupRef: &anywherev1.Ref{Name: "testMachineGroupRef"}}}
				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster_no_changes.yaml")
				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				datacenterSpec := &anywherev1.VSphereDatacenterConfig{}
				if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), datacenterSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = datacenterSpec.Spec
					assert.Equal(t, objectKey.Name, "testDataRef", "expected Name to be testDataRef")
				}).Return(nil)

				existingVSDatacenter := &anywherev1.VSphereDatacenterConfig{}
				existingVSDatacenter.Spec = datacenterSpec.Spec
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any()).Return(existingVSDatacenter, nil)

				machineSpec := &anywherev1.VSphereMachineConfig{}
				if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), machineSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "testMachineGroupRef-cp", "expected Name to be testMachineGroupRef-cp")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(name)
					cluster.SetNamespace(namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "testMachineGroupRef", "expected Name to be testMachineGroupRef")
				}).Return(nil)

				existingVSMachine := &anywherev1.VSphereMachineConfig{}
				existingVSMachine.Spec = machineSpec.Spec
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any()).Return(existingVSMachine, nil)

				kubeAdmControlPlane := &bootstrapv1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(kubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				mcDeployment := &clusterv1.MachineDeployment{}
				if err := yaml.Unmarshal([]byte(machineDeploymentFile), mcDeployment); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().MachineDeployment(ctx, gomock.Any()).Return(mcDeployment, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					println(template.GetName(), " : ", template.GetKind())
					switch template.GetKind() {
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedMachineDeploymentOnlyReplica), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			fetcher := mocks.NewMockResourceFetcher(mockCtrl)
			resourceUpdater := mocks.NewMockResourceUpdater(mockCtrl)
			tt.prepare(ctx, fetcher, resourceUpdater, tt.args.name, tt.args.namespace)

			cor := resource.NewClusterReconciler(fetcher, resourceUpdater, test.FakeNow, log.NullLogger{})

			if err := cor.Reconcile(ctx, tt.args.objectKey, false); (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
