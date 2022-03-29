package clustermanager_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	mocksdiagnostics "github.com/aws/eks-anywhere/pkg/diagnostics/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/features"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

var (
	eksaClusterResourceType           = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType    = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

func TestClusterManagerInstallNetworkingSuccess(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}

	networkingManifest := []byte("cilium")
	clusterSpec := test.NewClusterSpec()

	c, m := newClusterManager(t)
	m.provider.EXPECT().GetDeployments()
	m.networking.EXPECT().GenerateManifest(ctx, clusterSpec, []string{}).Return(networkingManifest, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, networkingManifest)

	if err := c.InstallNetworking(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallNetworking() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallNetworkingNetworkingError(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterSpec := test.NewClusterSpec()

	c, m := newClusterManager(t)
	m.provider.EXPECT().GetDeployments()
	m.networking.EXPECT().GenerateManifest(ctx, clusterSpec, []string{}).Return(nil, errors.New("error in networking"))

	if err := c.InstallNetworking(ctx, cluster, clusterSpec, m.provider); err == nil {
		t.Errorf("ClusterManager.InstallNetworking() error = nil, wantErr not nil")
	}
}

func TestClusterManagerInstallNetworkingClientError(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	networkingManifest := []byte("cilium")
	clusterSpec := test.NewClusterSpec()
	retries := 2

	c, m := newClusterManager(t)
	m.provider.EXPECT().GetDeployments()
	m.networking.EXPECT().GenerateManifest(ctx, clusterSpec, []string{}).Return(networkingManifest, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, networkingManifest).Return(errors.New("error from client")).Times(retries)

	c.Retrier = retrier.NewWithMaxRetries(retries, 1*time.Microsecond)
	if err := c.InstallNetworking(ctx, cluster, clusterSpec, m.provider); err == nil {
		t.Errorf("ClusterManager.InstallNetworking() error = nil, wantErr not nil")
	}
}

func TestClusterManagerInstallStorageClassSuccess(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}

	storageClassManifest := []byte("yaml: values")

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateStorageClass().Return(storageClassManifest)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, storageClassManifest)

	if err := c.InstallStorageClass(ctx, cluster, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallStorageClass() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallStorageClassProviderNothing(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateStorageClass().Return(nil)

	if err := c.InstallStorageClass(ctx, cluster, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallStorageClass() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallStorageClassClientError(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	storageClassManifest := []byte("yaml: values")
	retries := 2

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateStorageClass().Return(storageClassManifest)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, storageClassManifest).Return(
		errors.New("error from client")).Times(retries)

	c.Retrier = retrier.NewWithMaxRetries(retries, 1*time.Microsecond)
	if err := c.InstallStorageClass(ctx, cluster, m.provider); err == nil {
		t.Errorf("ClusterManager.InstallStorageClass() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCAPIWaitForDeploymentStackedEtcd(t *testing.T) {
	ctx := context.Background()
	clusterObj := &types.Cluster{}
	c, m := newClusterManager(t)
	clusterSpecStackedEtcd := test.NewClusterSpec()

	m.client.EXPECT().InitInfrastructure(ctx, clusterSpecStackedEtcd, clusterObj, m.provider)
	for namespace, deployments := range internal.CAPIDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m", "Available", deployment, namespace)
		}
	}
	providerDeployments := map[string][]string{}
	m.provider.EXPECT().GetDeployments().Return(providerDeployments)
	for namespace, deployments := range providerDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m", "Available", deployment, namespace)
		}
	}
	if err := c.InstallCAPI(ctx, clusterSpecStackedEtcd, clusterObj, m.provider); err != nil {
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
	m.client.EXPECT().InitInfrastructure(ctx, clusterSpecExternalEtcd, clusterObj, m.provider)
	for namespace, deployments := range internal.CAPIDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m", "Available", deployment, namespace)
		}
	}
	for namespace, deployments := range internal.ExternalEtcdDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m", "Available", deployment, namespace)
		}
	}
	providerDeployments := map[string][]string{}
	m.provider.EXPECT().GetDeployments().Return(providerDeployments)
	for namespace, deployments := range providerDeployments {
		for _, deployment := range deployments {
			m.client.EXPECT().WaitForDeployment(ctx, clusterObj, "30m", "Available", deployment, namespace)
		}
	}
	if err := c.InstallCAPI(ctx, clusterSpecExternalEtcd, clusterObj, m.provider); err != nil {
		t.Errorf("ClusterManager.InstallCAPI() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerSaveLogsSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
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
	m.diagnosticsFactory.EXPECT().DiagnosticBundleManagementCluster(bootstrapCluster.KubeconfigFile).Return(b, nil)
	b.EXPECT().CollectAndAnalyze(ctx, gomock.AssignableToTypeOf(&time.Time{}))

	m.diagnosticsFactory.EXPECT().DiagnosticBundleFromSpec(clusterSpec, m.provider, workloadCluster.KubeconfigFile).Return(b, nil)
	b.EXPECT().CollectAndAnalyze(ctx, gomock.AssignableToTypeOf(&time.Time{}))

	if err := c.SaveLogsManagementCluster(ctx, bootstrapCluster); err != nil {
		t.Errorf("ClusterManager.SaveLogsManagementCluster() error = %v, wantErr nil", err)
	}

	if err := c.SaveLogsWorkloadCluster(ctx, m.provider, clusterSpec, workloadCluster); err != nil {
		t.Errorf("ClusterManager.SaveLogsWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, cluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterWithExternalEtcdSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 2
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.client.EXPECT().WaitForManagedExternalEtcdReady(ctx, cluster, "60m", clusterName)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, cluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterSuccessWithExtraObjects(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
		s.VersionsBundle.KubeVersion = "1.20"
		s.VersionsBundle.KubeDistro.CoreDNS.Tag = "v1.8.3-eks-1-20-1"
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	wantKubeconfigFile := "folder/cluster-name-eks-a-cluster.kubeconfig"
	wantCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: wantKubeconfigFile,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, wantCluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, wantCluster, gomock.Any())

	got, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider)
	if err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}

	if !reflect.DeepEqual(got, wantCluster) {
		t.Errorf("ClusterManager.CreateWorkloadCluster() cluster = %#v, want %#v", got, wantCluster)
	}
}

func TestClusterManagerCreateWorkloadClusterErrorApplyingExtraObjects(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
		s.VersionsBundle.KubeVersion = "1.20"
		s.VersionsBundle.KubeDistro.CoreDNS.Tag = "v1.8.3-eks-1-20-1"
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	wantKubeconfigFile := "folder/cluster-name-eks-a-cluster.kubeconfig"
	wantCluster := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: wantKubeconfigFile,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, wantCluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, wantCluster, gomock.Any()).Return(errors.New("error applying"))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err == nil {
		t.Error("ClusterManager.CreateWorkloadCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateWorkloadClusterWaitForMachinesTimeout(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, cluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail once
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
	}}}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err == nil {
		t.Error("ClusterManager.CreateWorkloadCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateWorkloadClusterWaitForMachinesSuccessAfterRetries(t *testing.T) {
	retries := 10
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 1*time.Minute, 2*time.Minute))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().KubeconfigSecretAvailable(ctx, "", clusterName, constants.EksaSystemNamespace).Return(true, nil)
	m.provider.EXPECT().RunPostControlPlaneCreation(ctx, clusterSpec, cluster)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail a bunch of times
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Times(retries-5).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef  times
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Times(3).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
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
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
	}
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Times(1).Return(machines, nil)
	// Finally return two machines with node ref
	machines = []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
	}
	m.client.EXPECT().GetMachines(ctx, cluster, cluster.Name).Times(1).Return(machines, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterSuccess(t *testing.T) {
	clusterName := "cluster-name"
	mCluster := &types.Cluster{
		Name: clusterName,
	}
	wCluster := &types.Cluster{
		Name: clusterName,
	}

	tt := newSpecChangedTest(t)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "60m", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.provider.EXPECT().MachineDeploymentsToDelete(wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy()).Return([]string{})
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().ValidateWorkerNodes(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(nil)
	tt.mocks.provider.EXPECT().GetDeployments()
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(nil, nil)

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

	tt := newSpecChangedTest(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(nil, nil)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", clusterName)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	// Fail once
	tt.mocks.client.EXPECT().GetMachines(ctx, mCluster, mCluster.Name).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	tt.mocks.client.EXPECT().GetMachines(ctx, mCluster, mCluster.Name).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
	}}}, nil)

	if err := tt.clusterManager.UpgradeCluster(ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateWorkloadClusterWaitForMachinesFailedWithUnhealthyNode(t *testing.T) {
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
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
	}

	tt := newSpecChangedTest(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "60m", clusterName).MaxTimes(5)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	// Return a machine with no nodeRef the rest of the retries
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).MinTimes(1).Return(machines, nil)
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(nil, nil)

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
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.provider.EXPECT().GenerateCAPISpecForUpgrade(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy())
	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	tt.mocks.provider.EXPECT().RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, wCluster, mCluster)
	tt.mocks.client.EXPECT().WaitForControlPlaneReady(tt.ctx, mCluster, "60m", clusterName).MaxTimes(2)
	tt.mocks.client.EXPECT().WaitForControlPlaneNotReady(tt.ctx, mCluster, "1m", clusterName)
	tt.mocks.client.EXPECT().GetMachines(tt.ctx, mCluster, mCluster.Name).Return([]types.Machine{}, nil).Times(2)
	tt.mocks.provider.EXPECT().MachineDeploymentsToDelete(wCluster, tt.clusterSpec, tt.clusterSpec.DeepCopy()).Return([]string{})
	tt.mocks.client.EXPECT().WaitForDeployment(tt.ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).Return(errors.New("time out"))
	tt.mocks.client.EXPECT().ValidateControlPlaneNodes(tt.ctx, mCluster, wCluster.Name).Return(nil)
	tt.mocks.client.EXPECT().ValidateWorkerNodes(tt.ctx, wCluster.Name, mCluster.KubeconfigFile).Return(nil)
	tt.mocks.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(nil, nil)

	if err := tt.clusterManager.UpgradeCluster(tt.ctx, mCluster, wCluster, tt.clusterSpec, tt.mocks.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
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
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from, to.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	m.client.EXPECT().GetClusters(ctx, to).Return(clusters, nil)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", capiClusterName)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
	m.client.EXPECT().ValidateWorkerNodes(ctx, to.Name, to.KubeconfigFile)
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
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to).Return(errors.New("error moving"))

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCAPIErrorGetClusters(t *testing.T) {
	from := &types.Cluster{
		Name: "from-cluster",
	}
	to := &types.Cluster{
		Name: "to-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = to.Name
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	m.client.EXPECT().GetClusters(ctx, to).Return(nil, errors.New("error getting clusters"))

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
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().GetClusters(ctx, to).Return(clusters, nil)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "15m0s", capiClusterName).Return(errors.New("error waiting for control plane"))

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
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-wn"}}}
	})
	ctx := context.Background()

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(0, 10*time.Microsecond, 20*time.Microsecond))
	m.client.EXPECT().GetMachines(ctx, from, from.Name)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	m.client.EXPECT().GetClusters(ctx, to)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, to, to.Name)
	m.client.EXPECT().ValidateWorkerNodes(ctx, to.Name, to.KubeconfigFile)
	m.client.EXPECT().GetMachines(ctx, to, from.Name).Return(nil, errors.New("error getting machines")).AnyTimes()

	if err := c.MoveCAPI(ctx, from, to, from.Name, clusterSpec); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateEKSAResourcesSuccess(t *testing.T) {
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	os.Unsetenv(features.CloudStackProviderEnvVar)
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Components = "testdata/eksa_components.yaml"
	tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl = "testdata/eksa_components.yaml"

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	c, m := newClusterManager(t)

	m.client.EXPECT().ApplyKubeSpecFromBytesForce(ctx, tt.cluster, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, tt.cluster, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, tt.cluster, gomock.Any(), gomock.Any()).MaxTimes(2)
	if err := c.CreateEKSAResources(ctx, tt.cluster, tt.clusterSpec, datacenterConfig, machineConfigs); err != nil {
		t.Errorf("ClusterManager.CreateEKSAResources() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerPauseEKSAControllerReconcileSuccessWithoutMachineConfig(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"

	clusterObj := &types.Cluster{
		Name: clusterName,
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster-test",
			},
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "data-center-name",
				},
			},
		}
	})

	expectedPauseAnnotation := map[string]string{"anywhere.eks.amazonaws.com/paused": "true"}

	cm, m := newClusterManager(t)
	m.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	m.provider.EXPECT().MachineResourceType().Return("")
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaClusterResourceType, clusterSpec.Cluster.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)

	if err := cm.PauseEKSAControllerReconcile(ctx, clusterObj, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.PauseEKSAControllerReconcile() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerPauseEKSAControllerReconcileSuccessWithMachineConfig(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"

	clusterObj := &types.Cluster{
		Name: clusterName,
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "datasourcename",
				},
				ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
					MachineGroupRef: &v1alpha1.Ref{
						Name: clusterName + "-cp",
					},
				},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
					MachineGroupRef: &v1alpha1.Ref{
						Name: clusterName,
					},
				}},
			},
		}
	})

	expectedPauseAnnotation := map[string]string{"anywhere.eks.amazonaws.com/paused": "true"}

	cm, m := newClusterManager(t)
	m.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	m.provider.EXPECT().MachineResourceType().Return(eksaVSphereMachineResourceType).Times(3)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Cluster.Spec.DatacenterRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereMachineResourceType, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereMachineResourceType, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaClusterResourceType, clusterSpec.Cluster.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)

	if err := cm.PauseEKSAControllerReconcile(ctx, clusterObj, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.PauseEKSAControllerReconcile() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerResumeEKSAControllerReconcileSuccessWithoutMachineConfig(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"

	clusterObj := &types.Cluster{
		Name: clusterName,
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "data-center-name",
				},
			},
		}
	})
	clusterSpec.Cluster.PauseReconcile()

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
		},
	}
	pauseAnnotation := "anywhere.eks.amazonaws.com/paused"

	cm, m := newClusterManager(t)
	m.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	m.provider.EXPECT().MachineResourceType().Return("")
	m.provider.EXPECT().DatacenterConfig(clusterSpec).Return(datacenterConfig)
	m.client.EXPECT().RemoveAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Cluster.Spec.DatacenterRef.Name, pauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().RemoveAnnotationInNamespace(ctx, eksaClusterResourceType, clusterSpec.Cluster.Name, pauseAnnotation, clusterObj, "").Return(nil)

	if err := cm.ResumeEKSAControllerReconcile(ctx, clusterObj, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.ResumeEKSAControllerReconcile() error = %v, wantErr nil", err)
	}
	annotations := clusterSpec.Cluster.GetAnnotations()
	if _, ok := annotations[pauseAnnotation]; ok {
		t.Errorf("%s annotation exists, should be removed", pauseAnnotation)
	}
}

