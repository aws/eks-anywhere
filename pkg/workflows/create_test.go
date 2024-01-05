package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type createTestSetup struct {
	t                *testing.T
	packageInstaller *mocks.MockPackageInstaller
	bootstrapper     *mocks.MockBootstrapper
	clusterManager   *mocks.MockClusterManager
	gitOpsManager    *mocks.MockGitOpsManager
	provider         *providermocks.MockProvider
	writer           *writermocks.MockFileWriter
	validator        *mocks.MockValidator
	eksd             *mocks.MockEksdInstaller
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
	gitOpsManager := mocks.NewMockGitOpsManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	eksd := mocks.NewMockEksdInstaller(mockCtrl)
	packageInstaller := mocks.NewMockPackageInstaller(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	workflow := workflows.NewCreate(bootstrapper, provider, clusterManager, gitOpsManager, writer, eksd, packageInstaller)
	validator := mocks.NewMockValidator(mockCtrl)

	return &createTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		eksd:             eksd,
		packageInstaller: packageInstaller,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		clusterSpec:      test.NewClusterSpec(func(s *cluster.Spec) { s.Cluster.Name = "cluster-name"; s.Cluster.Annotations = map[string]string{} }),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func newCreateWorkloadClusterTest(t *testing.T) *createTestSetup {
	tt := newCreateTest(t)

	tt.bootstrapCluster = &types.Cluster{
		Name:           "management-cluster",
		KubeconfigFile: "kubeconfig.yaml",
	}

	tt.workloadCluster = &types.Cluster{
		Name:           "workload-cluster",
		KubeconfigFile: "wl-kubeconfig.yaml",
	}

	tt.clusterSpec.Cluster.Name = tt.workloadCluster.Name
	tt.clusterSpec.Cluster.SetManagedBy(tt.bootstrapCluster.Name)
	tt.clusterSpec.ManagementCluster = tt.bootstrapCluster

	return tt
}

func (c *createTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec)
	c.provider.EXPECT().Name()
	c.gitOpsManager.EXPECT().Validations(c.ctx, c.clusterSpec)
}

func (c *createTestSetup) expectCreateBootstrap() {
	opts := []bootstrapper.BootstrapClusterOption{bootstrapper.WithExtraDockerMounts()}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts(c.clusterSpec).Return(opts, nil),
		// Checking for not nil because in go you can't compare closures
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, c.clusterSpec, gomock.Not(gomock.Nil()),
		).Return(c.bootstrapCluster, nil),

		c.provider.EXPECT().PreCAPIInstallOnBootstrap(c.ctx, c.bootstrapCluster, c.clusterSpec),

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
		c.clusterManager.EXPECT().InstallMachineHealthChecks(
			c.ctx, c.clusterSpec, c.bootstrapCluster,
		),
		c.clusterManager.EXPECT().RunPostCreateWorkloadCluster(
			c.ctx, c.bootstrapCluster, c.workloadCluster, c.clusterSpec,
		),
		c.clusterManager.EXPECT().CreateEKSANamespace(
			c.ctx, c.workloadCluster,
		),
		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider,
		),
		c.provider.EXPECT().UpdateSecrets(c.ctx, c.workloadCluster, c.clusterSpec),
	)
}

func (c *createTestSetup) expectInstallResourcesOnManagementTask() {
	gomock.InOrder(
		c.provider.EXPECT().PostWorkloadInit(c.ctx, c.workloadCluster, c.clusterSpec),
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
		c.clusterManager.EXPECT().InstallMachineHealthChecks(
			c.ctx, c.clusterSpec, c.bootstrapCluster,
		),
		c.clusterManager.EXPECT().RunPostCreateWorkloadCluster(
			c.ctx, c.bootstrapCluster, c.workloadCluster, c.clusterSpec,
		),
	)
	c.clusterManager.EXPECT().InstallCAPI(
		c.ctx, c.clusterSpec, c.workloadCluster, c.provider,
	).Times(0)
	c.provider.EXPECT().UpdateSecrets(c.ctx, c.workloadCluster, c.clusterSpec).Times(0)
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
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider),

		c.eksd.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.workloadCluster),

		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),

		c.eksd.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.workloadCluster),

		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(c.ctx, c.workloadCluster, c.clusterSpec, c.provider),
	)
}

