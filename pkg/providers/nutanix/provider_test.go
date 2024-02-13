package nutanix

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	mockCrypto "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	mocknutanix "github.com/aws/eks-anywhere/pkg/providers/nutanix/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata/eksa-cluster.json
var nutanixClusterConfigSpecJSON string

//go:embed testdata/datacenterConfig.json
var nutanixDatacenterConfigSpecJSON string

//go:embed testdata/machineConfig.json
var nutanixMachineConfigSpecJSON string

//go:embed testdata/machineDeployment.json
var nutanixMachineDeploymentSpecJSON string

func thenErrorExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if expected != actual {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

func testDefaultNutanixProvider(t *testing.T) *Provider {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}

	mockWriter := filewritermocks.NewMockFileWriter(ctrl)

	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)
	return provider
}

func testNutanixProvider(t *testing.T, nutanixClient Client, kubectl *executables.Kubectl, certValidator crypto.TlsValidator, httpClient *http.Client, writer filewriter.FileWriter) *Provider {
	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpec), clusterConf)
	require.NoError(t, err)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConf)
	require.NoError(t, err)

	workerConfs := map[string]*anywherev1.NutanixMachineConfig{
		"eksa-unit-test": machineConf,
	}

	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")

	clientCache := &ClientCache{
		clients: make(map[string]Client),
	}
	clientCache.clients[dcConf.Name] = nutanixClient

	ctrl := gomock.NewController(t)
	mockIPValidator := mocknutanix.NewMockIPValidator(ctrl)
	mockIPValidator.EXPECT().ValidateControlPlaneIPUniqueness(gomock.Any()).Return(nil).AnyTimes()
	provider := NewProvider(dcConf, workerConfs, clusterConf, kubectl, writer, clientCache, mockIPValidator, certValidator, httpClient, time.Now, false)
	require.NotNil(t, provider)
	return provider
}

func givenManagementComponents() *cluster.ManagementComponents {
	return &cluster.ManagementComponents{
		Nutanix: releasev1alpha1.NutanixBundle{
			Version: "1.0.0",
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-nutanix/v1.0.0/infrastructure-components-development.yaml",
			},
			ClusterTemplate: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-nutanix/v1.0.0/cluster-template.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-nutanix/v1.0.0/metadata.yaml",
			},
		},
	}
}

func testNutanixProviderWithClusterSpec(t *testing.T, nutanixClient Client, kubectl *executables.Kubectl, certValidator crypto.TlsValidator, httpClient *http.Client, writer filewriter.FileWriter, clusterSpec *cluster.Spec) *Provider {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")

	clientCache := &ClientCache{
		clients: make(map[string]Client),
	}
	clientCache.clients[clusterSpec.NutanixDatacenter.Name] = nutanixClient
	ctrl := gomock.NewController(t)
	mockIPValidator := mocknutanix.NewMockIPValidator(ctrl)
	mockIPValidator.EXPECT().ValidateControlPlaneIPUniqueness(gomock.Any()).Return(nil).AnyTimes()
	provider := NewProvider(clusterSpec.NutanixDatacenter, clusterSpec.NutanixMachineConfigs, clusterSpec.Cluster, kubectl, writer, clientCache, mockIPValidator, certValidator, httpClient, time.Now, false)
	require.NotNil(t, provider)
	return provider
}

func TestNutanixProviderBootstrapClusterOpts(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	opts, err := provider.BootstrapClusterOpts(clusterSpec)
	assert.NoError(t, err)
	assert.Nil(t, opts)
}

func TestNutanixProviderBootstrapSetup(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.BootstrapSetup(context.Background(), provider.clusterConfig, &types.Cluster{Name: "eksa-unit-test"})
	assert.NoError(t, err)
}

func TestNutanixProviderPostBootstrapSetup(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.PostBootstrapSetup(context.Background(), provider.clusterConfig, &types.Cluster{Name: "eksa-unit-test"})
	assert.NoError(t, err)
}

func TestNutanixProviderPostBootstrapDeleteForUpgrade(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.PostBootstrapDeleteForUpgrade(context.Background(), &types.Cluster{Name: "eksa-unit-test"})
	assert.NoError(t, err)
}

