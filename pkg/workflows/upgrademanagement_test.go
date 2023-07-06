package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type upgradeManagementTestSetup struct {
	t                   *testing.T
	clusterManager      *mocks.MockClusterManager
	gitOpsManager       *mocks.MockGitOpsManager
	provider            *providermocks.MockProvider
	writer              *writermocks.MockFileWriter
	validator           *mocks.MockValidator
	eksdInstaller       *mocks.MockEksdInstaller
	eksdUpgrader        *mocks.MockEksdUpgrader
	capiManager         *mocks.MockCAPIManager
	managementUpgrader  *mocks.MockManagementUpgrader
	datacenterConfig    providers.DatacenterConfig
	machineConfigs      []providers.MachineConfig
	workflow            *workflows.UpgradeManagement
	ctx                 context.Context
	newClusterSpec      *cluster.Spec
	currentClusterSpec  *cluster.Spec
	forceCleanup        bool
	managementCluster   *types.Cluster
	managementStatePath string
}

func newUpgradeManagementTest(t *testing.T) *upgradeManagementTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	gitOpsManager := mocks.NewMockGitOpsManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	validator := mocks.NewMockValidator(mockCtrl)
	eksdInstaller := mocks.NewMockEksdInstaller(mockCtrl)
	eksdUpgrader := mocks.NewMockEksdUpgrader(mockCtrl)
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	capiUpgrader := mocks.NewMockCAPIManager(mockCtrl)
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	managementUpgrader := mocks.NewMockManagementUpgrader(mockCtrl)
	workflow := workflows.NewUpgradeManagement(
		provider,
		capiUpgrader,
		clusterManager,
		gitOpsManager,
		writer,
		eksdUpgrader,
		eksdInstaller,
		managementUpgrader,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &upgradeManagementTestSetup{
		t:                  t,
		clusterManager:     clusterManager,
		gitOpsManager:      gitOpsManager,
		provider:           provider,
		writer:             writer,
		validator:          validator,
		eksdInstaller:      eksdInstaller,
		eksdUpgrader:       eksdUpgrader,
		capiManager:        capiUpgrader,
		managementUpgrader: managementUpgrader,
		datacenterConfig:   datacenterConfig,
		machineConfigs:     machineConfigs,
		workflow:           workflow,
		ctx:                context.Background(),
		newClusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = "management"
			s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
		}),
		managementStatePath: fmt.Sprintf("cluster-state-backup-%s", time.Now().Format("2006-01-02T15_04_05")),
	}
}

func (c *upgradeManagementTestSetup) WithForceCleanup() *upgradeManagementTestSetup {
	c.forceCleanup = true
	return c
}

func newUpgradeManagementClusterTest(t *testing.T) *upgradeManagementTestSetup {
	tt := newUpgradeManagementTest(t)
	tt.managementCluster = &types.Cluster{Name: "management"}
	return tt
}

func (c *upgradeManagementTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateUpgradeCluster(c.ctx, gomock.Any(), c.newClusterSpec, c.currentClusterSpec)
	c.provider.EXPECT().Name()
	c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, gomock.Any(), c.managementCluster.Name).Return(c.currentClusterSpec, nil)
}

func (c *upgradeManagementTestSetup) expectSetupToFail() {
	c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, gomock.Any(), c.managementCluster.Name).Return(nil, errors.New("failed setup"))
}

