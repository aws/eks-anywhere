package clustermanager_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	tinkerbellv1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/capt/v1beta1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	mocksdiagnostics "github.com/aws/eks-anywhere/pkg/diagnostics/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

var (
	eksaClusterResourceType           = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType    = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
	expectedPauseAnnotation           = map[string]string{"anywhere.eks.amazonaws.com/paused": "true"}
	maxTime                           = time.Duration(math.MaxInt64)
	managementStatePath               = fmt.Sprintf("cluster-state-backup-%s", time.Now().Format("2006-01-02T15_04_05"))
)

func getKcpAndMdsForNodeCount(count int32) (*controlplanev1beta2.KubeadmControlPlane, []clusterv1beta2.MachineDeployment) {
	kcp := &controlplanev1beta2.KubeadmControlPlane{
		Spec: controlplanev1beta2.KubeadmControlPlaneSpec{
			Replicas: ptr.Int32(count),
		},
	}

	md := []clusterv1beta2.MachineDeployment{
		{
			Spec: clusterv1beta2.MachineDeploymentSpec{
				Replicas: ptr.Int32(count),
			},
		},
	}

	return kcp, md
}

func TestClusterManagerCAPIWaitForDeploymentStackedEtcd(t *testing.T) {
	ctx := context.Background()
	clusterObj := &types.Cluster{}
	c, m := newClusterManager(t)
	clusterSpecStackedEtcd := test.NewClusterSpec()
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpecStackedEtcd.Bundles)

	m.client.EXPECT().InitInfrastructure(ctx, managementComponents, clusterSpecStackedEtcd, clusterObj, m.provider)
	for namespace, deployments := range internal.CAPIDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m0s", "Available", deployment, namespace)
		}
	}
	providerDeployments := map[string][]string{}
	m.provider.EXPECT().GetDeployments().Return(providerDeployments)
	for namespace, deployments := range providerDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m0s", "Available", deployment, namespace)
		}
	}

	if err := c.InstallCAPI(ctx, managementComponents, clusterSpecStackedEtcd, clusterObj, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCAPIWaitForDeploymentExternalEtcd(t *testing.T) {
	ctx := context.Background()
	clusterObj := &types.Cluster{}
	c, m := newClusterManager(t)
	clusterSpecExternalEtcd := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 1}
	})
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpecExternalEtcd.Bundles)

	m.client.EXPECT().InitInfrastructure(ctx, managementComponents, clusterSpecExternalEtcd, clusterObj, m.provider)
	for namespace, deployments := range internal.CAPIDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m0s", "Available", deployment, namespace)
		}
	}
	for namespace, deployments := range internal.ExternalEtcdDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m0s", "Available", deployment, namespace)
		}
	}
	providerDeployments := map[string][]string{}
	m.provider.EXPECT().GetDeployments().Return(providerDeployments)
	for namespace, deployments := range providerDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m0s", "Available", deployment, namespace)
		}
	}
	if err := c.InstallCAPI(ctx, managementComponents, clusterSpecExternalEtcd, clusterObj, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerSaveLogsSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	bootstrapCluster := &types.Cluster{
		Name:           "bootstrap",
		KubeconfigFile: "bootstrap.kubeconfig",
	}

	workloadCluster := &types.Cluster{
		Name:           "workload",
		KubeconfigFile: "workload.kubeconfig",
	}

	c, m := newClusterManager(t)

	b := m.diagnosticsBundle
	m.diagnosticsFactory.EXPECT().DiagnosticBundleManagementCluster(clusterSpec, bootstrapCluster.KubeconfigFile).Return(b, nil)
	b.EXPECT().CollectAndAnalyze(ctx, gomock.AssignableToTypeOf(&time.Time{}))

	m.diagnosticsFactory.EXPECT().DiagnosticBundleWorkloadCluster(clusterSpec, m.provider, workloadCluster.KubeconfigFile, false).Return(b, nil)
	b.EXPECT().CollectAndAnalyze(ctx, gomock.AssignableToTypeOf(&time.Time{}))

	if err := c.SaveLogsManagementCluster(ctx, clusterSpec, bootstrapCluster); err != nil {
		t.Errorf("ClusterManager.SaveLogsManagementCluster() error = %v, wantErr nil", err)
	}

	if err := c.SaveLogsWorkloadCluster(ctx, m.provider, clusterSpec, workloadCluster); err != nil {
		t.Errorf("ClusterManager.SaveLogsWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerPauseCAPIWorkloadClusters(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)
	m.client.EXPECT().PauseCAPICluster(ctx, capiClusterName, mgmtCluster.KubeconfigFile).Return(nil)

	if err := c.PauseCAPIWorkloadClusters(ctx, mgmtCluster); err != nil {
		t.Errorf("ClusterManager.PauseCAPIWorkloadClusters() error = %v", err)
	}
}

func TestClusterManagerPauseCAPIWorkloadClustersErrorGetClusters(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(nil, errors.New("Error: failed to get clusters"))

	if err := c.PauseCAPIWorkloadClusters(ctx, mgmtCluster); err == nil {
		t.Error("ClusterManager.PauseCAPIWorkloadClusters() error = nil, wantErr not nil")
	}
}

func TestClusterManagerPauseCAPIWorkloadClustersErrorPause(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	c, m := newClusterManager(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)
	m.client.EXPECT().PauseCAPICluster(ctx, capiClusterName, mgmtCluster.KubeconfigFile).Return(errors.New("Error pausing cluster"))

	if err := c.PauseCAPIWorkloadClusters(ctx, mgmtCluster); err == nil {
		t.Error("ClusterManager.PauseCAPIWorkloadClusters() error = nil, wantErr not nil")
	}
}

func TestClusterManagerPauseCAPIWorkloadClustersSkipManagement(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: mgmtClusterName}}}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)

	if err := c.PauseCAPIWorkloadClusters(ctx, mgmtCluster); err != nil {
		t.Errorf("ClusterManager.PauseCAPIWorkloadClusters() error = %v", err)
	}
}