func TestNutanixProviderPostBootstrapSetupUpgrade(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.PostBootstrapSetupUpgrade(context.Background(), provider.clusterConfig, &types.Cluster{Name: "eksa-unit-test"})
	assert.NoError(t, err)
}

func TestNutanixProviderPostWorkloadInit(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.PostWorkloadInit(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderName(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	name := provider.Name()
	assert.Equal(t, "nutanix", name)
}

func TestNutanixProviderDatacenterResourceType(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	resource := provider.DatacenterResourceType()
	assert.Equal(t, "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", resource)
}

func TestNutanixProviderMachineResourceType(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	resource := provider.MachineResourceType()
	assert.Equal(t, "nutanixmachineconfigs.anywhere.eks.amazonaws.com", resource)
}

func TestNutanixProviderDeleteResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().Execute(gomock.Any(), "delete", []string{"nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "default", "--ignore-not-found=true"}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	executable.EXPECT().Execute(gomock.Any(), "delete", []string{"nutanixmachineconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "default", "--ignore-not-found=true"}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	kubectl := executables.NewKubectl(executable)

	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	clusterSpec.ManagementCluster = &types.Cluster{Name: "eksa-unit-test", KubeconfigFile: "testdata/kubeconfig.yaml"}
	err := provider.DeleteResources(context.Background(), clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderPostClusterDeleteValidate(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.PostClusterDeleteValidate(context.Background(), &types.Cluster{Name: "eksa-unit-test"})
	assert.NoError(t, err)
}

func TestNutanixProviderSetupAndValidateCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name            string
		clusterConfFile string
		expectErr       bool
		expectErrStr    string
	}{
		{
			name:            "valid cluster config",
			clusterConfFile: "testdata/eksa-cluster.yaml",
			expectErr:       false,
		},
		{
			name:            "valid cluster config with trust bundle",
			clusterConfFile: "testdata/cluster_nutanix_with_trust_bundle.yaml",
			expectErr:       false,
		},
		{
			name:            "valid cluster config with invalid trust bundle",
			clusterConfFile: "testdata/cluster_nutanix_with_invalid_trust_bundle.yaml",
			expectErr:       true,
			expectErrStr:    "failed to validate cluster spec: invalid cert",
		},
		{
			name:            "valid cluster config with invalid pe cluster name - same as pc name",
			clusterConfFile: "testdata/eksa-cluster-invalid-pe-cluster-pc.yaml",
			expectErr:       true,
			expectErrStr:    "failed to validate cluster spec: failed to validate machine config: failed to find cluster with name \"prism-central\": failed to find cluster by name \"prism-central\": <nil>",
		},
		{
			name:            "valid cluster config with invalid pe cluster name - non existent pe name",
			clusterConfFile: "testdata/eksa-cluster-invalid-pe-cluster-random-name.yaml",
			expectErr:       true,
			expectErrStr:    "failed to validate cluster spec: failed to validate machine config: failed to find cluster with name \"non-existent-cluster\": failed to find cluster by name \"non-existent-cluster\": <nil>",
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for cp",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_cp.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for md",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_md.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
	}

	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()
	clusters := &v3.ClusterListIntentResponse{
		Entities: []*v3.ClusterIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cda"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("4692a614-85e7-4abc-9bf3-8fb0f9d790bc"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-central"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("PRISM_CENTRAL")},
						},
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1abc"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster-2"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1xyz"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster-3"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
		},
	}
	mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil).AnyTimes()
	subnets := &v3.SubnetListIntentResponse{
		Entities: []*v3.SubnetIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				},
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet"),
				},
			},
		},
	}
	mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(subnets, nil).AnyTimes()
	images := &v3.ImageListIntentResponse{
		Entities: []*v3.ImageIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdc"),
				},
				Spec: &v3.Image{
					Name: utils.StringPtr("prism-image"),
				},
			},
		},
	}
	mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(images, nil).AnyTimes()
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockCertValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCertValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("invalid cert"))
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)
	assert.NotNil(t, provider)

	for _, tt := range tests {
		clusterSpec := test.NewFullClusterSpec(t, tt.clusterConfFile)
		err := provider.SetupAndValidateCreateCluster(context.Background(), clusterSpec)
		if tt.expectErr {
			assert.Error(t, err, tt.name)
			thenErrorExpected(t, tt.expectErrStr, err)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}

	sshKeyTests := []struct {
		name            string
		clusterConfFile string
		expectErr       bool
		performTest     func(t *testing.T, provider *Provider, clusterSpec *cluster.Spec) error
	}{
		{
			name:            "validate is ssh key gets generated for cp",
			clusterConfFile: "testdata/eksa-cluster-multiple-machineconfigs.yaml",
			expectErr:       false,
			performTest: func(t *testing.T, provider *Provider, clusterSpec *cluster.Spec) error {
				// Set the SSH Authorized Key to empty string
				controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
				clusterSpec.NutanixMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

				err := provider.SetupAndValidateCreateCluster(context.Background(), clusterSpec)
				if err != nil {
					return fmt.Errorf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
				}
				// Expect the SSH Authorized Key to be not empty
				if clusterSpec.NutanixMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
					return fmt.Errorf("sshAuthorizedKey has not changed for control plane machine")
				}
				return nil
			},
		},
	}

	for _, tc := range sshKeyTests {
		t.Run(tc.name, func(t *testing.T) {
			clusterSpec := test.NewFullClusterSpec(t, tc.clusterConfFile)
			// to avoid "because: there are no expected calls of the method "Write" for that receiver"
			// using test.NewWriter(t) instead of filewritermocks.NewMockFileWriter(ctrl)
			_, mockWriter := test.NewWriter(t)
			provider := testNutanixProviderWithClusterSpec(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter, clusterSpec)
			assert.NotNil(t, provider)

			err := tc.performTest(t, provider, clusterSpec)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("Test failed. %s", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Test failed. %s", err)
				}
			}
		})
	}
}

