package cloudstack

import (
	"context"
	"embed"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata
var configFS embed.FS

const (
	expectedCloudStackName              = "cloudstack"
	cloudStackCloudConfigWithInvalidUrl = "W0dsb2JhbF0KdmVyaWZ5LXNzbCA9IGZhbHNlCmFwaS1rZXkgPSB0ZXN0LWtleTEKc2VjcmV0LWtleSA9IHRlc3Qtc2VjcmV0MQphcGktdXJsID0geHh4Cg=="
	validCloudStackCloudConfig          = "W0dsb2JhbF0KYXBpLWtleSAgICA9IGZha2UtYXBpLWtleQpzZWNyZXQta2V5ID0gZmFrZS1zZWNy\nZXQta2V5CmFwaS11cmwgICAgPSBodHRwOi8vMTAuMTEuMC4yOjgwODAvY2xpZW50L2FwaQoKW0ds\nb2JhbDJdCmFwaS1rZXkgICAgPSBmYWtlLWFwaS1rZXkKc2VjcmV0LWtleSA9IGZha2Utc2VjcmV0\nLWtleQphcGktdXJsICAgID0gaHR0cDovLzEwLjEyLjAuMjo4MDgwL2NsaWVudC9hcGkKCg=="
	defaultCloudStackCloudConfigPath    = "testdata/cloudstack_config_valid.ini"
)

var notFoundError = apierrors.NewNotFound(schema.GroupResource{}, "")

var expectedSecret = &v1.Secret{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: v1.SchemeGroupVersion.Version,
	},
	ObjectMeta: metav1.ObjectMeta{
		Namespace: constants.EksaSystemNamespace,
		Name:      "global",
	},
	Data: map[string][]byte{
		"api-url":    []byte("http://127.16.0.1:8080/client/api"),
		"api-key":    []byte("test-key1"),
		"secret-key": []byte("test-secret1"),
		"verify-ssl": []byte("false"),
	},
}

func givenClusterConfig(t *testing.T, fileName string) *v1alpha1.Cluster {
	return givenClusterSpec(t, fileName).Cluster
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

// TODO: Validate against validator operations instead of using wildcard, now that it's mocked. https://github.com/aws/eks-anywhere/issues/3944
func givenWildcardValidator(mockCtrl *gomock.Controller, clusterSpec *cluster.Spec) *MockProviderValidator {
	validator := NewMockProviderValidator(mockCtrl)
	validator.EXPECT().ValidateClusterMachineConfigs(gomock.Any(), gomock.Any()).SetArg(1, *clusterSpec).AnyTimes()
	validator.EXPECT().ValidateCloudStackDatacenterConfig(gomock.Any(), clusterSpec.CloudStackDatacenter).AnyTimes()
	validator.EXPECT().ValidateControlPlaneEndpointUniqueness(gomock.Any()).AnyTimes()
	return validator
}

func fillClusterSpecWithClusterConfig(spec *cluster.Spec, clusterConfig *v1alpha1.Cluster) {
	spec.Cluster = clusterConfig
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.CloudStackDatacenterConfig {
	datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	datacenterConfig.SetDefaults()
	return datacenterConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.CloudStackMachineConfig {
	config, err := cluster.ParseConfigFromFile(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file: %v", err)
	}

	return config.CloudStackMachineConfigs
}

func givenProvider(t *testing.T) *cloudstackProvider {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterConfig := clusterSpec.Cluster
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		clusterConfig,
		nil,
		validator,
	)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	return provider
}

func givenManagementComponents() *cluster.ManagementComponents {
	return &cluster.ManagementComponents{
		CloudStack: releasev1alpha1.CloudStackBundle{
			Version: "v0.1.0",
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-cloudstack/v0.1.0/infrastructure-components-development.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-cloudstack/v0.1.0/metadata.yaml",
			},
		},
	}
}

func saveContext(t *testing.T, configPath string) {
	cloudStackCloudConfig, err := configFS.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read cloudstack cloud-config file from %s: %v", configPath, err)
	}
	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, base64.StdEncoding.EncodeToString(cloudStackCloudConfig))
	t.Setenv(decoder.CloudStackCloudConfigB64SecretKey, os.Getenv(decoder.EksacloudStackCloudConfigB64SecretKey))
}