func TestClusterManagerResumeCAPIWorkloadClustersErrorGetClusters(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(nil, errors.New("Error: failed to get clusters"))

	if err := c.ResumeCAPIWorkloadClusters(ctx, mgmtCluster); err == nil {
		t.Error("ClusterManager.ResumeCAPIWorkloadClusters() error = nil, wantErr not nil")
	}
}

func TestClusterManagerResumeCAPIWorkloadClustersErrorResume(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	c, m := newClusterManager(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)
	m.client.EXPECT().ResumeCAPICluster(ctx, capiClusterName, mgmtCluster.KubeconfigFile).Return(errors.New("Error pausing cluster"))

	if err := c.ResumeCAPIWorkloadClusters(ctx, mgmtCluster); err == nil {
		t.Error("ClusterManager.ResumeCAPIWorkloadClusters() error = nil, wantErr not nil")
	}
}

func TestClusterManagerResumeCAPIWorkloadClusters(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)
	m.client.EXPECT().ResumeCAPICluster(ctx, capiClusterName, mgmtCluster.KubeconfigFile).Return(nil)

	if err := c.ResumeCAPIWorkloadClusters(ctx, mgmtCluster); err != nil {
		t.Errorf("ClusterManager.ResumeCAPIWorkloadClusters() error = %v", err)
	}
}

func TestClusterManagerResumeCAPIWorkloadClustersSkipManagement(t *testing.T) {
	ctx := context.Background()
	mgmtClusterName := "cluster-name"
	mgmtCluster := &types.Cluster{
		Name:           mgmtClusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: mgmtClusterName}}}
	c, m := newClusterManager(t)
	m.client.EXPECT().GetClusters(ctx, mgmtCluster).Return(clusters, nil)

	if err := c.ResumeCAPIWorkloadClusters(ctx, mgmtCluster); err != nil {
		t.Errorf("ClusterManager.ResumeCAPIWorkloadClusters() error = %v", err)
	}
}