func TestNutanixProviderSetupAndValidateDeleteCluster(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	tests := []struct {
		name            string
		clusterConfFile string
		expectErr       bool
		expectErrStr    string
	}{
		{
			name:            "valid cluster config",
			clusterConfFile: "testdata/eksa-cluster.yaml",
			expectErr:       false,
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for cp",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_cp.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for md",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_md.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
	}

	for _, tt := range tests {
		clusterSpec := test.NewFullClusterSpec(t, tt.clusterConfFile)
		err := provider.SetupAndValidateDeleteCluster(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec)
		if tt.expectErr {
			assert.Error(t, err, tt.name)
			thenErrorExpected(t, tt.expectErrStr, err)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestNutanixProviderSetupAndValidateUpgradeCluster(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	tests := []struct {
		name            string
		clusterConfFile string
		expectErr       bool
		expectErrStr    string
	}{
		{
			name:            "valid cluster config",
			clusterConfFile: "testdata/eksa-cluster.yaml",
			expectErr:       false,
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for cp",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_cp.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
		{
			name:            "cluster config with unsupported upgrade strategy configuration for md",
			clusterConfFile: "testdata/cluster_nutanix_with_upgrade_strategy_md.yaml",
			expectErr:       true,
			expectErrStr:    "failed setup and validations: Upgrade rollout strategy customization is not supported for nutanix provider",
		},
	}

	for _, tt := range tests {
		clusterSpec := test.NewFullClusterSpec(t, tt.clusterConfFile)
		err := provider.SetupAndValidateUpgradeCluster(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec, clusterSpec)
		if tt.expectErr {
			assert.Error(t, err, tt.name)
			thenErrorExpected(t, tt.expectErrStr, err)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestNutanixProviderUpdateSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil).Times(2)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.UpdateSecrets(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)

	storedMarshal := jsonMarshal
	jsonMarshal = fakemarshal
	err = provider.UpdateSecrets(context.Background(), cluster, clusterSpec)
	assert.ErrorContains(t, err, "marshalling failed")
	restoremarshal(storedMarshal)

	clusterSpec.NutanixDatacenter.Spec.CredentialRef.Name = "capx-eksa-unit-test"
	err = provider.UpdateSecrets(context.Background(), cluster, clusterSpec)
	assert.ErrorContains(t, err, "NutanixDatacenterConfig CredentialRef name cannot be the same as the NutanixCluster CredentialRef name")
}

func TestNutanixProviderGenerateCAPISpecForCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil).Times(2)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, cpSpec)
	assert.NotNil(t, workerSpec)
}

func TestNutanixProviderGenerateCAPISpecForCreateWorkerVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil).Times(2)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-worker-version.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, cpSpec)
	assert.NotNil(t, workerSpec)
}

