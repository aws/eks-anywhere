package management_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
)

type upgradeManagementTestSetup struct {
	t                           *testing.T
	client                      kubernetes.Client
	clientFactory               *mocks.MockClientFactory
	clusterManager              *mocks.MockClusterManager
	gitOpsManager               *mocks.MockGitOpsManager
	provider                    *providermocks.MockProvider
	writer                      *writermocks.MockFileWriter
	validator                   *mocks.MockValidator
	eksdInstaller               *mocks.MockEksdInstaller
	eksdUpgrader                *mocks.MockEksdUpgrader
	capiManager                 *mocks.MockCAPIManager
	clusterUpgrader             *mocks.MockClusterUpgrader
	datacenterConfig            providers.DatacenterConfig
	machineConfigs              []providers.MachineConfig
	ctx                         context.Context
	newClusterSpec              *cluster.Spec
	currentClusterSpec          *cluster.Spec
	currentManagementComponents *cluster.ManagementComponents
	newManagementComponents     *cluster.ManagementComponents
	managementCluster           *types.Cluster
	managementStatePath         string
	management                  *management.Upgrade
}

func newUpgradeManagementTest(t *testing.T) *upgradeManagementTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	clientFactory := mocks.NewMockClientFactory(mockCtrl)
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
	clusterUpgrader := mocks.NewMockClusterUpgrader(mockCtrl)
	management := management.NewUpgrade(
		clientFactory,
		provider,
		capiUpgrader,
		clusterManager,
		gitOpsManager,
		writer,
		eksdUpgrader,
		eksdInstaller,
		clusterUpgrader,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	currentClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "management"
		s.Cluster.Namespace = "default"
		s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
		s.Cluster.SetManagementComponentsVersion("v0.19.0-dev+latest")
		s.Bundles = test.Bundle()
		s.EKSARelease = test.EKSARelease()
	})

	newClusterSpec := currentClusterSpec.DeepCopy()
	newClusterSpec.Cluster.Annotations = nil

	client := test.NewFakeKubeClient(currentClusterSpec.Cluster, currentClusterSpec.EKSARelease, currentClusterSpec.Bundles)

	return &upgradeManagementTestSetup{
		t:                           t,
		client:                      client,
		clientFactory:               clientFactory,
		clusterManager:              clusterManager,
		gitOpsManager:               gitOpsManager,
		provider:                    provider,
		writer:                      writer,
		validator:                   validator,
		eksdInstaller:               eksdInstaller,
		eksdUpgrader:                eksdUpgrader,
		capiManager:                 capiUpgrader,
		clusterUpgrader:             clusterUpgrader,
		datacenterConfig:            datacenterConfig,
		machineConfigs:              machineConfigs,
		management:                  management,
		ctx:                         context.Background(),
		currentManagementComponents: cluster.ManagementComponentsFromBundles(currentClusterSpec.Bundles),
		newManagementComponents:     cluster.ManagementComponentsFromBundles(newClusterSpec.Bundles),
		currentClusterSpec:          currentClusterSpec,
		newClusterSpec:              newClusterSpec,
		managementStatePath:         fmt.Sprintf("%s-backup-%s", "management", time.Now().Format("2006-01-02T15_04_05")),
	}
}

