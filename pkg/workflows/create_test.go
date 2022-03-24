package workflows_test

import (
	"context"
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

type createTestSetup struct {
	t                *testing.T
	bootstrapper     *mocks.MockBootstrapper
	clusterManager   *mocks.MockClusterManager
	addonManager     *mocks.MockAddonManager
	provider         *providermocks.MockProvider
	writer           *writermocks.MockFileWriter
	validator        *mocks.MockValidator
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
	workflow         *workflows.Create
	ctx              context.Context
	clusterSpec      *cluster.Spec
	forceCleanup     bool
	bootstrapCluster *types.Cluster
	workloadCluster  *types.Cluster
}

func newCreateTest(t *testing.T) *createTestSetup {
	mockCtrl := gomock.NewController(t)
	bootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	addonManager := mocks.NewMockAddonManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	workflow := workflows.NewCreate(bootstrapper, provider, clusterManager, addonManager, writer)
	validator := mocks.NewMockValidator(mockCtrl)

	return &createTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		addonManager:     addonManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		clusterSpec:      test.NewClusterSpec(func(s *cluster.Spec) { s.Cluster.Name = "cluster-name"; s.Cluster.Annotations = map[string]string{} }),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func (c *createTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec)
	c.provider.EXPECT().Name()
	c.addonManager.EXPECT().Validations(c.ctx, c.clusterSpec)
}

func (c *createTestSetup) expectCreateBootstrap() {
	opts := []bootstrapper.BootstrapClusterOption{
		bootstrapper.WithDefaultCNIDisabled(), bootstrapper.WithExtraDockerMounts(),
	}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts().Return(opts, nil),
		// Checking for not nil because in go you can't compare closures
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, c.clusterSpec, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil()),
		).Return(c.bootstrapCluster, nil),

		c.clusterManager.EXPECT().InstallCAPI(c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.provider.EXPECT().PostBootstrapSetup(c.ctx, c.clusterSpec.Cluster, c.bootstrapCluster),
	)
}

func (c *createTestSetup) expectCreateWorkload() {
	gomock.InOrder(
		c.clusterManager.EXPECT().CreateWorkloadCluster(
			c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider,
		).Return(c.workloadCluster, nil),

		c.clusterManager.EXPECT().InstallNetworking(
			c.ctx, c.workloadCluster, c.clusterSpec, c.provider,
		),
		c.clusterManager.EXPECT().InstallStorageClass(
			c.ctx, c.workloadCluster, c.provider,
		),
		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider,
		),
		c.provider.EXPECT().UpdateSecrets(c.ctx, c.workloadCluster),
	)
}

func (c *createTestSetup) expectCreateWorkloadSkipCAPI() {
	gomock.InOrder(
		c.clusterManager.EXPECT().CreateWorkloadCluster(
			c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider,
		).Return(c.workloadCluster, nil),

		c.clusterManager.EXPECT().InstallNetworking(
			c.ctx, c.workloadCluster, c.clusterSpec, c.provider,
		),
		c.clusterManager.EXPECT().InstallStorageClass(
			c.ctx, c.workloadCluster, c.provider,
		),
	)
	c.clusterManager.EXPECT().InstallCAPI(
		c.ctx, c.clusterSpec, c.workloadCluster, c.provider,
	).Times(0)
	c.provider.EXPECT().UpdateSecrets(c.ctx, c.workloadCluster).Times(0)
}

func (c *createTestSetup) expectMoveManagement() {
	c.clusterManager.EXPECT().MoveCAPI(
		c.ctx, c.bootstrapCluster, c.workloadCluster, c.workloadCluster.Name, c.clusterSpec, gomock.Any(),
	)
}

func (c *createTestSetup) skipMoveManagement() {
	c.clusterManager.EXPECT().MoveCAPI(
		c.ctx, c.bootstrapCluster, c.workloadCluster, gomock.Any(), c.clusterSpec,
	).Times(0)
}

func (c *createTestSetup) expectInstallEksaComponents() {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.workloadCluster),

		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),

		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(c.ctx, c.workloadCluster, c.clusterSpec, c.provider),
	)
}

func (c *createTestSetup) skipInstallEksaComponents() {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.workloadCluster).Times(0),

		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, c.bootstrapCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),

		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider),
	)
}

func (c *createTestSetup) expectInstallAddonManager() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.addonManager.EXPECT().InstallGitOps(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs),
	)
}

func (c *createTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),
		c.writer.EXPECT().Write("cluster-name-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),
	)
}

func (c *createTestSetup) expectDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any())
}

func (c *createTestSetup) expectNotDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any()).Times(0)
}

func (c *createTestSetup) expectInstallMHC() {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallMachineHealthChecks(
			c.ctx, c.bootstrapCluster, c.provider,
		),
	)
}

func (c *createTestSetup) run() error {
	return c.workflow.Run(c.ctx, c.clusterSpec, c.validator, c.forceCleanup)
}

func (c *createTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx).Return(nil)
}

func TestCreateRunSuccess(t *testing.T) {
	test := newCreateTest(t)

	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCreateWorkload()
	test.expectMoveManagement()
	test.expectInstallEksaComponents()
	test.expectInstallAddonManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectInstallMHC()
	test.expectPreflightValidationsToPass()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunSuccessForceCleanup(t *testing.T) {
	test := newCreateTest(t)
	test.forceCleanup = true
	test.bootstrapper.EXPECT().DeleteBootstrapCluster(test.ctx, &types.Cluster{Name: "cluster-name"}, gomock.Any())
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCreateWorkload()
	test.expectMoveManagement()
	test.expectInstallEksaComponents()
	test.expectInstallAddonManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectInstallMHC()
	test.expectPreflightValidationsToPass()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateWorkloadClusterRunSuccess(t *testing.T) {
	managementKubeconfig := "test.kubeconfig"
	test := newCreateTest(t)

	test.bootstrapCluster.ExistingManagement = true
	test.bootstrapCluster.KubeconfigFile = managementKubeconfig
	test.bootstrapCluster.Name = "cluster-name"

	test.clusterSpec.ManagementCluster = &types.Cluster{
		Name:               test.bootstrapCluster.Name,
		KubeconfigFile:     managementKubeconfig,
		ExistingManagement: true,
	}

	test.expectSetup()
	test.expectCreateWorkloadSkipCAPI()
	test.skipMoveManagement()
	test.skipInstallEksaComponents()
	test.expectInstallAddonManager()
	test.expectWriteClusterConfig()
	test.expectNotDeleteBootstrap()
	test.expectInstallMHC()
	test.expectPreflightValidationsToPass()

	if err := test.run(); err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}