func TestClusterManagerBackupCAPISuccess(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}

	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().BackupManagement(ctx, from, managementStatePath, from.Name)

	if err := c.BackupCAPI(ctx, from, managementStatePath, from.Name); err != nil {
		t.Errorf("ClusterManager.BackupCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerBackupCAPIRetrySuccess(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}

	ctx := context.Background()

	c, m := newClusterManager(t)
	firstTry := m.client.EXPECT().BackupManagement(ctx, from, managementStatePath, from.Name).Return(errors.New("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:61994/api?timeout=30s\": EOF"))
	secondTry := m.client.EXPECT().BackupManagement(ctx, from, managementStatePath, from.Name).Return(nil)
	gomock.InOrder(
		firstTry,
		secondTry,
	)
	if err := c.BackupCAPI(ctx, from, managementStatePath, from.Name); err != nil {
		t.Errorf("ClusterManager.BackupCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerBackupCAPIWaitForInfrastructureSuccess(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}

	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().BackupManagement(ctx, from, managementStatePath, from.Name)

	if err := c.BackupCAPIWaitForInfrastructure(ctx, from, managementStatePath, from.Name); err != nil {
		t.Errorf("ClusterManager.BackupCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterctlWaitRetryPolicy(t *testing.T) {
	connectionRefusedError := fmt.Errorf("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:53733/api?timeout=30s\": dial tcp 127.0.0.1:53733: connect: connection refused")
	ioTimeoutError := fmt.Errorf("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:61994/api?timeout=30s\": net/http: TLS handshake timeout")
	miscellaneousError := fmt.Errorf("Some other random miscellaneous error")

	_, wait := clustermanager.ClusterctlMoveRetryPolicy(1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for connection refused")
	}

	_, wait = clustermanager.ClusterctlMoveRetryPolicy(-1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly protect for total retries < 0")
	}

	_, wait = clustermanager.ClusterctlMoveRetryPolicy(2, connectionRefusedError)
	if wait != 15*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly protect for second retry wait")
	}

	_, wait = clustermanager.ClusterctlMoveRetryPolicy(1, ioTimeoutError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for ioTimeout")
	}

	retry, _ := clustermanager.ClusterctlMoveRetryPolicy(1, miscellaneousError)
	if retry != false {
		t.Errorf("ClusterctlMoveRetryPolicy didn't not-retry on non-network error")
	}
}

func TestClusterctlWaitForInfrastructureRetryPolicy(t *testing.T) {
	connectionRefusedError := fmt.Errorf("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:53733/api?timeout=30s\": dial tcp 127.0.0.1:53733: connect: connection refused")
	ioTimeoutError := fmt.Errorf("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:61994/api?timeout=30s\": net/http: TLS handshake timeout")
	infrastructureError := fmt.Errorf("Error: failed to get object graph: failed to check for provisioned infrastructure: cannot start the move operation while cluster is still provisioning the infrastructure")
	nodeError := fmt.Errorf("Error: failed to get object graph: failed to check for provisioned infrastructure: cannot start the move operation while machine is still provisioning the node")
	miscellaneousError := fmt.Errorf("Some other random miscellaneous error")

	_, wait := clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for connection refused")
	}

	_, wait = clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(-1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly protect for total retries < 0")
	}

	_, wait = clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(2, connectionRefusedError)
	if wait != 15*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly protect for second retry wait")
	}

	_, wait = clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(1, ioTimeoutError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for ioTimeout")
	}

	_, wait = clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(1, infrastructureError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for infrastructureError")
	}

	_, wait = clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(1, nodeError)
	if wait != 10*time.Second {
		t.Errorf("ClusterctlMoveRetryPolicy didn't correctly calculate first retry wait for nodeError")
	}

	retry, _ := clustermanager.ClusterctlMoveWaitForInfrastructureRetryPolicy(1, miscellaneousError)
	if retry != false {
		t.Errorf("ClusterctlMoveRetryPolicy didn't not-retry on non-network error")
	}
}

func TestClusterManagerBackupCAPIError(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}

	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().BackupManagement(ctx, from, managementStatePath, from.Name).Return(errors.New("backing up CAPI resources"))

	if err := c.BackupCAPI(ctx, from, managementStatePath, from.Name); err == nil {
		t.Errorf("ClusterManager.BackupCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerMoveCAPISuccess(t *testing.T) {
	for _, providerKind := range []string{"vsphere", "tinkerbell"} {
		t.Run(providerKind, func(t *testing.T) {
			from := &types.Cluster{
				Name: "from-cluster",
			}
			to := &types.Cluster{
				Name: "to-cluster",
			}
			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = to.Name
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
			})
			ctx := context.Background()

			c, m := newClusterManager(t)
			if providerKind == "tinkerbell" {
				clusterSpec.Cluster.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
				clusterSpec.TinkerbellDatacenter = &v1alpha1.TinkerbellDatacenterConfig{
					Spec: v1alpha1.TinkerbellDatacenterConfigSpec{TinkerbellIP: "2001:db8::5"},
				}
				readSource := m.client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, from.KubeconfigFile, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
						list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{addressHardware("2001:db8::10")}
						return nil
					})
				backup := m.writer.EXPECT().Write("tinkerbell-hardware-addresses.json", gomock.Any(), gomock.Any(), gomock.Any()).
					Return("test-cluster/tinkerbell-hardware-addresses.json", nil)
				readTarget := m.client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, to.KubeconfigFile, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
						list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{addressHardware("")}
						return nil
					})
				writeTarget := m.client.EXPECT().MergePatchResourceStatus(ctx, "hardware.tinkerbell.org", "test-hardware", gomock.Any(), to.KubeconfigFile, constants.EksaSystemNamespace)
				gomock.InOrder(readSource, backup, readTarget, writeTarget)
			}
			kcp, mds := getKcpAndMdsForNodeCount(0)
			m.client.EXPECT().GetKubeadmControlPlane(ctx,
				from,
				to.Name,
				gomock.AssignableToTypeOf(executables.WithCluster(from)),
				gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
			).Return(kcp, nil)
			m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
				to.Name,
				gomock.AssignableToTypeOf(executables.WithCluster(from)),
				gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
			).Return(mds, nil)
			m.client.EXPECT().GetMachines(ctx, from, to.Name)
			m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", to.Name)
			m.client.EXPECT().MoveManagement(ctx, from, to, to.Name)
			m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", to.Name)
			m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
			m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, to.Name, to.KubeconfigFile)
			m.client.EXPECT().GetKubeadmControlPlane(ctx,
				to,
				to.Name,
				gomock.AssignableToTypeOf(executables.WithCluster(from)),
				gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
			).Return(kcp, nil)
			m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
				to.Name,
				gomock.AssignableToTypeOf(executables.WithCluster(from)),
				gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
			).Return(mds, nil)
			m.client.EXPECT().GetMachines(ctx, to, to.Name)

			if err := c.MoveCAPI(ctx, from, to, to.Name, clusterSpec); err != nil {
				t.Errorf("ClusterManager.MoveCAPI() error = %v, wantErr nil", err)
			}
		})
	}
}

