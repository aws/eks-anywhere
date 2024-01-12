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
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
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

func TestProviderGenerateCAPISpecForCreate(t *testing.T) {
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
			testName:          "main with rollout strategies",
			clusterconfigFile: "cluster_main_with_rollout_strategy.yaml",
			wantCPFile:        "testdata/expected_results_main_rollout_strategy_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_rollout_strategy_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)

			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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
			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForCreateWithAutoscalingConfiguration(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	wng := &clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]
	ca := &v1alpha1.AutoScalingConfiguration{
		MaxCount: 5,
		MinCount: 3,
	}
	wng.AutoScalingConfiguration = ca
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_autoscaling_md.yaml")
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
	validator.EXPECT().ValidateSecretsUnchanged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf(changedSecretMsg))

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestProviderGenerateCAPISpecForCreateWithAffinity(t *testing.T) {
	clusterSpecManifest := "cluster_affinity.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

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

func TestProviderGenerateCAPISpecForCreateWithZoneIdAndNetworkId(t *testing.T) {
	clusterSpecManifest := "cluster_main.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	datacenterConfig.Spec.AvailabilityZones[0].Zone = v1alpha1.CloudStackZone{
		Id: "zoneId",
		Network: v1alpha1.CloudStackResourceIdentifier{
			Id: "networkId",
		},
	}
	clusterSpec.CloudStackDatacenter = datacenterConfig
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_resourceids_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

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
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

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

func TestProviderGenerateCAPISpecForCreateWithMirrorConfigInsecureSkipVerify(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = test.RegistryMirrorInsecureSkipVerifyEnabled()
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_with_insecure_skip_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_with_insecure_skip_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMirrorConfigInsecureSkipVerifyAndCert(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = test.RegistryMirrorInsecureSkipVerifyEnabledAndCACert()
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_with_insecure_skip_and_cert_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_with_insecure_skip_and_cert_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithProxyConfig(t *testing.T) {
	clusterSpecManifest := "cluster_minimal_proxy.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)

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
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main_multiple_worker_node_groups.yaml")

	datacenterConfig := givenDatacenterConfig(t, "cluster_main_multiple_worker_node_groups.yaml")
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VersionsBundles["1.21"].CloudStack.Version = cloudStackProviderVersion
	setupContext(t)

	result := provider.Version(clusterSpec)
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
	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
	}{
		{
			testName: "correct Overrides layer",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].CloudStack = releasev1alpha1.CloudStackBundle{
					Version: "v0.1.0",
					ClusterAPIController: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-cloudstack/release/manager:v0.1.0",
					},
					KubeRbacProxy: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
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
			bundle := tt.clusterSpec.RootVersionsBundle()
			wantManifests := []releasev1alpha1.Manifest{
				bundle.CloudStack.Components,
				bundle.CloudStack.Metadata,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

func TestGetDatacenterConfig(t *testing.T) {
	provider := givenProvider(t)

	providerConfig := provider.DatacenterConfig(givenClusterSpec(t, testClusterConfigMainFilename))
	if providerConfig.Kind() != "CloudStackDatacenterConfig" {
		t.Fatalf("Unexpected error DatacenterConfig: kind field not found: %s", providerConfig.Kind())
	}
}

func TestProviderDeleteResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.ManagementCluster = &types.Cluster{
		KubeconfigFile: "testKubeConfig",
	}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	for _, mc := range machineConfigs {
		kubectl.EXPECT().DeleteEksaCloudStackMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace)
	}
	kubectl.EXPECT().DeleteEksaCloudStackDatacenterConfig(ctx, provider.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, provider.datacenterConfig.Namespace)

	err := provider.DeleteResources(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestChangeDiffNoChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	assert.Nil(t, provider.ChangeDiff(clusterSpec, clusterSpec))
}

func TestChangeDiffWithChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].CloudStack.Version = "v0.2.0"
	})
	newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].CloudStack.Version = "v0.1.0"
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
			setupContext(t)
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
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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

