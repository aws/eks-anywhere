package resource_test

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	"github.com/aws/eks-anywhere/controllers/controllers/resource/mocks"
	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
)

//go:embed testdata/vsphereKubeadmcontrolplane.yaml
var vsphereKubeadmcontrolplaneFile string

//go:embed testdata/vsphereEtcdadmcluster.yaml
var vsphereEtcdadmclusterFile string

//go:embed testdata/vsphereMachineTemplate.yaml
var vsphereMachineTemplateFile string

//go:embed testdata/vsphereMachineDeployment.yaml
var vsphereMachineDeploymentFile string

//go:embed testdata/expectedVSphereMachineDeployment.yaml
var expectedVSphereMachineDeploymentFile string

//go:embed testdata/expectedVSphereMachineDeploymentOnlyReplica.yaml
var expectedVSphereMachineDeploymentOnlyReplica string

//go:embed testdata/expectedVSphereMachineDeploymentTemplateChanged.yaml
var expectedVSphereMachineDeploymentTemplateChanged string

//go:embed testdata/vsphereDatacenterConfigSpec.yaml
var vsphereDatacenterConfigSpecPath string

//go:embed testdata/vsphereMachineConfigSpec.yaml
var vsphereMachineConfigSpecPath string

//go:embed testdata/vsphereKubeadmconfigTemplateSpec.yaml
var vsphereKubeadmconfigTemplateSpecPath string

//go:embed testdata/cloudstackKubeadmcontrolplane.yaml
var cloudstackKubeadmcontrolplaneFile string

//go:embed testdata/cloudstackEtcdadmcluster.yaml
var cloudstackEtcdadmclusterFile string

//go:embed testdata/cloudstackMachineTemplate.yaml
var cloudstackMachineTemplateFile string

//go:embed testdata/cloudstackMachineDeployment.yaml
var cloudstackMachineDeploymentFile string

//go:embed testdata/expectedCloudStackMachineDeployment.yaml
var expectedCloudStackMachineDeploymentFile string

//go:embed testdata/expectedCloudStackMachineDeploymentOnlyReplica.yaml
var expectedCloudStackMachineDeploymentOnlyReplica string

//go:embed testdata/expectedCloudStackMachineDeploymentTemplateChanged.yaml
var expectedCloudStackMachineDeploymentTemplateChanged string

//go:embed testdata/cloudstackDatacenterConfigSpec.yaml
var cloudstackDatacenterConfigSpecPath string

//go:embed testdata/cloudstackMachineConfigSpec.yaml
var cloudstackMachineConfigSpecPath string

//go:embed testdata/cloudstackKubeadmconfigTemplateSpec.yaml
var cloudstackKubeadmconfigTemplateSpecPath string

