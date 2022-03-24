package workflows_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type upgradeTestSetup struct {
	t                  *testing.T
	bootstrapper       *mocks.MockBootstrapper
	clusterManager     *mocks.MockClusterManager
	addonManager       *mocks.MockAddonManager
	provider           *providermocks.MockProvider
	writer             *writermocks.MockFileWriter
	validator          *mocks.MockValidator
	capiManager        *mocks.MockCAPIManager
	datacenterConfig   providers.DatacenterConfig
	machineConfigs     []providers.MachineConfig
	workflow           *workflows.Upgrade
	ctx                context.Context
	newClusterSpec     *cluster.Spec
	currentClusterSpec *cluster.Spec
	forceCleanup       bool
	bootstrapCluster   *types.Cluster
	workloadCluster    *types.Cluster
}

func newUpgradeTest(t *testing.T) *upgradeTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	bootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	addonManager := mocks.NewMockAddonManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	validator := mocks.NewMockValidator(mockCtrl)
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	capiUpgrader := mocks.NewMockCAPIManager(mockCtrl)
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	workflow := workflows.NewUpgrade(bootstrapper, provider, capiUpgrader, clusterManager, addonManager, writer)

	for _, e := range featureEnvVars {
		if err := os.Setenv(e, "true"); err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		for _, e := range featureEnvVars {
			if err := os.Unsetenv(e); err != nil {
				t.Fatal(err)
			}
		}
	})

	return &upgradeTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		addonManager:     addonManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		capiManager:      capiUpgrader,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		newClusterSpec:   test.NewClusterSpec(func(s *cluster.Spec) { s.Cluster.Name = "cluster-name" }),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func (c *upgradeTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateUpgradeCluster(c.ctx, gomock.Any(), c.newClusterSpec)
	c.provider.EXPECT().Name()
}

func (c *upgradeTestSetup) expectUpdateSecrets(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.provider.EXPECT().UpdateSecrets(c.ctx, expectedCluster).Return(nil),
	)
}

func (c *upgradeTestSetup) expectEnsureEtcdCAPIComponentsExistTask(expectedCluster *types.Cluster) {
	currentSpec := c.currentClusterSpec
	c.capiManager.EXPECT().EnsureEtcdProvidersInstallation(c.ctx, expectedCluster, c.provider, currentSpec)
}

func (c *upgradeTestSetup) expectUpgradeCoreComponents(expectedCluster *types.Cluster) {
	currentSpec := c.currentClusterSpec
	networkingChangeDiff := types.NewChangeDiff(&types.ComponentChangeDiff{
		ComponentName: "cilium",
		OldVersion:    "v0.0.1",
		NewVersion:    "v0.0.2",
	})
	capiChangeDiff := types.NewChangeDiff(&types.ComponentChangeDiff{
		ComponentName: "vsphere",
		OldVersion:    "v0.0.1",
		NewVersion:    "v0.0.2",
	})
	fluxChangeDiff := types.NewChangeDiff(&types.ComponentChangeDiff{
		ComponentName: "Flux",
		OldVersion:    "v0.0.1",
		NewVersion:    "v0.0.2",
	})
	eksaChangeDiff := types.NewChangeDiff(&types.ComponentChangeDiff{
		ComponentName: "eks-a",
		OldVersion:    "v0.0.1",
		NewVersion:    "v0.0.2",
	})
	gomock.InOrder(
		c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, expectedCluster, c.newClusterSpec.Cluster.Name).Return(currentSpec, nil),
		c.clusterManager.EXPECT().UpgradeNetworking(c.ctx, expectedCluster, currentSpec, c.newClusterSpec, c.provider).Return(networkingChangeDiff, nil),
		c.capiManager.EXPECT().Upgrade(c.ctx, expectedCluster, c.provider, currentSpec, c.newClusterSpec).Return(capiChangeDiff, nil),
		c.addonManager.EXPECT().UpdateLegacyFileStructure(c.ctx, currentSpec, c.newClusterSpec),
		c.addonManager.EXPECT().Upgrade(c.ctx, expectedCluster, currentSpec, c.newClusterSpec).Return(fluxChangeDiff, nil),
		c.clusterManager.EXPECT().Upgrade(c.ctx, expectedCluster, currentSpec, c.newClusterSpec).Return(eksaChangeDiff, nil),
	)
}

func (c *upgradeTestSetup) expectCreateBootstrap() {
	opts := []bootstrapper.BootstrapClusterOption{
		bootstrapper.WithDefaultCNIDisabled(), bootstrapper.WithExtraDockerMounts(),
	}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts().Return(opts, nil),
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil()),
		).Return(c.bootstrapCluster, nil),

		c.clusterManager.EXPECT().InstallCAPI(c.ctx, gomock.Not(gomock.Nil()), c.bootstrapCluster, c.provider),
	)
}

