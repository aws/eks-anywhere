package management_test

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
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
)

type createTestSetup struct {
	t                *testing.T
	packageInstaller *mocks.MockPackageInstaller
	clusterManager   *mocks.MockClusterManager
	bootstrapper     *mocks.MockBootstrapper
	gitOpsManager    *mocks.MockGitOpsManager
	provider         *providermocks.MockProvider
	writer           *writermocks.MockFileWriter
	validator        *mocks.MockValidator
	eksdInstaller    *mocks.MockEksdInstaller
	clusterCreator   *mocks.MockClusterCreator
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
	ctx              context.Context
	clusterSpec      *cluster.Spec
	bootstrapCluster *types.Cluster
	workloadCluster  *types.Cluster
	workflow         *management.Create
}

func newCreateTest(t *testing.T) *createTestSetup {
	featureEnvVars := []string{}
	featureEnvVars = append(featureEnvVars, features.UseControllerForCli)
	mockCtrl := gomock.NewController(t)
	bootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	gitOpsManager := mocks.NewMockGitOpsManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	eksdInstaller := mocks.NewMockEksdInstaller(mockCtrl)
	packageInstaller := mocks.NewMockPackageInstaller(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	clusterCreator := mocks.NewMockClusterCreator(mockCtrl)
	validator := mocks.NewMockValidator(mockCtrl)

	workflow := management.NewCreate(
		bootstrapper,
		provider,
		clusterManager,
		gitOpsManager,
		writer,
		eksdInstaller,
		packageInstaller,
		clusterCreator,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &createTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		eksdInstaller:    eksdInstaller,
		packageInstaller: packageInstaller,
		clusterCreator:   clusterCreator,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
		clusterSpec:      test.NewClusterSpec(func(s *cluster.Spec) { s.Cluster.Name = "cluster-name" }),
	}
}

func (c *createTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec)
	c.provider.EXPECT().Name()
	c.gitOpsManager.EXPECT().Validations(c.ctx, c.clusterSpec)
}

func (c *createTestSetup) expectCreateBootstrap() {
	opts := []bootstrapper.BootstrapClusterOption{bootstrapper.WithExtraDockerMounts()}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts(
			c.clusterSpec).Return(opts, nil),
		// Checking for not nil because in go you can't compare closures
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, c.clusterSpec, gomock.Not(gomock.Nil()),
		).Return(c.bootstrapCluster, nil),
	)
}

func (c *createTestSetup) expectCAPIInstall() {
	gomock.InOrder(
		c.provider.EXPECT().PreCAPIInstallOnBootstrap(
			c.ctx, c.bootstrapCluster, c.clusterSpec),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.provider.EXPECT().PostCAPIInstallSetup(
			c.ctx, c.clusterSpec.Cluster, c.bootstrapCluster),
	)
}

func (c *createTestSetup) expectInstallEksaComponentsBootstrap() {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster),

		c.clusterManager.EXPECT().CreateEKSAReleaseBundle(
			c.ctx, c.bootstrapCluster, c.clusterSpec),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.bootstrapCluster),
	)
}

func (c *createTestSetup) expectInstallEksaComponentsWorkload(err1, err2, err3, err4, err5 error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider).Return(err1),

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.workloadCluster).Return(err2),

		c.clusterManager.EXPECT().CreateEKSAReleaseBundle(
			c.ctx, c.workloadCluster, c.clusterSpec).Return(err3),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.workloadCluster).Return(err4),

		c.clusterCreator.EXPECT().Run(
			c.ctx, c.clusterSpec, *c.workloadCluster).Return(err5),
	)
}

func (c *createTestSetup) expectInstallGitOpsManager() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(
			c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(
			c.clusterSpec).Return(c.machineConfigs),

		c.gitOpsManager.EXPECT().InstallGitOps(
			c.ctx, c.workloadCluster, c.clusterSpec, c.datacenterConfig, c.machineConfigs),
	)
}

func (c *createTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(
			c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(
			c.clusterSpec).Return(c.machineConfigs),

		c.writer.EXPECT().Write(
			"cluster-name-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),
	)
}

func (c *createTestSetup) expectDeleteBootstrap() {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any(), gomock.Any())
}

func (c *createTestSetup) run() error {
	return c.workflow.Run(c.ctx, c.clusterSpec, c.validator)
}

func (c *createTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx)
}

func (c *createTestSetup) expectInstallResourcesOnManagementTask() {
	gomock.InOrder(
		c.provider.EXPECT().PostWorkloadInit(c.ctx, c.workloadCluster, c.clusterSpec),
	)
}

func (c *createTestSetup) expectMoveManagement() {
	c.clusterManager.EXPECT().MoveCAPI(
		c.ctx, c.bootstrapCluster, c.workloadCluster, c.workloadCluster.Name, c.clusterSpec, gomock.Any(),
	)
}

