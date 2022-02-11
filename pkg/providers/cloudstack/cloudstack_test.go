package cloudstack

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	etcdv1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	kubeadmnv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	testClusterConfigMainFilename = "cluster_main.yaml"
	testDataDir                   = "testdata"
	expectedCloudStackName        = "cloudstack"
	expectedCloudStackUsername    = "cloudstack_username"
	expectedCloudStackPassword    = "cloudstack_password"
	expectedCloudStackCloudConfig = "go="
	expectedExpClusterResourceSet = "expClusterResourceSetKey"
	eksd119Release                = "kubernetes-1-19-eks-4"
	eksd119ReleaseTag             = "eksdRelease:kubernetes-1-19-eks-4"
	eksd121ReleaseTag             = "eksdRelease:kubernetes-1-21-eks-4"
	ubuntuOSTag                   = "os:ubuntu"
	testTemplate                  = "centos7-k8s-118"
)

type DummyProviderCloudMonkeyClient struct {
	osTag string
}

func NewDummyProviderCloudMonkeyClient() *DummyProviderCloudMonkeyClient {
	return &DummyProviderCloudMonkeyClient{osTag: ubuntuOSTag}
}

func (pc *DummyProviderCloudMonkeyClient) TemplateHasSnapshot(ctx context.Context, template string) (bool, error) {
	return false, nil
}

func (pc *DummyProviderCloudMonkeyClient) GetWorkloadAvailableSpace(ctx context.Context, machineConfig *v1alpha1.CloudStackMachineConfig) (float64, error) {
	return math.MaxFloat64, nil
}

func (pc *DummyProviderCloudMonkeyClient) DeployTemplate(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) ValidateCloudStackSetup(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, selfSigned *bool) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) ValidateCloudStackSetupMachineConfig(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfig *v1alpha1.CloudStackMachineConfig, selfSigned *bool) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) SearchTemplate(ctx context.Context, domain string, zone string, account string, template string) (string, error) {
	return template, nil
}

func (pc *DummyProviderCloudMonkeyClient) SearchDiskOffering(ctx context.Context, domain string, zone string, account string, diskOffering string) (string, error) {
	return diskOffering, nil
}

func (pc *DummyProviderCloudMonkeyClient) SearchComputeOffering(ctx context.Context, domain string, zone string, account string, computeOffering string) (string, error) {
	return computeOffering, nil
}

func (pc *DummyProviderCloudMonkeyClient) LibraryElementExists(ctx context.Context, library string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderCloudMonkeyClient) GetLibraryElementContentVersion(ctx context.Context, element string) (string, error) {
	return "", nil
}

func (pc *DummyProviderCloudMonkeyClient) DeleteLibraryElement(ctx context.Context, element string) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) CreateLibrary(ctx context.Context, datastore, library string) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, resourcePool string, resizeDisk2 bool) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) ResizeDisk(ctx context.Context, template, diskName string, diskSizeInGB int) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) ImportTemplate(ctx context.Context, library, ovaURL, name string) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) GetTags(ctx context.Context, path string) (tags []string, err error) {
	return []string{eksd119ReleaseTag, eksd121ReleaseTag, pc.osTag}, nil
}

func (pc *DummyProviderCloudMonkeyClient) ListTags(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (pc *DummyProviderCloudMonkeyClient) CreateTag(ctx context.Context, tag, category string) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) AddTag(ctx context.Context, path, tag string) error {
	return nil
}

func (pc *DummyProviderCloudMonkeyClient) ListCategories(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (pc *DummyProviderCloudMonkeyClient) CreateCategoryForVM(ctx context.Context, name string) error {
	return nil
}

type DummyNetClient struct{}

func (n *DummyNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "255.255.255.255:22" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}

func givenClusterConfig(t *testing.T, fileName string) *v1alpha1.Cluster {
	return givenClusterSpec(t, fileName).Cluster
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = "1.19"
		s.VersionsBundle.EksD.Name = eksd119Release
		s.Namespace = "test-namespace"
	})
}