func (c *upgradeTestSetup) expectNotToCreateBootstrap() {
	c.provider.EXPECT().BootstrapClusterOpts().Times(0)
	c.bootstrapper.EXPECT().CreateBootstrapCluster(
		c.ctx, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil()),
	).Times(0)

	c.clusterManager.EXPECT().InstallCAPI(c.ctx, gomock.Not(gomock.Nil()), c.bootstrapCluster, c.provider).Times(0)
}

func (c *upgradeTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.newClusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.newClusterSpec).Return(c.machineConfigs),
		c.writer.EXPECT().Write("cluster-name-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),
	)
}

func (c *upgradeTestSetup) expectDeleteBootstrap() {
	gomock.InOrder(
		c.bootstrapper.EXPECT().DeleteBootstrapCluster(
			c.ctx, c.bootstrapCluster,
			gomock.Any()).Return(nil),
	)
}

func (c *upgradeTestSetup) expectNotToDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any()).Times(0)
}

func (c *upgradeTestSetup) expectUpgradeWorkload(expectedCluster *types.Cluster) {
	c.expectUpgradeWorkloadToReturn(expectedCluster, nil)
	c.clusterManager.EXPECT().ApplyBundles(c.ctx, c.newClusterSpec, expectedCluster)
}

func (c *upgradeTestSetup) expectUpgradeWorkloadToReturn(expectedCluster *types.Cluster, err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().UpgradeCluster(
			c.ctx, c.bootstrapCluster, expectedCluster, c.newClusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeTestSetup) expectMoveManagementToBootstrap() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.workloadCluster, c.bootstrapCluster, gomock.Any(), c.newClusterSpec, gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectNotToMoveManagementToBootstrap() {
	c.clusterManager.EXPECT().MoveCAPI(c.ctx, c.workloadCluster, c.bootstrapCluster, gomock.Any(), c.newClusterSpec, gomock.Any()).Times(0)
}

func (c *upgradeTestSetup) expectMoveManagementToWorkload() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.bootstrapCluster, c.workloadCluster, gomock.Any(), c.newClusterSpec, gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectNotToMoveManagementToWorkload() {
	c.clusterManager.EXPECT().MoveCAPI(c.ctx, c.bootstrapCluster, c.workloadCluster, gomock.Any(), c.newClusterSpec, gomock.Any()).Times(0)
}

func (c *upgradeTestSetup) expectPauseEKSAControllerReconcile(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().PauseEKSAControllerReconcile(
			c.ctx, expectedCluster, c.currentClusterSpec, c.provider,
		),
	)
}

func (c *upgradeTestSetup) expectResumeEKSAControllerReconcile(expecteCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(
			c.ctx, expecteCluster, c.newClusterSpec, c.provider,
		),
	)
}

func (c *upgradeTestSetup) expectPauseGitOpsKustomization(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.addonManager.EXPECT().PauseGitOpsKustomization(
			c.ctx, expectedCluster, c.newClusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectDatacenterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.newClusterSpec).Return(c.datacenterConfig),
	)
}

func (c *upgradeTestSetup) expectMachineConfigs() {
	gomock.InOrder(
		c.provider.EXPECT().MachineConfigs(c.newClusterSpec).Return(c.machineConfigs),
	)
}

func (c *upgradeTestSetup) expectCreateEKSAResources(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, expectedCluster, c.newClusterSpec, c.datacenterConfig, c.machineConfigs,
		),
	)
}

func (c *upgradeTestSetup) expectUpdateGitEksaSpec() {
	gomock.InOrder(
		c.addonManager.EXPECT().UpdateGitEksaSpec(
			c.ctx, c.newClusterSpec, c.datacenterConfig, c.machineConfigs,
		),
	)
}

func (c *upgradeTestSetup) expectForceReconcileGitRepo(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.addonManager.EXPECT().ForceReconcileGitRepo(
			c.ctx, expectedCluster, c.newClusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectResumeGitOpsKustomization(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.addonManager.EXPECT().ResumeGitOpsKustomization(
			c.ctx, expectedCluster, c.newClusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectVerifyClusterSpecChanged(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.newClusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.newClusterSpec).Return(c.machineConfigs),
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, expectedCluster, c.newClusterSpec, c.datacenterConfig, c.machineConfigs).Return(true, nil),
	)
}

func (c *upgradeTestSetup) expectSaveLogs(expectedWorkloadCluster *types.Cluster) {
	gomock.InOrder(

		c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.bootstrapCluster).Return(nil),
		c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.newClusterSpec, expectedWorkloadCluster),
	)
}