func setupContext(t *testing.T) {
	saveContext(t, defaultCloudStackCloudConfigPath)
}

type providerTest struct {
	*WithT
	t                                  *testing.T
	ctx                                context.Context
	managementCluster, workloadCluster *types.Cluster
	provider                           *cloudstackProvider
	cluster                            *v1alpha1.Cluster
	clusterSpec                        *cluster.Spec
	datacenterConfig                   *v1alpha1.CloudStackDatacenterConfig
	machineConfigs                     map[string]*v1alpha1.CloudStackMachineConfig
	kubectl                            *mocks.MockProviderKubectlClient
	validator                          *MockProviderValidator
}

func newProviderTest(t *testing.T) *providerTest {
	setupContext(t)
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	spec := givenClusterSpec(t, testClusterConfigMainFilename)
	p := &providerTest{
		t:     t,
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
		cluster:          spec.Cluster,
		clusterSpec:      spec,
		datacenterConfig: spec.CloudStackDatacenter,
		machineConfigs:   spec.CloudStackMachineConfigs,
		kubectl:          kubectl,
		validator:        givenWildcardValidator(ctrl, spec),
	}
	p.buildNewProvider()
	return p
}

func (tt *providerTest) buildNewProvider() {
	tt.provider = newProvider(
		tt.t,
		tt.clusterSpec.CloudStackDatacenter,
		tt.clusterSpec.Cluster,
		tt.kubectl,
		tt.validator,
	)
}

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterConfig := clusterSpec.Cluster
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		clusterConfig,
		kubectl,
		validator,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
}

func newProviderWithKubectl(t *testing.T, datacenterConfig *v1alpha1.CloudStackDatacenterConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, validator ProviderValidator) *cloudstackProvider {
	return newProvider(
		t,
		datacenterConfig,
		clusterConfig,
		kubectl,
		validator,
	)
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.CloudStackDatacenterConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, validator ProviderValidator) *cloudstackProvider {
	_, writer := test.NewWriter(t)
	return NewProvider(datacenterConfig, clusterConfig, kubectl, validator, writer, test.FakeNow, test.NewNullLogger())
}

func TestProviderSetupAndValidateCreateClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	tt.Expect(err.Error()).To(Equal("validating environment variables: CloudStack instance global's managementApiEndpoint xxx is invalid: CloudStack managementApiEndpoint is invalid: #{err}"))
}

func TestProviderSetupAndValidateUpgradeClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	tt.Expect(err.Error()).To(Equal("validating environment variables: CloudStack instance global's managementApiEndpoint xxx is invalid: CloudStack managementApiEndpoint is invalid: #{err}"))
}

func TestProviderSetupAndValidateDeleteClusterFailureOnInvalidUrl(t *testing.T) {
	tt := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, cloudStackCloudConfigWithInvalidUrl)
	err := provider.SetupAndValidateDeleteCluster(ctx, cluster, nil)
	tt.Expect(err.Error()).To(Equal("validating environment variables: CloudStack instance global's managementApiEndpoint xxx is invalid: CloudStack managementApiEndpoint is invalid: #{err}"))
}

func TestProviderSetupAndValidateUpgradeClusterFailureOnGetSecretFailure(t *testing.T) {
	tt := NewWithT(t)
	clusterSpecManifest := "cluster_main.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	secretFailureMsg := "getting secret for profile global: test-error"
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(apierrors.NewBadRequest(secretFailureMsg))

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	tt.Expect(err.Error()).To(Equal("validating secrets unchanged: getting secret for profile global: test-error"))
}

