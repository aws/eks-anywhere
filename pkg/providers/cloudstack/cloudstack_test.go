package cloudstack

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	expectedCloudStackName = "cloudstack"
	eksd119Release         = "kubernetes-1-19-eks-4"

	/* Generated from ini file (like the following) then b64 encoded: `cat fake-cloud-config.ini | base64 | tr -d '\n'`
	[Global]
	verify-ssl = false
	api-key = test-key1
	secret-key = test-secret1
	api-url = http://127.16.0.1:8080/client/api
	*/
	expectedCloudStackCloudConfig       = "W0dsb2JhbF0KdmVyaWZ5LXNzbCA9IGZhbHNlCmFwaS1rZXkgPSB0ZXN0LWtleTEKc2VjcmV0LWtleSA9IHRlc3Qtc2VjcmV0MQphcGktdXJsID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCg=="
	cloudStackCloudConfigWithInvalidUrl = "W0dsb2JhbF0KdmVyaWZ5LXNzbCA9IGZhbHNlCmFwaS1rZXkgPSB0ZXN0LWtleTEKc2VjcmV0LWtleSA9IHRlc3Qtc2VjcmV0MQphcGktdXJsID0geHh4Cg=="
)

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
		s.Cluster.Namespace = "test-namespace"
	})
}

func givenWildcardCmk(mockCtrl *gomock.Controller) ProviderCmkClient {
	cmk := mocks.NewMockProviderCmkClient(mockCtrl)
	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateDiskOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateZoneAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateCloudStackConnection(gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateDomainAndGetId(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAccountPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateNetworkPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().GetManagementApiEndpoint(gomock.Any()).AnyTimes().Return("http://127.16.0.1:8080/client/api", nil)
	return cmk
}

func fillClusterSpecWithClusterConfig(spec *cluster.Spec, clusterConfig *v1alpha1.Cluster) {
	spec.Cluster = clusterConfig
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.CloudStackDatacenterConfig {
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return datacenterConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.CloudStackMachineConfig {
	machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file: %v", err)
	}
	return machineConfigs
}

func givenProvider(t *testing.T) *cloudstackProvider {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		nil,
		cmk,
	)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	return provider
}

type testContext struct {
	oldCloudStackCloudConfigSecretName   string
	isCloudStackCloudConfigSecretNameSet bool
}

func workerNodeGroup1MachineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-template-1234567890000",
						},
					},
				},
			},
		},
	}
}

func workerNodeGroup2MachineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-1-template-1234567890000",
						},
					},
				},
			},
		},
	}
}

func (tctx *testContext) SaveContext() {
	tctx.oldCloudStackCloudConfigSecretName, tctx.isCloudStackCloudConfigSecretNameSet = os.LookupEnv(decoder.EksacloudStackCloudConfigB64SecretKey)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, expectedCloudStackCloudConfig)
	os.Setenv(decoder.CloudStackCloudConfigB64SecretKey, os.Getenv(decoder.EksacloudStackCloudConfigB64SecretKey))
}

func setupContext() {
	var tctx testContext
	tctx.SaveContext()
}

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
		cmk,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
}

func newProviderWithKubectl(t *testing.T, datacenterConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, cmk ProviderCmkClient) *cloudstackProvider {
	return newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
		cmk,
	)
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, cmk ProviderCmkClient) *cloudstackProvider {
	_, writer := test.NewWriter(t)

	return NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
		cmk,
		writer,
		test.FakeNow,
		false,
	)
}

func TestProviderGenerateCAPISpecForCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
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

func TestProviderSetupAndValidateCreateClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderSetupAndValidateUpgradeClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderSetupAndValidateDeleteClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateDeleteCluster(ctx, cluster)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderSetupAndValidateCreateClusterFailureOnInvalidClusterSpec(t *testing.T) {
	tt := NewWithT(t)
	clusterSpecManifest := "cluster_invalid.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderSetupAndValidateUpgradeClusterFailureOnInvalidClusterSpec(t *testing.T) {
	tt := NewWithT(t)
	clusterSpecManifest := "cluster_invalid.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderGenerateCAPISpecForCreateWithAffinity(t *testing.T) {
	clusterSpecManifest := "cluster_affinity.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_affinity_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_affinity_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

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

func TestProviderGenerateCAPISpecForCreateWithMirrorAndCertConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_with_cert_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

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

func TestProviderGenerateCAPISpecForCreateWithProxyConfig(t *testing.T) {
	clusterSpecManifest := "cluster_minimal_proxy.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_minimal_proxy_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_minimal_proxy_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMultipleWorkerNodeGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main_multiple_worker_node_groups.yaml")

	datacenterConfig := givenDatacenterConfig(t, "cluster_main_multiple_worker_node_groups.yaml")
	machineConfigs := givenMachineConfigs(t, "cluster_main_multiple_worker_node_groups.yaml")
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	_, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_multiple_worker_node_groups.yaml")
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

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDatacenterConfig(context.TODO(), datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.CloudStackDatacenterConfig{}, nil)

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

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
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

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDatacenterConfig(context.TODO(), datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.CloudStackDatacenterConfig{datacenterConfig}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("CloudStackDatacenter %s already exists", datacenterConfig.Name), err)
}

