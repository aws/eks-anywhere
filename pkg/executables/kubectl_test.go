package executables_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	secretObjectType = "addons.cluster.x-k8s.io/resource-set"
	secretObjectName = "csi-vsphere-config"
)

var capiClustersResourceType = fmt.Sprintf("clusters.%s", clusterv1.GroupVersion.Group)

func newKubectl(t *testing.T) (*executables.Kubectl, context.Context, *types.Cluster, *mockexecutables.MockExecutable) {
	kubeconfigFile := "c.kubeconfig"
	cluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
		Name:           "test-cluster",
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)

	return executables.NewKubectl(executable), ctx, cluster, executable
}

type kubectlTest struct {
	t *testing.T
	*WithT
	k          *executables.Kubectl
	ctx        context.Context
	cluster    *types.Cluster
	e          *mockexecutables.MockExecutable
	namespace  string
	kubeconfig string
}

func newKubectlTest(t *testing.T) *kubectlTest {
	k, ctx, cluster, e := newKubectl(t)
	return &kubectlTest{
		t:          t,
		k:          k,
		ctx:        ctx,
		cluster:    cluster,
		e:          e,
		WithT:      NewWithT(t),
		namespace:  "namespace",
		kubeconfig: cluster.KubeconfigFile,
	}
}

func TestKubectlApplyKubeSpecSuccess(t *testing.T) {
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyKubeSpec(ctx, cluster, spec); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpec() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecError(t *testing.T) {
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyKubeSpec(ctx, cluster, spec); err == nil {
		t.Errorf("Kubectl.ApplyKubeSpec() error = nil, want not nil")
	}
}

func TestKubectlApplyKubeSpecFromBytesSuccess(t *testing.T) {
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyKubeSpecFromBytes(ctx, cluster, data); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecFromBytesError(t *testing.T) {
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyKubeSpecFromBytes(ctx, cluster, data); err == nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlDeleteKubeSpecFromBytesSuccess(t *testing.T) {
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteKubeSpecFromBytes(ctx, cluster, data); err != nil {
		t.Errorf("Kubectl.DeleteKubeSpecFromBytes() error = %v, want nil", err)
	}
}

func TestKubectlDeleteSpecFromBytesError(t *testing.T) {
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteKubeSpecFromBytes(ctx, cluster, data); err == nil {
		t.Errorf("Kubectl.DeleteKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlApplyKubeSpecFromBytesWithNamespaceSuccess(t *testing.T) {
	var data []byte
	var namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--namespace", namespace, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytesWithNamespace() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecFromBytesWithNamespaceError(t *testing.T) {
	var data []byte
	var namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--namespace", namespace, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace); err == nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlCreateNamespaceSuccess(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.CreateNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.CreateNamespace() error = %v, want nil", err)
	}
}

func TestKubectlCreateNamespaceError(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.CreateNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.CreateNamespace() error = nil, want not nil")
	}
}

func TestKubectlDeleteNamespaceSuccess(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.DeleteNamespace() error = %v, want nil", err)
	}
}

func TestKubectlDeleteNamespaceError(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.DeleteNamespace() error = nil, want not nil")
	}
}

func TestKubectlDeleteSecretSuccess(t *testing.T) {
	var secretName, namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "secret", secretName, "--kubeconfig", cluster.KubeconfigFile, "--namespace", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteSecret(ctx, cluster, secretName, namespace); err != nil {
		t.Errorf("Kubectl.DeleteNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetNamespaceSuccess(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.GetNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.GetNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetNamespaceError(t *testing.T) {
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.GetNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.GetNamespace() error = nil, want not nil")
	}
}

func TestKubectlWaitSuccess(t *testing.T) {
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"wait", "--timeout", timeout, "--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err != nil {
		t.Errorf("Kubectl.Wait() error = %v, want nil", err)
	}
}

func TestKubectlWaitError(t *testing.T) {
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"wait", "--timeout", timeout, "--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err == nil {
		t.Errorf("Kubectl.Wait() error = nil, want not nil")
	}
}

func TestKubectlSaveLogSuccess(t *testing.T) {
	filename := "testfile"
	_, writer := test.NewWriter(t)

	deployment := &types.Deployment{
		Namespace: "namespace",
		Name:      "testname",
		Container: "container",
	}

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"--kubeconfig", cluster.KubeconfigFile, "logs", fmt.Sprintf("deployment/%s", deployment.Name), "-n", deployment.Namespace, "-c", deployment.Container}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.SaveLog(ctx, cluster, deployment, filename, writer); err != nil {
		t.Errorf("Kubectl.SaveLog() error = %v, want nil", err)
	}
}

