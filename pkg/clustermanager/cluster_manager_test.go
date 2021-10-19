package clustermanager_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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
	m.networking.EXPECT().GenerateManifest(clusterSpec).Return(networkingManifest, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, networkingManifest)

	if err := c.InstallNetworking(ctx, cluster, clusterSpec); err != nil {
		t.Errorf("ClusterManager.InstallNetworking() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerInstallNetworkingNetworkingError(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{}
	clusterSpec := test.NewClusterSpec()

	c, m := newClusterManager(t)
	m.networking.EXPECT().GenerateManifest(clusterSpec).Return(nil, errors.New("error in networking"))

	if err := c.InstallNetworking(ctx, cluster, clusterSpec); err == nil {
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
	m.networking.EXPECT().GenerateManifest(clusterSpec).Return(networkingManifest, nil)
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, networkingManifest).Return(errors.New("error from client")).Times(retries)

	c.Retrier = retrier.NewWithMaxRetries(retries, 1*time.Microsecond)
	if err := c.InstallNetworking(ctx, cluster, clusterSpec); err == nil {
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
		s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 1}
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
	cluster := &types.Cluster{Name: "cluster-name"}

	c, m := newClusterManager(t)
	m.writer.EXPECT().WithDir(gomock.Any()).Return(m.writer, nil)
	for file, logCmd := range internal.ClusterDeployments {
		m.client.EXPECT().SaveLog(ctx, cluster, logCmd, file, m.writer).Times(1).Return(nil)
	}

	if err := c.SaveLogs(ctx, cluster); err != nil {
		t.Errorf("ClusterManager.SaveLogs() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
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
		s.Name = clusterName
		s.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Spec.ControlPlaneConfiguration.Count = 2
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForManagedExternalEtcdReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
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
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
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
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
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
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
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
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
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
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail once
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, cluster).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
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
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	cluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 1*time.Minute, 2*time.Minute))
	m.provider.EXPECT().GenerateCAPISpecForCreate(ctx, cluster, clusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail a bunch of times
	m.client.EXPECT().GetMachines(ctx, cluster).Times(retries-5).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef  times
	m.client.EXPECT().GetMachines(ctx, cluster).Times(3).Return([]types.Machine{{Metadata: types.MachineMetadata{
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
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(machines, nil)
	// Finally return two machines with node ref
	machines = []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: status},
	}
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(machines, nil)
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
	ctx := context.Background()
	clusterName := "test-cluster"
	mCluster := &types.Cluster{
		Name: clusterName,
	}

	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, wClusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", clusterName).MaxTimes(2)
	m.client.EXPECT().GetMachines(ctx, mCluster).Return([]types.Machine{}, nil).Times(2)
	m.client.EXPECT().WaitForDeployment(ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, mCluster).Return(nil)
	m.client.EXPECT().ValidateWorkerNodes(ctx, mCluster).Return(nil)
	m.provider.EXPECT().GetDeployments()
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMachinesTimeout(t *testing.T) {
	ctx := context.Background()
	clusterName := "test-cluster"
	mCluster := &types.Cluster{
		Name: clusterName,
	}

	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	m.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, wClusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", clusterName)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	// Fail once
	m.client.EXPECT().GetMachines(ctx, mCluster).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, mCluster).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
	}}}, nil)

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateWorkloadClusterWaitForMachinesFailedWithUnhealthyNode(t *testing.T) {
	ctx := context.Background()
	clusterName := "test-cluster"
	mCluster := &types.Cluster{
		Name: clusterName,
	}

	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

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

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	m.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, wClusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", clusterName).MaxTimes(5)
	m.client.EXPECT().WaitForDeployment(ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, mCluster).MinTimes(1).Return(machines, nil)

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForCAPITimeout(t *testing.T) {
	ctx := context.Background()
	clusterName := "test-cluster"
	mCluster := &types.Cluster{
		Name: clusterName,
	}

	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: clusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateCAPISpecForUpgrade(ctx, mCluster, wCluster, wClusterSpec)
	m.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(ctx, mCluster, test.OfType("[]uint8"), constants.EksaSystemNamespace).Times(2)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", clusterName).MaxTimes(2)
	m.client.EXPECT().GetMachines(ctx, mCluster).Return([]types.Machine{}, nil).Times(2)
	m.client.EXPECT().WaitForDeployment(ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).Return(errors.New("time out"))
	m.client.EXPECT().ValidateControlPlaneNodes(ctx, mCluster).Return(nil)
	m.client.EXPECT().ValidateWorkerNodes(ctx, mCluster).Return(nil)
	m.writer.EXPECT().Write(clusterName+"-eks-a-cluster.yaml", gomock.Any(), gomock.Not(gomock.Nil()))

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err == nil {
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
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	m.client.EXPECT().GetClusters(ctx, to).Return(clusters, nil)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "5m0s", capiClusterName)
	m.client.EXPECT().GetMachines(ctx, to)

	if err := c.MoveCAPI(ctx, from, to); err != nil {
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
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from)
	m.client.EXPECT().MoveManagement(ctx, from, to).Return(errors.New("error moving"))

	if err := c.MoveCAPI(ctx, from, to); err == nil {
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
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().GetMachines(ctx, from)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	m.client.EXPECT().GetClusters(ctx, to).Return(nil, errors.New("error getting clusters"))

	if err := c.MoveCAPI(ctx, from, to); err == nil {
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
	ctx := context.Background()

	c, m := newClusterManager(t)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	capiClusterName := "capi-cluster"
	clusters := []types.CAPICluster{{Metadata: types.Metadata{Name: capiClusterName}}}
	m.client.EXPECT().GetMachines(ctx, from)
	m.client.EXPECT().GetClusters(ctx, to).Return(clusters, nil)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, to, "5m0s", capiClusterName).Return(errors.New("error waiting for control plane"))

	if err := c.MoveCAPI(ctx, from, to); err == nil {
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
	ctx := context.Background()

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(0, 10*time.Microsecond, 20*time.Microsecond))
	m.client.EXPECT().GetMachines(ctx, from)
	m.client.EXPECT().MoveManagement(ctx, from, to)
	m.client.EXPECT().GetClusters(ctx, to)
	m.client.EXPECT().GetMachines(ctx, to).Return(nil, errors.New("error getting machines")).AnyTimes()

	if err := c.MoveCAPI(ctx, from, to); err == nil {
		t.Error("ClusterManager.MoveCAPI() error = nil, wantErr not nil")
	}
}

func TestClusterManagerCreateEKSAResourcesSuccess(t *testing.T) {
	clusterSpec := &cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				KubernetesVersion:             "1.19",
				ControlPlaneConfiguration:     v1alpha1.ControlPlaneConfiguration{Count: 1},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{Count: 1}},
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
				},
			},
		},
		Bundles: &anywherev1alpha1.Bundles{},
	}

	ctx := context.Background()
	clusterName := "cluster-name"
	cluster := &types.Cluster{
		Name: clusterName,
	}

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{}

	c, m := newClusterManager(t)

	m.client.EXPECT().ApplyKubeSpecFromBytesForce(ctx, cluster, gomock.Any())
	m.client.EXPECT().ApplyKubeSpecFromBytes(ctx, cluster, gomock.Any())
	if err := c.CreateEKSAResources(ctx, cluster, clusterSpec, datacenterConfig, machineConfigs); err != nil {
		t.Errorf("ClusterManager.CreateEKSAResources() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerPauseEKSAControllerReconcileSuccessWithoutMachineConfig(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"

	clusterObj := &types.Cluster{
		Name: clusterName,
	}

	clusterSpec := &cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "data-center-name",
				},
			},
		},
	}

	expectedPauseAnnotation := map[string]string{"anywhere.eks.amazonaws.com/paused": "true"}

	cm, m := newClusterManager(t)
	m.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	m.provider.EXPECT().MachineResourceType().Return("")
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Spec.DatacenterRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaClusterResourceType, clusterObj.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)

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

	clusterSpec := &cluster.Spec{
		Cluster: &v1alpha1.Cluster{
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
		},
	}

	expectedPauseAnnotation := map[string]string{"anywhere.eks.amazonaws.com/paused": "true"}

	cm, m := newClusterManager(t)
	m.provider.EXPECT().DatacenterResourceType().Return(eksaVSphereDatacenterResourceType)
	m.provider.EXPECT().MachineResourceType().Return(eksaVSphereMachineResourceType).Times(3)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Spec.DatacenterRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereMachineResourceType, clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaVSphereMachineResourceType, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().UpdateAnnotationInNamespace(ctx, eksaClusterResourceType, clusterObj.Name, expectedPauseAnnotation, clusterObj, "").Return(nil)

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

	clusterSpec := &cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "data-center-name",
				},
			},
		},
	}
	clusterSpec.PauseReconcile()

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
	m.provider.EXPECT().DatacenterConfig().Return(datacenterConfig)
	m.client.EXPECT().RemoveAnnotationInNamespace(ctx, eksaVSphereDatacenterResourceType, clusterSpec.Spec.DatacenterRef.Name, pauseAnnotation, clusterObj, "").Return(nil)
	m.client.EXPECT().RemoveAnnotationInNamespace(ctx, eksaClusterResourceType, clusterObj.Name, pauseAnnotation, clusterObj, "").Return(nil)

	if err := cm.ResumeEKSAControllerReconcile(ctx, clusterObj, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.ResumeEKSAControllerReconcile() error = %v, wantErr nil", err)
	}
	annotations := clusterSpec.GetAnnotations()
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
	tt := newTest(t)
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "testdata/testClusterSpec.yaml"
	tt.clusterManager.Retrier = retrier.NewWithMaxRetries(2, 0)

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Not(gomock.Nil())).Return(errors.New("error from apply")).Times(2)

	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("ClusterManager.InstallCustomComponents() error = nil, wantErr not nil")
	}
}