func TestProviderSetupAndValidateUpgradeClusterSuccessOnSecretNotFound(t *testing.T) {
	tt := NewWithT(t)
	clusterSpecManifest := "cluster_main.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	tt.Expect(err).To(BeNil())
}

func TestProviderSetupAndValidateUpgradeClusterFailureOnSecretChanged(t *testing.T) {
	tt := NewWithT(t)
	clusterSpecManifest := "cluster_main.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	modifiedSecret := expectedSecret.DeepCopy()
	modifiedSecret.Data["api-key"] = []byte("updated-api-key")
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	changedSecretMsg := "profile global is different from secret"
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(changedSecretMsg))

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
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
	clusterSpecManifest := "cluster_minimal_proxy.yaml"
	provider := givenProvider(t)
	provider.clusterConfig = givenClusterConfig(t, clusterSpecManifest)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	bootstrapClusterOps, err := provider.BootstrapClusterOpts(clusterSpec)
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
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kc.kubeconfig",
	}

	setupContext(t)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	provider.validator = givenWildcardValidator(mockCtrl, clusterSpec)

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDatacenterConfig(ctx, datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.CloudStackDatacenterConfig{}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfMachineExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kc.kubeconfig",
	}

	setupContext(t)

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	provider.validator = givenWildcardValidator(mockCtrl, clusterSpec)

	idx := 0
	var existingMachine string
	for _, config := range newMachineConfigs {
		if idx == 0 {
			kubectl.EXPECT().SearchCloudStackMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{config}, nil)
			existingMachine = config.Name
		} else {
			kubectl.EXPECT().SearchCloudStackMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil).MaxTimes(1)
		}
		idx++
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("CloudStackMachineConfig %s already exists", existingMachine), err)
}

func TestSetupAndValidateSelfManagedClusterSkipMachineNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kc.kubeconfig",
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
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kc.kubeconfig",
	}

	setupContext(t)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	provider.validator = givenWildcardValidator(mockCtrl, clusterSpec)

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchCloudStackMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchCloudStackDatacenterConfig(ctx, datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.CloudStackDatacenterConfig{datacenterConfig}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("CloudStackDatacenter %s already exists", datacenterConfig.Name), err)
}

func TestSetupAndValidateSelfManagedClusterSkipDatacenterNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kc.kubeconfig",
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
	setupContext(t)

	err := provider.SetupAndValidateDeleteCluster(ctx, nil, nil)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestCleanupProviderInfrastructure(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	setupContext(t)

	err := provider.CleanupProviderInfrastructure(ctx)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestVersion(t *testing.T) {
	cloudStackProviderVersion := "v4.14.1"
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	managementComponents.CloudStack.Version = cloudStackProviderVersion
	setupContext(t)

	result := provider.Version(managementComponents)
	if result != cloudStackProviderVersion {
		t.Fatalf("Unexpected version expected <%s> actual=<%s>", cloudStackProviderVersion, result)
	}
}

func TestPreCAPIInstallOnBootstrap(t *testing.T) {
	tests := []struct {
		testName                string
		configPath              string
		expectedSecretsYamlPath string
	}{
		{
			testName:                "valid single profile",
			configPath:              defaultCloudStackCloudConfigPath,
			expectedSecretsYamlPath: "testdata/expected_secrets_single.yaml",
		},
	}

	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	for _, test := range tests {
		saveContext(t, test.configPath)
		expectedSecretsYaml, err := configFS.ReadFile(test.expectedSecretsYamlPath)
		if err != nil {
			t.Fatalf("Failed to read embed eksd release: %s", err)
		}

		kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, apierrors.NewNotFound(schema.GroupResource{}, ""))
		kubectl.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), expectedSecretsYaml)
		_ = provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

		if err := provider.PreCAPIInstallOnBootstrap(ctx, cluster, clusterSpec); err != nil {
			t.Fatalf("provider.PreCAPIInstallOnBootstrap() err = %v, want err = nil", err)
		}
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey did not get generated for etcd machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyAllMachineConfigs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
	if clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
	if clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey not generated for etcd machines")
	}
	if clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and worker machines")
	}
	if clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and etcd machines")
	}
}

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	p := givenProvider(t)

	managementComponents := givenManagementComponents()
	wantInfraBundle := &types.InfrastructureBundle{
		FolderName: "infrastructure-cloudstack/v0.1.0/",
		Manifests: []releasev1alpha1.Manifest{
			managementComponents.CloudStack.Components,
			managementComponents.CloudStack.Metadata,
		},
	}

	infraBundle := p.GetInfrastructureBundle(managementComponents)
	assert.Equal(t, wantInfraBundle.FolderName, infraBundle.FolderName, "Incorrect folder name")
	assert.Equal(t, len(infraBundle.Manifests), 2, "Wrong number of files in the infrastructure bundle")
	assert.Equal(t, wantInfraBundle.Manifests, infraBundle.Manifests, "Incorrect manifests")
}