type isoURLMoveTestCase struct {
	name           string
	tinkerbellIP   string
	bootstrapIP    string
	controlPlaneIP string
	ipv6           bool
	isoBootURL     string
	want           string
}

func TestClusterManagerMoveCAPIISOURLs(t *testing.T) {
	for _, family := range []isoURLMoveTestCase{
		{name: "IPv4", tinkerbellIP: "192.0.2.10", bootstrapIP: "192.0.2.20", controlPlaneIP: "192.0.2.30", want: "http://192.0.2.10:7172/iso/hook.iso"},
		{name: "IPv6", tinkerbellIP: "2001:db8::10", bootstrapIP: "2001:db8::20", controlPlaneIP: "2001:db8::30", ipv6: true, want: "http://[2001:db8::10]:7172/iso6/hook.iso"},
		{name: "IPv6 with IPv4 media", tinkerbellIP: "2001:db8::10", bootstrapIP: "2001:db8::20", controlPlaneIP: "2001:db8::30", ipv6: true, isoBootURL: "http://192.0.2.61:7172/iso6/hook.iso", want: "http://192.0.2.61:7172/iso6/hook.iso"},
	} {
		for _, customTemplate := range []bool{false, true} {
			name := family.name + "/default"
			filename := "../providers/tinkerbell/testdata/cluster_hook_iso_boot.yaml"
			if customTemplate {
				name = family.name + "/override"
				filename = "../providers/tinkerbell/testdata/cluster_tinkerbell_awsiam.yaml"
			}
			t.Run(name, func(t *testing.T) {
				testClusterManagerMoveCAPIISOURL(t, family, filename, customTemplate)
			})
		}
	}
}

func testClusterManagerMoveCAPIISOURL(t *testing.T, family isoURLMoveTestCase, filename string, customTemplate bool) {
	t.Helper()
	g := NewWithT(t)
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, filename)
	spec.TinkerbellDatacenter.Spec.TinkerbellIP = family.tinkerbellIP
	spec.TinkerbellDatacenter.Spec.IsoBoot = true
	spec.TinkerbellDatacenter.Spec.IsoBootURL = family.isoBootURL
	spec.TinkerbellDatacenter.Spec.HookIsoURL = "http://example.invalid/source.iso"
	spec.TinkerbellDatacenter.Spec.OSImageURL = "http://example.invalid/os.tgz"
	spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = family.controlPlaneIP
	if family.ipv6 {
		spec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"2001:db8:100::/56"}
		spec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"2001:db8:200::/112"}
	}
	spec.Cluster.AddTinkerbellIPAnnotation(family.bootstrapIP)
	sourceClient := test.NewFakeKubeClient()
	log := test.NewNullLogger()
	cp, err := tinkerbell.ControlPlaneSpec(ctx, log, sourceClient, spec)
	g.Expect(err).NotTo(HaveOccurred())
	workers, err := tinkerbell.WorkersSpec(ctx, log, sourceClient, spec)
	g.Expect(err).NotTo(HaveOccurred())
	if family.isoBootURL != "" {
		g.Expect(cp.ControlPlaneMachineTemplate.Spec.Template.Spec.BootOptions.ISOURL).To(Equal(family.isoBootURL))
		for _, worker := range workers.Groups {
			g.Expect(worker.ProviderMachineTemplate.Spec.Template.Spec.BootOptions.ISOURL).To(Equal(family.isoBootURL))
		}
	}
	objects := []kubernetes.Object{cp.KubeadmControlPlane, cp.ControlPlaneMachineTemplate}
	objects = append(objects, workers.WorkerObjects()...)
	for _, obj := range objects {
		g.Expect(sourceClient.Create(ctx, obj)).To(Succeed())
	}

	c, m := newClusterManager(t)
	from := &types.Cluster{Name: "bootstrap", KubeconfigFile: "source.kubeconfig"}
	to := &types.Cluster{Name: spec.Cluster.Name, KubeconfigFile: "target.kubeconfig"}
	m.clientFactory.EXPECT().BuildClientFromKubeconfig(from.KubeconfigFile).Return(sourceClient, nil)
	expectMoveNodesReady(m, ctx, from, spec.Cluster.Name)
	expectMoveNodesReady(m, ctx, to, spec.Cluster.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", spec.Cluster.Name)
	m.client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, from.KubeconfigFile, gomock.Any()).Return(nil)
	m.client.EXPECT().MoveManagement(ctx, from, to, spec.Cluster.Name).Return(nil)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", spec.Cluster.Name)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, spec.Cluster.Name)
	m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, spec.Cluster.Name, to.KubeconfigFile)
	g.Expect(c.MoveCAPI(ctx, from, to, spec.Cluster.Name, spec)).To(Succeed())

	spec.Cluster.ClearTinkerbellIPAnnotation()
	workloadCP, err := tinkerbell.ControlPlaneSpec(ctx, log, sourceClient, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workloadCP.ControlPlaneMachineTemplate.Name).To(Equal(cp.ControlPlaneMachineTemplate.Name))
	workloadWorkers, err := tinkerbell.WorkersSpec(ctx, log, sourceClient, spec)
	g.Expect(err).NotTo(HaveOccurred())
	templates := []*tinkerbellv1.TinkerbellMachineTemplate{workloadCP.ControlPlaneMachineTemplate}
	for i, worker := range workloadWorkers.Groups {
		g.Expect(worker.ProviderMachineTemplate.Name).To(Equal(workers.Groups[i].ProviderMachineTemplate.Name))
		templates = append(templates, worker.ProviderMachineTemplate)
	}
	for _, desired := range templates {
		moved := &tinkerbellv1.TinkerbellMachineTemplate{}
		g.Expect(sourceClient.Get(ctx, desired.Name, desired.Namespace, moved)).To(Succeed())
		g.Expect(moved.Spec.Template.Spec.BootOptions.ISOURL).To(Equal(family.want))
		g.Expect(moved.Spec).To(Equal(desired.Spec))
		if customTemplate {
			g.Expect(moved.Spec.Template.Spec.TemplateOverride).To(Equal(cp.ControlPlaneMachineTemplate.Spec.Template.Spec.TemplateOverride))
		}
	}
	g.Expect(spec.TinkerbellDatacenter.Spec.IsoBootURL).To(Equal(family.isoBootURL))
	g.Expect(spec.TinkerbellDatacenter.Spec.HookIsoURL).To(Equal("http://example.invalid/source.iso"))
}