func (c *createTestSetup) skipInstallEksaComponents() {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider).Times(0),

		c.eksd.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.workloadCluster).Times(0),

		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.clusterManager.EXPECT().CreateEKSAResources(
			c.ctx, c.bootstrapCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs,
		),

		c.eksd.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.bootstrapCluster),

		c.clusterManager.EXPECT().ResumeEKSAControllerReconcile(c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider),
	)
}

func (c *createTestSetup) expectCuratedPackagesInstallation() {
	c.packageInstaller.EXPECT().InstallCuratedPackages(c.ctx).Times(1)
}

func (c *createTestSetup) expectInstallGitOpsManager() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),

		c.gitOpsManager.EXPECT().InstallGitOps(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs),
	)
}

func (c *createTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(c.clusterSpec).Return(c.datacenterConfig),
		c.provider.EXPECT().MachineConfigs(c.clusterSpec).Return(c.machineConfigs),
		c.writer.EXPECT().Write(c.clusterSpec.Cluster.Name+"-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),
	)
}

func (c *createTestSetup) expectDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any(), gomock.Any())
}

func (c *createTestSetup) expectNotDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any(), gomock.Any()).Times(0)
}

func (c *createTestSetup) expectInstallMHC() {
	c.clusterManager.EXPECT().InstallMachineHealthChecks(
		c.ctx, c.clusterSpec, c.bootstrapCluster,
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
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponents()
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunAWSIamConfigFail(t *testing.T) {
	wantError := errors.New("test error")
	test := newCreateTest(t)

	// Adding AWSIAMConfig to cluster spec.
	test.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.provider.EXPECT().BootstrapClusterOpts(test.clusterSpec).Return([]bootstrapper.BootstrapClusterOption{bootstrapper.WithExtraDockerMounts()}, nil)
	test.bootstrapper.EXPECT().CreateBootstrapCluster(test.ctx, test.clusterSpec, gomock.Not(gomock.Nil())).Return(test.bootstrapCluster, nil)
	test.provider.EXPECT().PreCAPIInstallOnBootstrap(test.ctx, test.bootstrapCluster, test.clusterSpec)
	test.clusterManager.EXPECT().InstallCAPI(test.ctx, test.clusterSpec, test.bootstrapCluster, test.provider)
	test.clusterManager.EXPECT().CreateAwsIamAuthCaSecret(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Name).Return(wantError)
	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	if err := test.run(); err == nil {
		t.Fatalf("Create.Run() err = %v, want err = %v", err, wantError)
	}
}

func TestCreateRunAWSIamConfigSuccess(t *testing.T) {
	test := newCreateTest(t)

	// Adding AWSIAMConfig to cluster spec.
	test.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}
	test.clusterManager.EXPECT().CreateAwsIamAuthCaSecret(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Name)
	test.clusterManager.EXPECT().InstallAwsIamAuth(test.ctx, test.bootstrapCluster, test.workloadCluster, test.clusterSpec)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponents()
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	// test.expectInstallMHC()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunSuccessForceCleanup(t *testing.T) {
	test := newCreateTest(t)
	test.forceCleanup = true
	test.bootstrapper.EXPECT().DeleteBootstrapCluster(test.ctx, &types.Cluster{Name: "cluster-name"}, gomock.Any(), gomock.Any())
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponents()
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateWorkloadClusterRunSuccess(t *testing.T) {
	test := newCreateWorkloadClusterTest(t)

	test.expectSetup()
	test.expectCreateWorkloadSkipCAPI()
	test.skipMoveManagement()
	test.skipInstallEksaComponents()
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectNotDeleteBootstrap()
	// test.expectInstallMHC()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	if err := test.run(); err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateWorkloadClusterRunAWSIamConfigSuccess(t *testing.T) {
	test := newCreateWorkloadClusterTest(t)

	// Adding AWSIAMConfig to cluster spec.
	test.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}
	test.clusterManager.EXPECT().CreateAwsIamAuthCaSecret(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Name)
	test.clusterManager.EXPECT().InstallAwsIamAuth(test.ctx, test.bootstrapCluster, test.workloadCluster, test.clusterSpec)
	test.expectSetup()
	test.expectCreateWorkloadSkipCAPI()
	test.skipMoveManagement()
	test.skipInstallEksaComponents()
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectNotDeleteBootstrap()
	// test.expectInstallMHC()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	if err := test.run(); err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateWorkloadClusterRunAWSIamConfigFail(t *testing.T) {
	wantError := errors.New("test error")
	test := newCreateWorkloadClusterTest(t)

	// Adding AWSIAMConfig to cluster spec.
	test.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.clusterManager.EXPECT().CreateAwsIamAuthCaSecret(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Name).Return(wantError)
	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	if err := test.run(); err == nil {
		t.Fatalf("Create.Run() err = %v, want err = %v", err, wantError)
	}
}

func TestCreateWorkloadClusterTaskCreateWorkloadClusterFailure(t *testing.T) {
	test := newCreateWorkloadClusterTest(t)
	commandContext := task.CommandContext{
		BootstrapCluster: test.bootstrapCluster,
		ClusterSpec:      test.clusterSpec,
		Provider:         test.provider,
		ClusterManager:   test.clusterManager,
	}

	gomock.InOrder(
		test.clusterManager.EXPECT().CreateWorkloadCluster(
			test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider,
		).Return(nil, errors.New("test")),
		test.clusterManager.EXPECT().SaveLogsManagementCluster(
			test.ctx, test.clusterSpec, test.bootstrapCluster,
		),
		test.clusterManager.EXPECT().SaveLogsWorkloadCluster(
			test.ctx, test.provider, test.clusterSpec, nil,
		),
		test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any()),
	)
	err := task.NewTaskRunner(&workflows.CreateWorkloadClusterTask{}, test.writer).RunTask(test.ctx, &commandContext)
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateWorkloadClusterTaskRunPostCreateWorkloadClusterFailure(t *testing.T) {
	test := newCreateWorkloadClusterTest(t)
	commandContext := task.CommandContext{
		BootstrapCluster: test.bootstrapCluster,
		ClusterSpec:      test.clusterSpec,
		Provider:         test.provider,
		ClusterManager:   test.clusterManager,
	}

	gomock.InOrder(
		test.clusterManager.EXPECT().CreateWorkloadCluster(
			test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider,
		).Return(test.workloadCluster, nil),
		test.clusterManager.EXPECT().InstallNetworking(
			test.ctx, test.workloadCluster, test.clusterSpec, test.provider,
		),
		test.clusterManager.EXPECT().InstallMachineHealthChecks(
			test.ctx, test.clusterSpec, test.bootstrapCluster,
		),
		test.clusterManager.EXPECT().RunPostCreateWorkloadCluster(
			test.ctx, test.bootstrapCluster, test.workloadCluster, test.clusterSpec,
		).Return(errors.New("test")),
		test.clusterManager.EXPECT().SaveLogsManagementCluster(
			test.ctx, test.clusterSpec, test.bootstrapCluster,
		),
		test.clusterManager.EXPECT().SaveLogsWorkloadCluster(
			test.ctx, test.provider, test.clusterSpec, test.workloadCluster,
		),
		test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any()),
	)
	err := task.NewTaskRunner(&workflows.CreateWorkloadClusterTask{}, test.writer).RunTask(test.ctx, &commandContext)
	if err == nil {
		t.Fatalf("expected error from task")
	}
}
