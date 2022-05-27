package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const testDataDir = "testdata"

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.TinkerbellDatacenterConfig {
	datacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return datacenterConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.TinkerbellMachineConfig {
	machineConfigs, err := v1alpha1.GetTinkerbellMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file: %v", err)
	}
	return machineConfigs
}

func newProvider(datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, docker Docker, kubectl ProviderKubectlClient) *Provider {
	reader, err := hardware.NewCSVReaderFromFile("./testdata/hardware.csv")
	if err != nil {
		panic(err)
	}

	return NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		reader,
		writer,
		docker,
		kubectl,
		test.FakeNow,
		true,
		false,
	)
}

func TestTinkerbellProviderGenerateDeploymentFileWithExternalEtcd(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")

	clusterSpecManifest := "cluster_tinkerbell_external_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_external_etcd.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestTinkerbellProviderMachineConfigsMissingUserSshKeys(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_missing_ssh_keys.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	keyGenerator := mocks.NewMockSSHAuthKeyGenerator(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	const sshKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="
	keyGenerator.EXPECT().GenerateSSHAuthKey(gomock.Any()).Return(sshKey, nil)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)

	// Hack: monkey patch the key generator directly for determinism.
	provider.keyGenerator = keyGenerator

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_missing_ssh_keys.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileWithStackedEtcd(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_stacked_etcd.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileMultipleWorkerNodeGroups(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_multiple_node_groups.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_external_etcd.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_tinkerbell_md_multiple_node_groups.yaml")
}

func TestTinkerbellProviderPreCAPIInstallOnBootstrapSuccess(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tinkManifestName := "tink.yaml"
	bootsImageURI := "public.ecr.aws/eks-anywhere/tinkerbell/boots:latest"
	hegelManifestName := "hegel.yaml"

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			Manifest: releasev1alpha1.Manifest{URI: tinkManifestName},
		},
		Boots: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: bootsImageURI},
		},
		Hegel: releasev1alpha1.TinkerbellServiceBundle{
			Manifest: releasev1alpha1.Manifest{URI: hegelManifestName},
		},
	}

	kubectl.EXPECT().ApplyKubeSpec(ctx, cluster, tinkManifestName)
	kubectl.EXPECT().WaitForDeployment(ctx, cluster, deploymentWaitTimeout, "Available", "tink-server", constants.EksaSystemNamespace)
	kubectl.EXPECT().WaitForDeployment(ctx, cluster, deploymentWaitTimeout, "Available", "tink-controller-manager", constants.EksaSystemNamespace)

	kubectl.EXPECT().ApplyKubeSpec(ctx, cluster, hegelManifestName)
	kubectl.EXPECT().WaitForDeployment(ctx, cluster, deploymentWaitTimeout, "Available", "hegel", constants.EksaSystemNamespace)

	docker.EXPECT().Run(ctx,
		"public.ecr.aws/eks-anywhere/tinkerbell/boots:latest",
		"boots",
		[]string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67"},
		"-v", gomock.Any(),
		"--network", "host",
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
	)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)

	err := provider.InstallTinkerbellStack(ctx, cluster, clusterSpec, true)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellProviderPreCAPIInstallOnBootstrapFailureApplyingManifest(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			Manifest: releasev1alpha1.Manifest{URI: "tink.yaml"},
		},
	}

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)

	kubectlError := "kubectl error"
	expectedError := fmt.Sprintf("applying tink manifest: %s", kubectlError)
	kubectl.EXPECT().ApplyKubeSpec(ctx, cluster, "tink.yaml").Return(errors.New(kubectlError))

	err := provider.InstallTinkerbellStack(ctx, cluster, clusterSpec, true)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}

func TestTinkerbellProviderPreCAPIInstallOnBootstrapFailureWaitingForDeployment(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			Manifest: releasev1alpha1.Manifest{URI: "tink.yaml"},
		},
	}

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, kubectl)

	kubectlError := "kubectl error"
	expectedError := fmt.Sprintf("waiting for deployment tink-server: %s", kubectlError)
	kubectl.EXPECT().ApplyKubeSpec(ctx, cluster, "tink.yaml")
	kubectl.EXPECT().WaitForDeployment(ctx, cluster, deploymentWaitTimeout, "Available", "tink-server", constants.EksaSystemNamespace).Return(errors.New(kubectlError))

	err := provider.InstallTinkerbellStack(ctx, cluster, clusterSpec, true)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}
