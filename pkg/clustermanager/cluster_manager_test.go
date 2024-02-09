package clustermanager_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	mocksdiagnostics "github.com/aws/eks-anywhere/pkg/diagnostics/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
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

func TestClusterManagerInstallNetworkingSuccess(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}

	clusterSpec := test.NewClusterSpec()

	c, m := newClusterManager(t)
	m.provider.EXPECT().GetDeployments()
	m.networking.EXPECT().Install(ctx, cluster, clusterSpec, []string{})

	if err := c.InstallNetworking(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallNetworking() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallNetworkingNetworkingError(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterSpec := test.NewClusterSpec()

	c, m := newClusterManager(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))
	m.provider.EXPECT().GetDeployments()
	m.networking.EXPECT().Install(ctx, cluster, clusterSpec, []string{}).Return(errors.New("error in networking"))

	if err := c.InstallNetworking(ctx, cluster, clusterSpec, m.provider); err == nil {
		t.Errorf("ClusterManager.InstallNetworking() error = nil, wantErr not nil")
	}
}

func getKcpAndMdsForNodeCount(count int32) (*controlplanev1.KubeadmControlPlane, []clusterv1.MachineDeployment) {
	kcp := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			Replicas: ptr.Int32(count),
		},
	}

	md := []clusterv1.MachineDeployment{
		{
			Spec: clusterv1.MachineDeploymentSpec{
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

	m.diagnosticsFactory.EXPECT().DiagnosticBundleWorkloadCluster(clusterSpec, m.provider, workloadCluster.KubeconfigFile).Return(b, nil)
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

func TestClusterManagerCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, mgmtCluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mgmtCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneAvailable(ctx, mgmtCluster, "1h0m0s", clusterName)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, mgmtCluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, mgmtCluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterErrorGetKubeconfig(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Name = tt.clusterName
	gomock.InOrder(
		tt.mocks.provider.EXPECT().GenerateCAPISpecForCreate(tt.ctx, tt.cluster, tt.clusterSpec),
		tt.mocks.writer.EXPECT().Write(tt.clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil())),
		tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace),
		tt.mocks.client.EXPECT().WaitForControlPlaneAvailable(tt.ctx, tt.cluster, "1h0m0s", tt.clusterName),
		tt.mocks.client.EXPECT().GetWorkloadKubeconfig(tt.ctx, tt.clusterName, tt.cluster).Return(nil, errors.New("get kubeconfig error")),
	)

	_, err := tt.clusterManager.CreateWorkloadCluster(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)
	tt.Expect(err).To(MatchError(ContainSubstring("get kubeconfig error")))
}

func TestClusterManagerCreateWorkloadClusterTimeoutOverrideSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}

	c, m := newClusterManager(t, clustermanager.WithControlPlaneWaitTimeout(20*time.Minute))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, mgmtCluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mgmtCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneAvailable(ctx, mgmtCluster, "20m0s", clusterName)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, mgmtCluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, mgmtCluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerRunPostCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	workloadCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "workload-kubeconfig",
	}

	kcp, mds := getKcpAndMdsForNodeCount(0)

	c, m := newClusterManager(t)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		mgmtCluster,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)

	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)

	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).AnyTimes().Return([]types.Machine{}, nil)

	if err := c.RunPostCreateWorkloadCluster(ctx, mgmtCluster, workloadCluster, clusterSpec); err != nil {
		t.Errorf("ClusterManager.RunPostCreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterWithExternalEtcdSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 2
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, mgmtCluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mgmtCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForManagedExternalEtcdReady(ctx, mgmtCluster, "1h0m0s", clusterName)
	m.client.EXPECT().WaitForControlPlaneAvailable(ctx, mgmtCluster, "1h0m0s", clusterName)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, mgmtCluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, mgmtCluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterWithExternalEtcdTimeoutOverrideSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 2
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}

	c, m := newClusterManager(t, clustermanager.WithExternalEtcdWaitTimeout(30*time.Minute))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, mgmtCluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mgmtCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForManagedExternalEtcdReady(ctx, mgmtCluster, "30m0s", clusterName)
	m.client.EXPECT().WaitForControlPlaneAvailable(ctx, mgmtCluster, "1h0m0s", clusterName)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, mgmtCluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, mgmtCluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerRunPostCreateWorkloadClusterWaitForMachinesTimeout(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	workloadCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "workload-kubeconfig",
	}

	c, m := newClusterManager(t, clustermanager.WithMachineBackoff(1*time.Nanosecond), clustermanager.WithMachineMaxWait(50*time.Microsecond), clustermanager.WithMachineMinWait(100*time.Microsecond))

	kcp, mds := getKcpAndMdsForNodeCount(1)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		mgmtCluster,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)

	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)

	// Fail once
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""},
	}}}, nil)
	if err := c.RunPostCreateWorkloadCluster(ctx, mgmtCluster, workloadCluster, clusterSpec); err == nil {
		t.Error("ClusterManager.RunPostCreateWorkloadCluster() error = nil, wantErr not nil", err)
	}
}