func TestNutanixProviderGenerateCAPISpecForCreate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, errors.New("test error"))
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	assert.EqualError(t, err, "updating Nutanix credentials: loading secrets object: executing apply: test error")
	assert.Nil(t, cpSpec)
	assert.Nil(t, workerSpec)
}

func TestNutanixProviderGenerateCAPISpecForUpgrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil).Times(2)
	executable.EXPECT().Execute(gomock.Any(), "get",
		"clusters.anywhere.eks.amazonaws.com", "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig", "testdata/kubeconfig.yaml", "--field-selector=metadata.name=eksa-unit-test").Return(*bytes.NewBufferString(nutanixClusterConfigSpecJSON), nil)
	executable.EXPECT().Execute(gomock.Any(), "get",
		"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test").Return(*bytes.NewBufferString(nutanixMachineConfigSpecJSON), nil).AnyTimes()
	executable.EXPECT().Execute(gomock.Any(), "get",
		"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test").Return(*bytes.NewBufferString(nutanixDatacenterConfigSpecJSON), nil)
	executable.EXPECT().Execute(gomock.Any(), "get",
		"machinedeployments.cluster.x-k8s.io", "eksa-unit-test-eksa-unit-test", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "eksa-system").Return(*bytes.NewBufferString(nutanixMachineDeploymentSpecJSON), nil).Times(2)
	executable.EXPECT().Execute(gomock.Any(), "get",
		"kubeadmcontrolplanes.controlplane.cluster.x-k8s.io", "eksa-unit-test", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "eksa-system").Return(*bytes.NewBufferString(nutanixMachineDeploymentSpecJSON), nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test", KubeconfigFile: "testdata/kubeconfig.yaml"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForUpgrade(context.Background(), cluster, cluster, clusterSpec, clusterSpec)
	assert.NoError(t, err)
	assert.NotEmpty(t, cpSpec)
	assert.NotEmpty(t, workerSpec)
}

func TestNutanixProviderGenerateCAPISpecForUpgrade_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, errors.New("test error"))
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test", KubeconfigFile: "testdata/kubeconfig.yaml"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForUpgrade(context.Background(), cluster, cluster, clusterSpec, clusterSpec)
	assert.EqualError(t, err, "updating Nutanix credentials: loading secrets object: executing apply: test error")
	assert.Nil(t, cpSpec)
	assert.Nil(t, workerSpec)
}

func TestNeedsNewControlPlaneTemplate(t *testing.T) {
	tests := []struct {
		name             string
		newClusterSpec   func(spec cluster.Spec) cluster.Spec
		newMachineConfig func(anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig
		expectedResult   bool
	}{
		{
			name: "kubernetes version changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Cluster.Spec.KubernetesVersion = "1.21.2"
				return *s
			},
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				return spec
			},
			expectedResult: true,
		},
		{
			name: "bundle spec number changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Bundles.Spec.Number = 42
				return *s
			},
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				return spec
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		oldClusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
		newClusterSpec := tt.newClusterSpec(*oldClusterSpec)

		oldMachineConf := &anywherev1.NutanixMachineConfig{}
		err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), oldMachineConf)
		require.NoError(t, err)
		newMachineConf := tt.newMachineConfig(*oldMachineConf)

		assert.Equal(t, tt.expectedResult, NeedsNewControlPlaneTemplate(oldClusterSpec, &newClusterSpec, oldMachineConf, &newMachineConf))
	}
}