func TestGetDatacenterConfig(t *testing.T) {
	provider := givenProvider(t)

	providerConfig := provider.DatacenterConfig(givenClusterSpec(t, testClusterConfigMainFilename))
	if providerConfig.Kind() != "CloudStackDatacenterConfig" {
		t.Fatalf("Unexpected error DatacenterConfig: kind field not found: %s", providerConfig.Kind())
	}
}

func TestChangeDiffNoChange(t *testing.T) {
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	assert.Nil(t, provider.ChangeDiff(managementComponents, managementComponents))
}

func TestChangeDiffWithChange(t *testing.T) {
	provider := givenProvider(t)
	managementComponents := givenManagementComponents()
	managementComponents.CloudStack.Version = "v0.1.0"
	newManagementComponents := givenManagementComponents()
	newManagementComponents.CloudStack.Version = "v0.2.0"

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "cloudstack",
		NewVersion:    "v0.2.0",
		OldVersion:    "v0.1.0",
	}

	assert.Equal(t, wantDiff, provider.ChangeDiff(managementComponents, newManagementComponents))
}

func TestSetupAndValidateUpgradeCluster(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{}
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, clusterSpec.CloudStackDatacenter, clusterSpec.Cluster,
		kubectl, validator)
	setupContext(t)

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterCPSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, clusterSpec.CloudStackDatacenter, clusterSpec.Cluster, kubectl, validator)
	clusterSpec.CloudStackMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	cluster := &types.Cluster{}

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterWorkerSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, clusterSpec.CloudStackDatacenter, clusterSpec.Cluster,
		kubectl, validator)
	clusterSpec.CloudStackMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterEtcdSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, clusterSpec.CloudStackDatacenter, clusterSpec.Cluster,
		kubectl, validator)
	clusterSpec.CloudStackMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsNameUniquenessSuccess(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	prevSpec := tt.clusterSpec.DeepCopy()
	prevSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "prev-test-cp"
	prevSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "prev-test-etcd"
	machineConfigs := tt.clusterSpec.CloudStackMachineConfigs
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.Name).Return(prevSpec.Cluster, nil)
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().SearchCloudStackMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{}, nil).AnyTimes()
	}

	err := tt.provider.validateMachineConfigsNameUniqueness(tt.ctx, cluster, tt.clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsNameUniquenessError(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	prevSpec := tt.clusterSpec.DeepCopy()
	prevSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "prev-test-cp"
	prevSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "prev-test-etcd"
	machineConfigs := tt.clusterSpec.CloudStackMachineConfigs
	dummyMachineConfig := &v1alpha1.CloudStackMachineConfig{
		Spec: v1alpha1.CloudStackMachineConfigSpec{Users: []v1alpha1.UserConfiguration{{Name: "capc"}}},
	}

	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.Name).Return(prevSpec.Cluster, nil)
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().SearchCloudStackMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.CloudStackMachineConfig{dummyMachineConfig}, nil).AnyTimes()
	}
	err := tt.provider.validateMachineConfigsNameUniqueness(tt.ctx, cluster, tt.clusterSpec)
	thenErrorExpected(t, fmt.Sprintf("machineconfig %s already exists", tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name), err)
}