func TestKubectlSaveLogError(t *testing.T) {
	filename := "testfile"
	_, writer := test.NewWriter(t)

	deployment := &types.Deployment{
		Namespace: "namespace",
		Name:      "testname",
		Container: "container",
	}

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"--kubeconfig", cluster.KubeconfigFile, "logs", fmt.Sprintf("deployment/%s", deployment.Name), "-n", deployment.Namespace, "-c", deployment.Container}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.SaveLog(ctx, cluster, deployment, filename, writer); err == nil {
		t.Errorf("Kubectl.SaveLog() error = nil, want not nil")
	}
}

func TestKubectlDeleteClusterSuccess(t *testing.T) {
	kubeconfigFile := "c.kubeconfig"
	managementCluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}
	clusterToDelete := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteCluster(ctx, managementCluster, clusterToDelete); err != nil {
		t.Errorf("Kubectl.DeleteCluster() error = %v, want nil", err)
	}
}

func TestKubectlDeleteClusterError(t *testing.T) {
	kubeconfigFile := "c.kubeconfig"
	managementCluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}
	clusterToDelete := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteCluster(ctx, managementCluster, clusterToDelete); err == nil {
		t.Errorf("Kubectl.DeleteCluster() error = nil, want not nil")
	}
}

func TestKubectlGetMachines(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantMachines     []types.Machine
	}{
		{
			testName:         "no machines",
			jsonResponseFile: "testdata/kubectl_no_machines.json",
			wantMachines:     []types.Machine{},
		},
		{
			testName:         "machines with no node ref",
			jsonResponseFile: "testdata/kubectl_machines_no_node_ref_no_labels.json",
			wantMachines: []types.Machine{
				{
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
				{
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
						},
					},
				},
			},
		},
		{
			testName:         "machines with no conditions",
			jsonResponseFile: "testdata/kubectl_machines_no_conditions.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "eksa-test-capd",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-control-plane-5nfdg",
						},
					},
				},
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":    "eksa-test-capd",
							"cluster.x-k8s.io/deployment-name": "eksa-test-capd-md-0",
							"machine-template-hash":            "663441929",
						},
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-md-0-bb7885f6f-gkb85",
						},
					},
				},
			},
		},
		{
			testName:         "machines with node ref",
			jsonResponseFile: "testdata/kubectl_machines_with_node_ref.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "eksa-test-capd",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-control-plane-5nfdg",
						},
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":    "eksa-test-capd",
							"cluster.x-k8s.io/deployment-name": "eksa-test-capd-md-0",
							"machine-template-hash":            "663441929",
						},
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-md-0-bb7885f6f-gkb85",
						},
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
						},
					},
				},
			},
		},
		{
			testName:         "etcd machines",
			jsonResponseFile: "testdata/kubectl_etcd_machines_no_node_ref.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name": "eksa-test-capd",
							"cluster.x-k8s.io/etcd-cluster": "",
						},
					},
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{
				"get", "machines", "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
				"--selector=cluster.x-k8s.io/cluster-name=" + cluster.Name,
				"--namespace", constants.EksaSystemNamespace,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotMachines, err := k.GetMachines(ctx, cluster, cluster.Name)
			if err != nil {
				t.Fatalf("Kubectl.GetMachines() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotMachines, tt.wantMachines) {
				t.Fatalf("Kubectl.GetMachines() machines = %+v, want %+v", gotMachines, tt.wantMachines)
			}
		})
	}
}

