package workflows_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type upgradeTestSetup struct {
	t                *testing.T
	bootstrapper     *mocks.MockBootstrapper
	clusterManager   *mocks.MockClusterManager
	addonManager     *mocks.MockAddonManager
	provider         *providermocks.MockProvider
	writer           *writermocks.MockFileWriter
	validator        *mocks.MockValidator
	capiUpgrader     *mocks.MockCAPIUpgrader
	datacenterConfig *providermocks.MockDatacenterConfig
	machineConfigs   []providers.MachineConfig
	workflow         *workflows.Upgrade
	ctx              context.Context
	clusterSpec      *cluster.Spec
	forceCleanup     bool
	bootstrapCluster *types.Cluster
	workloadCluster  *types.Cluster
}

func newUpgradeTest(t *testing.T) *upgradeTestSetup {
	os.Setenv(features.ComponentsUpgradesEnvVar, "true")
	t.Cleanup(func() {
		os.Unsetenv(features.ComponentsUpgradesEnvVar)
	})
	mockCtrl := gomock.NewController(t)
	bootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	addonManager := mocks.NewMockAddonManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	validator := mocks.NewMockValidator(mockCtrl)
	datacenterConfig := providermocks.NewMockDatacenterConfig(mockCtrl)
	capiUpgrader := mocks.NewMockCAPIUpgrader(mockCtrl)
	machineConfigs := []providers.MachineConfig{providermocks.NewMockMachineConfig(mockCtrl)}
	workflow := workflows.NewUpgrade(bootstrapper, provider, capiUpgrader, clusterManager, addonManager, writer)

	return &upgradeTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		addonManager:     addonManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		capiUpgrader:     capiUpgrader,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		clusterSpec:      test.NewClusterSpec(func(s *cluster.Spec) { s.Name = "cluster-name" }),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func (c *upgradeTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateUpgradeCluster(c.ctx, c.clusterSpec)
	c.provider.EXPECT().Name()
}

func (c *upgradeTestSetup) expectUpdateSecrets() {
	gomock.InOrder(
		c.provider.EXPECT().UpdateSecrets(c.ctx, c.workloadCluster).Return(nil),
	)
}

func (c *upgradeTestSetup) expectUpgradeCoreComponents() {
	currentSpec := &cluster.Spec{}
	gomock.InOrder(
		c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, c.workloadCluster).Return(currentSpec, nil),
		c.capiUpgrader.EXPECT().Upgrade(c.ctx, c.workloadCluster, c.provider, currentSpec, c.clusterSpec),
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

func (c *upgradeTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig().Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs().Return(c.machineConfigs),
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

func (c *upgradeTestSetup) expectUpgradeWorkload() {
	c.expectUpgradeWorkloadToReturn(nil)
}

func (c *upgradeTestSetup) expectUpgradeWorkloadToReturn(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().UpgradeCluster(
			c.ctx, c.bootstrapCluster, c.workloadCluster, c.clusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeTestSetup) expectMoveManagementToBootstrap() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.workloadCluster, c.bootstrapCluster, gomock.Any(), gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectMoveManagementToWorkload() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.bootstrapCluster, c.workloadCluster, gomock.Any(), gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectPauseEKSAControllerReconcile() {
	gomock.InOrder(
		c.clusterManager.EXPECT().PauseEKSAControllerReconcile(
			c.ctx, c.workloadCluster, c.clusterSpec, c.provider,
		),
	)
}

func (c *upgradeTestSetup) expectResumeEKSAControllerReconcile() {
	gomock.InOrder(
		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(
			c.ctx, c.workloadCluster, c.clusterSpec, c.provider,
		),
	)
}

func (c *upgradeTestSetup) expectPauseGitOpsKustomization() {
	gomock.InOrder(
		c.addonManager.EXPECT().PauseGitOpsKustomization(
			c.ctx, c.workloadCluster, c.clusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectDatacenterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig().Return(c.datacenterConfig),
	)
}

func (c *upgradeTestSetup) expectMachineConfigs() {
	gomock.InOrder(
		c.provider.EXPECT().MachineConfigs().Return(c.machineConfigs),
	)
}

func (c *upgradeTestSetup) expectCreateEKSAResources() {
	gomock.InOrder(
		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),
	)
}

func (c *upgradeTestSetup) expectUpdateGitEksaSpec() {
	gomock.InOrder(
		c.addonManager.EXPECT().UpdateGitEksaSpec(
			c.ctx, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),
	)
}

func (c *upgradeTestSetup) expectForceReconcileGitRepo() {
	gomock.InOrder(
		c.addonManager.EXPECT().ForceReconcileGitRepo(
			c.ctx, c.workloadCluster, c.clusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectUpgradeFluxComponents() {
	currentSpec := &cluster.Spec{}
	gomock.InOrder(
		c.addonManager.EXPECT().Upgrade(
			c.ctx, currentSpec, c.clusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectResumeGitOpsKustomization() {
	gomock.InOrder(
		c.addonManager.EXPECT().ResumeGitOpsKustomization(
			c.ctx, c.workloadCluster, c.clusterSpec,
		),
	)
}

func (c *upgradeTestSetup) expectVerifyClusterSpecChanged() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig().Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs().Return(c.machineConfigs),
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs).Return(true, nil),
	)
}

func (c *upgradeTestSetup) expectSaveLogs() {
	gomock.InOrder(
		c.clusterManager.EXPECT().SaveLogs(c.ctx, c.bootstrapCluster).Return(nil),
	)
}

func (c *upgradeTestSetup) run() error {
	// ctx context.Context, workloadCluster *types.Cluster, forceCleanup bool
	return c.workflow.Run(c.ctx, c.clusterSpec, c.workloadCluster, c.validator, c.forceCleanup)
}

func (c *upgradeTestSetup) expectVerifyClusterSpecNoChanges() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig().Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs().Return(c.machineConfigs),
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs).Return(false, nil),
	)
}

func (c *upgradeTestSetup) expectPauseEKSAControllerReconcileNotToBeCalled() {
	c.clusterManager.EXPECT().PauseEKSAControllerReconcile(c.ctx, c.workloadCluster, c.clusterSpec, c.provider).Times(0)
}

func (c *upgradeTestSetup) expectPauseGitOpsKustomizationNotToBeCalled() {
	c.addonManager.EXPECT().PauseGitOpsKustomization(c.ctx, c.workloadCluster, c.clusterSpec).Times(0)
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
	test.expectUpdateSecrets()
	test.expectUpgradeCoreComponents()
	test.expectUpgradeFluxComponents()
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
	test.expectUpdateSecrets()
	test.expectUpgradeCoreComponents()
	test.expectUpgradeFluxComponents()
	test.expectVerifyClusterSpecChanged()
	test.expectPauseEKSAControllerReconcile()
	test.expectPauseGitOpsKustomization()
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkload()
	test.expectMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources()
	test.expectResumeEKSAControllerReconcile()
	test.expectUpdateGitEksaSpec()
	test.expectForceReconcileGitRepo()
	test.expectResumeGitOpsKustomization()

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunFailedUpgrade(t *testing.T) {
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets()
	test.expectUpgradeCoreComponents()
	test.expectUpgradeFluxComponents()
	test.expectVerifyClusterSpecChanged()
	test.expectPauseEKSAControllerReconcile()
	test.expectPauseGitOpsKustomization()
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkloadToReturn(errors.New("failed upgrading"))
	test.expectMoveManagementToWorkload()
	test.expectSaveLogs()

	err := test.run()
	if err == nil {
		t.Fatal("Upgrade.Run() err = nil, want err not nil")
	}
}