func fillClusterSpecWithClusterConfig(spec *cluster.Spec, clusterConfig *v1alpha1.Cluster) {
	spec.Cluster = clusterConfig
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.CloudStackDeploymentConfig {
	deploymentConfig, err := v1alpha1.GetCloudStackDeploymentConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return deploymentConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.CloudStackMachineConfig {
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file")
	}
	return machineConfigs
}

func givenProvider(t *testing.T) *cloudstackProvider {
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(
		t,
		deploymentConfig,
		machineConfigs,
		clusterConfig,
		nil,
	)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	return provider
}

func givenCloudMonkeyMock(t *testing.T) *mocks.MockProviderCloudMonkeyClient {
	ctrl := gomock.NewController(t)
	return mocks.NewMockProviderCloudMonkeyClient(ctrl)
}

type testContext struct {
	oldUsername                          string
	isUsernameSet                        bool
	oldPassword                          string
	isPasswordSet                        bool
	oldCloudStackCloudConfigSecretName   string
	isCloudStackCloudConfigSecretNameSet bool
	oldExpClusterResourceSet             string
	isExpClusterResourceSetSet           bool
}

func (tctx *testContext) SaveContext() {
	tctx.oldUsername, tctx.isUsernameSet = os.LookupEnv(eksacloudStackUsernameKey)
	tctx.oldPassword, tctx.isPasswordSet = os.LookupEnv(eksacloudStackPasswordKey)
	tctx.oldCloudStackCloudConfigSecretName, tctx.isCloudStackCloudConfigSecretNameSet = os.LookupEnv(eksacloudStackCloudConfigB64SecretKey)
	tctx.oldExpClusterResourceSet, tctx.isExpClusterResourceSetSet = os.LookupEnv(cloudStackPasswordKey)
	os.Setenv(eksacloudStackUsernameKey, expectedCloudStackUsername)
	os.Setenv(cloudStackPasswordKey, os.Getenv(eksacloudStackPasswordKey))
	os.Setenv(eksacloudStackCloudConfigB64SecretKey, expectedCloudStackCloudConfig)
	os.Setenv(cloudStackCloudConfigB64SecretKey, os.Getenv(eksacloudStackCloudConfigB64SecretKey))
	os.Setenv(expClusterResourceSetKey, expectedExpClusterResourceSet)
}

func (tctx *testContext) RestoreContext() {
	if tctx.isUsernameSet {
		os.Setenv(eksacloudStackUsernameKey, tctx.oldUsername)
	} else {
		os.Unsetenv(eksacloudStackUsernameKey)
	}
	if tctx.isPasswordSet {
		os.Setenv(eksacloudStackPasswordKey, tctx.oldPassword)
	} else {
		os.Unsetenv(eksacloudStackPasswordKey)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

type providerTest struct {
	*WithT
	ctx                                context.Context
	managementCluster, workloadCluster *types.Cluster
	provider                           *cloudstackProvider
	cluster                            *v1alpha1.Cluster
	clusterSpec                        *cluster.Spec
	datacenterConfig                   *v1alpha1.CloudStackDeploymentConfig
	machineConfigs                     map[string]*v1alpha1.CloudStackMachineConfig
	kubectl                            *mocks.MockProviderKubectlClient
	govc                               *mocks.MockProviderCloudMonkeyClient
	resourceSetManager                 *mocks.MockClusterResourceSetManager
}

func newProviderTest(t *testing.T) *providerTest {
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	govc := mocks.NewMockProviderCloudMonkeyClient(ctrl)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(ctrl)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProvider(
		t,
		deploymentConfig,
		machineConfigs,
		clusterConfig,
		govc,
		kubectl,
		resourceSetManager,
	)
	return &providerTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		managementCluster: &types.Cluster{
			Name:           "m-cluster",
			KubeconfigFile: "kubeconfig-m.kubeconfig",
		},
		workloadCluster: &types.Cluster{
			Name:           "test",
			KubeconfigFile: "kubeconfig-w.kubeconfig",
		},
		provider:           provider,
		cluster:            clusterConfig,
		clusterSpec:        givenClusterSpec(t, testClusterConfigMainFilename),
		datacenterConfig:   deploymentConfig,
		machineConfigs:     machineConfigs,
		kubectl:            kubectl,
		govc:               govc,
		resourceSetManager: resourceSetManager,
	}
}

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		deploymentConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
}

func newProviderWithKubectl(t *testing.T, datacenterConfig *v1alpha1.CloudStackDeploymentConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient) *cloudstackProvider {
	ctrl := gomock.NewController(t)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(ctrl)
	return newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		NewDummyProviderCloudMonkeyClient(),
		kubectl,
		resourceSetManager,
	)
}

