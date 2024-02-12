package workload_test

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
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows/workload"
)

type upgradeTestSetup struct {
	t                     *testing.T
	clusterManager        *mocks.MockClusterManager
	gitOpsManager         *mocks.MockGitOpsManager
	provider              *providermocks.MockProvider
	writer                *writermocks.MockFileWriter
	validator             *mocks.MockValidator
	eksd                  *mocks.MockEksdInstaller
	packageInstaller      *mocks.MockPackageInstaller
	clusterUpgrader       *mocks.MockClusterUpgrader
	datacenterConfig      providers.DatacenterConfig
	machineConfigs        []providers.MachineConfig
	ctx                   context.Context
	currentClusterSpec    *cluster.Spec
	clusterSpec           *cluster.Spec
	workloadCluster       *types.Cluster
	workload              *workload.Upgrade
	backupClusterStateDir string
}

func newUpgradeTest(t *testing.T) *upgradeTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	gitOpsManager := mocks.NewMockGitOpsManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	eksd := mocks.NewMockEksdInstaller(mockCtrl)
	packageInstaller := mocks.NewMockPackageInstaller(mockCtrl)
	eksdInstaller := mocks.NewMockEksdInstaller(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	clusterUpgrader := mocks.NewMockClusterUpgrader(mockCtrl)

	validator := mocks.NewMockValidator(mockCtrl)

	workload := workload.NewUpgrade(
		provider,
		clusterManager,
		gitOpsManager,
		writer,
		clusterUpgrader,
		eksdInstaller,
		packageInstaller,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &upgradeTestSetup{
		t:                t,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		eksd:             eksd,
		packageInstaller: packageInstaller,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workload:         workload,
		ctx:              context.Background(),
		clusterUpgrader:  clusterUpgrader,
		currentClusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = "workload"
			s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
			s.ManagementCluster = &types.Cluster{Name: "management"}
			s.Cluster.Spec.KubernetesVersion = v1alpha1.Kube127
		}),
		clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = "workload"
			s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
			s.ManagementCluster = &types.Cluster{Name: "management"}
			s.Cluster.Spec.KubernetesVersion = v1alpha1.Kube128
		}),
		workloadCluster:       &types.Cluster{Name: "workload"},
		backupClusterStateDir: fmt.Sprintf("%s-backup-%s", "workload", time.Now().Format("2006-01-02T15_04_05")),
	}
}

func (c *upgradeTestSetup) expectSetup() {
	c.clusterManager.EXPECT().GetCurrentClusterSpec(c.ctx, c.clusterSpec.ManagementCluster, c.clusterSpec.Cluster.Name).Return(c.currentClusterSpec, nil)
	c.provider.EXPECT().SetupAndValidateUpgradeCluster(c.ctx, c.clusterSpec.ManagementCluster, c.clusterSpec, c.currentClusterSpec)
	c.provider.EXPECT().Name()
	c.gitOpsManager.EXPECT().Validations(c.ctx, c.clusterSpec)
}

func (c *upgradeTestSetup) expectUpgradeWorkloadCluster(err error) {
	c.clusterUpgrader.EXPECT().Run(c.ctx, c.clusterSpec, *c.clusterSpec.ManagementCluster).Return(err)
}

func (c *upgradeTestSetup) expectWriteWorkloadClusterConfig(err error) {
	gomock.InOrder(
		c.writer.EXPECT().Write("workload-eks-a-cluster.yaml", gomock.Any(), gomock.Any()).Return("workload-eks-a-cluster.yaml", err),
	)
}

func (c *upgradeTestSetup) expectDatacenterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig).AnyTimes(),
	)
}

func (c *upgradeTestSetup) expectMachineConfigs() {
	gomock.InOrder(
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs).AnyTimes(),
	)
}

func (c *upgradeTestSetup) run() error {
	return c.workload.Run(c.ctx, c.workloadCluster, c.clusterSpec, c.validator)
}

func (c *upgradeTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func (c *upgradeTestSetup) expectBackupWorkloadFromCluster(err error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().BackupCAPI(c.ctx, c.clusterSpec.ManagementCluster, c.backupClusterStateDir, c.workloadCluster.Name).Return(err),
	)
}

func (c *upgradeTestSetup) expectSaveLogsManagement() {
	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.clusterSpec.ManagementCluster)
	c.expectWrite()
}

func (c *upgradeTestSetup) expectWrite() {
	c.writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
}

func TestUpgradeRunSuccess(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectBackupWorkloadFromCluster(nil)
	test.expectUpgradeWorkloadCluster(nil)
	test.expectWriteWorkloadClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunUpgradeFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectBackupWorkloadFromCluster(nil)
	test.expectUpgradeWorkloadCluster(fmt.Errorf("boom"))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunGetCurrentClusterSpecFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.clusterManager.EXPECT().GetCurrentClusterSpec(test.ctx, test.clusterSpec.ManagementCluster, test.clusterSpec.Cluster.Name).Return(nil, fmt.Errorf("boom"))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunValidateFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.clusterManager.EXPECT().GetCurrentClusterSpec(test.ctx, test.clusterSpec.ManagementCluster, test.clusterSpec.Cluster.Name).AnyTimes().Return(test.currentClusterSpec, nil)
	test.provider.EXPECT().Name().AnyTimes()
	test.gitOpsManager.EXPECT().Validations(test.ctx, test.clusterSpec).AnyTimes()
	test.provider.EXPECT().SetupAndValidateUpgradeCluster(test.ctx, test.clusterSpec.ManagementCluster, test.clusterSpec, test.currentClusterSpec).Return(fmt.Errorf("boom"))
	test.expectPreflightValidationsToPass()
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeWorkloadRunBackupFailed(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectBackupWorkloadFromCluster(errors.New(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}

func TestUpgradeRunWriteClusterConfigFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newUpgradeTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectBackupWorkloadFromCluster(nil)
	test.expectUpgradeWorkloadCluster(nil)
	test.expectWriteWorkloadClusterConfig(fmt.Errorf("boom"))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Upgrade.Run() err = %v, want err = nil", err)
	}
}