func TestClusterManagerMoveCAPISnapshotRetries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		failures  int
		reads     int
		invalid   string
		saveError bool
		wantError string
	}{
		{name: "transient API failure", failures: 1, reads: 2},
		{name: "exhausted API retries", failures: 2, reads: 2, wantError: "API unavailable"},
		{name: "invalid addresses", reads: 1, invalid: "addresses", wantError: "reading Hardware"},
		{name: "invalid hardware identity", reads: 1, invalid: "mac", wantError: "invalid MAC"},
		{name: "snapshot save failure", reads: 1, saveError: true, wantError: "saving Tinkerbell hardware address snapshot"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			r := retrier.NewWithMaxRetries(2, 0)
			_, m := newClusterManager(t)
			// Match Factory.WithClusterManager: only Kubernetes operations retry.
			client := struct {
				clustermanager.CAPIClient
				clustermanager.KubernetesClient
			}{
				CAPIClient:       m.client,
				KubernetesClient: clustermanager.NewRetrierClient(m.client, r),
			}
			c := clustermanager.New(m.clientFactory, client, m.writer, m.diagnosticsFactory, m.eksaComponents, clustermanager.WithRetrier(r))
			from := &types.Cluster{Name: "bootstrap", KubeconfigFile: "source.kubeconfig"}
			to := &types.Cluster{Name: "workload", KubeconfigFile: "target.kubeconfig"}
			spec := test.NewClusterSpec()
			spec.Cluster.Name = to.Name
			spec.Cluster.Spec.WorkerNodeGroupConfigurations = nil
			spec.Cluster.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
			spec.TinkerbellDatacenter = &v1alpha1.TinkerbellDatacenterConfig{
				Spec: v1alpha1.TinkerbellDatacenterConfigSpec{TinkerbellIP: "2001:db8::5"},
			}
			m.clientFactory.EXPECT().BuildClientFromKubeconfig(from.KubeconfigFile).Return(test.NewFakeKubeClient(), nil)
			expectMoveNodesReady(m, ctx, from, to.Name)
			m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", to.Name)
			hw := addressHardware("2001:db8::10")
			switch tc.invalid {
			case "addresses":
				hw.Object["status"] = map[string]interface{}{"addresses": "invalid"}
			case "mac":
				g.Expect(unstructured.SetNestedSlice(hw.Object, []interface{}{
					map[string]interface{}{"dhcp": map[string]interface{}{"mac": "invalid"}},
				}, "spec", "interfaces")).To(Succeed())
			}
			attempts := 0
			read := m.client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, from.KubeconfigFile, gomock.Any()).
				Times(tc.reads).
				DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
					attempts++
					if attempts <= tc.failures {
						return errors.New("API unavailable")
					}
					list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{hw}
					return nil
				})
			if tc.invalid == "" && tc.failures < tc.reads {
				var saveErr error
				if tc.saveError {
					saveErr = errors.New("disk unavailable")
				}
				save := m.writer.EXPECT().Write("tinkerbell-hardware-addresses.json", gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, data []byte, _ ...filewriter.FileOptionsFunc) (string, error) {
						g.Expect(string(data)).To(MatchJSON(`[{"name":"test-hardware","namespace":"eksa-system","macs":["02:00:00:00:00:01"],"addresses":[{"type":"InternalIP","address":"2001:db8::10"}]}]`))
						return "snapshot.json", saveErr
					})
				gomock.InOrder(read, save)
				if !tc.saveError {
					move := m.client.EXPECT().MoveManagement(ctx, from, to, to.Name).Return(nil)
					restore := m.client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, to.KubeconfigFile, gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
							list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{addressHardware("")}
							return nil
						})
					patch := m.client.EXPECT().MergePatchResourceStatus(ctx, "hardware.tinkerbell.org", "test-hardware", gomock.Any(), to.KubeconfigFile, constants.EksaSystemNamespace)
					gomock.InOrder(save, move, restore, patch)
					m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", to.Name)
					m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
					m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, to.Name, to.KubeconfigFile)
					expectMoveNodesReady(m, ctx, to, to.Name)
				}
			}
			err := c.MoveCAPI(ctx, from, to, to.Name, spec)
			if tc.wantError != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tc.wantError)))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(attempts).To(Equal(tc.reads))
		})
	}
}