func newProvider(t *testing.T, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, govc ProviderCloudMonkeyClient, kubectl ProviderKubectlClient, resourceSetManager ClusterResourceSetManager) *cloudstackProvider {
	_, writer := test.NewWriter(t)
	return NewProviderCustomNet(
		deploymentConfig,
		machineConfigs,
		clusterConfig,
		govc,
		kubectl,
		writer,
		&DummyNetClient{},
		test.FakeNow,
		false,
		resourceSetManager,
	)
}

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineTemplate(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "minimal",
			clusterconfigFile: "cluster_minimal.yaml",
			wantCPFile:        "testdata/expected_results_minimal_cp.yaml",
			wantMDFile:        "testdata/expected_results_minimal_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDeploymentConfig{
				Spec: v1alpha1.CloudStackDeploymentConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{},
			}

			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			deploymentConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateExtraEndpointAsCertSANs(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "certSANs",
			clusterconfigFile: "cluster_extraEndpoints.yaml",
			wantCPFile:        "testdata/expected_results_extraEndpoints_cp.yaml",
			wantMDFile:        "testdata/expected_results_extraEndpoints_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDeploymentConfig{
				Spec: v1alpha1.CloudStackDeploymentConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{},
			}

			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			deploymentConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineTemplateExternalEtcd(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "main",
			clusterconfigFile: testClusterConfigMainFilename,
			wantCPFile:        "testdata/expected_results_main_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDeploymentConfig{
				Spec: v1alpha1.CloudStackDeploymentConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{},
			}

			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1alpha3.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))
			deploymentConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	oldCP := &kubeadmnv1alpha3.KubeadmControlPlane{
		Spec: kubeadmnv1alpha3.KubeadmControlPlaneSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-control-plane-template-original",
			},
		},
	}
	oldMD := &v1alpha3.MachineDeployment{
		Spec: v1alpha3.MachineDeploymentSpec{
			Template: v1alpha3.MachineTemplateSpec{
				Spec: v1alpha3.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-worker-node-template-original",
					},
				},
			},
		},
	}
	etcdadmCluster := &etcdv1alpha3.EtcdadmCluster{
		Spec: etcdv1alpha3.EtcdadmClusterSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-etcd-template-original",
			},
		},
	}

	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Name).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Namespace).Return(deploymentConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Namespace).Return(machineConfigs[controlPlaneMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Namespace).Return(machineConfigs[workerNodeMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Namespace).Return(machineConfigs[etcdMachineConfigName], nil)
	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterSpec.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldCP, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, cluster, clusterSpec.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldMD, nil)
	kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, clusterSpec.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadmCluster, nil)
	cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_no_machinetemplate_update_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_no_machinetemplate_update_md.yaml")
}

func TestProviderGenerateCAPISpecForCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestProviderGenerateDeploymentFileWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	deploymentConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cloudmonkey := NewDummyProviderCloudMonkeyClient()
	cloudmonkey.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_md.yaml")
}