func TestClusterManagerRunPostCreateWorkloadClusterWaitForMachinesSuccessAfterRetries(t *testing.T) {
	retries := 10
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	})

	mgmtCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "mgmt-kubeconfig",
	}
	workloadCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "workload-kubeconfig",
	}

	c, m := newClusterManager(t, clustermanager.WithMachineBackoff(1*time.Nanosecond), clustermanager.WithMachineMaxWait(1*time.Minute), clustermanager.WithMachineMinWait(2*time.Minute))

	kcp, mds := getKcpAndMdsForNodeCount(1)
	m.client.EXPECT().GetKubeadmControlPlane(ctx,
		mgmtCluster,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)

	m.client.EXPECT().GetMachineDeploymentsForCluster(ctx,
		mgmtCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mgmtCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)

	// Fail a bunch of times
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).Times(retries-5).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef  times
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).Times(3).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""},
	}}}, nil)
	//// Return a machine with nodeRef + NodeHealthy condition and another with it
	status := types.MachineStatus{
		NodeRef: &types.ResourceRef{},
		Conditions: types.Conditions{
			{
				Type:   "NodeHealthy",
				Status: "True",
			},
		},
	}
	machines := []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""}}},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""}}, Status: status},
	}
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).Times(1).Return(machines, nil)
	// Finally return two machines with node ref
	machines = []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""}}, Status: status},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""}}, Status: status},
	}
	m.client.EXPECT().GetMachines(ctx, mgmtCluster, mgmtCluster.Name).Times(1).Return(machines, nil)
	if err := c.RunPostCreateWorkloadCluster(ctx, mgmtCluster, workloadCluster, clusterSpec); err != nil {
		t.Errorf("ClusterManager.RunPostCreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeSelfManagedClusterFailClientError(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)
	mockCtrl := gomock.NewController(t)

	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(nil, errors.New("can't build client"))
	c := clustermanager.New(cf, tt.mocks.client, nil, nil, nil, nil, nil)
	tt.clusterManager = c

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Errorf("ClusterManager.UpgradeCluster() error is nil, wantErr %v", err)
	}
}

func TestClusterManagerUpgradeSelfManagedClusterSuccess(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt := newSpecChangedTest(t)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, tt.cluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeSelfManagedClusterWithUnstackedEtcdSuccess(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)

	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}
	tt.oldClusterConfig.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdReady(tt.ctx, mCluster, "1h0m0s", clusterName)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(8)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, tt.cluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeSelfManagedClusterWithUnstackedEtcdTimeoutNotReadySuccess(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)

	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}
	tt.oldClusterConfig.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdReady(tt.ctx, mCluster, "1h0m0s", clusterName)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdNotReady(tt.ctx, mCluster, "1m", clusterName).Return(errors.New("timed out waiting for the condition on clusters"))
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(8)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, tt.cluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeSelfManagedClusterWithUnstackedEtcdNotReadyError(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)

	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}
	tt.oldClusterConfig.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdNotReady(tt.ctx, mCluster, "1m", clusterName).Return(errors.New("etcd not ready"))
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	tt.Expect(tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider)).To(MatchError(ContainSubstring("etcd not ready")))
}

func TestClusterManagerUpgradeSelfManagedClusterWithUnstackedEtcdErrorRemovingAnnotation(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))

	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}
	tt.oldClusterConfig.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().WaitForManagedExternalEtcdReady(tt.ctx, mCluster, "1h0m0s", clusterName).Return(errors.New("timed out"))
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any(), mCluster, constants.EksaSystemNamespace).Return(errors.New("removing annotation"))
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	tt.Expect(tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider)).To(MatchError(ContainSubstring("removing annotation")))
}

