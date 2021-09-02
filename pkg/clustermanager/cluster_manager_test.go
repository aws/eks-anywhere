package clustermanager_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var clusterDeployments = map[string]*types.Deployment{
	"kubeadm-bootstrap-controller-manager.log":         {Name: "capi-kubeadm-bootstrap-controller-manager", Namespace: "capi-kubeadm-bootstrap-system", Container: "manager"},
	"kubeadm-control-plane-controller-manager.log":     {Name: "capi-kubeadm-control-plane-controller-manager", Namespace: "capi-kubeadm-control-plane-system", Container: "manager"},
	"capi-controller-manager.log":                      {Name: "capi-controller-manager", Namespace: "capi-system", Container: "manager"},
	"wh-capi-controller-manager.log":                   {Name: "capi-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"wh-capi-kubeadm-bootstrap-controller-manager.log": {Name: "capi-kubeadm-bootstrap-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"wh-kubeadm-control-plane-controller-manager.log":  {Name: "capi-kubeadm-control-plane-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"cert-manager.log":                                 {Name: "cert-manager", Namespace: "cert-manager"},
	"cert-manager-cainjector.log":                      {Name: "cert-manager-cainjector", Namespace: "cert-manager"},
	"cert-manager-webhook.log":                         {Name: "cert-manager-webhook", Namespace: "cert-manager"},
	"coredns.log":                                      {Name: "coredns", Namespace: "kube-system"},
	"local-path-provisioner.log":                       {Name: "local-path-provisioner", Namespace: "local-path-storage"},
	"capv-controller-manager.log":                      {Name: "capv-controller-manager", Namespace: "capv-system", Container: "manager"},
	"wh-capv-controller-manager.log":                   {Name: "capv-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
}

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

func TestClusterManagerSaveLogsSuccess(t *testing.T) {
	ctx := context.Background()
	cluster := &types.Cluster{Name: "cluster-name"}

	c, m := newClusterManager(t)
	m.writer.EXPECT().WithDir(gomock.Any()).Return(m.writer, nil)
	for file, logCmd := range clusterDeployments {
		m.client.EXPECT().SaveLog(ctx, cluster, logCmd, file, m.writer).Times(1).Return(nil)
	}

	if err := c.SaveLogs(ctx, cluster); err != nil {
		t.Errorf("ClusterManager.SaveLogs() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerGenerateDeploymentFileSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
	})

	workloadCluster := &types.Cluster{Name: clusterName}
	bootstrapCluster := &types.Cluster{Name: "eks-a-bootstrap"}
	c, m := newClusterManager(t)
	fileName := fmt.Sprintf("%s-eks-a-cluster.yaml", clusterSpec.ObjectMeta.Name)
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, workloadCluster, clusterSpec, fileName).Times(1).Return("", nil)

	if _, err := c.GenerateDeploymentFile(ctx, bootstrapCluster, workloadCluster, clusterSpec, m.provider, false); err != nil {
		t.Errorf("ClusterManager.GenerateDeploymentFile() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerGenerateDeploymentFileOverrideSuccess(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = clusterName
		s.Spec.OverrideClusterSpecFile = "testdata/testOverrideClusterSpec.yaml"
	})

	workloadCluster := &types.Cluster{Name: clusterName}
	bootstrapCluster := &types.Cluster{Name: "eks-a-bootstrap"}

	fileName := fmt.Sprintf("%s-eks-a-cluster.yaml", clusterSpec.ObjectMeta.Name)
	fileContent, err := clusterSpec.ReadOverrideClusterSpecFile()
	if err != nil {
		t.Errorf("Unable to read test file: %v", err)
	}
	c, m := newClusterManager(t)
	m.writer.EXPECT().Write(fileName, []byte(fileContent)).Return("", nil)

	if _, err := c.GenerateDeploymentFile(ctx, bootstrapCluster, workloadCluster, clusterSpec, m.provider, false); err != nil {
		t.Errorf("ClusterManager.GenerateDeploymentFile() error = %v, wantErr nil", err)
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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write("cluster-name-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))

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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForManagedExternalEtcdReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write("cluster-name-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))

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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write("cluster-name-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	m.client.EXPECT().GetMachines(ctx, cluster).Return([]types.Machine{}, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write("cluster-name-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil())).Return(wantKubeconfigFile, nil)
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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail once
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef the rest of the retries
	m.client.EXPECT().GetMachines(ctx, cluster).MinTimes(1).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
	}}}, nil)

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
	m.provider.EXPECT().GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, "cluster-name-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, cluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, cluster, "60m", clusterName)
	// Fail a bunch of times
	m.client.EXPECT().GetMachines(ctx, cluster).Times(retries-5).Return(nil, errors.New("error get machines"))
	// Return a machine with no nodeRef  times
	m.client.EXPECT().GetMachines(ctx, cluster).Times(3).Return([]types.Machine{{Metadata: types.MachineMetadata{
		Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""},
	}}}, nil)
	//// Return a machine with nodeRef and another with it
	machines := []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: types.MachineStatus{NodeRef: &types.ResourceRef{}}},
	}
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(machines, nil)
	// Finally return two machines with node ref
	machines = []types.Machine{
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: types.MachineStatus{NodeRef: &types.ResourceRef{}}},
		{Metadata: types.MachineMetadata{Labels: map[string]string{clusterv1.MachineControlPlaneLabelName: ""}}, Status: types.MachineStatus{NodeRef: &types.ResourceRef{}}},
	}
	m.client.EXPECT().GetMachines(ctx, cluster).Times(1).Return(machines, nil)
	kubeconfig := []byte("content")
	m.client.EXPECT().GetWorkloadKubeconfig(ctx, clusterName, cluster).Return(kubeconfig, nil)
	m.provider.EXPECT().UpdateKubeConfig(&kubeconfig, clusterName)
	m.writer.EXPECT().Write("cluster-name-eks-a-cluster.kubeconfig", gomock.Any(), gomock.Not(gomock.Nil()))

	if _, err := c.CreateWorkloadCluster(ctx, cluster, clusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.CreateWorkloadCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	mClusterName := "management-cluster"
	mCluster := &types.Cluster{
		Name: mClusterName,
	}

	wClusterName := "workload-cluster"
	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = wClusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: wClusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateDeploymentFileForUpgrade(ctx, mCluster, wCluster, wClusterSpec, "workload-cluster-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, mCluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", wClusterName).MaxTimes(2)
	m.client.EXPECT().GetMachines(ctx, mCluster).Return([]types.Machine{}, nil)
	m.client.EXPECT().WaitForDeployment(ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).MaxTimes(10)
	m.provider.EXPECT().GetDeployments()

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err != nil {
		t.Errorf("ClusterManager.UpgradeCluster() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerUpgradeWorkloadClusterWaitForMachinesTimeout(t *testing.T) {
	ctx := context.Background()
	mClusterName := "management-cluster"
	mCluster := &types.Cluster{
		Name: mClusterName,
	}

	wClusterName := "workload-cluster"
	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = wClusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: wClusterName,
	}

	c, m := newClusterManager(t, clustermanager.WithWaitForMachines(1*time.Nanosecond, 50*time.Microsecond, 100*time.Microsecond))
	m.provider.EXPECT().GenerateDeploymentFileForUpgrade(ctx, mCluster, wCluster, wClusterSpec, "workload-cluster-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, mCluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", wClusterName)
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

func TestClusterManagerUpgradeWorkloadClusterWaitForCapiTimeout(t *testing.T) {
	ctx := context.Background()
	mClusterName := "management-cluster"
	mCluster := &types.Cluster{
		Name: mClusterName,
	}

	wClusterName := "workload-cluster"
	wClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = wClusterName
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
	})

	wCluster := &types.Cluster{
		Name: wClusterName,
	}

	c, m := newClusterManager(t)
	m.provider.EXPECT().GenerateDeploymentFileForUpgrade(ctx, mCluster, wCluster, wClusterSpec, "workload-cluster-eks-a-cluster.yaml")
	m.client.EXPECT().ApplyKubeSpecWithNamespace(ctx, mCluster, test.OfType("string"), constants.EksaSystemNamespace)
	m.client.EXPECT().WaitForControlPlaneReady(ctx, mCluster, "60m", wClusterName).MaxTimes(2)
	m.client.EXPECT().GetMachines(ctx, mCluster).Return([]types.Machine{}, nil)
	m.client.EXPECT().WaitForDeployment(ctx, wCluster, "30m", "Available", gomock.Any(), gomock.Any()).Return(errors.New("time out"))

	if err := c.UpgradeCluster(ctx, mCluster, wCluster, wClusterSpec, m.provider); err == nil {
		t.Error("ClusterManager.UpgradeCluster() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCapiSuccess(t *testing.T) {
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

	if err := c.MoveCapi(ctx, from, to); err != nil {
		t.Errorf("ClusterManager.MoveCapi() error = %v, wantErr nil", err)
	}
}

func TestClusterManagerMoveCapiErrorMove(t *testing.T) {
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

	if err := c.MoveCapi(ctx, from, to); err == nil {
		t.Error("ClusterManager.MoveCapi() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCapiErrorGetClusters(t *testing.T) {
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

	if err := c.MoveCapi(ctx, from, to); err == nil {
		t.Error("ClusterManager.MoveCapi() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCapiErrorWaitForControlPlane(t *testing.T) {
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

	if err := c.MoveCapi(ctx, from, to); err == nil {
		t.Error("ClusterManager.MoveCapi() error = nil, wantErr not nil")
	}
}

func TestClusterManagerMoveCapiErrorGetMachines(t *testing.T) {
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

	if err := c.MoveCapi(ctx, from, to); err == nil {
		t.Error("ClusterManager.MoveCapi() error = nil, wantErr not nil")
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

	m.client.EXPECT().GetNamespace(ctx, cluster.KubeconfigFile, clusterSpec.Namespace)
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
	tt := newTest(t)
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "testdata/testOverrideClusterSpec.yaml"

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Not(gomock.Nil())).Return(nil)

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
	tt.clusterSpec.VersionsBundle.Eksa.Components.URI = "testdata/testOverrideClusterSpec.yaml"
	tt.clusterManager.Retrier = retrier.NewWithMaxRetries(2, 0)

	tt.mocks.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Not(gomock.Nil())).Return(errors.New("error from apply")).Times(2)

	if err := tt.clusterManager.InstallCustomComponents(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("ClusterManager.InstallCustomComponents() error = nil, wantErr not nil")
	}
}

func TestClusterManagerClusterSpecChangedNoChanges(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	clusterSpec := v1alpha1.ClusterSpec{
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
				Name: clusterName,
			},
		}},
		DatacenterRef: v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: clusterName,
		},
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: clusterSpec,
		},
	}
	spec.Spec = clusterSpec

	// clusterConfig := &v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{KubernetesVersion: "1.19", WorkerNodeReplicas: 3, ControlPlaneReplicas: 3}}

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

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: clusterSpec,
		}, nil,
	)
	m.client.EXPECT().GetEksaVSphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, gomock.Any()).Return(datacenterConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any()).Return(machineConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any()).Return(machineConfig, nil)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &spec, datacenterConfig, []providers.MachineConfig{machineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have been detected")
}