func TestProviderGenerateDeploymentFileWithMirrorAndCertConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_with_cert_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	deploymentConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cloudmonkey := NewDummyProviderCloudMonkeyClient()
	cloudmonkey.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterSpec.Cluster, kubectl)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_with_cert_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_with_cert_md.yaml")
}

func TestUpdateKubeConfig(t *testing.T) {
	provider := givenProvider(t)
	content := []byte{}

	err := provider.UpdateKubeConfig(&content, "clusterName")
	if err != nil {
		t.Fatalf("failed UpdateKubeConfig: %v", err)
	}
}

func TestBootstrapClusterOpts(t *testing.T) {
	provider := givenProvider(t)

	bootstrapClusterOps, err := provider.BootstrapClusterOpts()
	if err != nil {
		t.Fatalf("failed BootstrapClusterOpts: %v", err)
	}
	if bootstrapClusterOps == nil {
		t.Fatalf("expected BootstrapClusterOpts")
	}
}

func TestName(t *testing.T) {
	provider := givenProvider(t)

	if provider.Name() != expectedCloudStackName {
		t.Fatalf("unexpected Name %s!=%s", provider.Name(), expectedCloudStackName)
	}
}

func TestSetupAndValidateCreateCluster(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func thenErrorPrefixExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("Expected=<%s...> actual=<%s...>", expected, actual)
	}
}

func thenErrorExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if expected != actual {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

func thenErrorContainsExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if !strings.Contains(actual, expected) {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

func TestSetupAndValidateCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDeploymentConfig(context.TODO(), datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Namespace).Return([]*v1alpha1.CloudStackDeploymentConfig{}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfMachineExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	idx := 0
	var existingMachine string
	for _, config := range newMachineConfigs {
		if idx == 0 {
			kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{config}, nil)
			existingMachine = config.Name
		} else {
			kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil).MaxTimes(1)
		}
		idx++
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("CloudStackMachineConfig %s already exists", existingMachine), err)
}

func TestSetupAndValidateSelfManagedClusterSkipMachineNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfDatacenterExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDeploymentConfig(context.TODO(), deploymentConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Namespace).Return([]*v1alpha1.CloudStackDeploymentConfig{deploymentConfig}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("CloudStackDeployment %s already exists", deploymentConfig.Name), err)
}

func TestSetupAndValidateSelfManagedClusterSkipDatacenterNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	kubectl.EXPECT().SearchCloudStackDeploymentConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateDeleteCluster(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	err := provider.SetupAndValidateDeleteCluster(ctx)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeCluster(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterIpExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtl)
	provider.providerKubectlClient = kubectl
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterCPSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterWorkerSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterEtcdSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestCleanupProviderInfrastructure(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	err := provider.CleanupProviderInfrastructure(ctx)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestVersion(t *testing.T) {
	cloudStackProviderVersion := "v4.14.1"
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	clusterSpec.VersionsBundle.CloudStack.Version = cloudStackProviderVersion
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	result := provider.Version(clusterSpec)
	if result != cloudStackProviderVersion {
		t.Fatalf("Unexpected version expected <%s> actual=<%s>", cloudStackProviderVersion, result)
	}
}

func TestSetupAndValidateCreateClusterInsecure(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.deploymentConfig.Spec.Insecure = true
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.deploymentConfig.Spec.Thumbprint != "" {
		t.Fatalf("Expected=<> actual=<%s>", provider.deploymentConfig.Spec.Thumbprint)
	}
}

func TestSetupAndValidateCreateClusterNoControlPlaneEndpointIP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoDatacenter(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.deploymentConfig.Spec.ManagementApiEndpoint = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: CloudStackDeploymentConfig managementApiEndpoint is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoNetwork(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.deploymentConfig.Spec.Network = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "CloudStackDeploymentConfig network is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterEndpointPortNotSpecified(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host = "host1"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	assert.Nil(t, err)
	assert.Equal(t, "host1:6443", clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
}

func TestSetupAndValidateCreateClusterEndpointPortSpecified(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host = "host1:443"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	assert.Nil(t, err)
	assert.Equal(t, "host1:443", clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: provided CloudStackMachineConfig sshAuthorizedKey is invalid: ssh: no key found", err)
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey did not get generated for etcd machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyAllMachineConfigs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey not generated for etcd machines")
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and worker machines")
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and etcd machines")
	}
}

func TestSetupAndValidateUsersNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateSshAuthorizedKeysNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for control plane", err)
	}
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for worker nodes", err)
	}
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for etcd machines", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for control plane", err)
	}
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for worker nodes", err)
	}
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find CloudStackMachineConfig nonexistent for etcd machines", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyDifferent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = "redhat"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and worker nodes must have the same osFamily specified", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyDifferentForEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily = "ubuntu"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and etcd machines must have the same osFamily specified", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyEmpty(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	deploymentConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cloudmonkey := NewDummyProviderCloudMonkeyClient()
	cloudmonkey.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, deploymentConfig, machineConfigs, clusterConfig, nil)
	provider.providerCloudMonkeyClient = cloudmonkey
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].Name = ""
	workerNodeMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].Name = ""
	etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].Name = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily != v1alpha1.Redhat {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Redhat)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.OSFamily != v1alpha1.Redhat {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Redhat)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily != v1alpha1.Redhat {
		t.Fatalf("got osFamily for etcd machine as %v, want %v", provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily, v1alpha1.Redhat)
	}
}

