package workload_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

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

type createTestSetup struct {
	t                    *testing.T
	clusterManager       *mocks.MockClusterManager
	gitOpsManager        *mocks.MockGitOpsManager
	provider             *providermocks.MockProvider
	writer               *writermocks.MockFileWriter
	validator            *mocks.MockValidator
	eksd                 *mocks.MockEksdInstaller
	packageInstaller     *mocks.MockPackageInstaller
	clusterCreator       *mocks.MockClusterCreator
	datacenterConfig     providers.DatacenterConfig
	machineConfigs       []providers.MachineConfig
	ctx                  context.Context
	clusterSpec          *cluster.Spec
	workloadCluster      *types.Cluster
	workload             *workload.Create
	managementComponents *cluster.ManagementComponents
}

func newCreateTest(t *testing.T) *createTestSetup {
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
	clusterUpgrader := mocks.NewMockClusterCreator(mockCtrl)

	validator := mocks.NewMockValidator(mockCtrl)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
	})
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	workload := workload.NewCreate(
		provider,
		clusterManager,
		gitOpsManager,
		writer,
		eksdInstaller,
		packageInstaller,
		clusterUpgrader,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &createTestSetup{
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
		clusterCreator:   clusterUpgrader,
		clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = "workload"
			s.Cluster.Namespace = "test-ns"
			s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
			s.ManagementCluster = &types.Cluster{Name: "management"}
		}),
		workloadCluster:      &types.Cluster{Name: "workload"},
		managementComponents: managementComponents,
	}
}

func (c *createTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec)
	c.provider.EXPECT().Name()
	c.gitOpsManager.EXPECT().Validations(c.ctx, c.clusterSpec)
}

func (c *createTestSetup) expectCreateWorkloadCluster(err1, err2 error) {
	c.clusterManager.EXPECT().CreateNamespace(c.ctx, c.clusterSpec.ManagementCluster, c.clusterSpec.Cluster.Namespace).Return(err1)
	c.clusterCreator.EXPECT().CreateSync(c.ctx, c.clusterSpec, c.clusterSpec.ManagementCluster).Return(c.workloadCluster, err2)
}

func (c *createTestSetup) expectWriteWorkloadClusterConfig(err error) {
	gomock.InOrder(
		c.writer.EXPECT().Write("workload-eks-a-cluster.yaml", gomock.Any(), gomock.Any()).Return("workload-eks-a-cluster.yaml", err),
	)
}

func (c *createTestSetup) expectDatacenterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig).AnyTimes(),
	)
}

func (c *createTestSetup) expectMachineConfigs() {
	gomock.InOrder(
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs).AnyTimes(),
	)
}

func (c *createTestSetup) run() error {
	return c.workload.Run(c.ctx, c.clusterSpec, c.validator)
}

func (c *createTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func (c *createTestSetup) expectSaveLogsManagement() {
	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.clusterSpec.ManagementCluster)
	c.expectWrite()
}

func (c *createTestSetup) expectInstallGitOpsManager(err error) {
	c.gitOpsManager.EXPECT().InstallGitOps(
		c.ctx, c.workloadCluster, c.managementComponents, c.clusterSpec, c.datacenterConfig, c.machineConfigs).Return(err)
}

func (c *createTestSetup) expectWrite() {
	c.writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
}

func TestCreateRunSuccess(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateWorkloadCluster(nil, nil)
	test.expectInstallGitOpsManager(nil)
	test.expectWriteWorkloadClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateWorkloadCluster(nil, fmt.Errorf("Failure"))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateNamespaceFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.clusterManager.EXPECT().CreateNamespace(test.ctx, test.clusterSpec.ManagementCluster, test.clusterSpec.Cluster.Namespace).Return(fmt.Errorf(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunValidateFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.provider.EXPECT().Name()
	test.gitOpsManager.EXPECT().Validations(test.ctx, test.clusterSpec)
	test.provider.EXPECT().SetupAndValidateCreateCluster(test.ctx, test.clusterSpec).Return(fmt.Errorf("Failure"))
	test.expectPreflightValidationsToPass()
	test.expectWrite()

	err := test.run()
	if err == nil || !strings.Contains(err.Error(), "validations failed") {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunGitOpsConfigFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateWorkloadCluster(nil, nil)
	test.expectInstallGitOpsManager(fmt.Errorf("Failure"))
	test.expectWriteWorkloadClusterConfig(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunWriteClusterConfigFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectDatacenterConfig()
	test.expectMachineConfigs()
	test.expectCreateWorkloadCluster(nil, nil)
	test.expectInstallGitOpsManager(nil)
	test.expectWriteWorkloadClusterConfig(fmt.Errorf("Failure"))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}
