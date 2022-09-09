package tinkerbell

import (
	"context"
	"path"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	stackmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	testDataDir = "testdata"
	testIP      = "5.6.7.8"
)

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

func newProvider(datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, docker stack.Docker, helm stack.Helm, kubectl ProviderKubectlClient, forceCleanup bool) *Provider {
	hardwareFile := "./testdata/hardware.csv"
	provider, err := NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		hardwareFile,
		writer,
		docker,
		helm,
		kubectl,
		testIP,
		test.FakeNow,
		forceCleanup,
		false,
	)
	if err != nil {
		panic(err)
	}

	return provider
}

func TestTinkerbellProviderGenerateDeploymentFileWithExternalEtcd(t *testing.T) {
	t.Skip("External etcd unsupported for GA")
	clusterSpecManifest := "cluster_tinkerbell_external_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

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
	clusterSpecManifest := "cluster_tinkerbell_missing_ssh_keys.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	keyGenerator := mocks.NewMockSSHAuthKeyGenerator(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	const sshKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="
	keyGenerator.EXPECT().GenerateSSHAuthKey(gomock.Any()).Return(sshKey, nil)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)

	// Hack: monkey patch the key generator and the stack installer directly for determinism.
	provider.keyGenerator = keyGenerator
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

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
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

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

func TestTinkerbellProviderGenerateDeploymentFileWithNodeLabels(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_node_labels.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_node_labels.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md_node_labels.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileWithNodeTaints(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_node_taints.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_node_taints.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md_node_taints.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileMultipleWorkerNodeGroups(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_multiple_node_groups.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

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

func TestPreCAPIInstallOnBootstrapSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test", KubeconfigFile: "test.kubeconfig"}
	ctx := context.Background()
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().Install(
		ctx,
		clusterSpec.VersionsBundle.Tinkerbell,
		testIP,
		"test.kubeconfig",
		"",
		gomock.Any(),
		gomock.Any(),
	)

	err := provider.PreCAPIInstallOnBootstrap(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed PreCAPIInstallOnBootstrap: %v", err)
	}
}

func TestPostWorkloadInitSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test", KubeconfigFile: "test.kubeconfig"}
	ctx := context.Background()
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().Install(
		ctx,
		clusterSpec.VersionsBundle.Tinkerbell,
		testIP,
		"test.kubeconfig",
		"",
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	)
	stackInstaller.EXPECT().UninstallLocal(ctx)

	err := provider.PostWorkloadInit(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed PostWorkloadInit: %v", err)
	}
}

func TestPostBootstrapSetupSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test", KubeconfigFile: "test.kubeconfig"}
	ctx := context.Background()
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)

	kubectl.EXPECT().ApplyKubeSpecFromBytesForce(ctx, cluster, gomock.Any())
	kubectl.EXPECT().WaitForBaseboardManagements(ctx, cluster, "5m", "Contactable", gomock.Any()).MaxTimes(2)

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)

	err := provider.PostBootstrapSetup(ctx, provider.clusterConfig, cluster)
	if err != nil {
		t.Fatalf("failed PostBootstrapSetup: %v", err)
	}
}

func TestTinkerbellProviderGenerateDeploymentFileWithFullOIDC(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_full_oidc.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_full_oidc.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileWithMinimalOIDC(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_minimal_oidc.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_minimal_oidc.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileWithAWSIamConfig(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_awsiam.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_awsiam.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestProviderGenerateDeploymentFileForWithMinimalRegistryMirror(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_minimal_registry_mirror.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_minimal_registry_mirror.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md_minimal_registry_mirror.yaml")
}

func TestProviderGenerateDeploymentFileForWithRegistryMirrorWithCert(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_registry_mirror_with_cert.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp_registry_mirror_with_cert.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md_registry_mirror_with_cert.yaml")
}

func TestProviderGenerateDeploymentFileForWithBottlerocketMinimalRegistryMirror(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_bottlerocket_minimal_registry_mirror.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_bottlerocket_cp_minimal_registry_mirror.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_bottlerocket_md_minimal_registry_mirror.yaml")
}

func TestProviderGenerateDeploymentFileForWithBottlerocketRegistryMirrorWithCert(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_bottlerocket_registry_mirror_with_cert.yaml"
	mockCtrl := gomock.NewController(t)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	forceCleanup := false

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	provider := newProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl, forceCleanup)
	provider.stackInstaller = stackInstaller

	stackInstaller.EXPECT().CleanupLocalBoots(ctx, forceCleanup)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_bottlerocket_cp_registry_mirror_with_cert.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_bottlerocket_md_registry_mirror_with_cert.yaml")
}