// NeedsNewWorkloadTemplate determines if a new workload template is needed.
func TestNeedsNewWorkloadTemplate(t *testing.T) {
	tests := []struct {
		name             string
		newClusterSpec   func(spec cluster.Spec) cluster.Spec
		newMachineConfig func(anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig
		expectedResult   bool
	}{
		{
			name: "kubernetes version changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Cluster.Spec.KubernetesVersion = "1.21.2"
				return *s
			},
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				return spec
			},
			expectedResult: true,
		},
		{
			name: "bundle spec number changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Bundles.Spec.Number = 42
				return *s
			},
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				return spec
			},
			expectedResult: true,
		},
		{
			name: "woker node config labels changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Labels = map[string]string{"foo": "bar"}
				return *s
			},
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				return spec
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		oldClusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
		newClusterSpec := tt.newClusterSpec(*oldClusterSpec)

		oldMachineConf := &anywherev1.NutanixMachineConfig{}
		err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), oldMachineConf)
		require.NoError(t, err)
		newMachineConf := tt.newMachineConfig(*oldMachineConf)

		newWorkerConfig := newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]
		oldWorkerConfig := oldClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]

		assert.Equal(t, tt.expectedResult, NeedsNewWorkloadTemplate(oldClusterSpec, &newClusterSpec, oldMachineConf, &newMachineConf, newWorkerConfig, oldWorkerConfig))
	}
}

func TestAnyImmutableFieldChanged(t *testing.T) {
	tests := []struct {
		name             string
		newMachineConfig func(anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig
		expectedResult   bool
	}{
		{
			name: "machine image changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.Image.Name = utils.StringPtr("new-image")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine image identifier type changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.Image.Type = anywherev1.NutanixIdentifierUUID
				conf.Spec.Image.Name = utils.StringPtr("49ab2c64-72a1-4637-9673-e2f13b1463cb")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine memory size changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.MemorySize = resource.MustParse("4Gi")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine system disk size changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.SystemDiskSize = resource.MustParse("20Gi")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine VCPU sockets changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.VCPUSockets = 2
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine vcpus per socket changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.VCPUsPerSocket = 2
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine cluster changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.Cluster.Name = utils.StringPtr("new-cluster")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine subnet changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.Subnet.Name = utils.StringPtr("new-subnet")
				return *conf
			},
			expectedResult: true,
		},
		{
			name: "machine OS Family changed",
			newMachineConfig: func(spec anywherev1.NutanixMachineConfig) anywherev1.NutanixMachineConfig {
				conf := spec.DeepCopy()
				conf.Spec.OSFamily = "new-os-family"
				return *conf
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		oldMachineConf := &anywherev1.NutanixMachineConfig{}
		err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), oldMachineConf)
		require.NoError(t, err)
		newMachineConf := tt.newMachineConfig(*oldMachineConf)

		assert.Equal(t, tt.expectedResult, AnyImmutableFieldChanged(oldMachineConf, &newMachineConf))
	}
}

func TestNutanixProviderGenerateStorageClass(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	sc := provider.GenerateStorageClass()
	assert.Nil(t, sc)
}

func TestNutanixProviderGenerateMHC(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	mhc, err := provider.GenerateMHC(clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, mhc)
}

func TestNutanixProviderUpdateKubeconfig(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	err := provider.UpdateKubeConfig(nil, "test")
	assert.NoError(t, err)
}

func TestNutanixProviderVersion(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	managementComponents := givenManagementComponents()
	v := provider.Version(managementComponents)
	assert.NotNil(t, v)
}

func TestNutanixProviderEnvMap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	managementComponents := givenManagementComponents()

	t.Run("required envs not set", func(t *testing.T) {
		os.Clearenv()
		envMap, err := provider.EnvMap(managementComponents, clusterSpec)
		assert.Error(t, err)
		assert.Nil(t, envMap)
	})

	t.Run("required envs set", func(t *testing.T) {
		t.Setenv(constants.NutanixUsernameKey, "nutanix")
		t.Setenv(constants.NutanixPasswordKey, "nutanix")
		t.Setenv(nutanixEndpointKey, "prism.nutanix.com")
		t.Setenv(expClusterResourceSetKey, "true")

		envMap, err := provider.EnvMap(managementComponents, clusterSpec)
		assert.NoError(t, err)
		assert.NotNil(t, envMap)
	})
}