type clusterChangedTest struct {
	*testSetup
	oldClusterConfig, newClusterConfig                         *v1alpha1.Cluster
	oldDatacenterConfig, newDatacenterConfig                   *v1alpha1.VSphereDatacenterConfig
	oldControlPlaneMachineConfig, newControlPlaneMachineConfig *v1alpha1.VSphereMachineConfig
	oldWorkerMachineConfig, newWorkerMachineConfig             *v1alpha1.VSphereMachineConfig
}

func newClusterChangedTest(t *testing.T) *clusterChangedTest {
	testSetup := newTest(t)
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

	changedTest := &clusterChangedTest{
		testSetup:                    testSetup,
		oldClusterConfig:             clusterConfig,
		newClusterConfig:             newClusterConfig,
		oldDatacenterConfig:          datacenterConfig,
		newDatacenterConfig:          datacenterConfig.DeepCopy(),
		oldControlPlaneMachineConfig: machineConfig,
		newControlPlaneMachineConfig: machineConfig.DeepCopy(),
		oldWorkerMachineConfig:       workerMachineConfig,
		newWorkerMachineConfig:       workerMachineConfig.DeepCopy(),
	}

	var err error
	changedTest.clusterSpec, err = cluster.BuildSpecFromBundles(newClusterConfig, test.Bundles(t))
	if err != nil {
		t.Fatalf("Failed setting up cluster spec for ClusterChanged test: %v", err)
	}

	return changedTest
}