func (c *upgradeManagementTestSetup) expectUpdateSecrets(err error) {
	gomock.InOrder(
		c.provider.EXPECT().UpdateSecrets(c.ctx, c.managementCluster, c.newClusterSpec).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectEnsureManagementEtcdCAPIComponentsExist(err error) {
	gomock.InOrder(
		c.capiManager.EXPECT().EnsureEtcdProvidersInstallation(c.ctx, c.managementCluster, c.provider, c.currentClusterSpec).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectPauseGitOpsReconcile(err error) {
	gomock.InOrder(
		c.gitOpsManager.EXPECT().PauseClusterResourcesReconcile(
			c.ctx, c.managementCluster, c.newClusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectUpgradeCoreComponents() {
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
		c.provider.EXPECT().PreCoreComponentsUpgrade(gomock.Any(), gomock.Any(), gomock.Any()),
		c.clusterManager.EXPECT().UpgradeNetworking(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec, c.provider).Return(networkingChangeDiff, nil),
		c.capiManager.EXPECT().Upgrade(c.ctx, c.managementCluster, c.provider, currentSpec, c.newClusterSpec).Return(capiChangeDiff, nil),
		c.gitOpsManager.EXPECT().Install(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec).Return(nil),
		c.gitOpsManager.EXPECT().Upgrade(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec).Return(fluxChangeDiff, nil),
		c.clusterManager.EXPECT().Upgrade(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec).Return(eksaChangeDiff, nil),
		c.eksdUpgrader.EXPECT().Upgrade(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec).Return(eksdChangeDiff, nil),
		c.clusterManager.EXPECT().ApplyBundles(c.ctx, c.newClusterSpec, c.managementCluster),
		c.clusterManager.EXPECT().ApplyReleases(c.ctx, c.newClusterSpec, c.managementCluster),
	)
}

func (c *upgradeManagementTestSetup) expectProviderNoUpgradeNeeded() {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec, c.managementCluster).Return(false, nil)
}

func (c *upgradeManagementTestSetup) expectProviderUpgradeNeeded() {
	c.provider.EXPECT().UpgradeNeeded(c.ctx, c.newClusterSpec, c.currentClusterSpec, c.managementCluster).Return(true, nil)
}

func (c *upgradeManagementTestSetup) expectVerifyClusterSpecChanged(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.managementCluster, c.newClusterSpec).Return(true, err),
	)
}

func (c *upgradeManagementTestSetup) expectBackupManagementFromCluster(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().BackupCAPIWaitForInfrastructure(c.ctx, c.managementCluster, c.managementStatePath, c.managementCluster.Name).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectPauseCAPIWorkloadClusters(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().PauseCAPIWorkloadClusters(c.ctx, c.managementCluster).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectDatacenterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.newClusterSpec).Return(c.datacenterConfig).AnyTimes(),
	)
}

func (c *upgradeManagementTestSetup) expectMachineConfigs() {
	gomock.InOrder(
		c.provider.EXPECT().MachineConfigs(c.newClusterSpec).Return(c.machineConfigs).AnyTimes(),
	)
}

func (c *upgradeManagementTestSetup) expectUpgradeManagementEKSAResources(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().UpgradeEKSAResources(
			c.ctx, c.managementCluster, c.newClusterSpec, c.datacenterConfig, c.machineConfigs,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectInstallEksdManifest(err error) {
	gomock.InOrder(
		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.newClusterSpec, c.managementCluster,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectUpgradeManagementCluster(err error) {
	c.managementUpgrader.EXPECT().UpgradeManagementCluster(c.ctx, c.managementCluster).Return(err)
}

func (c *upgradeManagementTestSetup) expectResumeCAPIWorkloadClustersAPI(err error) {
	c.clusterManager.EXPECT().ResumeCAPIWorkloadClusters(c.ctx, c.managementCluster).Return(err)
}

func (c *upgradeManagementTestSetup) expectUpdateGitEksaSpec(err error) {
	gomock.InOrder(
		c.gitOpsManager.EXPECT().UpdateGitEksaSpec(
			c.ctx, c.newClusterSpec, c.datacenterConfig, c.machineConfigs,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectForceReconcileGitRepo(err error) {
	gomock.InOrder(
		c.gitOpsManager.EXPECT().ForceReconcileGitRepo(
			c.ctx, c.managementCluster, c.newClusterSpec,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectResumeGitOpsReconcile(err error) {
	gomock.InOrder(
		c.gitOpsManager.EXPECT().ResumeClusterResourcesReconcile(
			c.ctx, c.managementCluster, c.newClusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectWriteManagementClusterConfig(err error) {
	gomock.InOrder(
		c.writer.EXPECT().Write("management-eks-a-cluster.yaml", gomock.Any(), gomock.Any()).Return("management-eks-a-cluster.yaml", err),
	)
}

func (c *upgradeManagementTestSetup) expectSaveLogs() {
	gomock.InOrder(
		c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.newClusterSpec, c.managementCluster).Return(nil),
	)
}

func (c *upgradeManagementTestSetup) expectWriteCheckpointFile() {
	gomock.InOrder(
		c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.newClusterSpec.Cluster.Name), gomock.Any()),
	)
}

func (c *upgradeManagementTestSetup) run() error {
	return c.workflow.Run(c.ctx, c.newClusterSpec, c.managementCluster, c.validator)
}

func (c *upgradeManagementTestSetup) expectVerifyClusterSpecNoChanges() {
	gomock.InOrder(
		c.clusterManager.EXPECT().EKSAClusterSpecChanged(c.ctx, c.managementCluster, c.newClusterSpec).Return(false, nil),
	)
}

func (c *upgradeManagementTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func TestUpgradeManagementWithCheckpointFirstRunFailed(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")

	test := newUpgradeManagementClusterTest(t)
	test.expectSetupToFail()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunUpdateSecretFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunEnsureETCDFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunPauseGitOpsReconcileUpgradeFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunVerifySpecChangeUpgradeFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedBackup(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectBackupManagementFromCluster(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunPauseWorkloadCAPIFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunUpgradeEKSAFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgrade(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(errors.New("failed upgrading"))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunResumeCAPIWorkloadFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectResumeCAPIWorkloadClustersAPI(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunUpdateGitEksaSpecFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunForceReconcileGitRepoFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunResumeClusterResourcesReconcileFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(errors.New(""))
	test.expectWriteManagementClusterConfig(nil)
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunWriteManagementClusterConfigFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectWriteManagementClusterConfig(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestSkipUpgradeManagementRunSuccess(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecNoChanges()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("UpgradeManagement.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeManagementRunSuccess(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderNoUpgradeNeeded()
	test.expectVerifyClusterSpecChanged(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectWriteManagementClusterConfig(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("UpgradeManagement.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeManagementRunProviderNeedsUpgradeSuccess(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectProviderUpgradeNeeded()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectUpgradeManagementEKSAResources(nil)
	test.expectInstallEksdManifest(nil)
	test.expectUpgradeManagementCluster(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectWriteManagementClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("UpgradeManagement.Run() err = %v, want err = nil", err)
	}
}