func TestNutanixProviderGetDeployments(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	deps := provider.GetDeployments()
	assert.NotNil(t, deps)
	assert.Contains(t, deps, "capx-system")
}

func TestNutanixProviderGetInfrastructureBundle(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	managementComponents := givenManagementComponents()
	wantInfraBundle := &types.InfrastructureBundle{
		FolderName: "infrastructure-nutanix/1.0.0/",
		Manifests: []releasev1alpha1.Manifest{
			managementComponents.Nutanix.Components,
			managementComponents.Nutanix.Metadata,
			managementComponents.Nutanix.ClusterTemplate,
		},
	}

	assert.Equal(t, wantInfraBundle, provider.GetInfrastructureBundle(managementComponents))
}

func TestNutanixProviderDatacenterConfig(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	dc := provider.DatacenterConfig(clusterSpec)
	assert.Equal(t, provider.datacenterConfig, dc)
}

func TestNutanixProviderMachineConfigs(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	confs := provider.MachineConfigs(clusterSpec)
	require.NotEmpty(t, confs)
	assert.Len(t, confs, 1)
}

func TestNutanixProviderValidateNewSpec(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.ValidateNewSpec(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderChangeDiffWithChange(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	managementComponents := givenManagementComponents()
	managementComponents.Nutanix.Version = "v0.5.2"

	newManagementComponents := givenManagementComponents()
	newManagementComponents.Nutanix.Version = "v1.0.0"
	want := &types.ComponentChangeDiff{
		ComponentName: "nutanix",
		NewVersion:    "v1.0.0",
		OldVersion:    "v0.5.2",
	}

	got := provider.ChangeDiff(managementComponents, newManagementComponents)
	assert.Equal(t, got, want)
}

func TestNutanixProviderRunPostControlPlaneUpgrade(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.RunPostControlPlaneUpgrade(context.Background(), clusterSpec, clusterSpec, cluster, cluster)
	assert.NoError(t, err)
}

func TestNutanixProviderUpgradeNeeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(*bytes.NewBufferString(nutanixDatacenterConfigSpecJSON), nil)
	executable.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(*bytes.NewBufferString(nutanixMachineConfigSpecJSON), nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test", KubeconfigFile: "testdata/kubeconfig.yaml"}
	upgrade, err := provider.UpgradeNeeded(context.Background(), clusterSpec, clusterSpec, cluster)
	assert.NoError(t, err)
	assert.False(t, upgrade)
}

func TestNutanixProviderRunPostControlPlaneCreation(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.RunPostControlPlaneCreation(context.Background(), clusterSpec, cluster)
	assert.NoError(t, err)
}

func TestNutanixProviderMachineDeploymentsToDelete(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	deps := provider.MachineDeploymentsToDelete(cluster, clusterSpec, clusterSpec)
	assert.NotNil(t, deps)
	assert.Len(t, deps, 0)
}

func TestNutanixProviderInstallCustomProviderComponents(t *testing.T) {
	provider := testDefaultNutanixProvider(t)

	kubeConfigFile := "test"
	err := provider.InstallCustomProviderComponents(context.Background(), kubeConfigFile)
	assert.NoError(t, err)
}

func TestNutanixProviderPreCAPIInstallOnBootstrap(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockCertValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()
	mockHTTPClient := &http.Client{Transport: mockTransport}
	mockWriter := filewritermocks.NewMockFileWriter(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl, mockCertValidator, mockHTTPClient, mockWriter)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.PreCAPIInstallOnBootstrap(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)

	storedMarshal := jsonMarshal
	jsonMarshal = fakemarshal
	err = provider.PreCAPIInstallOnBootstrap(context.Background(), cluster, clusterSpec)
	assert.ErrorContains(t, err, "marshalling failed")
	restoremarshal(storedMarshal)
}

func TestNutanixProviderPostMoveManagementToBootstrap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.PostMoveManagementToBootstrap(context.Background(), cluster)
	assert.NoError(t, err)
}