func expectMoveNodesReady(m *clusterManagerMocks, ctx context.Context, cluster *types.Cluster, clusterName string) {
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterName, gomock.Any(), gomock.Any()).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx, clusterName, gomock.Any(), gomock.Any()).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, cluster, clusterName)
}

func TestClusterManagerMoveCAPIRetrySuccess(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, to.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", to.Name)
	firstTry := m.client.EXPECT().MoveManagement(ctx, from, to, to.Name).Return(errors.New("Error: failed to connect to the management cluster: action failed after 9 attempts: Get \"https://127.0.0.1:61994/api?timeout=30s\": EOF"))
	secondTry := m.client.EXPECT().MoveManagement(ctx, from, to, to.Name).Return(nil)
	gomock.InOrder(
		firstTry,
		secondTry,
	)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", to.Name)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
	m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, to.Name, to.KubeconfigFile)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		to,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, to, to.Name)

	if err := c.MoveCAPI(ctx, from, to, to.Name, clusterSpec); err != nil {
		t.Errorf("ClusterManager.MoveCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerMoveKubectlWaitRetrySuccess(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, to.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", to.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to, to.Name)
	firstTry := m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", to.Name).Return(errors.New("executing wait: error: the server doesn't have a resource type \"clusters\""))
	secondTry := m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", to.Name).Return(nil)
	gomock.InOrder(
		firstTry,
		secondTry,
	)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
	m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, to.Name, to.KubeconfigFile)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		to,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		to.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, to, to.Name)

	if err := c.MoveCAPI(ctx, from, to, to.Name, clusterSpec); err != nil {
		t.Errorf("ClusterManager.MoveCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerMoveCAPIErrorMove(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to, from.Name).Return(errors.New("error moving"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorWaitForClusterReady(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", from.Name).Return(errors.New("error waiting for cluster to be ready"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorWaitForControlPlane(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to, from.Name)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", from.Name).Return(errors.New("error waiting for control plane"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorGetMachines(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()

	c, m := newClusterManager(t, clustermanager.WithMachineBackoff(0), clustermanager.WithMachineMaxWait(10*time.Microsecond), clustermanager.WithMachineMinWait(20*time.Microsecond))
	kcp, mds := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().WaitForClusterReady(ctx, from, "1h0m0s", from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to, from.Name)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", from.Name)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
	m.client.EXPECT().CountMachineDeploymentReplicasReady(ctx, to.Name, to.KubeconfigFile)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		to,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	m.client.EXPECT().GetMachines(ctx, to, from.Name).Return(nil, errors.New("error getting machines")).AnyTimes()

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorGetKubeadmControlPlane(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()
	c, m := newClusterManager(t, clustermanager.WithMachineBackoff(0), clustermanager.WithMachineMaxWait(10*time.Microsecond), clustermanager.WithMachineMinWait(20*time.Microsecond))
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(nil, errors.New("error getting KubeadmControlPlane"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorGetMachineDeploymentsForCluster(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()
	c, m := newClusterManager(t, clustermanager.WithMachineBackoff(0), clustermanager.WithMachineMaxWait(10*time.Microsecond), clustermanager.WithMachineMinWait(20*time.Microsecond))
	kcp, _ := getKcpAndMdsForNodeCount(0)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		from,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		from.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(from)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(nil, errors.New("error getting MachineDeployments"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestPauseEKSAControllerReconcileWorkloadCluster(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "data-center-name",
			},
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "mgmt-cluster",
			},
		},
	}

	tt.expectPauseClusterReconciliation()

	tt.Expect(tt.clusterManager.PauseEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).To(Succeed())
}

func TestResumeEKSAControllerReconcileWorkloadClusterUpdateAnnotationError(t *testing.T) {
	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))
	tt.clusterSpec.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "data-center-name",
			},
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "mgmt-cluster",
			},
		},
	}

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
		},
	}
	pauseAnnotation := "anywhere.eks.amazonaws.com/paused"

	tt.mocks.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	tt.mocks.provider.EXPECT().MachineResourceType().Return("")
	tt.mocks.provider.EXPECT().DatacenterConfig(tt.clusterSpec).Return(datacenterConfig)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, pauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, pauseAnnotation, tt.cluster, "").Return(errors.New("pause eksa cluster error"))

	tt.Expect(tt.clusterManager.ResumeEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).NotTo(Succeed())
}

