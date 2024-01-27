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

func (c *createTestSetup) run() error {
	return c.workflow.Run(c.ctx, c.clusterSpec, c.validator)
}

func (c *createTestSetup) expectPreflightValidationsToPass() {
	c.validator.EXPECT().PreflightValidations(c.ctx)
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

func (c *createTestSetup) expectCAPIInstall(err1, err2, err3 error) {
	gomock.InOrder(
		c.provider.EXPECT().PreCAPIInstallOnBootstrap(
			c.ctx, c.bootstrapCluster, c.clusterSpec).Return(err1),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.clusterSpec, c.bootstrapCluster, c.provider).Return(err2),

		c.provider.EXPECT().PostBootstrapSetup(
			c.ctx, c.clusterSpec.Cluster, c.bootstrapCluster).Return(err3),
	)
}

func TestCreateRunSuccess(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)

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
	err := errors.New("test")
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, err)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}