func TestClusterManagerUpgradeWorkloadClusterSuccess(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterAWSIamConfigSuccess(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	// Adding AWSIamConfig to the cluster spec.
	oldIamConfig := &v1alpha1.AWSIamConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSIamConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            tt.clusterName,
			ResourceVersion: "999",
		},
	}
	tt.oldClusterConfig.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.AWSIamConfigKind, Name: oldIamConfig.Name}}
	tt.newClusterConfig = tt.oldClusterConfig.DeepCopy()

	r := []eksdv1.Release{
		*test.EksdRelease("1-19"),
	}
	er := test.EKSARelease()
	er.ResourceVersion = "999"

	config := &cluster.Config{
		Cluster:               tt.oldClusterConfig,
		VSphereDatacenter:     tt.oldDatacenterConfig,
		OIDCConfigs:           map[string]*v1alpha1.OIDCConfig{},
		VSphereMachineConfigs: map[string]*v1alpha1.VSphereMachineConfig{},
		AWSIAMConfigs: map[string]*v1alpha1.AWSIamConfig{
			oldIamConfig.Name: oldIamConfig,
		},
	}

	cs, _ := cluster.NewSpec(config, tt.clusterSpec.Bundles, r, er)

	tt.clusterSpec = cs

	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)
	tt.mocks.awsIamAuth.EXPECT().UpgradeAWSIAMAuth(tt.ctx, wCluster, tt.clusterSpec).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeCloudStackWorkloadClusterSuccess(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMDReadyErrorOnce(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	// Fail once
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Times(1).Return(0, 0, errors.New("error counting MD replicas"))
	// Return 1 and 1 for ready and total replicas
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Times(1).Return(1, 1, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMDReadyUnreadyOnce(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	// Return 0 and 1 for ready and total replicas once
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Times(1).Return(0, 1, nil)
	// Return 1 and 1 for ready and total replicas
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, mCluster.Name, mCluster.KubeconfigFile).Times(1).Return(1, 1, nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMachinesTimeout(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}

	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t, clustermanager.WithMachineBackoff(1*time.Nanosecond), clustermanager.WithMachineMaxWait(50*time.Microsecond), clustermanager.WithMachineMinWait(100*time.Microsecond))
	kcp, _ := getKcpAndMdsForNodeCount(1)

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "1h0m0s", clusterName)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	// Fail once
	tt.mocks.client.EXPECT().GetMachines(ctx, mCluster, mCluster.Name).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	tt.mocks.client.EXPECT().GetMachines(ctx, mCluster, mCluster.Name).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""},
	}}}, nil)

	if err := tt.clusterManager.UpgradeCluster(ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerUpgradeWorkloadClusterGetMachineDeploymentError(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, _ := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(nil, errors.New("get md err"))
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	tt.Expect(tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider)).To(MatchError(ContainSubstring("md err")))
}

func TestClusterManagerUpgradeWorkloadClusterRemoveOldWorkerNodeGroupsError(t *testing.T) {
	mgmtClusterName := "cluster-name"
	workClusterName := "cluster-name-w"

	mCluster := &types.Cluster{
		Name: mgmtClusterName,
	}
	wCluster := &types.Cluster{
		Name: workClusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, mCluster, mgmtClusterName).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, mCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", mgmtClusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", mgmtClusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile).Return(errors.New("delete wng error"))
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, mCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, mCluster.Name).Return(nil)
	tt.mocks.writer.EXPECT().Write(mgmtClusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	tt.Expect(tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider)).To(MatchError(ContainSubstring("wng err")))
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMachinesFailedWithUnhealthyNode(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	status := types.MachineStatus{
		NodeRef: &types.ResourceRef{},
		Conditions: types.Conditions{
			{
				Type:   "NodeHealthy",
				Status: "False",
			},
		},
	}
	machines := []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneNameLabel: ""}}, Status: status},
	}

	tt := newSpecChangedTest(t, clustermanager.WithMachineBackoff(1*time.Nanosecond), clustermanager.WithMachineMaxWait(50*time.Microsecond), clustermanager.WithMachineMinWait(100*time.Microsecond))
	kcp, _ := getKcpAndMdsForNodeCount(1)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", clusterName).MaxTimes(5)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	// Return a machine with no nodeRef the rest of the retries
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).MinTimes(1).Return(machines, nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForCAPITimeout(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)
	kcp, mds := getKcpAndMdsForNodeCount(0)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "1h0m0s", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetKubeadmControlPlane(tt.ctx,
		mCluster,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(kcp, nil)
	tt.mocks.client.EXPECT().GetMachineDeploymentsForCluster(tt.ctx,
		mCluster.Name,
		gomock.AssignableToTypeOf(executables.WithCluster(mCluster)),
		gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace)),
	).Return(mds, nil)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.client.EXPECT().GetMachineDeployment(tt.ctx, "cluster-name-md-0", gomock.AssignableToTypeOf(executables.WithKubeconfig(mCluster.KubeconfigFile)), gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(&mds[0], nil)
	tt.mocks.client.EXPECT().DeleteOldWorkerNodeGroup(tt.ctx, &mds[0], mCluster.KubeconfigFile)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m0s", "Available", gomock.Any(), gomock.Any()).Return(errors.New("time out"))
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().CountMachineDeploymentReplicasReady(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(0, 0, nil)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.networking.EXPECT().RunPostControlPlaneUpgradeSetup(tt.ctx, wCluster).Return(nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
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

func TestClusterManagerCreateEKSAResourcesSuccess(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"v1.28": "test-image",
		},
	}

	c, m := newClusterManager(t)

	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, gomock.Any())
	// ApplyKubeSpecFromBytes is called twice. Once for Bundles and again for EKSARelease.
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, gomock.Any()).MaxTimes(2)
	m.client.EXPECT().GetConfigMap(ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(configMap, nil)
	m.client.EXPECT().Apply(ctx, tt.cluster.KubeconfigFile, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, tt.cluster, gomock.Any(), gomock.Any()).MaxTimes(2)
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).To(Succeed())
	_, ok := datacenterConfig.GetAnnotations()["anywhere.eks.amazonaws.com/paused"]
	tt.Expect(ok).To(BeTrue())
	_, ok = tt.clusterSpec.Cluster.GetAnnotations()["anywhere.eks.amazonaws.com/paused"]
	tt.Expect(ok).To(BeTrue())
}