func TestProviderGenerateCAPISpecForUpgradeIncompleteClusterSpec(t *testing.T) {
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.CloudStackDatacenter = nil
	templateBuilder := NewTemplateBuilder(time.Now)
	if _, err := templateBuilder.GenerateCAPISpecControlPlane(clusterSpec); err == nil {
		t.Fatalf("Expected error for incomplete cluster spec, but no error occurred")
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
			testName:          "main with rollout strategies",
			clusterconfigFile: "cluster_main_with_rollout_strategy.yaml",
			wantCPFile:        "testdata/expected_results_main_rollout_strategy_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_rollout_strategy_md.yaml",
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
			setupContext(t)
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
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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
	setupContext(t)
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
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineGroupRefs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
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
					Name: "test-control-plane-template-1234567890000",
				},
			},
		},
	}
	oldMD := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-md-0-1234567890000",
					},
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-template-1234567890000",
						},
					},
				},
			},
		},
	}
	etcdadmCluster := &etcdv1.EtcdadmCluster{
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-etcd-template-1234567890000",
			},
		},
	}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	validator := givenWildcardValidator(mockCtrl, clusterSpec)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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

	wnMachineConfig := machineConfigs[workerNodeMachineConfigName]

	newClusterSpec := clusterSpec.DeepCopy()
	newWorkersMachineConfigName := "new-test-wn"
	newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = newWorkersMachineConfigName
	newWorkerMachineConfig := wnMachineConfig.DeepCopy()
	newWorkerMachineConfig.Name = newWorkersMachineConfigName
	newClusterSpec.CloudStackMachineConfigs[newWorkersMachineConfigName] = newWorkerMachineConfig

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(datacenterConfig, nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[controlPlaneMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[workerNodeMachineConfigName], nil)
	kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[etcdMachineConfigName], nil)
	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldCP, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldMD, nil).Times(2)
	kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadmCluster, nil)

	provider.templateBuilder.now = test.NewFakeNow

	_, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, newClusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
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
			setupContext(t)
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
			newConfig := v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(1), MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "CloudStackMachineConfig"}, Name: "md-2"}
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
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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