func TestResumeEKSAControllerReconcileManagementCluster(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "data-center-name",
			},
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: tt.clusterName,
			},
		},
	}

	tt.clusterSpec.Cluster.PauseReconcile()

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
		},
	}
	pauseAnnotation := "anywhere.eks.amazonaws.com/paused"

	tt.mocks.client.EXPECT().
		ListObjects(tt.ctx, eksaClusterResourceType, "", "", &v1alpha1.ClusterList{}).
		DoAndReturn(func(_ context.Context, _, _, _ string, obj *v1alpha1.ClusterList) error {
			obj.Items = []v1alpha1.Cluster{
				*tt.clusterSpec.Cluster,
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workload-cluster-1",
					},
					Spec: v1alpha1.ClusterSpec{
						DatacenterRef: v1alpha1.Ref{
							Kind: v1alpha1.VSphereDatacenterKind,
							Name: "data-center-name",
						},
						ManagementCluster: v1alpha1.ManagementCluster{
							Name: tt.clusterName,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workload-cluster-2",
					},
					Spec: v1alpha1.ClusterSpec{
						DatacenterRef: v1alpha1.Ref{
							Kind: v1alpha1.VSphereDatacenterKind,
							Name: "data-center-name",
						},
						ManagementCluster: v1alpha1.ManagementCluster{
							Name: "mgmt-cluster-2",
						},
					},
				},
			}
			return nil
		})
	tt.mocks.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType).Times(2)
	tt.mocks.provider.EXPECT().MachineResourceType().Return("").Times(2)
	tt.mocks.provider.EXPECT().DatacenterConfig(tt.clusterSpec).Return(datacenterConfig)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, pauseAnnotation, tt.cluster, "").Return(nil).Times(2)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, pauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		tt.clusterSpec.Cluster.Name,
		v1alpha1.ManagedByCLIAnnotation,
		tt.cluster,
		"",
	).Return(nil)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaClusterResourceType, "workload-cluster-1", pauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		"workload-cluster-1",
		v1alpha1.ManagedByCLIAnnotation,
		tt.cluster,
		"",
	).Return(nil)

	tt.Expect(tt.clusterManager.ResumeEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).To(Succeed())
}

func TestResumeEKSAControllerReconcileManagementClusterListObjectsError(t *testing.T) {
	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))
	tt.clusterSpec.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: tt.clusterName,
			},
		},
	}

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: tt.clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
		},
	}

	tt.mocks.provider.EXPECT().DatacenterConfig(tt.clusterSpec).Return(datacenterConfig)

	tt.mocks.client.EXPECT().ListObjects(tt.ctx, eksaClusterResourceType, "", "", &v1alpha1.ClusterList{}).Return(errors.New("list error"))

	tt.Expect(tt.clusterManager.ResumeEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).NotTo(Succeed())
}

func TestPauseEKSAControllerReconcileWorkloadClusterWithMachineConfig(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster = &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "datasourcename",
			},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{
					Name: tt.clusterName + "-cp",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				MachineGroupRef: &v1alpha1.Ref{
					Name: tt.clusterName,
				},
			}},
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "mgmt-cluster",
			},
		},
	}

	tt.mocks.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	tt.mocks.provider.EXPECT().MachineResourceType().Return(eksaVSphereMachineResourceType).Times(3)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereMachineResourceType, tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereMachineResourceType, tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		tt.clusterSpec.Cluster.Name,
		map[string]string{
			v1alpha1.ManagedByCLIAnnotation: "true",
		},
		tt.cluster,
		"",
	).Return(nil)

	tt.Expect(tt.clusterManager.PauseEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).To(Succeed())
}

type testSetup struct {
	*WithT
	clusterManager       *clustermanager.ClusterManager
	mocks                *clusterManagerMocks
	ctx                  context.Context
	managementComponents *cluster.ManagementComponents
	clusterSpec          *cluster.Spec
	cluster              *types.Cluster
	clusterName          string
}