func (c *upgradeTestSetup) run() error {
	// ctx context.Context, workloadCluster *types.Cluster, forceCleanup bool
	return c.workflow.Run(c.ctx, c.newClusterSpec, c.workloadCluster, c.validator, c.forceCleanup)
}

func (c *upgradeTestSetup) expectProviderNoUpgradeNeeded() {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec).Return(false, nil)
}

func (c *upgradeTestSetup) expectProviderUpgradeNeeded() {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec).Return(true, nil)
}

func (c *upgradeTestSetup) expectVerifyClusterSpecNoChanges() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.newClusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.newClusterSpec).Return(c.machineConfigs),
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.workloadCluster, c.newClusterSpec, c.datacenterConfig, c.machineConfigs).Return(false, nil),
	)
}

func (c *upgradeTestSetup) expectPauseEKSAControllerReconcileNotToBeCalled() {
	c.clusterManager.EXPECT().PauseEKSAControllerReconcile(c.ctx, c.workloadCluster, c.newClusterSpec, c.provider).Times(0)
}

func (c *upgradeTestSetup) expectPauseGitOpsKustomizationNotToBeCalled() {
	c.addonManager.EXPECT().PauseGitOpsKustomization(c.ctx, c.workloadCluster, c.newClusterSpec).Times(0)
}

func (c *upgradeTestSetup) expectCreateBootstrapNotToBeCalled() {
	c.provider.EXPECT().BootstrapClusterOpts().Times(0)
	c.bootstrapper.EXPECT().CreateBootstrapCluster(c.ctx, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).Times(0)
	c.clusterManager.EXPECT().InstallCAPI(c.ctx, gomock.Not(gomock.Nil()), c.bootstrapCluster, c.provider).Times(0)
}

func (c *upgradeTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func TestSkipUpgradeRunSuccess(t *testing.T) {
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster)
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecNoChanges()
	test.expectPauseEKSAControllerReconcileNotToBeCalled()
	test.expectPauseGitOpsKustomizationNotToBeCalled()
	test.expectCreateBootstrapNotToBeCalled()

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunSuccess(t *testing.T) {
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster)
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(test.workloadCluster)
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkload(test.workloadCluster)
	test.expectMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.workloadCluster)
	test.expectResumeEKSAControllerReconcile(test.workloadCluster)
	test.expectUpdateGitEksaSpec()
	test.expectForceReconcileGitRepo(test.workloadCluster)
	test.expectResumeGitOpsKustomization(test.workloadCluster)

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunProviderNeedsUpgradeSuccess(t *testing.T) {
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster)
	test.expectProviderUpgradeNeeded()
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkload(test.workloadCluster)
	test.expectMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.workloadCluster)
	test.expectResumeEKSAControllerReconcile(test.workloadCluster)
	test.expectUpdateGitEksaSpec()
	test.expectForceReconcileGitRepo(test.workloadCluster)
	test.expectResumeGitOpsKustomization(test.workloadCluster)

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunFailedUpgrade(t *testing.T) {
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster)
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(test.workloadCluster)
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkloadToReturn(test.workloadCluster, errors.New("failed upgrading"))
	test.expectMoveManagementToWorkload()
	test.expectSaveLogs(test.workloadCluster)

	err := test.run()
	if err == nil {
		t.Fatal("Upgrade.Run() err = nil, want err not nil")
	}
}

func TestUpgradeWorkloadRunSuccess(t *testing.T) {
	test := newUpgradeTest(t)
	test.newClusterSpec.Cluster.SetSelfManaged()

	test.bootstrapCluster.Name = "management-cluster"
	test.bootstrapCluster.ExistingManagement = true
	test.bootstrapCluster.KubeconfigFile = "kubeconfig.yaml"

	test.workloadCluster.KubeconfigFile = "wl-kubeconfig.yaml"

	test.newClusterSpec.ManagementCluster = &types.Cluster{
		Name:               test.bootstrapCluster.Name,
		KubeconfigFile:     "kubeconfig.yaml",
		ExistingManagement: true,
	}

	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.bootstrapCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.bootstrapCluster)
	test.expectUpgradeCoreComponents(test.bootstrapCluster)
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(test.bootstrapCluster)
	test.expectPauseEKSAControllerReconcile(test.bootstrapCluster)
	test.expectPauseGitOpsKustomization(test.bootstrapCluster)
	test.expectNotToCreateBootstrap()
	test.expectNotToMoveManagementToBootstrap()
	test.expectNotToMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectNotToDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.bootstrapCluster)
	test.expectResumeEKSAControllerReconcile(test.bootstrapCluster)
	test.expectUpdateGitEksaSpec()
	test.expectForceReconcileGitRepo(test.bootstrapCluster)
	test.expectResumeGitOpsKustomization(test.bootstrapCluster)
	test.expectUpgradeWorkload(test.bootstrapCluster)

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}