func TestClusterManagerClusterSpecChangedClusterChanged(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	clusterSpec := v1alpha1.ClusterSpec{
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
				Name: clusterName,
			},
		}},
		DatacenterRef: v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: clusterName,
		},
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: clusterSpec,
		},
	}
	spec.Spec = clusterSpec
	modifiedSpec := spec
	modifiedSpec.Spec.KubernetesVersion = "1.20"
	// clusterConfig := &v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{KubernetesVersion: "1.19", WorkerNodeReplicas: 3, ControlPlaneReplicas: 3}}

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

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: clusterSpec,
		}, nil,
	)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &modifiedSpec, datacenterConfig, []providers.MachineConfig{machineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesDatacenterSpecChanged(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	clusterSpec := v1alpha1.ClusterSpec{
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
				Name: clusterName,
			},
		}},
		DatacenterRef: v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: clusterName,
		},
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: clusterSpec,
		},
	}
	spec.Spec = clusterSpec

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
	modifiedDatacenterSpec := *datacenterConfig
	modifiedDatacenterSpec.Spec.Insecure = false

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: clusterSpec,
		}, nil,
	)
	m.client.EXPECT().GetEksaVSphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, gomock.Any()).Return(datacenterConfig, nil)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &spec, &modifiedDatacenterSpec, []providers.MachineConfig{machineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesControlPlaneMachineConfigSpecChanged(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	clusterSpec := v1alpha1.ClusterSpec{
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
				Name: clusterName,
			},
		}},
		DatacenterRef: v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: clusterName,
		},
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: clusterSpec,
		},
	}
	spec.Spec = clusterSpec

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
	modifiedMachineConfigSpec := *machineConfig
	modifiedMachineConfigSpec.Spec.NumCPUs = 4

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: clusterSpec,
		}, nil,
	)
	m.client.EXPECT().GetEksaVSphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, gomock.Any()).Return(datacenterConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any()).Return(machineConfig, nil)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &spec, datacenterConfig, []providers.MachineConfig{&modifiedMachineConfigSpec})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedNoChangesWorkerNodeMachineConfigSpecChanged(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	controlPlaneMachineConfigName := "control-plane"
	workerNodeMachineConfigName := "worker-nodes"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	clusterSpec := v1alpha1.ClusterSpec{
		KubernetesVersion: "1.19",
		ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
			Count: 1,
			Endpoint: &v1alpha1.Endpoint{
				Host: "1.1.1.1",
			},
			MachineGroupRef: &v1alpha1.Ref{
				Name: controlPlaneMachineConfigName,
			},
		},
		WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
			Count: 1,
			MachineGroupRef: &v1alpha1.Ref{
				Name: workerNodeMachineConfigName,
			},
		}},
		DatacenterRef: v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: clusterName,
		},
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			Spec: clusterSpec,
		},
	}
	spec.Spec = clusterSpec

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Insecure: true,
		},
	}
	controlPlaneMachineConfig := &v1alpha1.VSphereMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: controlPlaneMachineConfigName,
		},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			DiskGiB:   20,
			MemoryMiB: 8192,
			NumCPUs:   2,
		},
	}
	workerNodeMachineConfig := &v1alpha1.VSphereMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: workerNodeMachineConfigName,
		},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			DiskGiB:   20,
			MemoryMiB: 8192,
			NumCPUs:   2,
		},
	}
	modifiedMachineConfigSpec := *workerNodeMachineConfig
	modifiedMachineConfigSpec.Spec.NumCPUs = 4

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: clusterSpec,
		}, nil,
	)
	m.client.EXPECT().GetEksaVSphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, gomock.Any()).Return(datacenterConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any()).Return(controlPlaneMachineConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any()).Return(workerNodeMachineConfig, nil)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &spec, datacenterConfig, []providers.MachineConfig{controlPlaneMachineConfig, &modifiedMachineConfigSpec})
	assert.Nil(t, err, "Error should be nil")
	assert.True(t, diff, "Changes should have been detected")
}

