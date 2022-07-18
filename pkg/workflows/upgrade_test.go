package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	t                  *testing.T
	bootstrapper       *mocks.MockBootstrapper
	clusterManager     *mocks.MockClusterManager
	addonManager       *mocks.MockAddonManager
	provider           *providermocks.MockProvider
	writer             *writermocks.MockFileWriter
	validator          *mocks.MockValidator
	eksdInstaller      *mocks.MockEksdInstaller
	eksdUpgrader       *mocks.MockEksdUpgrader
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
	managementCluster  *types.Cluster
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
	eksdInstaller := mocks.NewMockEksdInstaller(mockCtrl)
	eksdUpgrader := mocks.NewMockEksdUpgrader(mockCtrl)
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	capiUpgrader := mocks.NewMockCAPIManager(mockCtrl)
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	workflow := workflows.NewUpgrade(bootstrapper, provider, capiUpgrader, clusterManager, addonManager, writer, eksdUpgrader, eksdInstaller)

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
		eksdInstaller:    eksdInstaller,
		eksdUpgrader:     eksdUpgrader,
		capiManager:      capiUpgrader,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		newClusterSpec:   test.NewClusterSpec(func(s *cluster.Spec) { s.Cluster.Name = "cluster-name" }),
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func newUpgradeSelfManagedClusterTest(t *testing.T) *upgradeTestSetup {
	tt := newUpgradeTest(t)
	tt.bootstrapCluster = &types.Cluster{
		Name:               "bootstrap",
		ExistingManagement: false,
		KubeconfigFile:     "kubeconfig.yaml",
	}
	tt.managementCluster = tt.workloadCluster
	return tt
}

func newUpgradeManagedClusterTest(t *testing.T) *upgradeTestSetup {
	tt := newUpgradeTest(t)
	tt.managementCluster = &types.Cluster{
		Name:               "management-cluster",
		ExistingManagement: true,
		KubeconfigFile:     "kubeconfig.yaml",
	}
	tt.workloadCluster.KubeconfigFile = "wl-kubeconfig.yaml"

	tt.newClusterSpec.Cluster.SetSelfManaged()
	tt.newClusterSpec.ManagementCluster = tt.managementCluster
	return tt
}

func (c *upgradeTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateUpgradeCluster(c.ctx, gomock.Any(), c.newClusterSpec, c.currentClusterSpec)
	c.provider.EXPECT().Name()
	c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, gomock.Any(), c.newClusterSpec.Cluster.Name).Return(c.currentClusterSpec, nil)
}

func (c *upgradeTestSetup) expectSetupToFail() {
	c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, gomock.Any(), c.newClusterSpec.Cluster.Name).Return(nil, errors.New("failed setup"))
}

func (c *upgradeTestSetup) expectUpdateSecrets(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.provider.EXPECT().UpdateSecrets(c.ctx, expectedCluster).Return(nil),
	)
}

func (c *upgradeTestSetup) expectEnsureEtcdCAPIComponentsExistTask(expectedCluster *types.Cluster) {
	currentSpec := c.currentClusterSpec
	gomock.InOrder(
		c.capiManager.EXPECT().EnsureEtcdProvidersInstallation(c.ctx, expectedCluster, c.provider, currentSpec),
	)
}