func TestClusterManagerInstallCustomComponentsSuccess(t *testing.T) {
	ctx := context.Background()
	tt := newTest(t)
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "testdata/testClusterSpec.yaml"

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Not(gomock.Nil())).Return(nil)

	for namespace, deployments := range internal.EksaDeployments {
		for _, deployment := range deployments {
			tt.mocks.client.EXPECT().WaitForDeployment(ctx, tt.cluster, "30m", "Available", deployment, namespace)
		}
	}
	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("ClusterManager.InstallCustomComponents() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallCustomComponentsErrorReadingManifest(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "fake.yaml"

	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("ClusterManager.InstallCustomComponents() error = nil, wantErr not nil")
	}
}

func TestClusterManagerInstallCustomComponentsErrorApplying(t *testing.T) {
	tt := newTest(t, clustermanager.WithRetrier(retrier.NewWithMaxRetries(2, 0)))
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "testdata/testClusterSpec.yaml"

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Not(gomock.Nil())).Return(errors.New("error from apply")).Times(2)

	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
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
				Count: 1,
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
		},
	}
	newClusterConfig := clusterConfig.DeepCopy()
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
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
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.OIDCConfigSpec{
			ClientId: "test",
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

	var err error

	changedTest.clusterSpec, err = cluster.BuildSpecFromBundles(newClusterConfig, test.Bundles(t))
	if err != nil {
		t.Fatalf("Failed setting up cluster spec for ClusterChanged test: %v", err)
	}

	return changedTest
}

