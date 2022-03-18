package tinkerbell

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	testDataDir                         = "testdata"
	expectedTinkerbellIP                = "1.2.3.4"
	expectedTinkerbellGRPCAuth          = "1.2.3.4:42113"
	expectedTinkerbellCertURL           = "1.2.3.4:42114/cert"
	expectedTinkerbellPBnJGRPCAuthority = "1.2.3.4:42000"
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

func newProviderWithKubectlWithTink(t *testing.T, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, kubectl ProviderKubectlClient, tinkerbellClients TinkerbellClients) *tinkerbellProvider {
	return newProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		writer,
		kubectl,
		tinkerbellClients,
	)
}

func newProvider(datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, kubectl ProviderKubectlClient, tinkerbellClients TinkerbellClients) *tinkerbellProvider {
	return NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		writer,
		kubectl,
		tinkerbellClients,
		test.FakeNow,
		true,
		"testdata/hardware_config.yaml",
		false,
	)
}

type testContext struct {
	oldTinkerbellIP                  string
	isTinkerbellIPSet                bool
	oldTinkerbellCertURL             string
	isTinkerbellCertURLSet           bool
	oldtinkerbellGRPCAuth            string
	isTinkerbellGRPCAuthSet          bool
	oldTinkerbellPBnJGRPCAuthority   string
	isTinkerbellPBnJGRPCAuthoritySet bool
}

func (tctx *testContext) SaveContext() {
	tctx.oldTinkerbellIP, tctx.isTinkerbellIPSet = os.LookupEnv(tinkerbellIPKey)
	tctx.oldTinkerbellCertURL, tctx.isTinkerbellCertURLSet = os.LookupEnv(tinkerbellCertURLKey)
	tctx.oldtinkerbellGRPCAuth, tctx.isTinkerbellGRPCAuthSet = os.LookupEnv(tinkerbellGRPCAuthKey)
	tctx.oldTinkerbellPBnJGRPCAuthority, tctx.isTinkerbellPBnJGRPCAuthoritySet = os.LookupEnv(tinkerbellPBnJGRPCAuthorityKey)
	os.Setenv(tinkerbellIPKey, expectedTinkerbellIP)
	os.Setenv(tinkerbellCertURLKey, expectedTinkerbellCertURL)
	os.Setenv(tinkerbellGRPCAuthKey, expectedTinkerbellGRPCAuth)
	os.Setenv(tinkerbellPBnJGRPCAuthorityKey, expectedTinkerbellPBnJGRPCAuthority)
	os.Setenv(features.TinkerbellProviderEnvVar, "true")
}

func (tctx *testContext) RestoreContext() {
	if tctx.isTinkerbellIPSet {
		os.Setenv(tinkerbellIPKey, tctx.oldTinkerbellIP)
	} else {
		os.Unsetenv(tinkerbellIPKey)
	}
	if tctx.isTinkerbellCertURLSet {
		os.Setenv(tinkerbellCertURLKey, tctx.oldTinkerbellCertURL)
	} else {
		os.Unsetenv(tinkerbellCertURLKey)
	}
	if tctx.isTinkerbellGRPCAuthSet {
		os.Setenv(tinkerbellGRPCAuthKey, tctx.oldtinkerbellGRPCAuth)
	} else {
		os.Unsetenv(tinkerbellGRPCAuthKey)
	}
	if tctx.isTinkerbellPBnJGRPCAuthoritySet {
		os.Setenv(tinkerbellPBnJGRPCAuthorityKey, tctx.oldTinkerbellPBnJGRPCAuthority)
	} else {
		os.Unsetenv(tinkerbellPBnJGRPCAuthorityKey)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

func setupHardware() []*tinkhardware.Hardware {
	var hardwares []*tinkhardware.Hardware
	hardwares = append(hardwares, &tinkhardware.Hardware{Id: "b14d7f5b-8903-4a4c-b38d-55889ba820ba"})
	hardwares = append(hardwares, &tinkhardware.Hardware{Id: "b14d7f5b-8903-4a4c-b38d-55889ba820bb"})
	hardwares = append(hardwares, &tinkhardware.Hardware{Id: "d2c14d26-640a-48f0-baee-a737c68a75f5"})
	hardwares = append(hardwares, &tinkhardware.Hardware{Id: "0c9d1701-f884-499e-80b8-6dcfc0973e85"})

	return hardwares
}

func TestTinkerbellProviderGenerateDeploymentFileWithExternalEtcd(t *testing.T) {
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell_external_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tink := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	tinkerbellClients := TinkerbellClients{tink, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tink.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	tink.EXPECT().GetWorkflow(ctx).Return([]*tinkworkflow.Workflow{}, nil)

	pbnjClient.EXPECT().GetPowerState(ctx, gomock.Any()).Return(pbnj.PowerStateOff, nil).Times(4)

	provider := newProviderWithKubectlWithTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, kubectl, tinkerbellClients)
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
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell_missing_ssh_keys.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tink := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	keyGenerator := mocks.NewMockSSHAuthKeyGenerator(mockCtrl)
	tinkerbellClients := TinkerbellClients{tink, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tink.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	tink.EXPECT().GetWorkflow(ctx).Return([]*tinkworkflow.Workflow{}, nil)

	pbnjClient.EXPECT().GetPowerState(ctx, gomock.Any()).Return(pbnj.PowerStateOff, nil).Times(4)

	const sshKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="
	keyGenerator.EXPECT().GenerateSSHAuthKey(gomock.Any()).Return(sshKey, nil)

	provider := newProviderWithKubectlWithTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, kubectl, tinkerbellClients)

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
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tink := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	tinkerbellClients := TinkerbellClients{tink, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tink.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	tink.EXPECT().GetWorkflow(ctx).Return([]*tinkworkflow.Workflow{}, nil)
	pbnjClient.EXPECT().GetPowerState(ctx, gomock.Any()).Return(pbnj.PowerStateOff, nil).Times(4)

	provider := newProviderWithKubectlWithTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, kubectl, tinkerbellClients)
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
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell_multiple_node_groups.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tink := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	tinkerbellClients := TinkerbellClients{tink, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tink.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	tink.EXPECT().GetWorkflow(ctx).Return([]*tinkworkflow.Workflow{}, nil)

	pbnjClient.EXPECT().GetPowerState(ctx, gomock.Any()).Return(pbnj.PowerStateOff, nil).Times(4)
	provider := newProviderWithKubectlWithTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, kubectl, tinkerbellClients)
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