func TestSetupAndValidateCreateClusterTemplateDifferent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Template = "test"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and worker nodes must have the same template specified", err)
	}
}

func TestSetupAndValidateCreateClusterTemplateDoesNotExist(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	cloudmonkey := givenCloudMonkeyMock(t)
	provider.providerCloudMonkeyClient = cloudmonkey
	setupContext(t)

	cloudmonkey.EXPECT().ValidateCloudStackSetup(ctx, provider.deploymentConfig, &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)

	template := provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Template
	templateNotFoundError := fmt.Errorf("%s not found. Has the template been imported?", template)
	cloudmonkey.EXPECT().SearchTemplate(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Template).Return("", templateNotFoundError).AnyTimes()
	cloudmonkey.EXPECT().SearchDiskOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.DiskOffering).Return("", nil).AnyTimes()
	cloudmonkey.EXPECT().SearchComputeOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.ComputeOffering).Return("", nil).AnyTimes()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorContainsExpected(t, "error validating template: "+template+" not found. Has the template been imported?", err)
}

func TestSetupAndValidateCreateClusterComputeOfferingDoesNotExist(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	cloudmonkey := givenCloudMonkeyMock(t)
	provider.providerCloudMonkeyClient = cloudmonkey
	setupContext(t)

	cloudmonkey.EXPECT().ValidateCloudStackSetup(ctx, provider.deploymentConfig, &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)

	computeOffering := provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.ComputeOffering
	computeOfferingNotFoundError := fmt.Errorf("%s not found. Has the compute offering been imported?", computeOffering)
	cloudmonkey.EXPECT().SearchTemplate(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Template).Return("", nil).AnyTimes()
	cloudmonkey.EXPECT().SearchDiskOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.DiskOffering).Return("", nil).AnyTimes()
	cloudmonkey.EXPECT().SearchComputeOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.ComputeOffering).Return("", computeOfferingNotFoundError).AnyTimes()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorContainsExpected(t, "error validating computeOffering: "+computeOffering+" not found. Has the compute offering been imported?", err)
}