func TestClusterManagerClusterSpecChangedNoChanges(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

func TestClusterManagerClusterSpecChangedClusterChanged(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.newClusterConfig.Spec.KubernetesVersion = "1.20"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedEksDReleaseChanged(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Name = "kubernetes-1-19-eks-5"

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesDatacenterSpecChanged(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.newDatacenterConfig.Spec.Insecure = false

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesControlPlaneMachineConfigSpecChanged(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.newControlPlaneMachineConfig.Spec.NumCPUs = 4

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesWorkerNodeMachineConfigSpecChanged(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.newWorkerMachineConfig.Spec.NumCPUs = 4

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedGitOpsDefault(t *testing.T) {
	tt := newClusterChangedTest(t)
	tt.clusterSpec.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{Kind: v1alpha1.GitOpsConfigKind}
	tt.oldClusterConfig = tt.clusterSpec.Cluster.DeepCopy()
	tt.clusterSpec.SetDefaultGitOps()

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.oldClusterConfig, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.cluster.Name, "").Return(test.Bundles(t), nil)
	tt.mocks.client.EXPECT().GetEksaVSphereDatacenterConfig(tt.ctx, tt.oldClusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldDatacenterConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldControlPlaneMachineConfig, nil)
	tt.mocks.client.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, tt.oldClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(tt.oldWorkerMachineConfig, nil)
	diff, err := tt.clusterManager.EKSAClusterSpecChanged(tt.ctx, tt.cluster, tt.clusterSpec, tt.newDatacenterConfig, []providers.MachineConfig{tt.newControlPlaneMachineConfig, tt.newWorkerMachineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

type testSetup struct {
	*WithT
	clusterManager *clustermanager.ClusterManager
	mocks          *mocks
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

type mocks struct {
	writer     *mockswriter.MockFileWriter
	networking *mocksmanager.MockNetworking
	client     *mocksmanager.MockClusterClient
	provider   *mocksprovider.MockProvider
}

func newClusterManager(t *testing.T, opts ...clustermanager.ClusterManagerOpt) (*clustermanager.ClusterManager, *mocks) {
	mockCtrl := gomock.NewController(t)
	m := &mocks{
		writer:     mockswriter.NewMockFileWriter(mockCtrl),
		networking: mocksmanager.NewMockNetworking(mockCtrl),
		client:     mocksmanager.NewMockClusterClient(mockCtrl),
		provider:   mocksprovider.NewMockProvider(mockCtrl),
	}

	c := clustermanager.New(m.client, m.networking, m.writer, opts...)

	return c, m
}

func TestClusterManagerGetCurrentClusterSpecGetClusterError(t *testing.T) {
	tt := newTest(t)

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(nil, errors.New("error from client"))

	_, err := tt.clusterManager.GetCurrentClusterSpec(tt.ctx, tt.cluster)
	tt.Expect(err).ToNot(BeNil())
}

func TestClusterManagerGetCurrentClusterSpecGetBundlesError(t *testing.T) {
	tt := newTest(t)

	tt.mocks.client.EXPECT().GetEksaCluster(tt.ctx, tt.cluster).Return(tt.clusterSpec.Cluster, nil)
	tt.mocks.client.EXPECT().GetBundles(tt.ctx, tt.cluster.KubeconfigFile, tt.clusterSpec.Cluster.Name, "").Return(nil, errors.New("error from client"))

	_, err := tt.clusterManager.GetCurrentClusterSpec(tt.ctx, tt.cluster)
	tt.Expect(err).ToNot(BeNil())
}