func (c *upgradeTestSetup) expectUpgradeCoreComponents(managementCluster *types.Cluster, workloadCluster *types.Cluster) {
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
	eksdChangeDiff := types.NewChangeDiff(&types.ComponentChangeDiff{
		ComponentName: "eks-d",
		OldVersion:    "v0.0.1",
		NewVersion:    "v0.0.2",
	})
	gomock.InOrder(
		c.clusterManager.EXPECT().UpgradeNetworking(c.ctx, workloadCluster, currentSpec, c.newClusterSpec, c.provider).Return(networkingChangeDiff, nil),
		c.capiManager.EXPECT().Upgrade(c.ctx, managementCluster, c.provider, currentSpec, c.newClusterSpec).Return(capiChangeDiff, nil),
		c.addonManager.EXPECT().UpdateLegacyFileStructure(c.ctx, currentSpec, c.newClusterSpec),
		c.addonManager.EXPECT().Upgrade(c.ctx, managementCluster, currentSpec, c.newClusterSpec).Return(fluxChangeDiff, nil),
		c.clusterManager.EXPECT().Upgrade(c.ctx, managementCluster, currentSpec, c.newClusterSpec).Return(eksaChangeDiff, nil),
		c.eksdUpgrader.EXPECT().Upgrade(c.ctx, managementCluster, currentSpec, c.newClusterSpec).Return(eksdChangeDiff, nil),
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

		c.provider.EXPECT().PreCAPIInstallOnBootstrap(c.ctx, c.bootstrapCluster, c.newClusterSpec),
		c.provider.EXPECT().PostBootstrapSetupUpgrade(c.ctx, c.newClusterSpec.Cluster, c.bootstrapCluster),
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

func (c *upgradeTestSetup) expectUpgradeWorkload(managementCluster *types.Cluster, workloadCluster *types.Cluster) {
	c.expectUpgradeWorkloadToReturn(managementCluster, workloadCluster, nil)
	if managementCluster != nil && managementCluster.ExistingManagement {
		c.clusterManager.EXPECT().ApplyBundles(c.ctx, c.newClusterSpec, managementCluster)
	} else {
		c.clusterManager.EXPECT().ApplyBundles(c.ctx, c.newClusterSpec, workloadCluster)
	}
}

func (c *upgradeTestSetup) expectUpgradeWorkloadToReturn(managementCluster *types.Cluster, workloadCluster *types.Cluster, err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().UpgradeCluster(
			c.ctx, managementCluster, workloadCluster, c.newClusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeTestSetup) expectMoveManagementToBootstrap() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.managementCluster, c.bootstrapCluster, gomock.Any(), c.newClusterSpec, gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectNotToMoveManagementToBootstrap() {
	c.clusterManager.EXPECT().MoveCAPI(c.ctx, c.managementCluster, c.bootstrapCluster, gomock.Any(), c.newClusterSpec, gomock.Any()).Times(0)
}

func (c *upgradeTestSetup) expectMoveManagementToWorkload() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.bootstrapCluster, c.managementCluster, gomock.Any(), c.newClusterSpec, gomock.Any(),
		),
	)
}

func (c *upgradeTestSetup) expectNotToMoveManagementToWorkload() {
	c.clusterManager.EXPECT().MoveCAPI(c.ctx, c.bootstrapCluster, c.managementCluster, gomock.Any(), c.newClusterSpec, gomock.Any()).Times(0)
}

func (c *upgradeTestSetup) expectPauseEKSAControllerReconcile(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().PauseEKSAControllerReconcile(
			c.ctx, expectedCluster, c.currentClusterSpec, c.provider,
		),
	)
}

func (c *upgradeTestSetup) expectResumeEKSAControllerReconcile(expectedCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(
			c.ctx, expectedCluster, c.newClusterSpec, c.provider,
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

func (c *upgradeTestSetup) expectInstallEksdManifest(expectedCLuster *types.Cluster) {
	gomock.InOrder(
		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.newClusterSpec, expectedCLuster,
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
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, expectedCluster, c.newClusterSpec).Return(true, nil),
	)
}

func (c *upgradeTestSetup) expectSaveLogs(expectedWorkloadCluster *types.Cluster) {
	gomock.InOrder(
		c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.newClusterSpec, c.bootstrapCluster).Return(nil),
		c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.newClusterSpec, expectedWorkloadCluster),
	)
}

func (c *upgradeTestSetup) expectWriteCheckpointFile() {
	gomock.InOrder(
		c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.newClusterSpec.Cluster.Name), gomock.Any()),
	)
}

func (c *upgradeTestSetup) run() error {
	return c.workflow.Run(c.ctx, c.newClusterSpec, c.managementCluster, c.workloadCluster, c.validator, c.forceCleanup)
}

func (c *upgradeTestSetup) expectProviderNoUpgradeNeeded(expectedCluster *types.Cluster) {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec, expectedCluster).Return(false, nil)
}

func (c *upgradeTestSetup) expectProviderUpgradeNeeded() {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec, c.workloadCluster).Return(true, nil)
}