func TestClusterManagerClusterSpecChangedNoChanges(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.OIDCConfigKind, Name: tt.clusterName}}
	tt.clusterSpec.OIDCConfig = tt.oldOIDCConfig.DeepCopy()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.oldClusterConfig.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

func TestClusterManagerClusterSpecChangedClusterChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.newClusterConfig.Spec.KubernetesVersion = "1.20"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedEksDReleaseChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Name = "kubernetes-1-19-eks-5"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesDatacenterSpecChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.newDatacenterConfig.Spec.Insecure = false
	tt.clusterSpec.OIDCConfig = tt.oldOIDCConfig.DeepCopy()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesControlPlaneMachineConfigSpecChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.newControlPlaneMachineConfig.Spec.NumCPUs = 4
	tt.clusterSpec.OIDCConfig = tt.oldOIDCConfig.DeepCopy()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})

	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesWorkerNodeMachineConfigSpecChanged(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.newWorkerMachineConfig.Spec.NumCPUs = 4
	tt.clusterSpec.OIDCConfig = tt.oldOIDCConfig.DeepCopy()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})

	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedGitOpsDefault(t *testing.T) {
	tt := newSpecChangedTest(t)
	tt.clusterSpec.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{Kind: v1alpha1.GitOpsConfigKind}
	tt.oldClusterConfig = tt.clusterSpec.Cluster.DeepCopy()
	oldGitOpsConfig := tt.clusterSpec.GitOpsConfig.DeepCopy()
	tt.clusterSpec.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{{Kind: v1alpha1.OIDCConfigKind, Name: tt.clusterName}}
	tt.clusterSpec.OIDCConfig = tt.oldOIDCConfig.DeepCopy()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterSpec.Cluster.Name).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaGitOpsConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.GitOpsRef.Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(oldGitOpsConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksdRelease(tt.ctx, gomock.Any(), constants.EksaSystemNamespace, gomock.Any())
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaOIDCConfig(tt.ctx, tt.clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Namespace).Return(tt.oldOIDCConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})

	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

type testSetup struct {
	*WithT
	clusterManager *clustermanager.ClusterManager
	mocks          *clusterManagerMocks
	ctx            context.Context
	clusterSpec    *cluster.Spec
	cluster        *types.Cluster
	clusterName    string
}

func newTest(t *testing.T, opts ...clustermanager.ClusterManagerOpt) *testSetup {
	c, m := newClusterManager(t, opts...)
	clusterName := "cluster-name"
	return &testSetup{
		WithT:          NewWithT(t),
		clusterManager: c,
		mocks:          m,
		ctx:            context.Background(),
		clusterSpec:    test.NewClusterSpec(),
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
	}

	c := clustermanager.New(m.client, m.networking, m.writer, m.diagnosticsFactory, m.awsIamAuth, opts...)

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

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster, tt.clusterName).Return(tt.clusterSpec.Cluster, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Name, "").Return(nil, errors.New("error from client"))

	_, err := tt.clusterManager.GetCurrentClusterSpec(tt.ctx, tt.cluster, tt.clusterName)
	tt.Expect(err).ToNot(BeNil())
}