func TestInstallCustomProviderComponentsKubeVipEnabled(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, dcConfig, cc, kubectl, nil)
	kubeConfigFile := "test"

	oldCloudstackKubeVipDisabledVal := os.Getenv(features.CloudStackKubeVipDisabledEnvVar)
	os.Unsetenv(features.CloudStackKubeVipDisabledEnvVar)
	defer os.Setenv(features.CloudStackKubeVipDisabledEnvVar, oldCloudstackKubeVipDisabledVal)
	kubectl.EXPECT().SetEksaControllerEnvVar(ctx, features.CloudStackKubeVipDisabledEnvVar, "false", kubeConfigFile).Return(nil)
	if err := provider.InstallCustomProviderComponents(ctx, kubeConfigFile); err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestProviderUpdateSecrets(t *testing.T) {
	tests := []struct {
		testName                string
		configPath              string
		expectedSecretsYamlPath string
		getSecretError          error
		applyError              error
		wantErr                 bool
	}{
		{
			testName:                "valid single profile",
			configPath:              defaultCloudStackCloudConfigPath,
			expectedSecretsYamlPath: "testdata/expected_secrets_single.yaml",
			getSecretError:          notFoundError,
			applyError:              nil,
			wantErr:                 false,
		},
		{
			testName:                "valid multiple profiles",
			configPath:              "testdata/cloudstack_config_multiple_profiles.ini",
			expectedSecretsYamlPath: "testdata/expected_secrets_multiple.yaml",
			getSecretError:          notFoundError,
			applyError:              nil,
			wantErr:                 false,
		},
		{
			testName:                "secret already present",
			configPath:              defaultCloudStackCloudConfigPath,
			expectedSecretsYamlPath: "testdata/expected_secrets_single.yaml",
			getSecretError:          nil,
			applyError:              nil,
			wantErr:                 false,
		},
		{
			testName:                "valid single profile",
			configPath:              defaultCloudStackCloudConfigPath,
			expectedSecretsYamlPath: "testdata/expected_secrets_single.yaml",
			getSecretError:          notFoundError,
			applyError:              errors.New("exception"),
			wantErr:                 true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tt := NewWithT(t)
			mockCtrl := gomock.NewController(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{}
			clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

			datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}
			saveContext(t, test.configPath)
			expectedSecretsYaml, err := configFS.ReadFile(test.expectedSecretsYamlPath)
			if err != nil {
				t.Fatalf("Failed to read embed eksd release: %s", err)
			}

			kubectl.EXPECT().GetSecretFromNamespace(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, test.getSecretError)
			if test.getSecretError != nil {
				kubectl.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), expectedSecretsYaml).Return(test.applyError)
			}

			if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
				t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
			}

			err = provider.UpdateSecrets(ctx, cluster, nil)
			if test.wantErr {
				tt.Expect(err).NotTo(BeNil())
			} else {
				tt.Expect(err).To(BeNil())
			}
		})
	}
}

func TestValidateNewSpecMachineConfigImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	newAffinityGroupIds := [1]string{"different"}
	newClusterSpec.CloudStackMachineConfigs[workerMachineConfigName].Spec.AffinityGroupIds = newAffinityGroupIds[:]

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.CloudStackDatacenter, nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.CloudStackMachineConfigs[workerMachineConfigName], nil).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "field is immutable")
}

func TestValidateNewSpecMachineConfigNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	apierr := apierrors.NewNotFound(schema.GroupResource{}, workerMachineConfigName)

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.CloudStackDatacenter, nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(nil, apierr).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "not found")
}