func TestKubectlGetEksaCloudStackMachineConfig(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantMachines     *v1alpha1.CloudStackMachineConfig
	}{
		{
			testName:         "no machines",
			jsonResponseFile: "testdata/kubectl_no_cs_machineconfigs.json",
			wantMachines: &v1alpha1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
			},
		},
		{
			testName:         "one machineconfig",
			jsonResponseFile: "testdata/kubectl_eksa_cs_machineconfig.json",
			wantMachines: &v1alpha1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudStackMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "test-etcd"},
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Template: v1alpha1.CloudStackResourceRef{
						Type:  "name",
						Value: "testTemplate",
					},
					ComputeOffering: v1alpha1.CloudStackResourceRef{
						Type:  "name",
						Value: "testOffering",
					},
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "maxdrib",
							SshAuthorizedKeys: []string{"ssh-rsa test123 hi"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			machineConfigName := "testMachineConfig"
			e.EXPECT().Execute(ctx, []string{
				"get", "--namespace", constants.EksaSystemNamespace,
				"cloudstackmachineconfigs.anywhere.eks.amazonaws.com",
				machineConfigName,
				"-o", "json", "--kubeconfig", cluster.KubeconfigFile,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotMachines, err := k.GetEksaCloudStackMachineConfig(ctx, machineConfigName, cluster.KubeconfigFile, constants.EksaSystemNamespace)
			if err != nil {
				t.Fatalf("Kubectl.GetEksaCloudStackMachineConfig() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotMachines, tt.wantMachines) {
				t.Fatalf("Kubectl.GetEksaCloudStackMachineConfig() machines = %+v, want %+v", gotMachines, tt.wantMachines)
			}
		})
	}
}

func TestKubectlGetEksaCloudStackDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantDatacenter   *v1alpha1.CloudStackDatacenterConfig
	}{
		{
			testName:         "no datacenter",
			jsonResponseFile: "testdata/kubectl_no_cs_datacenterconfigs.json",
			wantDatacenter: &v1alpha1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
			},
		},
		{
			testName:         "one datacenter",
			jsonResponseFile: "testdata/kubectl_eksa_cs_datacenterconfig.json",
			wantDatacenter: &v1alpha1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudStackDatacenterConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{
					Insecure: true,
					Network: v1alpha1.CloudStackResourceRef{
						Type:  "name",
						Value: "testNetwork",
					},
					Zone: v1alpha1.CloudStackResourceRef{
						Type:  "name",
						Value: "testZone",
					},
					Domain:  "testDomain",
					Account: "testAccount",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			datacenterConfigName := "testDatacenterConfig"
			e.EXPECT().Execute(ctx, []string{
				"get", "--namespace", constants.EksaSystemNamespace,
				"cloudstackdatacenterconfigs.anywhere.eks.amazonaws.com",
				datacenterConfigName,
				"-o", "json", "--kubeconfig", cluster.KubeconfigFile,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDatacenter, err := k.GetEksaCloudStackDatacenterConfig(ctx, datacenterConfigName, cluster.KubeconfigFile, constants.EksaSystemNamespace)
			if err != nil {
				t.Fatalf("Kubectl.GetEksaCloudStackDatacenterConfig() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotDatacenter, tt.wantDatacenter) {
				t.Fatalf("Kubectl.GetEksaCloudStackDatacenterConfig() machines = %+v, want %+v", gotDatacenter, tt.wantDatacenter)
			}
		})
	}
}

func TestKubectlLoadSecret(t *testing.T) {
	tests := []struct {
		testName string
		params   []string
		wantErr  error
	}{
		{
			testName: "SuccessScenario",
			params:   []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", "test_cluster", "--kubeconfig", "test_cluster.kind.kubeconfig", "--namespace", constants.EksaSystemNamespace},
			wantErr:  nil,
		},
		{
			testName: "ErrorScenario",
			params:   []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", "test_cluster", "--kubeconfig", "test_cluster.kind.kubeconfig", "--namespace", constants.EksaSystemNamespace},
			wantErr:  errors.New("error loading secret: "),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(tt *testing.T) {
			k, ctx, _, e := newKubectl(t)
			e.EXPECT().Execute(ctx, tc.params).Return(bytes.Buffer{}, tc.wantErr)

			err := k.LoadSecret(ctx, "test_cluster", secretObjectType, secretObjectName, "test_cluster.kind.kubeconfig")

			if (tc.wantErr != nil && err == nil) && !reflect.DeepEqual(tc.wantErr, err) {
				t.Errorf("%v got = %v, want %v", tc.testName, err, tc.wantErr)
			}
		})
	}
}

func TestKubectlGetSecret(t *testing.T) {
	tests := []struct {
		testName     string
		responseFile string
		wantSecret   *corev1.Secret
		params       []string
		wantErr      error
	}{
		{
			testName:     "SuccessScenario",
			responseFile: "testdata/kubectl_secret.json",
			wantSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vsphere-csi-controller",
					Namespace: "eksa-system",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				Data: map[string][]byte{
					"data": []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: vsphere-csi-controller
  namespace: kube-system
`),
				},
				Type: corev1.SecretType("addons.cluster.x-k8s.io/resource-set"),
			},
			params:  []string{"get", "secret", secretObjectName, "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", "c.kubeconfig"},
			wantErr: nil,
		},
		{
			testName:     "ErrorScenario",
			responseFile: "testdata/kubectl_secret.json",
			wantSecret:   nil,
			params:       []string{"get", "secret", secretObjectName, "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", "c.kubeconfig"},
			wantErr:      errors.New("error from kubectl client"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(tt *testing.T) {
			response := test.ReadFile(t, tc.responseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, tc.params).Return(*bytes.NewBufferString(response), tc.wantErr)

			secret, err := k.GetSecret(ctx, secretObjectName, executables.WithNamespace(constants.EksaSystemNamespace), executables.WithCluster(cluster))

			g := NewWithT(t)
			if tc.wantErr != nil {
				g.Expect(err.Error()).To(HaveSuffix(tc.wantErr.Error()))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(secret).To(Equal(tc.wantSecret))
			}
		})
	}
}

func TestKubectlGetClusters(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantClusters     []types.CAPICluster
	}{
		{
			testName:         "no clusters",
			jsonResponseFile: "testdata/kubectl_no_clusters.json",
			wantClusters:     []types.CAPICluster{},
		},
		{
			testName:         "machines with node ref",
			jsonResponseFile: "testdata/kubectl_clusters_one.json",
			wantClusters: []types.CAPICluster{
				{
					Metadata: types.Metadata{
						Name: "eksa-test-capd",
					},
					Status: types.ClusterStatus{Phase: "Provisioned"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{"get", "clusters.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotClusters, err := k.GetClusters(ctx, cluster)
			if err != nil {
				t.Fatalf("Kubectl.GetClusters() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotClusters, tt.wantClusters) {
				t.Fatalf("Kubectl.GetClusters() clusters = %+v, want %+v", gotClusters, tt.wantClusters)
			}
		})
	}
}

func TestKubectlGetEKSAClusters(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		expectedSpec     v1alpha1.ClusterSpec
		clusterName      string
	}{
		{
			testName:         "EKS-A cluster found",
			jsonResponseFile: "testdata/kubectl_eksa_cluster.json",
			expectedSpec: v1alpha1.ClusterSpec{
				KubernetesVersion:             "1.19",
				ControlPlaneConfiguration:     v1alpha1.ControlPlaneConfiguration{Count: 3},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3}},
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "test-cluster",
				},
			},
			clusterName: "test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{"get", "clusters.anywhere.eks.amazonaws.com", "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig", cluster.KubeconfigFile, "--field-selector=metadata.name=" + tt.clusterName}).Return(*bytes.NewBufferString(fileContent), nil)

			gotCluster, err := k.GetEksaCluster(ctx, cluster, tt.clusterName)
			if err != nil {
				t.Fatalf("Kubectl.GetEKSAClusters() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotCluster.Spec, tt.expectedSpec) {
				t.Fatalf("Kubectl.GetEKSAClusters() clusters = %+v, want %+v", gotCluster.Spec, tt.expectedSpec)
			}
		})
	}
}

func TestKubectlGetGetApiServerUrlSuccess(t *testing.T) {
	wantUrl := "https://127.0.0.1:37479"
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"},
	).Return(*bytes.NewBufferString(wantUrl), nil)

	gotUrl, err := k.GetApiServerUrl(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.GetApiServerUrl() error = %v, want nil", err)
	}

	if gotUrl != wantUrl {
		t.Fatalf("Kubectl.GetApiServerUrl() url = %s, want %s", gotUrl, wantUrl)
	}
}

func TestKubectlGetGetApiServerUrlError(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"},
	).Return(bytes.Buffer{}, errors.New("error in command"))

	_, err := k.GetApiServerUrl(ctx, cluster)
	if err == nil {
		t.Fatal("Kubectl.GetApiServerUrl() error = nil, want not nil")
	}
}

func TestKubectlGetPodsWithAllNamespaces(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantPodNames     []string
	}{
		{
			testName:         "no pods",
			jsonResponseFile: "testdata/kubectl_no_pods.json",
			wantPodNames:     []string{},
		},
		{
			testName:         "multiple pods",
			jsonResponseFile: "testdata/kubectl_pods.json",
			wantPodNames: []string{
				"coredns-74ff55c5b-cnbwh",
				"coredns-74ff55c5b-zlbph",
				"etcd-lol-control-plane",
				"kindnet-xzddb",
				"kube-apiserver-lol-control-plane",
				"kube-controller-manager-lol-control-plane",
				"kube-proxy-27v6c",
				"kube-scheduler-lol-control-plane",
				"local-path-provisioner-78776bfc44-s9ggt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "pods", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

			gotPods, err := k.GetPods(ctx, executables.WithCluster(cluster), executables.WithAllNamespaces())
			if err != nil {
				t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotPods))
			for _, p := range gotPods {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantPodNames) {
				t.Fatalf("Kubectl.GetPods() pod names = %+v, want %+v", gotNames, tt.wantPodNames)
			}
		})
	}
}

func TestKubectlGetPodsWithNamespace(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantPodNames     []string
		namespace        string
	}{
		{
			testName:         "no pods",
			jsonResponseFile: "testdata/kubectl_no_pods.json",
			wantPodNames:     []string{},
			namespace:        "kube-system",
		},
		{
			testName:         "multiple pods",
			jsonResponseFile: "testdata/kubectl_pods.json",
			wantPodNames: []string{
				"coredns-74ff55c5b-cnbwh",
				"coredns-74ff55c5b-zlbph",
				"etcd-lol-control-plane",
				"kindnet-xzddb",
				"kube-apiserver-lol-control-plane",
				"kube-controller-manager-lol-control-plane",
				"kube-proxy-27v6c",
				"kube-scheduler-lol-control-plane",
				"local-path-provisioner-78776bfc44-s9ggt",
			},
			namespace: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "pods", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", tt.namespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotPods, err := k.GetPods(ctx, executables.WithCluster(cluster), executables.WithNamespace(tt.namespace))
			if err != nil {
				t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotPods))
			for _, p := range gotPods {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantPodNames) {
				t.Fatalf("Kubectl.GetPods() pod names = %+v, want %+v", gotNames, tt.wantPodNames)
			}
		})
	}
}

func TestKubectlGetPodsWithServerSkipTLSAndToken(t *testing.T) {
	k, ctx, _, e := newKubectl(t)
	server := "https://127.0.0.1:37479"
	token := "token"
	fileContent := test.ReadFile(t, "testdata/kubectl_no_pods.json")
	e.EXPECT().Execute(
		ctx, []string{"get", "pods", "-o", "json", "--server", server, "--token", token, "--insecure-skip-tls-verify=true", "-A"},
	).Return(*bytes.NewBufferString(fileContent), nil)

	gotPods, err := k.GetPods(ctx,
		executables.WithServer(server), executables.WithToken(token), executables.WithSkipTLSVerify(), executables.WithAllNamespaces(),
	)
	if err != nil {
		t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
	}

	if len(gotPods) != 0 {
		t.Fatalf("Kubectl.GetPods() num pod  = %d, want 0", len(gotPods))
	}
}

func TestKubectlGetDeploymentsWithAllNamespaces(t *testing.T) {
	tests := []struct {
		testName            string
		jsonResponseFile    string
		wantDeploymentNames []string
	}{
		{
			testName:            "no deployments",
			jsonResponseFile:    "testdata/kubectl_no_deployments.json",
			wantDeploymentNames: []string{},
		},
		{
			testName:         "multiple deployments",
			jsonResponseFile: "testdata/kubectl_deployments.json",
			wantDeploymentNames: []string{
				"coredns",
				"local-path-provisioner",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "deployments", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetDeployments(ctx, executables.WithCluster(cluster), executables.WithAllNamespaces())
			if err != nil {
				t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantDeploymentNames) {
				t.Fatalf("Kubectl.GetDeployments() deployments = %+v, want %+v", gotNames, tt.wantDeploymentNames)
			}
		})
	}
}

func TestKubectlGetDeploymentsWithServerSkipTLSAndToken(t *testing.T) {
	server := "https://127.0.0.1:37479"
	token := "token"
	k, ctx, _, e := newKubectl(t)
	fileContent := test.ReadFile(t, "testdata/kubectl_no_deployments.json")
	e.EXPECT().Execute(
		ctx,
		[]string{"get", "deployments", "-o", "json", "--server", server, "--token", token, "--insecure-skip-tls-verify=true", "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

	gotDeployments, err := k.GetDeployments(
		ctx,
		executables.WithServer(server), executables.WithToken(token), executables.WithSkipTLSVerify(), executables.WithAllNamespaces(),
	)
	if err != nil {
		t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
	}

	if len(gotDeployments) != 0 {
		t.Fatalf("Kubectl.GetDeployments() num deployments = %d, want 0", len(gotDeployments))
	}
}

func TestKubectlGetDeploymentsWithNamespace(t *testing.T) {
	tests := []struct {
		testName            string
		jsonResponseFile    string
		wantDeploymentNames []string
		deploymentNamespace string
	}{
		{
			testName:            "no deployments",
			jsonResponseFile:    "testdata/kubectl_no_deployments.json",
			wantDeploymentNames: []string{},
			deploymentNamespace: "kube-system",
		},
		{
			testName:         "multiple deployments",
			jsonResponseFile: "testdata/kubectl_deployments.json",
			wantDeploymentNames: []string{
				"coredns",
				"local-path-provisioner",
			},
			deploymentNamespace: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "deployments", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", tt.deploymentNamespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetDeployments(ctx, executables.WithCluster(cluster), executables.WithNamespace(tt.deploymentNamespace))
			if err != nil {
				t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantDeploymentNames) {
				t.Fatalf("Kubectl.GetDeployments() deployments = %+v, want %+v", gotNames, tt.wantDeploymentNames)
			}
		})
	}
}

func TestKubectlGetMachineDeployments(t *testing.T) {
	tests := []struct {
		testName                   string
		jsonResponseFile           string
		wantMachineDeploymentNames []string
	}{
		{
			testName:                   "no machine deployments",
			jsonResponseFile:           "testdata/kubectl_no_machine_deployments.json",
			wantMachineDeploymentNames: []string{},
		},
		{
			testName:         "multiple machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments.json",
			wantMachineDeploymentNames: []string{
				"test0-md-0",
				"test1-md-0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "machinedeployments.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetMachineDeployments(ctx, executables.WithCluster(cluster))
			if err != nil {
				t.Fatalf("Kubectl.GetMachineDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantMachineDeploymentNames) {
				t.Fatalf("Kubectl.GetMachineDeployments() deployments = %+v, want %+v", gotNames, tt.wantMachineDeploymentNames)
			}
		})
	}
}

func TestKubectlGetKubeAdmControlPlanes(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantCpNames      []string
	}{
		{
			testName:         "no control planes",
			jsonResponseFile: "testdata/kubectl_no_kubeadmcps.json",
			wantCpNames:      []string{},
		},
		{
			testName:         "multiple control planes",
			jsonResponseFile: "testdata/kubectl_kubeadmcps.json",
			wantCpNames: []string{
				"test0-control-plane",
				"test1-control-plane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "kubeadmcontrolplanes.controlplane.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)

			gotCps, err := k.GetKubeadmControlPlanes(ctx, executables.WithCluster(cluster))
			if err != nil {
				t.Fatalf("Kubectl.GetKubeadmControlPlanes() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotCps))
			for _, p := range gotCps {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantCpNames) {
				t.Fatalf("Kubectl.GetKubeadmControlPlanes() controlPlanes = %+v, want %+v", gotNames, tt.wantCpNames)
			}
		})
	}
}

func TestKubectlVersion(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	fileContent := test.ReadFile(t, "testdata/kubectl_version.json")
	e.EXPECT().Execute(ctx, []string{"version", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)
	gotVersions, err := k.Version(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.Version() error = %v, want nil", err)
	}
	wantVersions := &executables.VersionResponse{
		ClientVersion: version.Info{
			Major:        "1",
			Minor:        "21",
			GitVersion:   "v1.21.2",
			GitCommit:    "092fbfbf53427de67cac1e9fa54aaa09a28371d7",
			GitTreeState: "clean",
			BuildDate:    "2021-06-16T12:59:11Z",
			GoVersion:    "go1.16.5",
			Compiler:     "gc",
			Platform:     "darwin/amd64",
		},
		ServerVersion: version.Info{
			Major:        "1",
			Minor:        "18+",
			GitVersion:   "v1.18.16-eks-1-18-4",
			GitCommit:    "3cdb4c9ab835e2964c8eaeb3ee77d088c7fa36aa",
			GitTreeState: "clean",
			BuildDate:    "2021-05-05T13:09:23Z",
			GoVersion:    "go1.13.15",
			Compiler:     "gc",
			Platform:     "linux/amd64",
		},
	}
	if !reflect.DeepEqual(gotVersions, wantVersions) {
		t.Fatalf("Kubectl.Version() versionResponse = %+v, want %+v", gotVersions, wantVersions)
	}
}

func TestKubectlValidateClustersCRDSuccess(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "crd", "clusters.cluster.x-k8s.io", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, nil)
	err := k.ValidateClustersCRD(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.ValidateClustersCRD() error = %v, want nil", err)
	}
}

func TestKubectlValidateClustersCRDNotFound(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "crd", "clusters.cluster.x-k8s.io", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, errors.New("CRD not found"))
	err := k.ValidateClustersCRD(ctx, cluster)
	if err == nil {
		t.Fatalf("Kubectl.ValidateClustersCRD() error == nil, want CRD not found")
	}
}

func TestKubectlUpdateEnvironmentVariablesInNamespace(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	envMap := map[string]string{
		"key": "val",
	}
	e.EXPECT().Execute(ctx, []string{
		"set", "env", "deployment",
		"eksa-controller-manager", "key=val",
		"--kubeconfig", cluster.KubeconfigFile,
		"--namespace", "eksa-system",
	})

	err := k.UpdateEnvironmentVariablesInNamespace(ctx, "deployment", "eksa-controller-manager", envMap, cluster, "eksa-system")
	if err != nil {
		t.Fatalf("Kubectl.UpdateEnvironmentVariablesInNamespace() error = %v, want nil", err)
	}
}

func TestKubectlUpdateAnnotation(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "gitrepositories", "flux-system",
		"key1=val1", "--overwrite",
		"--kubeconfig", cluster.KubeconfigFile,
	})

	a := map[string]string{
		"key1": "val1",
	}
	err := k.UpdateAnnotation(ctx, "gitrepositories", "flux-system", a, executables.WithOverwrite(), executables.WithCluster(cluster))
	if err != nil {
		t.Fatalf("Kubectl.UpdateAnnotation() error = %v, want nil", err)
	}
}

func TestKubectlRemoveAnnotation(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "cluster", "test-cluster", "key1-", "--kubeconfig", cluster.KubeconfigFile,
	})

	err := k.RemoveAnnotation(ctx, "cluster", "test-cluster", "key1", executables.WithCluster(cluster))
	if err != nil {
		t.Fatalf("Kubectl.RemoveAnnotation() error = %v, want nil", err)
	}
}

func TestKubectlRemoveAnnotationInNamespace(t *testing.T) {
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "cluster", "test-cluster", "key1-", "--kubeconfig", cluster.KubeconfigFile, "--namespace", "",
	})

	err := k.RemoveAnnotationInNamespace(ctx, "cluster", "test-cluster", "key1", cluster, "")
	if err != nil {
		t.Fatalf("Kubectl.RemoveAnnotationInNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetBundles(t *testing.T) {
	tt := newKubectlTest(t)
	wantBundles := test.Bundles(t)
	bundleName := "Bundle-name"
	bundlesJson, err := json.Marshal(wantBundles)
	if err != nil {
		t.Fatalf("Failed marshalling Bundles: %s", err)
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "bundles.anywhere.eks.amazonaws.com", bundleName, "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(*bytes.NewBuffer(bundlesJson), nil)

	gotBundles, err := tt.k.GetBundles(tt.ctx, tt.cluster.KubeconfigFile, bundleName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotBundles).To(Equal(wantBundles))
}

func TestKubectlGetClusterResourceSet(t *testing.T) {
	tt := newKubectlTest(t)
	resourceSetJson := test.ReadFile(t, "testdata/kubectl_clusterresourceset.json")
	resourceSetName := "Bundle-name"
	wantResourceSet := &addons.ClusterResourceSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "addons.cluster.x-k8s.io/v1beta1",
			Kind:       "ClusterResourceSet",
		},
		Spec: addons.ClusterResourceSetSpec{
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster.x-k8s.io/cluster-name": "cluster-1",
				},
			},
			Strategy: "ApplyOnce",
			Resources: []addons.ResourceRef{
				{
					Kind: "Secret",
					Name: "vsphere-csi-controller",
				},
				{
					Kind: "ConfigMap",
					Name: "vsphere-csi-controller-role",
				},
			},
		},
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "clusterresourcesets.addons.cluster.x-k8s.io", resourceSetName, "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(*bytes.NewBufferString(resourceSetJson), nil)

	gotResourceSet, err := tt.k.GetClusterResourceSet(tt.ctx, tt.cluster.KubeconfigFile, resourceSetName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotResourceSet).To(Equal(wantResourceSet))
}

func TestKubectlGetConfigMap(t *testing.T) {
	tt := newKubectlTest(t)
	configmapJson := test.ReadFile(t, "testdata/kubectl_configmap.json")
	configmapName := "csi.vsphere.vmware.com"
	wantConfigmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: "eksa-system",
		},
		Data: map[string]string{
			"data": `apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: csi.vsphere.vmware.com
spec:
  attachRequired: true
`,
		},
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "configmap", configmapName, "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(*bytes.NewBufferString(configmapJson), nil)

	gotConfigmap, err := tt.k.GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, configmapName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotConfigmap).To(Equal(wantConfigmap))
}

func TestKubectlSetDaemonSetImage(t *testing.T) {
	tt := newKubectlTest(t)
	daemonSetName := "ds-1"
	container := "cont1"
	image := "public.ecr.aws/image2"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"set", "image", "daemonset/ds-1", "cont1=public.ecr.aws/image2", "--namespace", tt.namespace, "--kubeconfig", tt.cluster.KubeconfigFile,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.SetDaemonSetImage(tt.ctx, tt.cluster.KubeconfigFile, daemonSetName, tt.namespace, container, image)).To(Succeed())
}

func TestKubectlCheckCAPIProviderExistsNotInstalled(t *testing.T) {
	tt := newKubectlTest(t)
	providerName := "providerName"
	providerNs := "providerNs"

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", providerNs), "--kubeconfig", tt.cluster.KubeconfigFile}).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, providerName, providerNs))
}

func TestKubectlCheckCAPIProviderExistsInstalled(t *testing.T) {
	tt := newKubectlTest(t)
	providerName := "providerName"
	providerNs := "providerNs"

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", providerNs), "--kubeconfig", tt.cluster.KubeconfigFile}).Return(*bytes.NewBufferString("namespace"), nil)

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "provider", "--namespace", providerNs, fmt.Sprintf("--field-selector=metadata.name=%s", providerName), "--kubeconfig", tt.cluster.KubeconfigFile})

	tt.Expect(tt.k.CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, providerName, providerNs))
}

func TestKubectlGetDeploymentSuccess(t *testing.T) {
	var replicas int32 = 2
	newKubectlGetterTest(t).withResourceType(
		"deployment",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDeployment(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).withJsonFromFile(
		"testdata/kubectl_deployment.json",
	).andWant(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: "kube-system",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "k8s.gcr.io/coredns:1.7.0",
								Name:  "coredns",
							},
						},
					},
				},
			},
		},
	).testSuccess()
}

func TestKubectlGetDeploymentError(t *testing.T) {
	newKubectlGetterTest(t).withResourceType(
		"deployment",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDeployment(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).testError()
}

func TestKubectlGetDaemonSetSuccess(t *testing.T) {
	newKubectlGetterTest(t).withResourceType(
		"daemonset",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDaemonSet(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).withJsonFromFile(
		"testdata/kubectl_daemonset.json",
	).andWant(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cilium",
				Namespace: "kube-system",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Command: []string{"cilium-agent"},
								Image:   "public.ecr.aws/isovalent/cilium:v1.9.11-eksa.1",
								Name:    "cilium-agent",
							},
						},
					},
				},
			},
		},
	).testSuccess()
}

func TestKubectlGetDaemonSetError(t *testing.T) {
	newKubectlGetterTest(t).withResourceType(
		"daemonset",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDaemonSet(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).testError()
}

func TestApplyTolerationsFromTaints(t *testing.T) {
	tt := newKubectlTest(t)
	params := []string{
		"get", "ds", "test",
		"-o", "jsonpath={range .spec.template.spec}{.tolerations} {end}",
		"-n", "testNs", "--kubeconfig", tt.cluster.KubeconfigFile,
	}
	tt.e.EXPECT().Execute(
		tt.ctx, gomock.Eq(params)).Return(bytes.Buffer{}, nil)
	var taints []corev1.Taint
	tt.Expect(tt.k.ApplyTolerationsFromTaints(tt.ctx, taints, taints, "ds", "test", tt.cluster.KubeconfigFile, "testNs", "/test")).To(Succeed())
}