func TestSetupAndValidateSelfManagedClusterSkipDatacenterNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	kubectl.EXPECT().SearchCloudStackMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	kubectl.EXPECT().SearchCloudStackDatacenterConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

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

	err := provider.SetupAndValidateDeleteCluster(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestCleanupProviderInfrastructure(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()

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

	result := provider.Version(clusterSpec)
	if result != cloudStackProviderVersion {
		t.Fatalf("Unexpected version expected <%s> actual=<%s>", cloudStackProviderVersion, result)
	}
}

func TestSetupAndValidateCreateClusterEndpointPortNotSpecified(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "host1"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	assert.Nil(t, err)
	assert.Equal(t, "host1:6443", clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
}

func TestSetupAndValidateCreateClusterEndpointPortSpecified(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "host1:443"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	assert.Nil(t, err)
	assert.Equal(t, "host1:443", clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
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
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
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
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
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
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
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
					KubeVip: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.2-2093eaeda5a4567f0e516d652e0b25b1d7abc774",
					},
					Metadata: releasev1alpha1.Manifest{
						URI: "Metadata.yaml",
					},
					Components: releasev1alpha1.Manifest{
						URI: "Components.yaml",
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
			assert.Equal(t, len(infraBundle.Manifests), 2, "Wrong number of files in the infrastructure bundle")
			wantManifests := []releasev1alpha1.Manifest{
				tt.clusterSpec.VersionsBundle.CloudStack.Components,
				tt.clusterSpec.VersionsBundle.CloudStack.Metadata,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

func TestGetDatacenterConfig(t *testing.T) {
	provider := givenProvider(t)
	provider.datacenterConfig.TypeMeta.Kind = "kind"

	providerConfig := provider.DatacenterConfig(givenEmptyClusterSpec())
	if providerConfig.Kind() != "kind" {
		t.Fatal("Unexpected error DatacenterConfig: kind field not found")
	}
}

func TestGetMHCSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	mhcTemplate := fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
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
apiVersion: cluster.x-k8s.io/v1beta1
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

	mch, err := provider.GenerateMHC(nil)
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
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDatacenterConfig{
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			cmk := givenWildcardCmk(mockCtrl)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
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
		{
			testName:          "main_with_taints",
			clusterconfigFile: "cluster_main_with_taints.yaml",
			wantCPFile:        "testdata/expected_results_main_with_taints_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_with_taints_md.yaml",
		},
		{
			testName:          "main with node labels",
			clusterconfigFile: "cluster_main_with_node_labels.yaml",
			wantCPFile:        "testdata/expected_results_main_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_node_labels_md.yaml",
		},
		{
			testName:          "main with cp node labels",
			clusterconfigFile: "cluster_main_with_cp_node_labels.yaml",
			wantCPFile:        "testdata/expected_results_main_node_labels_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			var tctx testContext
			tctx.SaveContext()
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDatacenterConfig{
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			cmk := givenWildcardCmk(mockCtrl)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
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
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	oldCP := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					Name: "test-control-plane-template-original",
				},
			},
		},
	}
	oldMD := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-md-0-original",
					},
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-template-original",
						},
					},
				},
			},
		},
	}
	etcdadmCluster := &etcdv1.EtcdadmCluster{
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-etcd-template-original",
			},
		},
	}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	cmk := givenWildcardCmk(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineDeploymentName := fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(datacenterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[controlPlaneMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[workerNodeMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[etcdMachineConfigName], nil)
	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldCP, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldMD, nil).Times(2)
	kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadmCluster, nil)
	cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_no_machinetemplate_update_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_no_machinetemplate_update_md.yaml")
}

func TestProviderGenerateCAPISpecForUpgradeMultipleWorkerNodeGroups(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantMDFile        string
	}{
		{
			testName:          "adding a worker node group",
			clusterconfigFile: "cluster_main_multiple_worker_node_groups.yaml",
			wantMDFile:        "testdata/expected_results_minimal_add_worker_node_group.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext()
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			cloudstackDatacenter := &v1alpha1.CloudStackDatacenterConfig{
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{},
			}
			cloudstackMachineConfig := &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}
			newClusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			newConfig := v1alpha1.WorkerNodeGroupConfiguration{Count: 1, MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "CloudStackMachineConfig"}, Name: "md-2"}
			newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newConfig)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup2MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil).AnyTimes()
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))

			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			cmk := givenWildcardCmk(mockCtrl)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, cmk)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			_, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, newClusterSpec)
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
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

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
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
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
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
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

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
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "setting up SSH keys: ssh: no key found", err)
}

