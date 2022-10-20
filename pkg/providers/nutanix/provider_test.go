package nutanix

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/pkg/constants"

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

	t.Setenv(constants.NutanixUsernameKey, "admin")
	t.Setenv(constants.NutanixPasswordKey, "password")

	provider := NewProvider(dcConf, workerConfs, clusterConf, kubectl, nutanixClient, time.Now)
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
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().Execute(gomock.Any(), "delete", []string{"nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "default", "--ignore-not-found=true"}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	executable.EXPECT().Execute(gomock.Any(), "delete", []string{"nutanixmachineconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", "testdata/kubeconfig.yaml", "--namespace", "default", "--ignore-not-found=true"}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	kubectl := executables.NewKubectl(executable)

	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)
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
	assert.NoError(t, err)
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
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
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
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	cluster := &types.Cluster{Name: "eksa-unit-test", KubeconfigFile: "testdata/kubeconfig.yaml"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cpSpec, workerSpec, err := provider.GenerateCAPISpecForUpgrade(context.Background(), cluster, cluster, clusterSpec, clusterSpec)
	assert.NoError(t, err)
	assert.NotEmpty(t, cpSpec)
	assert.NotEmpty(t, workerSpec)
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
			name: "control plane config endpoint changed",
			newClusterSpec: func(spec cluster.Spec) cluster.Spec {
				s := spec.DeepCopy()
				s.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "anotherprism.nutanix.com"
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

		assert.Equal(t, tt.expectedResult, NeedsNewWorkloadTemplate(oldClusterSpec, &newClusterSpec, oldMachineConf, &newMachineConf))
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
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	v := provider.Version(clusterSpec)
	assert.NotNil(t, v)
}

func TestNutanixProviderEnvMap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")

	t.Run("required envs not set", func(t *testing.T) {
		os.Clearenv()
		envMap, err := provider.EnvMap(clusterSpec)
		assert.Error(t, err)
		assert.Nil(t, envMap)
	})

	t.Run("required envs set", func(t *testing.T) {
		t.Setenv(constants.NutanixUsernameKey, "nutanix")
		t.Setenv(constants.NutanixPasswordKey, "nutanix")
		t.Setenv(nutanixEndpointKey, "prism.nutanix.com")

		envMap, err := provider.EnvMap(clusterSpec)
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
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(*bytes.NewBufferString(nutanixDatacenterConfigSpecJSON), nil)
	executable.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", "testdata/kubeconfig.yaml", "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "--namespace", "default", "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(*bytes.NewBufferString(nutanixMachineConfigSpecJSON), nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

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

func TestNutanixProviderPreCAPIInstallOnBootstrap(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	executable.EXPECT().ExecuteWithStdin(gomock.Any(), gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, nil)
	kubectl := executables.NewKubectl(executable)
	mockClient := NewMockClient(ctrl)
	provider := testNutanixProvider(t, mockClient, kubectl)

	cluster := &types.Cluster{Name: "eksa-unit-test"}
	clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	err := provider.PreCAPIInstallOnBootstrap(context.Background(), cluster, clusterSpec)
	assert.NoError(t, err)
}

func TestNutanixProviderPostMoveManagementToBootstrap(t *testing.T) {
	provider := testDefaultNutanixProvider(t)
	cluster := &types.Cluster{Name: "eksa-unit-test"}
	err := provider.PostMoveManagementToBootstrap(context.Background(), cluster)
	assert.NoError(t, err)
}