func (c *upgradeTestSetup) expectVerifyClusterSpecNoChanges() {
	gomock.InOrder(
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.workloadCluster, c.newClusterSpec).Return(false, nil),
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
	test := newUpgradeSelfManagedClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster, test.workloadCluster)
	test.expectProviderNoUpgradeNeeded(test.workloadCluster)
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
	test := newUpgradeSelfManagedClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster, test.workloadCluster)
	test.expectProviderNoUpgradeNeeded(test.workloadCluster)
	test.expectVerifyClusterSpecChanged(test.workloadCluster)
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkload(test.bootstrapCluster, test.workloadCluster)
	test.expectMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.workloadCluster)
	test.expectInstallEksdManifest(test.workloadCluster)
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
	test := newUpgradeSelfManagedClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster, test.workloadCluster)
	test.expectProviderUpgradeNeeded()
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkload(test.bootstrapCluster, test.workloadCluster)
	test.expectMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.workloadCluster)
	test.expectInstallEksdManifest(test.workloadCluster)
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
	test := newUpgradeSelfManagedClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster, test.workloadCluster)
	test.expectProviderNoUpgradeNeeded(test.workloadCluster)
	test.expectVerifyClusterSpecChanged(test.workloadCluster)
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkloadToReturn(test.bootstrapCluster, test.workloadCluster, errors.New("failed upgrading"))
	test.expectSaveLogs(test.workloadCluster)
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("Upgrade.Run() err = nil, want err not nil")
	}
}

func TestUpgradeWorkloadRunSuccess(t *testing.T) {
	test := newUpgradeManagedClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.managementCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.managementCluster)
	test.expectUpgradeCoreComponents(test.managementCluster, test.workloadCluster)
	test.expectProviderNoUpgradeNeeded(test.managementCluster)
	test.expectVerifyClusterSpecChanged(test.managementCluster)
	test.expectPauseEKSAControllerReconcile(test.managementCluster)
	test.expectPauseGitOpsKustomization(test.managementCluster)
	test.expectNotToCreateBootstrap()
	test.expectNotToMoveManagementToBootstrap()
	test.expectNotToMoveManagementToWorkload()
	test.expectWriteClusterConfig()
	test.expectNotToDeleteBootstrap()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateEKSAResources(test.managementCluster)
	test.expectInstallEksdManifest(test.managementCluster)
	test.expectResumeEKSAControllerReconcile(test.managementCluster)
	test.expectUpdateGitEksaSpec()
	test.expectForceReconcileGitRepo(test.managementCluster)
	test.expectResumeGitOpsKustomization(test.managementCluster)
	test.expectUpgradeWorkload(test.managementCluster, test.workloadCluster)

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeWithCheckpointFirstRunFailed(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.CheckpointEnabledEnvVar, "true")

	test := newUpgradeSelfManagedClusterTest(t)
	test.writer.EXPECT().TempDir()
	test.expectSetupToFail()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("Upgrade.Run() err = nil, want err not nil")
	}
}

func TestUpgradeWithCheckpointSecondRunSuccess(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.CheckpointEnabledEnvVar, "true")

	test := newUpgradeSelfManagedClusterTest(t)
	test.writer.EXPECT().TempDir()
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(test.workloadCluster)
	test.expectEnsureEtcdCAPIComponentsExistTask(test.workloadCluster)
	test.expectUpgradeCoreComponents(test.workloadCluster, test.workloadCluster)
	test.expectProviderNoUpgradeNeeded(test.workloadCluster)
	test.expectVerifyClusterSpecChanged(test.workloadCluster)
	test.expectPauseEKSAControllerReconcile(test.workloadCluster)
	test.expectPauseGitOpsKustomization(test.workloadCluster)
	test.expectCreateBootstrap()
	test.expectMoveManagementToBootstrap()
	test.expectUpgradeWorkloadToReturn(test.bootstrapCluster, test.workloadCluster, errors.New("failed upgrading"))
	test.expectSaveLogs(test.workloadCluster)
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("Upgrade.Run() err = nil, want err not nil")
	}

	test2 := newUpgradeSelfManagedClusterTest(t)
	test2.writer.EXPECT().TempDir().Return("testdata")
	test2.expectSetup()
	test2.expectUpgradeWorkload(test2.bootstrapCluster, test2.workloadCluster)
	test2.expectMoveManagementToWorkload()
	test2.expectWriteClusterConfig()
	test2.expectDeleteBootstrap()
	test2.expectDatacenterConfig()
	test2.expectMachineConfigs()
	test2.expectCreateEKSAResources(test2.workloadCluster)
	test2.expectInstallEksdManifest(test2.workloadCluster)
	test2.expectResumeEKSAControllerReconcile(test2.workloadCluster)
	test2.expectUpdateGitEksaSpec()
	test2.expectForceReconcileGitRepo(test2.workloadCluster)
	test2.expectResumeGitOpsKustomization(test2.workloadCluster)

	err = test2.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want nil", err)
	}
}
