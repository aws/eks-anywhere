package nutanix

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func testNutanixProvider(t *testing.T, nutanixClient Client, kubectl *executables.Kubectl) *Provider {
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

	os.Setenv(nutanixUsernameKey, "admin")
	defer os.Unsetenv(nutanixUsernameKey)
	os.Setenv(nutanixPasswordKey, "password")
	defer os.Unsetenv(nutanixPasswordKey)

	validator, err := NewValidator(nutanixClient)
	require.NoError(t, err)

	provider, err := NewProvider(dcConf, workerConfs, clusterConf, kubectl, nutanixClient, validator, time.Now)
	require.NoError(t, err)
	return provider
}

func TestNutanixProviderSetupAndValidateCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	mockClient := NewMockClient(ctrl)
	clusters := &v3.ClusterListIntentResponse{
		Entities: []*v3.ClusterIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cda"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster"),
				},
			},
		},
	}
	mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
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
	mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(subnets, nil)
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
	mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(images, nil)
	provider := testNutanixProvider(t, mockClient, kubectl)
	assert.NotNil(t, provider)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.SetupAndValidateCreateCluster(context.Background(), clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderGenerateCAPISpecForCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)
	assert.NotNil(t, provider)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")

	cpSpec, workerSpec, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, cpSpec)
	assert.NotNil(t, workerSpec)
}

func TestNutanixProviderUpdateSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.UpdateSecrets(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderGenerateMHC(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	mhc, err := provider.GenerateMHC(clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, mhc)
}

func TestNutanixProviderEnvMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	envMap, err := provider.EnvMap(clusterSpec)
	assert.Error(t, err)
	assert.Nil(t, envMap)

	os.Setenv("NUTANIX_USER", "nutanix")
	defer os.Unsetenv("NUTANIX_USER")
	os.Setenv("NUTANIX_PASSWORD", "nutanix")
	defer os.Unsetenv("NUTANIX_PASSWORD")
	os.Setenv("NUTANIX_ENDPOINT", "prism.nutanix.com")
	defer os.Unsetenv("NUTANIX_ENDPOINT")

	envMap, err = provider.EnvMap(clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, envMap)
}

func TestNutanixProviderGetInfrastructureBundle(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	bundle := provider.GetInfrastructureBundle(clusterSpec)
	assert.NotNil(t, bundle)
}

func TestNutanixProviderMachineConfigs(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	confs := provider.MachineConfigs(clusterSpec)
	require.NotEmpty(t, confs)
	assert.Len(t, confs, 1)
}