func TestClusterUpgradeNeededNoChanges(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)
	for _, value := range machineConfigsMap {
		kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, value.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(value, nil)
	}
	provider := newProviderWithKubectl(t, dcConfig, machineConfigsMap, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if specChanged {
		t.Fatalf("expected no spec change to be detected")
	}
}

func TestClusterUpgradeNeededDatacenterConfigChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	shinyModifiedDcConfig := dcConfig.DeepCopy()
	shinyModifiedDcConfig.Spec.ManagementApiEndpoint = "shiny-new-api-endpoint"
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	provider := newProviderWithKubectl(t, dcConfig, machineConfigsMap, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(shinyModifiedDcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestClusterUpgradeNeededMachineConfigsChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)
	modifiedMachineConfig := machineConfigsMap[cc.MachineConfigRefs()[0].Name].DeepCopy()
	modifiedMachineConfig.Spec.Affinity = "shiny-new-affinity"
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(modifiedMachineConfig, nil)
	provider := newProviderWithKubectl(t, dcConfig, machineConfigsMap, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestClusterUpgradeNeededMachineConfigsChangedDiskOffering(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)
	getEksaCloudStackMachineConfig := kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).AnyTimes()
	getEksaCloudStackMachineConfig.DoAndReturn(
		func(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error) {
			if cloudstackMachineConfigName == "test" {
				modifiedMachineConfig := machineConfigsMap["test"].DeepCopy()
				modifiedMachineConfig.Spec.DiskOffering.Name = "shiny-new-diskoffering"
				return modifiedMachineConfig, nil
			}
			return machineConfigsMap[cloudstackMachineConfigName], nil
		})

	provider := newProviderWithKubectl(t, dcConfig, machineConfigsMap, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestAnyImmutableFieldChangedDiskOfferingNoChange(t *testing.T) {
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newDcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	assert.False(t, AnyImmutableFieldChanged(dcConfig, newDcConfig, machineConfigsMap["test"], newMachineConfigsMap["test"]), "Should not have any immutable fields changes")
}

func TestAnyImmutableFieldChangedDiskOfferingNameChange(t *testing.T) {
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newDcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newMachineConfigsMap["test"].Spec.DiskOffering.Name = "newDiskOffering"
	assert.True(t, AnyImmutableFieldChanged(dcConfig, newDcConfig, machineConfigsMap["test"], newMachineConfigsMap["test"]), "Should not have any immutable fields changes")
}

func TestAnyImmutableFieldChangedSymlinksAdded(t *testing.T) {
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newDcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newMachineConfigsMap["test"].Spec.Symlinks["/new/folder"] = "/data/new/folder"
	assert.True(t, AnyImmutableFieldChanged(dcConfig, newDcConfig, machineConfigsMap["test"], newMachineConfigsMap["test"]), "Should not have any immutable fields changes")
}

func TestAnyImmutableFieldChangedSymlinksChange(t *testing.T) {
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	newDcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)

	for k, v := range newMachineConfigsMap["test"].Spec.Symlinks {
		newMachineConfigsMap["test"].Spec.Symlinks[k] = "/new" + v
	}
	assert.True(t, AnyImmutableFieldChanged(dcConfig, newDcConfig, machineConfigsMap["test"], newMachineConfigsMap["test"]), "Should not have any immutable fields changes")
}

func TestInstallCustomProviderComponentsKubeVipEnabled(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenEmptyClusterSpec()
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigsMap := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, dcConfig, machineConfigsMap, cc, kubectl, nil)
	kubeConfigFile := "test"

	oldCloudstackKubeVipDisabledVal := os.Getenv(features.CloudStackKubeVipDisabledEnvVar)
	os.Unsetenv(features.CloudStackKubeVipDisabledEnvVar)
	defer os.Setenv(features.CloudStackKubeVipDisabledEnvVar, oldCloudstackKubeVipDisabledVal)
	kubectl.EXPECT().SetEksaControllerEnvVar(ctx, features.CloudStackProviderEnvVar, "true", kubeConfigFile).Return(nil)
	kubectl.EXPECT().SetEksaControllerEnvVar(ctx, features.CloudStackKubeVipDisabledEnvVar, "false", kubeConfigFile).Return(nil)
	if err := provider.InstallCustomProviderComponents(ctx, kubeConfigFile); err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}
