package tinkerbell

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
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

func newProviderWithKubectlTink(t *testing.T, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, tinkerbellClients TinkerbellClients) *tinkerbellProvider {
	return newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
		tinkerbellClients,
	)
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, tinkerbellClients TinkerbellClients) *tinkerbellProvider {
	return NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
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

func TestTinkerbellProviderGenerateDeploymentFile(t *testing.T) {
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tinkctl := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	tinkerbellClients := TinkerbellClients{tinkctl, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()

	tinkctl.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	pbnjClient.EXPECT().ValidateBMCSecretCreds(ctx, gomock.Any()).Return(nil).Times(4)

	provider := newProviderWithKubectlTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, tinkerbellClients)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_cluster_tinkerbell_md.yaml")
}

func TestTinkerbellProviderGenerateDeploymentFileMultipleWorkerNodeGroups(t *testing.T) {
	setupContext(t)
	clusterSpecManifest := "cluster_tinkerbell_multiple_node_groups.yaml"
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	tinkctl := mocks.NewMockProviderTinkClient(mockCtrl)
	pbnjClient := mocks.NewMockProviderPbnjClient(mockCtrl)
	tinkerbellClients := TinkerbellClients{tinkctl, pbnjClient}
	cluster := &types.Cluster{Name: "test"}
	hardwares := setupHardware()
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	tinkctl.EXPECT().GetHardware(ctx).Return(hardwares, nil)
	pbnjClient.EXPECT().ValidateBMCSecretCreds(ctx, gomock.Any()).Return(nil).Times(4)
	provider := newProviderWithKubectlTink(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl, tinkerbellClients)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_tinkerbell_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_tinkerbell_md_multiple_node_groups.yaml")
}