func newUpgradeManagementClusterTest(t *testing.T) *upgradeManagementTestSetup {
	tt := newUpgradeManagementTest(t)
	tt.managementCluster = &types.Cluster{Name: "management", KubeconfigFile: "kubeconfig"}
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
		c.capiManager.EXPECT().EnsureEtcdProvidersInstallation(c.ctx, c.managementCluster, c.provider, c.currentManagementComponents, c.currentClusterSpec).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectPauseGitOpsReconcile(err error) {
	gomock.InOrder(
		c.gitOpsManager.EXPECT().PauseClusterResourcesReconcile(
			c.ctx, c.managementCluster, c.newClusterSpec, c.provider,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectPreCoreComponentsUpgrade() {
	c.provider.EXPECT().PreCoreComponentsUpgrade(gomock.Any(), gomock.Any(), c.newManagementComponents, gomock.Any())
}

func (c *upgradeManagementTestSetup) expectUpgradeCoreComponents() {
	currentSpec := c.currentClusterSpec
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
		c.provider.EXPECT().PreCoreComponentsUpgrade(gomock.Any(), gomock.Any(), c.newManagementComponents, gomock.Any()),
		c.clientFactory.EXPECT().BuildClientFromKubeconfig(c.managementCluster.KubeconfigFile).Return(c.client, nil),
		c.capiManager.EXPECT().Upgrade(c.ctx, c.managementCluster, c.provider, c.currentManagementComponents, c.newManagementComponents, c.newClusterSpec).Return(capiChangeDiff, nil),
		c.gitOpsManager.EXPECT().Install(c.ctx, c.managementCluster, c.newManagementComponents, currentSpec, c.newClusterSpec).Return(nil),
		c.gitOpsManager.EXPECT().Upgrade(c.ctx, c.managementCluster, c.currentManagementComponents, c.newManagementComponents, currentSpec, c.newClusterSpec).Return(fluxChangeDiff, nil),
		c.clusterManager.EXPECT().Upgrade(c.ctx, c.managementCluster, c.currentManagementComponents, c.newManagementComponents, c.newClusterSpec).Return(eksaChangeDiff, nil),
		c.eksdUpgrader.EXPECT().Upgrade(c.ctx, c.managementCluster, currentSpec, c.newClusterSpec).Return(nil),
	)
}

func (c *upgradeManagementTestSetup) expectBuildClientFromKubeconfig(err error) {
	c.clientFactory.EXPECT().BuildClientFromKubeconfig(c.managementCluster.KubeconfigFile).Return(c.client, err)
}

func (c *upgradeManagementTestSetup) expectBackupManagementFromCluster(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().BackupCAPI(c.ctx, c.managementCluster, c.managementStatePath, "").Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectBackupManagementInfrastructureFromCluster(err error) {
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

func (c *upgradeManagementTestSetup) expectInstallEksdManifest(err error) {
	gomock.InOrder(
		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.newClusterSpec, c.managementCluster,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectApplyBundles(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().ApplyBundles(
			c.ctx, c.newClusterSpec, c.managementCluster,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectApplyReleases(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().ApplyReleases(
			c.ctx, c.newClusterSpec, c.managementCluster,
		).Return(err),
	)
}

func (c *upgradeManagementTestSetup) expectUpgradeManagementCluster() {
	gomock.InOrder(
		c.clusterUpgrader.EXPECT().Run(c.ctx, c.newClusterSpec, *c.managementCluster).Return(nil),
		c.clientFactory.EXPECT().BuildClientFromKubeconfig(c.managementCluster.KubeconfigFile).Return(c.client, nil),
	)
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
	return c.management.Run(c.ctx, c.newClusterSpec, c.managementCluster, c.validator)
}

func (c *upgradeManagementTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func TestUpgradeManagementRunUpdateSetupFailed(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
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

func TestUpgradeManagementRunFailedBackup(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(errors.New(""))
	test.expectBackupManagementInfrastructureFromCluster(errors.New(""))
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgradeBuildClientFromKubeconfig(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectPreCoreComponentsUpgrade()
	test.expectBuildClientFromKubeconfig(errors.New(""))
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgradeGetManagementComponents(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	u := newUpgradeManagementClusterTest(t)
	u.client = test.NewFakeKubeClient()

	u.expectSetup()
	u.expectPreflightValidationsToPass()
	u.expectUpdateSecrets(nil)
	u.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	u.expectPauseGitOpsReconcile(nil)
	u.expectPreCoreComponentsUpgrade()
	u.expectBuildClientFromKubeconfig(nil)
	u.expectDatacenterConfig()
	u.expectMachineConfigs()
	u.expectSaveLogs()
	u.expectWriteCheckpointFile()

	err := u.run()
	g := NewWithT(t)
	g.Expect(err).To(MatchError(ContainSubstring("\"eksa-v0-19-0-dev-plus-latest\" not found")))
}

func TestUpgradeManagementRunFailedUpgradeClientGet(t *testing.T) {
	c := newUpgradeManagementClusterTest(t)
	c.client = test.NewFakeKubeClient(test.Bundle(), test.EKSARelease())

	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	c.expectSetup()
	c.expectPreflightValidationsToPass()
	c.expectUpdateSecrets(nil)
	c.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	c.expectPauseGitOpsReconcile(nil)
	c.expectUpgradeCoreComponents()
	c.expectDatacenterConfig()
	c.expectMachineConfigs()
	c.expectSaveLogs()
	c.expectWriteCheckpointFile()

	err := c.run()
	g := NewWithT(t)
	g.Expect(err).To(MatchError(ContainSubstring("\"management\" not found")))
}

func TestUpgradeManagementRunFailedUpgradeInstallEksd(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectInstallEksdManifest(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgradeApplyBundles(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectApplyBundles(errors.New(""))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgradeApplyReleases(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectApplyBundles(nil)
	test.expectApplyReleases(errors.New(""))
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.clusterUpgrader.EXPECT().Run(test.ctx, test.newClusterSpec, *test.managementCluster).Return(errors.New("failed upgrading"))
	test.expectSaveLogs()
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunFailedUpgradeClusterBuildClientFromKubeconfig(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectPauseGitOpsReconcile(nil)
	test.expectUpgradeCoreComponents()
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.clusterUpgrader.EXPECT().Run(test.ctx, test.newClusterSpec, *test.managementCluster).Return(errors.New("failed upgrading"))
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectWriteManagementClusterConfig(nil)
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
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
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(errors.New(""))
	test.expectWriteManagementClusterConfig(nil)
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectWriteCheckpointFile()

	err := test.run()
	if err == nil {
		t.Fatal("UpgradeManagement.Run() err = nil, want err not nil")
	}
}

func TestUpgradeManagementRunSuccess(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectWriteManagementClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("UpgradeManagement.Run() err = %v, want err = nil", err)
	}
}

func TestTinkerbellUpgradeManagementRunSuccess(t *testing.T) {
	os.Unsetenv(features.CheckpointEnabledEnvVar)
	features.ClearCache()
	test := newUpgradeManagementClusterTest(t)
	test.newClusterSpec.Cluster.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
	test.newClusterSpec.TinkerbellDatacenter = &v1alpha1.TinkerbellDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-datacenter",
			Namespace: "default",
		},
	}
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectUpdateSecrets(nil)
	test.expectEnsureManagementEtcdCAPIComponentsExist(nil)
	test.expectUpgradeCoreComponents()
	test.expectPauseGitOpsReconcile(nil)
	test.expectBackupManagementFromCluster(nil)
	test.expectPauseCAPIWorkloadClusters(nil)
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectInstallEksdManifest(nil)
	test.expectApplyBundles(nil)
	test.expectApplyReleases(nil)
	test.expectUpgradeManagementCluster()
	test.expectResumeCAPIWorkloadClustersAPI(nil)
	test.expectUpdateGitEksaSpec(nil)
	test.expectForceReconcileGitRepo(nil)
	test.expectResumeGitOpsReconcile(nil)
	test.expectWriteManagementClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("UpgradeManagement.Run() err = %v, want err = nil", err)
	}
}