func (tt *testSetup) expectPauseClusterReconciliation() *gomock.Call {
	lastCall := tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		tt.clusterSpec.Cluster.Name,
		map[string]string{
			v1alpha1.ManagedByCLIAnnotation: "true",
		},
		tt.cluster,
		"",
	).Return(nil)
	gomock.InOrder(
		tt.mocks.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType),
		tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil),
		tt.mocks.provider.EXPECT().MachineResourceType().Return(""),
		tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil),
		lastCall,
	)

	return lastCall
}

func newTest(t *testing.T, opts ...clustermanager.ClusterManagerOpt) *testSetup {
	c, m := newClusterManager(t, opts...)
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec()
	return &testSetup{
		WithT:                NewWithT(t),
		clusterManager:       c,
		mocks:                m,
		ctx:                  context.Background(),
		managementComponents: cluster.ManagementComponentsFromBundles(clusterSpec.Bundles),
		clusterSpec:          clusterSpec,
		cluster: &types.Cluster{
			Name: clusterName,
		},
		clusterName: clusterName,
	}
}

type clusterManagerMocks struct {
	clientFactory      *mocksmanager.MockClientFactory
	writer             *mockswriter.MockFileWriter
	client             *mocksmanager.MockClusterClient
	provider           *mocksprovider.MockProvider
	diagnosticsBundle  *mocksdiagnostics.MockDiagnosticBundle
	diagnosticsFactory *mocksdiagnostics.MockDiagnosticBundleFactory
	eksaComponents     *mocksmanager.MockEKSAComponents
}

func newClusterManager(t *testing.T, opts ...clustermanager.ClusterManagerOpt) (*clustermanager.ClusterManager, *clusterManagerMocks) {
	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		client:             mocksmanager.NewMockClusterClient(mockCtrl),
		provider:           mocksprovider.NewMockProvider(mockCtrl),
		diagnosticsFactory: mocksdiagnostics.NewMockDiagnosticBundleFactory(mockCtrl),
		diagnosticsBundle:  mocksdiagnostics.NewMockDiagnosticBundle(mockCtrl),
		eksaComponents:     mocksmanager.NewMockEKSAComponents(mockCtrl),
	}

	clusterName := "cluster-name"
	dc := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	oc := &v1alpha1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	b := test.Bundle()
	r := test.EksdRelease("1-19")
	ac := &v1alpha1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	gc := &v1alpha1.GitOpsConfig{}
	er := test.EKSARelease()

	fakeClient := test.NewFakeKubeClient(dc, oc, b, r, ac, gc, er)
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	m.clientFactory = cf
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.writer, m.diagnosticsFactory, m.eksaComponents, opts...)

	return c, m
}

func TestClusterManagerGetCurrentClusterSpecGetClusterError(t *testing.T) {
	tt := newTest(t)

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterName).Return(nil, errors.New("error from client"))

	_, err := tt.clusterManager.GetCurrentClusterSpec(tt.ctx, tt.cluster, tt.clusterName)
	tt.Expect(err).ToNot(BeNil())
}

func TestClusterManagerGetCurrentClusterSpecGetBundlesError(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.BundlesRef = &v1alpha1.BundlesRef{}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterName).Return(tt.clusterSpec.Cluster, nil)

	_, err := tt.clusterManager.GetCurrentClusterSpec(tt.ctx, tt.cluster, tt.clusterName)
	tt.Expect(err).ToNot(BeNil())
}

func TestCreateRegistryCredSecretSuccess(t *testing.T) {
	tt := newTest(t)

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      "registry-credentials",
		},
		Data: map[string][]byte{
			"username": []byte(""),
			"password": []byte(""),
		},
	}

	tt.mocks.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, secret).Return(nil)

	err := tt.clusterManager.CreateRegistryCredSecret(tt.ctx, tt.cluster)
	tt.Expect(err).To(BeNil())
}

func TestAllowDeleteWhilePaused(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "success allow delete while paused",
			err:  nil,
		},
		{
			name: "fail allow delete while paused",
			err:  fmt.Errorf("failure"),
		},
	}
	allowDelete := map[string]string{v1alpha1.AllowDeleteWhenPausedAnnotation: "true"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tt := newTest(t)
			cluster := tt.clusterSpec.Cluster
			tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, cluster.ResourceType(), cluster.Name, allowDelete, tt.cluster, cluster.Namespace).Return(test.err)
			err := tt.clusterManager.AllowDeleteWhilePaused(tt.ctx, tt.cluster, tt.clusterSpec)
			expectedErr := fmt.Errorf("updating paused annotation in cluster reconciliation: %v", test.err)
			tt.Expect(err).To(Or(BeNil(), MatchError(expectedErr)))
		})
	}
}