func TestClusterManagerClusterSpecChangedGitOpsDefault(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster-name"
	cl := &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: "sample.kubeconfig",
	}
	spec := cluster.Spec{
		Cluster: &v1alpha1.Cluster{
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
						Kind: v1alpha1.VSphereMachineConfigKind,
					},
				},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Name: clusterName,
						Kind: v1alpha1.VSphereMachineConfigKind,
					},
				}},
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: clusterName,
				},
				GitOpsRef: &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
				},
			},
		},
	}
	spec.SetDefaultGitOps()
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

	c, m := newClusterManager(t)
	m.client.EXPECT().GetEksaCluster(ctx, cl).Return(
		&v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				KubernetesVersion: "1.19",
				ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
					Count: 1,
					Endpoint: &v1alpha1.Endpoint{
						Host: "1.1.1.1",
					},
					MachineGroupRef: &v1alpha1.Ref{
						Name: clusterName,
						Kind: v1alpha1.VSphereMachineConfigKind,
					},
				},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Name: clusterName,
						Kind: v1alpha1.VSphereMachineConfigKind,
					},
				}},
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: clusterName,
				},
				GitOpsRef: &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
				},
			},
		}, nil,
	)
	m.client.EXPECT().GetEksaVSphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, gomock.Any()).Return(datacenterConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any()).Return(machineConfig, nil)
	m.client.EXPECT().GetEksaVSphereMachineConfig(ctx, spec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, gomock.Any()).Return(machineConfig, nil)
	diff, err := c.EKSAClusterSpecChanged(ctx, cl, &spec, datacenterConfig, []providers.MachineConfig{machineConfig})
	assert.Nil(t, err, "Error should be nil")
	assert.False(t, diff, "No changes should have not been detected")
}

type testSetup struct {
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