func TestClusterReconcilerReconcileVSphere(t *testing.T) {
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
				cluster := &anywherev1.Cluster{}
				cluster.SetName(name)
				cluster.SetNamespace(namespace)

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(vsphereKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(vsphereEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				machineDeployment := &clusterv1.MachineDeployment{}
				if err := yaml.Unmarshal([]byte(vsphereMachineDeploymentFile), machineDeployment); err != nil {
					t.Errorf("unmarshalling machinedeployment failed: %v", err)
				}
				workerNodeMachineConfig := &anywherev1.VSphereMachineConfig{
					Spec: anywherev1.VSphereMachineConfigSpec{
						Users: []anywherev1.UserConfiguration{
							{
								Name:              "capv",
								SshAuthorizedKeys: []string{"ssh-rsa ssh_key_value"},
							},
						},
					},
				}
				fetcher.EXPECT().MachineDeployment(ctx, gomock.Any(), gomock.Any()).Return(machineDeployment, nil)

				fetcher.EXPECT().Etcd(ctx, gomock.Any()).Return(etcdadmCluster, nil)
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.VSphereDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereEtcdMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(workerNodeMachineConfig, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.WorkerNodeGroupConfiguration{}, nil)
				fetcher.EXPECT().VSphereCredentials(ctx).Return(&corev1.Secret{
					Data: map[string][]byte{"username": []byte("username"), "password": []byte("password")},
				}, nil)
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
						if err := yaml.Unmarshal([]byte(expectedVSphereMachineDeploymentFile), expectedMCDeployment); err != nil {
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
				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster_no_changes.yaml")
				cluster.Spec = spec.Cluster.Spec
				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				datacenterSpec := &anywherev1.VSphereDatacenterConfig{}
				if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), datacenterSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Name)
					cluster.Spec = datacenterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				existingVSDatacenter := &anywherev1.VSphereDatacenterConfig{}
				existingVSDatacenter.Spec = datacenterSpec.Spec
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(existingVSDatacenter, nil)

				machineSpec := &anywherev1.VSphereMachineConfig{}
				if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), machineSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
				}
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)

				existingVSMachine := &anywherev1.VSphereMachineConfig{}
				existingVSMachine.Spec = machineSpec.Spec
				workerNodeMachineConfig := &anywherev1.VSphereMachineConfig{
					Spec: anywherev1.VSphereMachineConfigSpec{
						Users: []anywherev1.UserConfiguration{
							{
								Name:              "capv",
								SshAuthorizedKeys: []string{"ssh-rsa ssh_key_value"},
							},
						},
					},
				}
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(workerNodeMachineConfig, nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(vsphereKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				machineDeployment := &clusterv1.MachineDeployment{}
				if err := yaml.Unmarshal([]byte(vsphereMachineDeploymentFile), machineDeployment); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}
				fetcher.EXPECT().MachineDeployment(ctx, gomock.Any(), gomock.Any()).Return(machineDeployment, nil)

				fetcher.EXPECT().VSphereCredentials(ctx).Return(&corev1.Secret{
					Data: map[string][]byte{"username": []byte("username"), "password": []byte("password")},
				}, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					println(template.GetName(), " : ", template.GetKind())
					switch template.GetKind() {
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedVSphereMachineDeploymentOnlyReplica), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Vsphere provider) - worker node taints have changed",
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

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(vsphereKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(vsphereEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
					Taints: []corev1.Taint{
						{
							Key:    "key1",
							Value:  "val1",
							Effect: "PreferNoSchedule",
						},
					},
				}

				fetcher.EXPECT().Etcd(ctx, gomock.Any()).Return(etcdadmCluster, nil)
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.VSphereDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereEtcdMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)
				fetcher.EXPECT().VSphereCredentials(ctx).Return(&corev1.Secret{
					Data: map[string][]byte{"username": []byte("username"), "password": []byte("password")},
				}, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ApplyPatch(ctx, gomock.Any(), false).Return(nil)
				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "KubeadmConfigTemplate":
						existingKubeadmConfigTemplate := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(vsphereKubeadmconfigTemplateSpecPath), existingKubeadmConfigTemplate); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, existingKubeadmConfigTemplate, template, "values", existingKubeadmConfigTemplate, template)
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedVSphereMachineDeploymentTemplateChanged), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Vsphere provider) - worker node labels have changed",
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

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(vsphereDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.VSphereMachineConfig{}
					if err := yaml.Unmarshal([]byte(vsphereMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.VSphereMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(vsphereKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(vsphereEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
					Labels: map[string]string{
						"Key1": "Val1",
						"Key2": "Val2",
					},
				}

				fetcher.EXPECT().Etcd(ctx, gomock.Any()).Return(etcdadmCluster, nil)
				fetcher.EXPECT().ExistingVSphereDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.VSphereDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereEtcdMachineConfig(ctx, gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingVSphereWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.VSphereMachineConfig{}, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)
				fetcher.EXPECT().VSphereCredentials(ctx).Return(&corev1.Secret{
					Data: map[string][]byte{"username": []byte("username"), "password": []byte("password")},
				}, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ApplyPatch(ctx, gomock.Any(), false).Return(nil)
				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "KubeadmConfigTemplate":
						existingKubeadmConfigTemplate := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(vsphereKubeadmconfigTemplateSpecPath), existingKubeadmConfigTemplate); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, existingKubeadmConfigTemplate, template, "values", existingKubeadmConfigTemplate, template)
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedVSphereMachineDeploymentTemplateChanged), expectedMCDeployment); err != nil {
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

func TestClusterReconcilerReconcileCloudStack(t *testing.T) {
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
			name: "worker node reconcile (Cloudstack provider) - worker nodes has changes",
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

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-cloudstack.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(cloudstackKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(cloudstackEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				machineDeployment := &clusterv1.MachineDeployment{}
				if err := yaml.Unmarshal([]byte(cloudstackMachineDeploymentFile), machineDeployment); err != nil {
					t.Errorf("unmarshalling machinedeployment failed: %v", err)
				}
				fetcher.EXPECT().MachineDeployment(ctx, gomock.Any(), gomock.Any()).Return(machineDeployment, nil)

				fetcher.EXPECT().ExistingCloudStackDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.WorkerNodeGroupConfiguration{}, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "CloudStackMachineTemplate":
						if strings.Contains(template.GetName(), "worker-node") {
							expectedMachineTemplate := &unstructured.Unstructured{}
							if err := yaml.Unmarshal([]byte(cloudstackMachineTemplateFile), expectedMachineTemplate); err != nil {
								t.Errorf("unmarshal failed: %v", err)
							}
							assert.Equal(t, expectedMachineTemplate, template, "values", expectedMachineTemplate, template)
						}
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedCloudStackMachineDeploymentFile), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Cloudstack provider) - worker nodes has NO machine-template changes",
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
				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-cloudstack_no_changes.yaml")
				cluster.Spec = spec.Cluster.Spec
				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				datacenterSpec := &anywherev1.CloudStackDatacenterConfig{}
				if err := yaml.Unmarshal([]byte(cloudstackDatacenterConfigSpecPath), datacenterSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.CloudStackDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Name)
					cluster.Spec = datacenterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				machineSpec := &anywherev1.CloudStackMachineConfig{}
				if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), machineSpec); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = machineSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
				}
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)

				existingCSMachine := &anywherev1.CloudStackMachineConfig{}
				existingCSMachine.Spec = machineSpec.Spec
				fetcher.EXPECT().ExistingCloudStackDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(cloudstackKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				machineDeployment := &clusterv1.MachineDeployment{}
				if err := yaml.Unmarshal([]byte(cloudstackMachineDeploymentFile), machineDeployment); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}
				fetcher.EXPECT().MachineDeployment(ctx, gomock.Any(), gomock.Any()).Return(machineDeployment, nil)

				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					println(template.GetName(), " : ", template.GetKind())
					switch template.GetKind() {
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedCloudStackMachineDeploymentOnlyReplica), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Cloudstack provider) - worker node taints have changed",
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

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-cloudstack.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(cloudstackKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(cloudstackEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
					Taints: []corev1.Taint{
						{
							Key:    "key1",
							Value:  "val1",
							Effect: "PreferNoSchedule",
						},
					},
				}

				oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
				os.Unsetenv(features.CloudStackProviderEnvVar)
				defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)

				fetcher.EXPECT().ExistingCloudStackDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "KubeadmConfigTemplate":
						existingKubeadmConfigTemplate := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(cloudstackKubeadmconfigTemplateSpecPath), existingKubeadmConfigTemplate); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, existingKubeadmConfigTemplate, template, "values", existingKubeadmConfigTemplate, template)
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedCloudStackMachineDeploymentTemplateChanged), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
		{
			name: "worker node reconcile (Cloudstack provider) - worker node labels have changed",
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

				fetcher.EXPECT().FetchCluster(gomock.Any(), gomock.Any()).Return(cluster, nil)

				spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-cloudstack.yaml")
				cluster.Spec = spec.Cluster.Spec

				fetcher.EXPECT().FetchAppliedSpec(ctx, gomock.Any()).Return(spec, nil)

				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackDatacenterConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackDatacenterConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackDatacenterConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)
				fetcher.EXPECT().FetchObject(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, objectKey types.NamespacedName, obj client.Object) {
					clusterSpec := &anywherev1.CloudStackMachineConfig{}
					if err := yaml.Unmarshal([]byte(cloudstackMachineConfigSpecPath), clusterSpec); err != nil {
						t.Errorf("unmarshal failed: %v", err)
					}
					cluster := obj.(*anywherev1.CloudStackMachineConfig)
					cluster.SetName(objectKey.Name)
					cluster.SetNamespace(objectKey.Namespace)
					cluster.Spec = clusterSpec.Spec
					assert.Equal(t, objectKey.Name, "test_cluster", "expected Name to be test_cluster")
				}).Return(nil)

				kubeAdmControlPlane := &controlplanev1.KubeadmControlPlane{}
				if err := yaml.Unmarshal([]byte(cloudstackKubeadmcontrolplaneFile), kubeAdmControlPlane); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				etcdadmCluster := &etcdv1.EtcdadmCluster{}
				if err := yaml.Unmarshal([]byte(cloudstackEtcdadmclusterFile), etcdadmCluster); err != nil {
					t.Errorf("unmarshal failed: %v", err)
				}

				existingWorkerNodeGroupConfiguration := &anywherev1.WorkerNodeGroupConfiguration{
					Name:            "md-0",
					Count:           3,
					MachineGroupRef: nil,
					Labels: map[string]string{
						"Key1": "Val1",
						"Key2": "Val2",
					},
				}

				fetcher.EXPECT().ExistingCloudStackDatacenterConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackDatacenterConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackControlPlaneMachineConfig(ctx, gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingCloudStackWorkerMachineConfig(ctx, gomock.Any(), gomock.Any()).Return(&anywherev1.CloudStackMachineConfig{}, nil)
				fetcher.EXPECT().ExistingWorkerNodeGroupConfig(ctx, gomock.Any(), gomock.Any()).Return(existingWorkerNodeGroupConfiguration, nil)
				fetcher.EXPECT().Fetch(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.NewNotFound(schema.GroupResource{Group: "testgroup", Resource: "testresource"}, ""))

				resourceUpdater.EXPECT().ForceApplyTemplate(ctx, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, template *unstructured.Unstructured, dryRun bool) {
					assert.Equal(t, false, dryRun, "Expected dryRun didn't match")
					switch template.GetKind() {
					case "KubeadmConfigTemplate":
						existingKubeadmConfigTemplate := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(cloudstackKubeadmconfigTemplateSpecPath), existingKubeadmConfigTemplate); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, existingKubeadmConfigTemplate, template, "values", existingKubeadmConfigTemplate, template)
					case "MachineDeployment":
						expectedMCDeployment := &unstructured.Unstructured{}
						if err := yaml.Unmarshal([]byte(expectedCloudStackMachineDeploymentTemplateChanged), expectedMCDeployment); err != nil {
							t.Errorf("unmarshal failed: %v", err)
						}
						assert.Equal(t, expectedMCDeployment, template, "values", expectedMCDeployment, template)
					}
				}).AnyTimes().Return(nil)
			},
		},
	}
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	os.Setenv(features.CloudStackProviderEnvVar, "true")
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)
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
