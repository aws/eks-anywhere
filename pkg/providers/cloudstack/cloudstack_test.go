package cloudstack

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
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
	api-key    = test-key
	secret-key = test-secret
	api-url    = http://127.16.0.1:8080/client/api
	verify-ssl = true
	*/
	expectedCloudStackCloudConfig = "W0dsb2JhbF0KYXBpLWtleSAgICA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAphcGktdXJsICAgID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgPSB0cnVlCg=="
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
	cmk.EXPECT().ValidateTemplatePresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateServiceOfferingPresent(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateZonesPresent(gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAffinityGroupsPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateCloudStackConnection(gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateDomainPresent(gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateAccountPresent(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	cmk.EXPECT().ValidateNetworkPresent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
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
	return NewProviderCustomNet(
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

func TestProviderGenerateDeploymentFileWithMirrorConfig(t *testing.T) {
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

func TestProviderGenerateDeploymentFileWithMirrorAndCertConfig(t *testing.T) {
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

func TestProviderGenerateDeploymentFileWithProxyConfig(t *testing.T) {
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

	err := provider.SetupAndValidateDeleteCluster(ctx)
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
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
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
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
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
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
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
				s.VersionsBundle.CloudStack.ClusterAPIController.ImageDigest = tt.oldManager
			})

			newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.CloudStack.ClusterAPIController.ImageDigest = tt.newManager
			})

			g := NewWithT(t)
			g.Expect(provider.UpgradeNeeded(context.Background(), clusterSpec, newClusterSpec)).To(Equal(tt.want))
		})
	}
}