func TestClusterManagerCreateEKSAResourcesFailure(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Namespace = "test_namespace"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	c, m := newClusterManager(t)

	m.client.EXPECT().CreateNamespaceIfNotPresent(ctx, gomock.Any(), tt.clusterSpec.Cluster.Namespace).Return(errors.New(""))
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).NotTo(Succeed())
}

func TestClusterManagerCreateEKSAResourcesFailureBundles(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Namespace = "test_namespace"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
		client:             mocksmanager.NewMockClusterClient(mockCtrl),
		provider:           mocksprovider.NewMockProvider(mockCtrl),
		diagnosticsFactory: mocksdiagnostics.NewMockDiagnosticBundleFactory(mockCtrl),
		diagnosticsBundle:  mocksdiagnostics.NewMockDiagnosticBundle(mockCtrl),
		eksaComponents:     mocksmanager.NewMockEKSAComponents(mockCtrl),
	}
	fakeClient := test.NewFakeKubeClient()
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents)

	m.client.EXPECT().CreateNamespaceIfNotPresent(ctx, gomock.Any(), tt.clusterSpec.Cluster.Namespace).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(errors.New(""))
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).NotTo(Succeed())
}

func TestClusterManagerCreateEKSAResourcesFailureEKSARelease(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Namespace = "test_namespace"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"v1.28": "test-image",
		},
	}

	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
		client:             mocksmanager.NewMockClusterClient(mockCtrl),
		provider:           mocksprovider.NewMockProvider(mockCtrl),
		diagnosticsFactory: mocksdiagnostics.NewMockDiagnosticBundleFactory(mockCtrl),
		diagnosticsBundle:  mocksdiagnostics.NewMockDiagnosticBundle(mockCtrl),
		eksaComponents:     mocksmanager.NewMockEKSAComponents(mockCtrl),
	}
	fakeClient := test.NewFakeKubeClient()
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents)

	m.client.EXPECT().CreateNamespaceIfNotPresent(ctx, gomock.Any(), tt.clusterSpec.Cluster.Namespace).Return(nil)
	m.client.EXPECT().GetConfigMap(ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(configMap, nil)
	m.client.EXPECT().Apply(ctx, tt.cluster.KubeconfigFile, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(errors.New(""))
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).NotTo(Succeed())
}

func TestClusterManagerCreateEKSAResourcesNewUpgraderConfigMap(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Namespace = "test_namespace"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
		client:             mocksmanager.NewMockClusterClient(mockCtrl),
		provider:           mocksprovider.NewMockProvider(mockCtrl),
		diagnosticsFactory: mocksdiagnostics.NewMockDiagnosticBundleFactory(mockCtrl),
		diagnosticsBundle:  mocksdiagnostics.NewMockDiagnosticBundle(mockCtrl),
		eksaComponents:     mocksmanager.NewMockEKSAComponents(mockCtrl),
	}
	fakeClient := test.NewFakeKubeClient()
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents)

	m.client.EXPECT().CreateNamespaceIfNotPresent(ctx, gomock.Any(), tt.clusterSpec.Cluster.Namespace).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, gomock.Any())
	// ApplyKubeSpecFromBytes is called twice. Once for Bundles and again for EKSARelease.
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, gomock.Any()).MaxTimes(2)
	m.client.EXPECT().GetConfigMap(ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New("NotFound"))
	m.client.EXPECT().Apply(ctx, tt.cluster.KubeconfigFile, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, tt.cluster, gomock.Any(), gomock.Any()).MaxTimes(2)
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).To(Succeed())
	_, ok := datacenterConfig.GetAnnotations()["anywhere.eks.amazonaws.com/paused"]
	tt.Expect(ok).To(BeTrue())
	_, ok = tt.clusterSpec.Cluster.GetAnnotations()["anywhere.eks.amazonaws.com/paused"]
	tt.Expect(ok).To(BeTrue())
}