func TestSetupAndValidateCreateClusterDiskOfferingDoesNotExist(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	cloudmonkey := givenCloudMonkeyMock(t)
	provider.providerCloudMonkeyClient = cloudmonkey
	setupContext(t)

	cloudmonkey.EXPECT().ValidateCloudStackSetup(ctx, provider.deploymentConfig, &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)

	diskOffering := provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.DiskOffering
	diskOfferingNotFoundError := fmt.Errorf("%s not found. Has the disk offering been imported?", diskOffering)
	cloudmonkey.EXPECT().SearchTemplate(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Template).Return("", nil).AnyTimes()
	cloudmonkey.EXPECT().SearchDiskOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.DiskOffering).Return("", diskOfferingNotFoundError).AnyTimes()
	cloudmonkey.EXPECT().SearchComputeOffering(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.ComputeOffering).Return("", nil).AnyTimes()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorContainsExpected(t, "error validating diskOffering: "+diskOffering+" not found. Has the disk offering been imported?", err)
}

func TestSetupAndValidateCreateClusterErrorCheckingTemplate(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	cloudmonkey := givenCloudMonkeyMock(t)
	provider.providerCloudMonkeyClient = cloudmonkey
	setupContext(t)

	cloudmonkey.EXPECT().ValidateCloudStackSetup(ctx, provider.deploymentConfig, &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().ValidateCloudStackSetupMachineConfig(ctx, provider.deploymentConfig, provider.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name], &provider.selfSigned).Return(nil)
	cloudmonkey.EXPECT().SearchTemplate(ctx, provider.deploymentConfig.Spec.Domain, provider.deploymentConfig.Spec.Zone, provider.deploymentConfig.Spec.Account, provider.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Template).Return("", errors.New("failed to get template"))

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorContainsExpected(t, "error validating template: failed to get template", err)
}

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
	}{
		{
			testName: "correct Overrides layer",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.CloudStack = releasev1alpha1.CloudStackBundle{
					Version: "v0.1.0",
					ClusterAPIController: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-cloudstack/release/manager:v0.1.0",
					},
					Manager: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes/cloud-provider-cloudstack/cpi/manager:v0.1.0",
					},
					Metadata: releasev1alpha1.Manifest{
						URI: "Metadata.yaml",
					},
					Components: releasev1alpha1.Manifest{
						URI: "Components.yaml",
					},
					ClusterTemplate: releasev1alpha1.Manifest{
						URI: "ClusterTemplate.yaml",
					},
				}
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			p := givenProvider(t)

			infraBundle := p.GetInfrastructureBundle(tt.clusterSpec)
			if infraBundle == nil {
				t.Fatalf("provider.GetInfrastructureBundle() should have an infrastructure bundle")
			}
			assert.Equal(t, "infrastructure-cloudstack/v0.1.0/", infraBundle.FolderName, "Incorrect folder name")
			assert.Equal(t, len(infraBundle.Manifests), 3, "Wrong number of files in the infrastructure bundle")
			wantManifests := []releasev1alpha1.Manifest{
				tt.clusterSpec.VersionsBundle.CloudStack.Components,
				tt.clusterSpec.VersionsBundle.CloudStack.Metadata,
				tt.clusterSpec.VersionsBundle.CloudStack.ClusterTemplate,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

func TestGetDatacenterConfig(t *testing.T) {
	provider := givenProvider(t)
	provider.deploymentConfig.TypeMeta.Kind = "kind"

	providerConfig := provider.DatacenterConfig()
	if providerConfig.Kind() != "kind" {
		t.Fatal("Unexpected error DatacenterConfig: kind field not found")
	}
}

func TestValidateNewSpecSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})
	c := &types.Cluster{}

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	err := provider.ValidateNewSpec(context.TODO(), c, clusterSpec)
	assert.NoError(t, err, "No error should be returned when previous spec == new spec")
}

func TestValidateNewSpecDatacenterImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.ManagementApiEndpoint = newProviderConfig.Spec.ManagementApiEndpoint + "-new"

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "spec.managementApiEndpoint is immutable. Previous value https://127.0.0.1:8080/client/api-new, new value https://127.0.0.1:8080/client/api")
}

func TestValidateNewSpecMachineConfigNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.ManagementApiEndpoint = newProviderConfig.Spec.ManagementApiEndpoint + "-new"

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig.DeepCopy()
		s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "missing-machine-group"
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "missing-machine-group"
		s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "missing-machine-group"
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Errorf(t, err, "can't find machine config missing-machine-group in cloudstack provider machine configs")
}

func TestValidateNewSpecTLSInsecureImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Insecure = !newProviderConfig.Spec.Insecure
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Insecure should be immutable")
}

func TestValidateNewSpecTLSThumbprintImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Thumbprint = "new-" + newProviderConfig.Spec.Thumbprint

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Thumbprint should be immutable")
}

func TestValidateNewSpecMachineConfigSshUsersImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.ManagementApiEndpoint = newProviderConfig.Spec.ManagementApiEndpoint + "-new"

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	newMachineConfigs["test-cp"].Spec.Users[0].Name = "newNameShouldNotBeAllowed"

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "User should be immutable")
}

func TestValidateNewSpecMachineConfigSshAuthKeysImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.ManagementApiEndpoint = newProviderConfig.Spec.ManagementApiEndpoint + "-new"

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackDeploymentConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "SSH Authorized Keys should be immutable")
}

func TestGetMHCSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	mhcTemplate := fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineHealthCheck
metadata:
  name: test-node-unhealthy-5m
  namespace: %[1]s
spec:
  clusterName: test
  maxUnhealthy: 40%%
  nodeStartupTimeout: 10m
  selector:
    matchLabels:
      cluster.x-k8s.io/deployment-name: "test-md-0"
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: 300s
    - type: Ready
      status: "False"
      timeout: 300s
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineHealthCheck
metadata:
  name: test-kcp-unhealthy-5m
  namespace: %[1]s
spec:
  clusterName: test
  maxUnhealthy: 100%%
  selector:
    matchLabels:
      cluster.x-k8s.io/control-plane: ""
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: 300s
    - type: Ready
      status: "False"
      timeout: 300s
`, constants.EksaSystemNamespace)

	mch, err := provider.GenerateMHC()
	assert.NoError(t, err, "Expected successful execution of GenerateMHC() but got an error", "error", err)
	assert.Equal(t, string(mch), mhcTemplate, "generated MachineHealthCheck is different from the expected one")
}

func TestChangeDiffNoChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	assert.Nil(t, provider.ChangeDiff(clusterSpec, clusterSpec))
}

func TestChangeDiffWithChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.CloudStack.Version = "v0.2.0"
	})
	newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.CloudStack.Version = "v0.1.0"
	})

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "cloudstack",
		NewVersion:    "v0.1.0",
		OldVersion:    "v0.2.0",
	}

	assert.Equal(t, wantDiff, provider.ChangeDiff(clusterSpec, newClusterSpec))
}

func TestCloudStackProviderRunPostUpgrade(t *testing.T) {
	tt := newProviderTest(t)

	tt.resourceSetManager.EXPECT().ForceUpdate(tt.ctx, "test-crs-0", "eksa-system", tt.managementCluster, tt.workloadCluster)

	tt.Expect(tt.provider.RunPostUpgrade(tt.ctx, tt.clusterSpec, tt.managementCluster, tt.workloadCluster)).To(Succeed())
}

func TestProviderUpgradeNeeded(t *testing.T) {
	testCases := []struct {
		testName               string
		newManager, oldManager string
		want                   bool
	}{
		{
			testName:   "different manager",
			newManager: "a", oldManager: "b",
			want: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			provider := givenProvider(t)
			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.CloudStack.Manager.ImageDigest = tt.oldManager
			})

			newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.CloudStack.Manager.ImageDigest = tt.newManager
			})

			g := NewWithT(t)
			g.Expect(provider.UpgradeNeeded(context.Background(), clusterSpec, newClusterSpec)).To(Equal(tt.want))
		})
	}
}
