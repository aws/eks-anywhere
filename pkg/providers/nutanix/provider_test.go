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

func testDefaultNutanixProvider(t *testing.T) *Provider {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)
	return provider
}

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
	err := provider.PostBootstrapDeleteForUpgrade(context.Background())
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
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
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

func TestNutanixProviderSetupAndValidateDeleteCluster(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.SetupAndValidateDeleteCluster(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderSetupAndValidateUpgradeCluster(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.SetupAndValidateUpgradeCluster(context.Background(), &types.Cluster{Name: "eksa-unit-test"}, clusterSpec, clusterSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upgrade for nutanix provider isn't currently supported")
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

func TestNutanixProviderGenerateCAPISpecForCreate(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
	assert.NotNil(t, cpSpec)
	assert.NotNil(t, workerSpec)
}

func TestNutanixProviderGenerateCAPISpecForUpgrade(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForUpgrade(context.Background(), cluster, cluster, clusterSpec, clusterSpec)
	assert.NoError(t, err)
	assert.Nil(t, cpSpec)
	assert.Nil(t, workerSpec)
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
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	v := provider.Version(clusterSpec)
	assert.NotNil(t, v)
}

func TestNutanixProviderEnvMap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
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

func TestNutanixProviderGetDeployments(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	deps := provider.GetDeployments()
	assert.NotNil(t, deps)
	assert.Contains(t, deps, "capx-system")
}

func TestNutanixProviderGetInfrastructureBundle(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	bundle := provider.GetInfrastructureBundle(clusterSpec)
	assert.NotNil(t, bundle)
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

func TestNutanixProviderChangeDiff(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cd := provider.ChangeDiff(clusterSpec, clusterSpec)
	assert.Nil(t, cd)
}

func TestNutanixProviderRunPostControlPlaneUpgrade(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.RunPostControlPlaneUpgrade(context.Background(), clusterSpec, clusterSpec, cluster, cluster)
	assert.NoError(t, err)
}

func TestNutanixProviderUpgradeNeeded(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
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

func TestNutanixProviderPreCAPIInstallOnBootstrap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.PreCAPIInstallOnBootstrap(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderPostMoveManagementToBootstrap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.PostMoveManagementToBootstrap(context.Background(), cluster)
	assert.NoError(t, err)
}