func (c *createTestSetup) expectCreateWorkload() {
	gomock.InOrder(
		c.clusterCreator.EXPECT().Run(
			c.ctx, c.clusterSpec, *c.bootstrapCluster),

		c.clusterManager.EXPECT().GetWorkloadCluster(
			c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider).Return(c.workloadCluster, nil),

		c.clusterManager.EXPECT().CreateEKSANamespace(
			c.ctx, c.workloadCluster),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.workloadCluster, c.provider),

		c.provider.EXPECT().UpdateSecrets(
			c.ctx, c.workloadCluster, c.clusterSpec),
	)
}

func (c *createTestSetup) expectCuratedPackagesInstallation() {
	c.packageInstaller.EXPECT().InstallCuratedPackages(c.ctx).Times(1)
}

func TestCreateRunSuccess(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap()
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

func TestCreateBootstrapOptsFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectPreflightValidationsToPass()

	err := errors.New("test")

	opts := []bootstrapper.BootstrapClusterOption{}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts(
			c.clusterSpec).Return(opts, err),
	)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateValidationsFailure(t *testing.T) {
	c := newCreateTest(t)
	err := errors.New("test")

	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec).Return(err)
	c.provider.EXPECT().Name()
	c.gitOpsManager.EXPECT().Validations(c.ctx, c.clusterSpec)

	c.validator.EXPECT().PreflightValidations(c.ctx)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateBootstrapFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectPreflightValidationsToPass()

	err := errors.New("test")

	opts := []bootstrapper.BootstrapClusterOption{}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts(
			c.clusterSpec).Return(opts, nil),
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, c.clusterSpec, gomock.Not(gomock.Nil()),
		).Return(nil, err),
	)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreatePreCAPIFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()

	c.provider.EXPECT().PreCAPIInstallOnBootstrap(
		c.ctx, c.bootstrapCluster, c.clusterSpec).Return(errors.New("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallCAPIFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()

	gomock.InOrder(
		c.provider.EXPECT().PreCAPIInstallOnBootstrap(
			c.ctx, c.bootstrapCluster, c.clusterSpec),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider).Return(errors.New("test")),
	)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreatePostCAPIFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()

	gomock.InOrder(
		c.provider.EXPECT().PreCAPIInstallOnBootstrap(
			c.ctx, c.bootstrapCluster, c.clusterSpec),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.provider.EXPECT().PostCAPIInstallSetup(
			c.ctx, c.clusterSpec.Cluster, c.bootstrapCluster).Return(errors.New("test")),
	)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreatePostWorkloadInitFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall()
	c.expectInstallEksaComponentsBootstrap()
	c.expectCreateWorkload()

	c.provider.EXPECT().PostWorkloadInit(
		c.ctx, c.workloadCluster, c.clusterSpec).Return(errors.New("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallComponentsFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall()

	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster).Return(errors.New("test")),
	)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateEKSAReleaseBundleFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall()

	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster),

		c.clusterManager.EXPECT().CreateEKSAReleaseBundle(
			c.ctx, c.bootstrapCluster, c.clusterSpec).Return(errors.New("test")),
	)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallEksdManifestFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall()

	gomock.InOrder(
		c.clusterManager.EXPECT().InstallCustomComponents(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider),

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster),

		c.clusterManager.EXPECT().CreateEKSAReleaseBundle(
			c.ctx, c.bootstrapCluster, c.clusterSpec),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.bootstrapCluster).Return(errors.New("test")),
	)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateMoveCAPIFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall()
	c.expectInstallEksaComponentsBootstrap()
	c.expectCreateWorkload()
	c.expectInstallResourcesOnManagementTask()

	c.clusterManager.EXPECT().MoveCAPI(
		c.ctx, c.bootstrapCluster, c.workloadCluster, "workload", c.clusterSpec, gomock.Any()).Return(errors.New("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallEKSAWorkloadFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()

	test.clusterManager.EXPECT().InstallCustomComponents(
		test.ctx, test.clusterSpec, test.workloadCluster, test.provider).Return(errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() error expected")
	}
}

func TestCreateInstallEKSACustomComponentsFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectCAPIInstall()

	err := errors.New("test")
	c.clusterManager.EXPECT().InstallCustomComponents(
		c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider).Return(err)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(
		c.ctx, c.clusterSpec, c.bootstrapCluster,
	)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(
		c.ctx, c.provider, c.clusterSpec, nil,
	)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())
	c.expectPreflightValidationsToPass()

	err = c.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateCreatorWorkloadFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()

	err := errors.New("test")
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, err)

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err = test.run()
	if err == nil {
		t.Fatalf("Create.Run() error expected")
	}
}

func TestCreateRunCreatorFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	test.clusterCreator.EXPECT().Run(
		test.ctx, test.clusterSpec, *test.bootstrapCluster).Return(errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateGetWorkloadClusterFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	test.clusterCreator.EXPECT().Run(
		test.ctx, test.clusterSpec, *test.bootstrapCluster)

	test.clusterManager.EXPECT().GetWorkloadCluster(
		test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider).Return(nil, errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, nil)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateClusterCreateEKSANamespaceFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	test.clusterCreator.EXPECT().Run(
		test.ctx, test.clusterSpec, *test.bootstrapCluster)

	test.clusterManager.EXPECT().GetWorkloadCluster(
		test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider).Return(test.workloadCluster, nil)

	test.clusterManager.EXPECT().CreateEKSANamespace(
		test.ctx, test.workloadCluster).Return(errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateInstallCAPIWorkloadFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	test.clusterCreator.EXPECT().Run(
		test.ctx, test.clusterSpec, *test.bootstrapCluster)

	test.clusterManager.EXPECT().GetWorkloadCluster(
		test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider).Return(test.workloadCluster, nil)

	test.clusterManager.EXPECT().CreateEKSANamespace(
		test.ctx, test.workloadCluster)

	test.clusterManager.EXPECT().InstallCAPI(
		test.ctx, test.clusterSpec, test.workloadCluster, test.provider).Return(errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateUpdateSecretsFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	test.clusterCreator.EXPECT().Run(
		test.ctx, test.clusterSpec, *test.bootstrapCluster)

	test.clusterManager.EXPECT().GetWorkloadCluster(
		test.ctx, test.bootstrapCluster, test.clusterSpec, test.provider).Return(test.workloadCluster, nil)

	test.clusterManager.EXPECT().CreateEKSANamespace(
		test.ctx, test.workloadCluster)

	test.clusterManager.EXPECT().InstallCAPI(
		test.ctx, test.clusterSpec, test.workloadCluster, test.provider)

	test.provider.EXPECT().UpdateSecrets(
		test.ctx, test.workloadCluster, test.clusterSpec).Return(errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreateGitOPsFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap()
	test.expectPreflightValidationsToPass()

	gomock.InOrder(
		test.provider.EXPECT().DatacenterConfig(
			test.clusterSpec).Return(test.datacenterConfig),

		test.provider.EXPECT().MachineConfigs(
			test.clusterSpec).Return(test.machineConfigs),

		test.gitOpsManager.EXPECT().InstallGitOps(
			test.ctx, test.workloadCluster, test.clusterSpec, test.datacenterConfig, test.machineConfigs).Return(errors.New("test")),
	)
	test.expectWriteClusterConfig()
	test.expectCuratedPackagesInstallation()
	test.expectDeleteBootstrap()

	err := test.run()
	if err != nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunAWSIAMSuccess(t *testing.T) {
	test := newCreateTest(t)
	test.clusterSpec.AWSIamConfig = &v1alpha1.AWSIamConfig{}

	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap()
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

func TestCreateWriteConfigFailure(t *testing.T) {
	test := newCreateTest(t)

	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap()
	test.expectInstallGitOpsManager()
	test.expectPreflightValidationsToPass()

	gomock.InOrder(
		test.provider.EXPECT().DatacenterConfig(
			test.clusterSpec).Return(test.datacenterConfig),

		test.provider.EXPECT().MachineConfigs(
			test.clusterSpec).Return(test.machineConfigs),

		test.writer.EXPECT().Write(
			"cluster-name-eks-a-cluster.yaml", gomock.Any(), gomock.Any()).Return("", errors.New("test")),
	)

	test.clusterManager.EXPECT().SaveLogsManagementCluster(
		test.ctx, test.clusterSpec, test.bootstrapCluster,
	)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(
		test.ctx, test.provider, test.clusterSpec, test.workloadCluster,
	)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}

func TestCreateRunDeleteBootstrapFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall()
	test.expectCreateWorkload()
	test.expectInstallResourcesOnManagementTask()
	test.expectMoveManagement()
	test.expectInstallEksaComponentsWorkload(nil, nil, nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap()
	test.expectInstallGitOpsManager()
	test.expectPreflightValidationsToPass()
	test.expectCuratedPackagesInstallation()

	gomock.InOrder(
		test.provider.EXPECT().DatacenterConfig(
			test.clusterSpec).Return(test.datacenterConfig),

		test.provider.EXPECT().MachineConfigs(
			test.clusterSpec).Return(test.machineConfigs),

		test.writer.EXPECT().Write(
			"cluster-name-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),

		test.writer.EXPECT().Write(
			"cluster-name-checkpoint.yaml", gomock.Any(), gomock.Any()),
	)
	test.bootstrapper.EXPECT().DeleteBootstrapCluster(test.ctx, test.bootstrapCluster, gomock.Any(), gomock.Any()).Return(errors.New("test"))

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}