func TestClusterManagerCreateEKSAResourcesFailureApplyUpgraderConfigMap(t *testing.T) {
	features.ClearCache()
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Namespace = "test_namespace"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"v1.28": "test-image",
		},
	}

	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
		client:             mocksmanager.NewMockClusterClient(mockCtrl),
		provider:           mocksprovider.NewMockProvider(mockCtrl),
		diagnosticsFactory: mocksdiagnostics.NewMockDiagnosticBundleFactory(mockCtrl),
		diagnosticsBundle:  mocksdiagnostics.NewMockDiagnosticBundle(mockCtrl),
		eksaComponents:     mocksmanager.NewMockEKSAComponents(mockCtrl),
	}
	fakeClient := test.NewFakeKubeClient()
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents)

	m.client.EXPECT().CreateNamespaceIfNotPresent(ctx, gomock.Any(), tt.clusterSpec.Cluster.Namespace).Return(nil)
	m.client.EXPECT().GetConfigMap(ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(configMap, nil)
	m.client.EXPECT().Apply(ctx, tt.cluster.KubeconfigFile, gomock.Any()).Return(errors.New(""))
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any()).Return(nil)
	tt.Expect(c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs)).NotTo(Succeed())
}

func expectedMachineHealthCheck(unhealthyMachineTimeout, nodeStartupTimeout time.Duration) []byte {
	healthCheck := fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  creationTimestamp: null
  name: fluxTestCluster-worker-1-worker-unhealthy
  namespace: eksa-system
spec:
  clusterName: fluxTestCluster
  maxUnhealthy: 40%%
  nodeStartupTimeout: %[2]s
  selector:
    matchLabels:
      cluster.x-k8s.io/deployment-name: fluxTestCluster-worker-1
  unhealthyConditions:
  - status: Unknown
    timeout: %[1]s
    type: Ready
  - status: "False"
    timeout: %[1]s
    type: Ready
status:
  currentHealthy: 0
  expectedMachines: 0
  remediationsAllowed: 0

---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  creationTimestamp: null
  name: fluxTestCluster-kcp-unhealthy
  namespace: eksa-system
spec:
  clusterName: fluxTestCluster
  maxUnhealthy: 100%%
  nodeStartupTimeout: %[2]s
  selector:
    matchLabels:
      cluster.x-k8s.io/control-plane: ""
  unhealthyConditions:
  - status: Unknown
    timeout: %[1]s
    type: Ready
  - status: "False"
    timeout: %[1]s
    type: Ready
status:
  currentHealthy: 0
  expectedMachines: 0
  remediationsAllowed: 0

---
`, unhealthyMachineTimeout, nodeStartupTimeout)
	return []byte(healthCheck)
}

func TestInstallMachineHealthChecks(t *testing.T) {
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name = "worker-1"
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: constants.DefaultUnhealthyMachineTimeout,
		},
		NodeStartupTimeout: &metav1.Duration{
			Duration: constants.DefaultNodeStartupTimeout,
		},
	}
	wantMHC := expectedMachineHealthCheck(constants.DefaultUnhealthyMachineTimeout, constants.DefaultNodeStartupTimeout)
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, wantMHC)

	if err := tt.clusterManager.InstallMachineHealthChecks(ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("ClusterManager.InstallMachineHealthChecks() error = %v, wantErr nil", err)
	}
}

func TestInstallMachineHealthChecksWithTimeoutOverride(t *testing.T) {
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: (30 * time.Minute),
		},
		NodeStartupTimeout: &metav1.Duration{
			Duration: (30 * time.Minute),
		},
	}
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name = "worker-1"
	wantMHC := expectedMachineHealthCheck(30*time.Minute, 30*time.Minute)
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, wantMHC)

	if err := tt.clusterManager.InstallMachineHealthChecks(ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("ClusterManager.InstallMachineHealthChecks() error = %v, wantErr nil", err)
	}
}

func TestInstallMachineHealthChecksWithNoTimeout(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name = "worker-1"
	maxTime := time.Duration(math.MaxInt64)
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: maxTime,
		},
		NodeStartupTimeout: &metav1.Duration{
			Duration: maxTime,
		},
	}
	wantMHC := expectedMachineHealthCheck(maxTime, maxTime)

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, wantMHC)

	tt.Expect(tt.clusterManager.InstallMachineHealthChecks(tt.ctx, tt.clusterSpec, tt.cluster)).To(Succeed())
}

func TestInstallMachineHealthChecksApplyError(t *testing.T) {
	ctx := context.Background()
	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(2, 0)))
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name = "worker-1"
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: constants.DefaultUnhealthyMachineTimeout,
		},
		NodeStartupTimeout: &metav1.Duration{
			Duration: constants.DefaultNodeStartupTimeout,
		},
	}
	wantMHC := expectedMachineHealthCheck(clustermanager.DefaultUnhealthyMachineTimeout, clustermanager.DefaultNodeStartupTimeout)
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, wantMHC).Return(errors.New("apply error")).MaxTimes(2)

	if err := tt.clusterManager.InstallMachineHealthChecks(ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("ClusterManager.InstallMachineHealthChecks() error = nil, wantErr apply error")
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

func TestPauseEKSAControllerReconcileWorkloadClusterUpdateAnnotationError(t *testing.T) {
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

	tt.mocks.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	tt.mocks.provider.EXPECT().MachineResourceType().Return("")
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, expectedPauseAnnotation, tt.cluster, "").Return(errors.New("pause eksa cluster error"))

	tt.Expect(tt.clusterManager.PauseEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).NotTo(Succeed())
}

func TestPauseEKSAControllerReconcileManagementCluster(t *testing.T) {
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
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaVSphereDatacenterResourceType, tt.clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, tt.cluster, "").Return(nil).Times(2)
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
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(tt.ctx, eksaClusterResourceType, "workload-cluster-1", expectedPauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().UpdateAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		"workload-cluster-1",
		map[string]string{
			v1alpha1.ManagedByCLIAnnotation: "true",
		},
		tt.cluster,
		"",
	).Return(nil)

	tt.Expect(tt.clusterManager.PauseEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).To(Succeed())
}

func TestPauseEKSAControllerReconcileManagementClusterListObjectsError(t *testing.T) {
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

	tt.mocks.client.EXPECT().ListObjects(tt.ctx, eksaClusterResourceType, "", "", &v1alpha1.ClusterList{}).Return(errors.New("list error"))

	tt.Expect(tt.clusterManager.PauseEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).NotTo(Succeed())
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

func TestResumeEKSAControllerReconcileWorkloadCluster(t *testing.T) {
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
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, eksaClusterResourceType, tt.clusterSpec.Cluster.Name, pauseAnnotation, tt.cluster, "").Return(nil)
	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(
		tt.ctx,
		eksaClusterResourceType,
		tt.clusterSpec.Cluster.Name,
		v1alpha1.ManagedByCLIAnnotation,
		tt.cluster,
		"",
	).Return(nil)

	tt.Expect(tt.clusterManager.ResumeEKSAControllerReconcile(tt.ctx, tt.cluster, tt.clusterSpec, tt.mocks.provider)).To(Succeed())
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
	annotations := tt.clusterSpec.Cluster.GetAnnotations()
	if _, ok := annotations[pauseAnnotation]; ok {
		t.Errorf("%s annotation exists, should be removed", pauseAnnotation)
	}
	if _, ok := annotations[v1alpha1.ManagedByCLIAnnotation]; ok {
		t.Errorf("%s annotation exists, should be removed", v1alpha1.ManagedByCLIAnnotation)
	}
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

func TestClusterManagerInstallCustomComponentsSuccess(t *testing.T) {
	features.ClearCache()
	tt := newTest(t)

	tt.mocks.eksaComponents.EXPECT().Install(tt.ctx, logger.Get(), tt.cluster, tt.managementComponents, tt.clusterSpec)
	tt.mocks.provider.EXPECT().InstallCustomProviderComponents(tt.ctx, tt.cluster.KubeconfigFile)
	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.managementComponents, tt.clusterSpec, tt.cluster, tt.mocks.provider); err != nil {
		t.Errorf("ClusterManager.InstallCustomComponents() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallCustomComponentsErrorInstalling(t *testing.T) {
	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(2, 0)))

	tt.mocks.eksaComponents.EXPECT().Install(tt.ctx, logger.Get(), tt.cluster, tt.managementComponents, tt.clusterSpec).Return(errors.New("error from apply"))

	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.managementComponents, tt.clusterSpec, tt.cluster, nil); err == nil {
		t.Error("ClusterManager.InstallCustomComponents() error = nil, wantErr not nil")
	}
}

type specChangedTest struct {
	*testSetup
	oldClusterConfig, newClusterConfig                         *v1alpha1.Cluster
	oldDatacenterConfig, newDatacenterConfig                   *v1alpha1.VSphereDatacenterConfig
	oldControlPlaneMachineConfig, newControlPlaneMachineConfig *v1alpha1.VSphereMachineConfig
	oldWorkerMachineConfig, newWorkerMachineConfig             *v1alpha1.VSphereMachineConfig
	oldOIDCConfig                                              *v1alpha1.OIDCConfig
}

func newSpecChangedTest(t *testing.T, opts ...clustermanager.ClusterManagerOpt) *specChangedTest {
	testSetup := newTest(t, opts...)
	clusterName := testSetup.clusterName
	version := test.DevEksaVersion()
	clusterConfig := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: "1.19",
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Name: clusterName,
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Count: ptr.Int(1),
				MachineGroupRef: &v1alpha1.Ref{
					Name: clusterName + "-worker",
				},
			}},
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: clusterName,
			},
			IdentityProviderRefs: []v1alpha1.Ref{{
				Kind: v1alpha1.OIDCConfigKind,
				Name: clusterName,
			}},
			EksaVersion: &version,
		},
	}
	newClusterConfig := clusterConfig.DeepCopy()
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            clusterName,
			ResourceVersion: "999",
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: false,
		},
	}
	machineConfig := &v1alpha1.VSphereMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			DiskGiB:   20,
			MemoryMiB: 8192,
			NumCPUs:   2,
		},
	}

	workerMachineConfig := machineConfig.DeepCopy()
	workerMachineConfig.Name = clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name

	oidcConfig := &v1alpha1.OIDCConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OIDCConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "cluster-name",
			ResourceVersion: "999",
		},
	}

	changedTest := &specChangedTest{
		testSetup:                    testSetup,
		oldClusterConfig:             clusterConfig,
		newClusterConfig:             newClusterConfig,
		oldDatacenterConfig:          datacenterConfig,
		newDatacenterConfig:          datacenterConfig.DeepCopy(),
		oldControlPlaneMachineConfig: machineConfig,
		newControlPlaneMachineConfig: machineConfig.DeepCopy(),
		oldWorkerMachineConfig:       workerMachineConfig,
		newWorkerMachineConfig:       workerMachineConfig.DeepCopy(),
		oldOIDCConfig:                oidcConfig,
	}

	b := test.Bundle()
	b.ResourceVersion = "999"
	er := test.EKSARelease()
	er.ResourceVersion = "999"

	config := &cluster.Config{
		Cluster:               newClusterConfig,
		VSphereDatacenter:     datacenterConfig,
		OIDCConfigs:           map[string]*v1alpha1.OIDCConfig{clusterName: oidcConfig},
		VSphereMachineConfigs: map[string]*v1alpha1.VSphereMachineConfig{},
		AWSIAMConfigs:         map[string]*v1alpha1.AWSIamConfig{},
	}

	r := []eksdv1.Release{
		*test.EksdRelease("1-19"),
	}

	cs, err := cluster.NewSpec(config, b, r, er)
	if err != nil {
		t.Fatalf("could not create clusterSpec")
	}
	changedTest.clusterSpec = cs

	return changedTest
}

func TestClusterManagerClusterSpecChangedNoChanges(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.OIDCConfigKind, Name: tt.clusterName}}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNewEksdRelease(t *testing.T) {
	tt := newSpecChangedTest(t)

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	diffBundle := test.VersionBundle()
	diffBundle.EksD.Name = "different"
	tt.clusterSpec.VersionsBundles[v1alpha1.Kube119] = diffBundle

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNewEksdReleaseWorkers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	m := &clusterManagerMocks{
		writer:             mockswriter.NewMockFileWriter(mockCtrl),
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
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
	b.Spec.VersionsBundles[1].EksD.Name = "test2"
	r := test.EksdRelease("1-19")
	r2 := test.EksdRelease("1-22")
	r2.Name = "test2"
	ac := &v1alpha1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	gc := &v1alpha1.GitOpsConfig{}
	er := test.EKSARelease()

	fakeClient := test.NewFakeKubeClient(dc, oc, b, r, ac, gc, er, r2)
	cf := mocksmanager.NewMockClientFactory(mockCtrl)
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents)

	tt := newSpecChangedTest(t)
	tt.clusterManager = c
	tt.mocks = m

	kube122 := v1alpha1.KubernetesVersion("1.22")
	tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	diffBundle := test.VersionBundle()
	diffBundle.EksD.Name = "different"
	tt.clusterSpec.VersionsBundles[v1alpha1.Kube122] = diffBundle

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedClusterChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.newClusterConfig.Spec.KubernetesVersion = "1.20"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedEksDReleaseChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = "1.22"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedGitOpsDefault(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{Kind: v1alpha1.GitOpsConfigKind}
	tt.oldClusterConfig = tt.clusterSpec.Cluster.DeepCopy()
	tt.clusterSpec.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.OIDCConfigKind, Name: tt.clusterName}}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)

	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

func TestClusterManagerClusterSpecChangedAWSIamConfigChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.AWSIamConfigKind, Name: tt.clusterName}}
	tt.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}
	tt.oldClusterConfig = tt.clusterSpec.Cluster.DeepCopy()
	tt.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{Spec: v1alpha1.AWSIamConfigSpec{
		MapRoles: []v1alpha1.MapRoles{},
	}}

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)

	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec)

	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
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
	writer             *mockswriter.MockFileWriter
	networking         *mocksmanager.MockNetworking
	awsIamAuth         *mocksmanager.MockAwsIamAuth
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
		networking:         mocksmanager.NewMockNetworking(mockCtrl),
		awsIamAuth:         mocksmanager.NewMockAwsIamAuth(mockCtrl),
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
	cf.EXPECT().BuildClientFromKubeconfig("").Return(fakeClient, nil).AnyTimes()
	c := clustermanager.New(cf, m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, m.eksaComponents, opts...)

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

func TestClusterManagerDeletePackageResources(t *testing.T) {
	tt := newTest(t)

	tt.mocks.client.EXPECT().DeletePackageResources(tt.ctx, tt.cluster, tt.clusterName).Return(nil)

	err := tt.clusterManager.DeletePackageResources(tt.ctx, tt.cluster, tt.clusterName)
	tt.Expect(err).To(BeNil())
}

func TestCreateAwsIamAuthCaSecretSuccess(t *testing.T) {
	tt := newTest(t)

	tt.mocks.awsIamAuth.EXPECT().CreateAndInstallAWSIAMAuthCASecret(tt.ctx, tt.cluster, tt.clusterName).Return(nil)

	err := tt.clusterManager.CreateAwsIamAuthCaSecret(tt.ctx, tt.cluster, tt.clusterName)
	tt.Expect(err).To(BeNil())
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
		StringData: map[string]string{
			"username": "",
			"password": "",
		},
	}

	tt.mocks.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, secret).Return(nil)

	err := tt.clusterManager.CreateRegistryCredSecret(tt.ctx, tt.cluster)
	tt.Expect(err).To(BeNil())
}

func TestClusterManagerDeleteClusterSelfManagedCluster(t *testing.T) {
	tt := newTest(t)
	managementCluster := &types.Cluster{
		Name: "m-cluster",
	}

	tt.mocks.client.EXPECT().DeleteCluster(tt.ctx, managementCluster, tt.cluster)
	tt.mocks.provider.EXPECT().PostClusterDeleteValidate(tt.ctx, managementCluster)

	tt.Expect(
		tt.clusterManager.DeleteCluster(tt.ctx, managementCluster, tt.cluster, tt.mocks.provider, tt.clusterSpec),
	).To(Succeed())
}

func TestClusterManagerDeleteClusterManagedCluster(t *testing.T) {
	tt := newTest(t)
	managementCluster := &types.Cluster{
		Name: "m-cluster",
	}
	tt.clusterSpec.Cluster.SetManagedBy("m-cluster")
	tt.clusterSpec.GitOpsConfig = &v1alpha1.GitOpsConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config-git",
			Namespace: "my-ns",
		},
	}
	tt.clusterSpec.OIDCConfig = &v1alpha1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config-oidc",
			Namespace: "my-ns",
		},
	}
	tt.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config-aws",
			Namespace: "my-ns",
		},
	}

	gomock.InOrder(
		tt.expectPauseClusterReconciliation(),
		tt.mocks.client.EXPECT().DeleteEKSACluster(tt.ctx, managementCluster, tt.clusterSpec.Cluster.Name, tt.clusterSpec.Cluster.Namespace),
		tt.mocks.client.EXPECT().DeleteGitOpsConfig(tt.ctx, managementCluster, "my-config-git", "my-ns"),
		tt.mocks.client.EXPECT().DeleteOIDCConfig(tt.ctx, managementCluster, "my-config-oidc", "my-ns"),
		tt.mocks.client.EXPECT().DeleteAWSIamConfig(tt.ctx, managementCluster, "my-config-aws", "my-ns"),
		tt.mocks.provider.EXPECT().DeleteResources(tt.ctx, tt.clusterSpec),
		tt.mocks.client.EXPECT().DeleteCluster(tt.ctx, managementCluster, tt.cluster),
		tt.mocks.provider.EXPECT().PostClusterDeleteValidate(tt.ctx, managementCluster),
	)

	tt.Expect(
		tt.clusterManager.DeleteCluster(tt.ctx, managementCluster, tt.cluster, tt.mocks.provider, tt.clusterSpec),
	).To(Succeed())
}

func TestClusterManagerRemoveManagedByCLIAnnotationSuccess(t *testing.T) {
	clusterName := "cluster-name"
	cluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))

	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any(), cluster, gomock.Any()).Return(nil)

	tt.Expect(tt.clusterManager.RemoveManagedByCLIAnnotationForCluster(tt.ctx, cluster, tt.clusterSpec, tt.mocks.provider)).To(BeNil())
}

func TestClusterManagerRemoveManagedByCLIAnnotationError(t *testing.T) {
	clusterName := "cluster-name"
	cluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(1, 0)))

	tt.mocks.client.EXPECT().RemoveAnnotationInNamespace(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any(), cluster, gomock.Any()).Return(errors.New("removing annotation"))

	tt.Expect(tt.clusterManager.RemoveManagedByCLIAnnotationForCluster(tt.ctx, cluster, tt.clusterSpec, tt.mocks.provider)).To(MatchError(ContainSubstring("removing annotation")))
}