func TestProviderGenerateCAPISpecForUpgradeWorkerNodeGroupsKubernetesVersion(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantMDFile        string
		wantCPFile        string
	}{
		{
			testName:          "adding a worker node group",
			clusterconfigFile: "cluster_main_worker_node_group_kubernetes_version.yaml",
			wantMDFile:        "testdata/expected_results_main_md_worker_kubernetes_version.yaml",
			wantCPFile:        "testdata/expected_results_main_cp_kubernetes_version.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
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
			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup2MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackDatacenter, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[1].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().GetEksaCloudStackMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(cloudstackMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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

func TestClusterUpgradeNeededNoChanges(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
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
	provider := newProviderWithKubectl(t, dcConfig, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if specChanged {
		t.Fatalf("expected no spec change to be detected")
	}
}

func TestClusterNeedsNewWorkloadTemplateFalse(t *testing.T) {
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfig := givenMachineConfigs(t, testClusterConfigMainFilename)[cc.MachineConfigRefs()[0].Name]
	wng := &clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]

	assert.False(t, NeedsNewWorkloadTemplate(clusterSpec, clusterSpec, dcConfig, dcConfig, machineConfig, machineConfig, wng, wng, test.NewNullLogger()), "expected no spec change to be detected")
}

func TestClusterUpgradeNeededDatacenterConfigChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cc := givenClusterConfig(t, testClusterConfigMainFilename)
	fillClusterSpecWithClusterConfig(clusterSpec, cc)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	newClusterSpec := clusterSpec.DeepCopy()
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	shinyModifiedDcConfig := dcConfig.DeepCopy()
	shinyModifiedDcConfig.Spec.AvailabilityZones[0].ManagementApiEndpoint = "shiny-new-api-endpoint"
	newClusterSpec.CloudStackDatacenter = shinyModifiedDcConfig

	provider := newProviderWithKubectl(t, nil, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(givenDatacenterConfig(t, testClusterConfigMainFilename), nil)

	specChanged, err := provider.UpgradeNeeded(ctx, newClusterSpec, clusterSpec, cluster)
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
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
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
	provider := newProviderWithKubectl(t, dcConfig, cc, kubectl, nil)
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
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
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
				modifiedMachineConfig.Spec.DiskOffering = (*machineConfigsMap["test"].Spec.DiskOffering).DeepCopy()
				modifiedMachineConfig.Spec.DiskOffering.Name = "shiny-new-diskoffering"
				return modifiedMachineConfig, nil
			}
			return machineConfigsMap[cloudstackMachineConfigName], nil
		})

	provider := newProviderWithKubectl(t, dcConfig, cc, kubectl, nil)
	kubectl.EXPECT().GetEksaCloudStackDatacenterConfig(ctx, cc.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestNeedNewMachineTemplate(t *testing.T) {
	for _, tc := range []struct {
		Name                string
		ConfigureDatacenter func(old, nw *v1alpha1.CloudStackDatacenterConfig)
		ConfigureMachines   func(old, nw *v1alpha1.CloudStackMachineConfig)
		Expect              bool
	}{
		{
			Name: "Equivalent",
		},
		{
			// We can't retrieve the ManagementApiEndpoint for an availability zone from the
			// cloudstackv1.CloudStackCluster resource so a difference should be ignored.
			//
			// The criteria for changing this test and the context under which it was written
			// is unclear.
			Name: "AvailabilityZones_MissingManagementAPIEndpoint",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				nw.Spec.AvailabilityZones[0].ManagementApiEndpoint = ""
			},
		},
		{
			Name: "Datacenter_AvailabilityZones_Add",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				az := old.Spec.AvailabilityZones[0].DeepCopy()
				az.Name = "shinyNewAz"
				nw.Spec.AvailabilityZones = append(nw.Spec.AvailabilityZones, *az)
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_Removed",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{}
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_ChangeCredentialsRef",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				nw.Spec.AvailabilityZones[0].CredentialsRef = "new_credentials_ref"
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_ChangeZoneName",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				nw.Spec.AvailabilityZones[0].Zone.Name = "new_credentials_ref"
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_ChangeZoneNetwork",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				nw.Spec.AvailabilityZones[0].Zone.Network = v1alpha1.CloudStackResourceIdentifier{
					Name: "new_name",
				}
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_ChangeDomain",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				nw.Spec.AvailabilityZones[0].Domain = "shinyNewDomain"
			},
			Expect: true,
		},
		{
			Name: "Datacenter_AvailabilityZones_ChangeAccount",
			ConfigureDatacenter: func(old, nw *v1alpha1.CloudStackDatacenterConfig) {
				old.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
					{
						Name:           "name",
						CredentialsRef: "credentials_ref",
						Zone: v1alpha1.CloudStackZone{
							Name: "name",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "name",
							},
						},
						Domain:                "domain",
						Account:               "account",
						ManagementApiEndpoint: "management_api_endpoint",
					},
				}

				nw.Spec.AvailabilityZones = append(
					[]v1alpha1.CloudStackAvailabilityZone{},
					old.Spec.AvailabilityZones...,
				)

				nw.Spec.AvailabilityZones[0].Account = "new_account"
			},
			Expect: true,
		},
		{
			Name: "Machine_Symlinks_Add",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.Symlinks = map[string]string{
					"foo": "bar",
				}
				nw.Spec.Symlinks = map[string]string{
					"foo": "bar",
					"qux": "baz",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_Symlinks_Remove",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.Symlinks = map[string]string{
					"foo": "bar",
				}
				nw.Spec.Symlinks = map[string]string{}
			},
			Expect: true,
		},
		{
			Name: "Machine_Symlinks_Changed",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.Symlinks = map[string]string{
					"foo": "bar",
					"qux": "baz",
				}
				nw.Spec.Symlinks = map[string]string{
					"foo": "bar_changed",
					"qux": "baz_changed",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_NameChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Name: "name",
					},
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.Name = "name_changed"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_IDChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.Id = "id_changed"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_SizeChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
					CustomSize: 1,
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.CustomSize = 2
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_MountPathChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
					MountPath: "mount_path",
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.MountPath = "new_mount_path"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_DeviceChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
					Device: "device",
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.Device = "new_device_path"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_FilesystemChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
					Filesystem: "filesystem",
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.Filesystem = "new_filesystem"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_LabelChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
						Id: "id",
					},
					Label: "label",
				}
				nw.Spec.DiskOffering = old.Spec.DiskOffering.DeepCopy()
				nw.Spec.DiskOffering.Label = "new_label"
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_ToNil",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					MountPath:  "test",
					Device:     "test",
					Filesystem: "test",
				}
				nw.Spec.DiskOffering = nil
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_ToZeroValue",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
					MountPath:  "test",
					Device:     "test",
					Filesystem: "test",
				}
				nw.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{}
			},
			Expect: true,
		},
		{
			Name: "Machine_DiskOffering_Nil",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.DiskOffering = nil
				nw.Spec.DiskOffering = nil
			},
			Expect: false,
		},
		{
			Name: "Machine_ComputeOffering_NewID",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{}
				nw.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Id: "test",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_ComputeOffering_NewName",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{}
				nw.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Name: "test",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_ComputeOffering_IDChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Id: "test",
				}
				nw.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Id: "changed",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_ComputeOffering_NameChanged",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Name: "test",
				}
				nw.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Name: "changed",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_ComputeOffering_ToZeroValue",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
					Id: "test",
				}
				nw.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{}
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_Add",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{}

				nw.Spec.UserCustomDetails = map[string]string{
					"foo": "bar",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_Remove",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{
					"foo": "bar",
				}

				nw.Spec.UserCustomDetails = map[string]string{}
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_ToNil",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{
					"foo": "bar",
				}

				nw.Spec.UserCustomDetails = nil
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_Replace",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{
					"foo": "bar",
				}

				nw.Spec.UserCustomDetails = map[string]string{
					"qux": "baz",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_ReplaceEmptyValue",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{
					"foo": "",
					"qux": "baz",
				}

				nw.Spec.UserCustomDetails = map[string]string{
					"bar": "",
					"qux": "baz",
				}
			},
			Expect: true,
		},
		{
			Name: "Machine_UserCustomDetails_ReplaceEmptyValue",
			ConfigureMachines: func(old, nw *v1alpha1.CloudStackMachineConfig) {
				old.Spec.UserCustomDetails = map[string]string{
					"foo": "",
					"qux": "baz",
				}

				nw.Spec.UserCustomDetails = map[string]string{
					"bar": "",
					"qux": "baz",
				}
			},
			Expect: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			oldDatacenter := givenDatacenterConfig(t, testClusterConfigMainFilename)
			newDatacenter := oldDatacenter.DeepCopy()

			if tc.ConfigureDatacenter != nil {
				tc.ConfigureDatacenter(oldDatacenter, newDatacenter)
			}

			oldMachines := givenMachineConfigs(t, testClusterConfigMainFilename)
			oldMachine, newMachine := oldMachines["test"], oldMachines["test"].DeepCopy()

			if tc.ConfigureMachines != nil {
				tc.ConfigureMachines(oldMachine, newMachine)
			}

			result := NeedNewMachineTemplate(
				oldDatacenter,
				newDatacenter,
				oldMachine,
				newMachine,
				test.NewNullLogger(),
			)

			if result != tc.Expect {
				t.Fatalf("Expected: %v; Received: %v", tc.Expect, result)
			}
		})
	}
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

func TestNeedsNewWorkloadTemplateK8sVersion(t *testing.T) {
	oldSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newK8sSpec := oldSpec.DeepCopy()
	newK8sSpec.Cluster.Spec.KubernetesVersion = "1.25"
	wng := &oldSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]
	assert.True(t, NeedsNewWorkloadTemplate(oldSpec, newK8sSpec, nil, nil, nil, nil, wng, wng, test.NewNullLogger()))
}

func TestNeedsNewWorkloadTemplateBundleNumber(t *testing.T) {
	oldSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newK8sSpec := oldSpec.DeepCopy()
	newK8sSpec.Bundles.Spec.Number = 10000
	assert.True(t, NeedsNewWorkloadTemplate(oldSpec, newK8sSpec, nil, nil, nil, nil, nil, nil, test.NewNullLogger()))
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

func TestProviderGenerateCAPISpecForUpgradeEtcdEncryption(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "etcd-encryption",
			clusterconfigFile: "cluster_etcd_encryption.yaml",
			wantCPFile:        "testdata/expected_results_encryption_config_cp.yaml",
			wantMDFile:        "testdata/expected_results_minimal_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
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
			validator := givenWildcardValidator(mockCtrl, clusterSpec)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, validator)
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
