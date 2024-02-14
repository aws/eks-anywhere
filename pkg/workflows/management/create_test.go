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
	clientmocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
)

type createTestSetup struct {
	t                    *testing.T
	packageInstaller     *mocks.MockPackageInstaller
	clusterManager       *mocks.MockClusterManager
	bootstrapper         *mocks.MockBootstrapper
	gitOpsManager        *mocks.MockGitOpsManager
	provider             *providermocks.MockProvider
	writer               *writermocks.MockFileWriter
	validator            *mocks.MockValidator
	eksdInstaller        *mocks.MockEksdInstaller
	eksaInstaller        *mocks.MockEksaInstaller
	clusterCreator       *mocks.MockClusterCreator
	datacenterConfig     providers.DatacenterConfig
	machineConfigs       []providers.MachineConfig
	ctx                  context.Context
	managementComponents *cluster.ManagementComponents
	clusterSpec          *cluster.Spec
	bootstrapCluster     *types.Cluster
	workloadCluster      *types.Cluster
	workflow             *management.Create
	client               *clientmocks.MockClient
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
	eksaInstaller := mocks.NewMockEksaInstaller(mockCtrl)

	packageInstaller := mocks.NewMockPackageInstaller(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	clusterCreator := mocks.NewMockClusterCreator(mockCtrl)
	validator := mocks.NewMockValidator(mockCtrl)
	client := clientmocks.NewMockClient(mockCtrl)

	workflow := management.NewCreate(
		bootstrapper,
		provider,
		clusterManager,
		gitOpsManager,
		writer,
		eksdInstaller,
		packageInstaller,
		clusterCreator,
		eksaInstaller,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Namespace = "test-ns"
	})
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)

	return &createTestSetup{
		t:                t,
		bootstrapper:     bootstrapper,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		provider:         provider,
		writer:           writer,
		validator:        validator,
		eksdInstaller:    eksdInstaller,
		eksaInstaller:    eksaInstaller,
		packageInstaller: packageInstaller,
		clusterCreator:   clusterCreator,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workflow:         workflow,
		ctx:              context.Background(),
		bootstrapCluster: &types.Cluster{
			Name: "test-cluster",
		},
		workloadCluster:      &types.Cluster{},
		managementComponents: managementComponents,
		clusterSpec:          clusterSpec,
		client:               client,
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

func (c *createTestSetup) expectCreateRegistrySecret(err error) {
	c.clusterManager.EXPECT().CreateRegistryCredSecret(c.ctx, c.bootstrapCluster).Return(err)
}

func (c *createTestSetup) expectCAPIInstall(err1, err2, err3 error) {
	gomock.InOrder(
		c.provider.EXPECT().PreCAPIInstallOnBootstrap(
			c.ctx, c.bootstrapCluster, c.clusterSpec).Return(err1),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.managementComponents, c.clusterSpec, c.bootstrapCluster, c.provider).Return(err2),

		c.provider.EXPECT().PostBootstrapSetup(
			c.ctx, c.clusterSpec.Cluster, c.bootstrapCluster).Return(err3),
	)
}

func (c *createTestSetup) expectInstallEksaComponentsBootstrap(err1, err2, err3, err4 error) {
	gomock.InOrder(

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster).Return(err1),

		c.eksaInstaller.EXPECT().Install(
			c.ctx, logger.Get(), c.bootstrapCluster, c.managementComponents, c.clusterSpec).Return(err2),

		c.provider.EXPECT().InstallCustomProviderComponents(
			c.ctx, c.bootstrapCluster.KubeconfigFile).Return(err3),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.bootstrapCluster).Return(err4),
	)
}

func (c *createTestSetup) expectCreateWorkload(err1, err2, err3, err4, err5, err6 error) {
	gomock.InOrder(
		c.clusterManager.EXPECT().CreateNamespace(c.ctx, c.bootstrapCluster, c.clusterSpec.Cluster.Namespace).Return(err1),

		c.clusterCreator.EXPECT().CreateSync(c.ctx, c.clusterSpec, c.bootstrapCluster).Return(c.workloadCluster, err2),

		c.clusterManager.EXPECT().CreateEKSANamespace(
			c.ctx, c.workloadCluster).Return(err3),

		c.clusterManager.EXPECT().InstallCAPI(
			c.ctx, c.managementComponents, c.clusterSpec, c.workloadCluster, c.provider).Return(err4),

		c.provider.EXPECT().UpdateSecrets(
			c.ctx, c.workloadCluster, c.clusterSpec).Return(err5),

		c.clusterManager.EXPECT().CreateRegistryCredSecret(c.ctx, c.workloadCluster).Return(err6),
	)
}

func (c *createTestSetup) expectInstallResourcesOnManagementTask(err error) {
	gomock.InOrder(
		c.provider.EXPECT().PostWorkloadInit(c.ctx, c.workloadCluster, c.clusterSpec).Return(err),
	)
}

func (c *createTestSetup) expectPauseReconcile(err error) {
	c.clusterManager.EXPECT().PauseEKSAControllerReconcile(
		c.ctx, c.bootstrapCluster, c.clusterSpec, c.provider).Return(err)
}

func (c *createTestSetup) expectMoveManagement(err error) {
	c.clusterManager.EXPECT().MoveCAPI(
		c.ctx, c.bootstrapCluster, c.workloadCluster, c.workloadCluster.Name, c.clusterSpec, gomock.Any()).Return(err)
}

func (c *createTestSetup) expectInstallEksaComponentsWorkload(err1, err2, err3 error) {
	gomock.InOrder(

		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.workloadCluster).Return(err1),

		c.eksaInstaller.EXPECT().Install(
			c.ctx, logger.Get(), c.workloadCluster, c.managementComponents, c.clusterSpec),

		c.provider.EXPECT().InstallCustomProviderComponents(
			c.ctx, c.workloadCluster.KubeconfigFile),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.workloadCluster),

		c.clusterManager.EXPECT().CreateNamespace(c.ctx, c.workloadCluster, c.clusterSpec.Cluster.Namespace).Return(err3),

		c.clusterCreator.EXPECT().Run(c.ctx, c.clusterSpec, *c.workloadCluster).Return(err2),
	)
}

func (c *createTestSetup) expectInstallGitOpsManager() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(
			c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(
			c.clusterSpec).Return(c.machineConfigs),

		c.gitOpsManager.EXPECT().InstallGitOps(
			c.ctx, c.workloadCluster, c.managementComponents, c.clusterSpec, c.datacenterConfig, c.machineConfigs),
	)
}

func (c *createTestSetup) expectWriteClusterConfig() {
	gomock.InOrder(
		c.provider.EXPECT().DatacenterConfig(
			c.clusterSpec).Return(c.datacenterConfig),

		c.provider.EXPECT().MachineConfigs(
			c.clusterSpec).Return(c.machineConfigs),

		c.writer.EXPECT().Write(
			"test-cluster-eks-a-cluster.yaml", gomock.Any(), gomock.Any()),
	)
}

func (c *createTestSetup) expectDeleteBootstrap(err error) {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, gomock.Any(), gomock.Any()).Return(err)
}

func (c *createTestSetup) expectCuratedPackagesInstallation() {
	c.packageInstaller.EXPECT().InstallCuratedPackages(c.ctx).Times(1)
}

func TestCreateRunSuccess(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	test.expectInstallEksaComponentsWorkload(nil, nil, nil)
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap(nil)
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

func TestCreateRegistrySecretFailure(t *testing.T) {
	c := newCreateTest(t)
	c.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{Authenticate: true}
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()

	c.expectCreateRegistrySecret(fmt.Errorf(""))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
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
			c.ctx, c.managementComponents, c.clusterSpec, c.bootstrapCluster, c.provider).Return(errors.New("test")),
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

func TestCreateInstallCRDFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)

	gomock.InOrder(
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

func TestCreateInstallCustomComponentsFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)

	err := errors.New("test")

	c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster)

	c.eksaInstaller.EXPECT().Install(
		c.ctx, logger.Get(), c.bootstrapCluster, c.managementComponents, c.clusterSpec).Return(err)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallCustomProviderComponentsFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)

	err := errors.New("test")

	c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster)

	c.eksaInstaller.EXPECT().Install(
		c.ctx, logger.Get(), c.bootstrapCluster, c.managementComponents, c.clusterSpec)

	c.provider.EXPECT().InstallCustomProviderComponents(
		c.ctx, c.bootstrapCluster.KubeconfigFile).Return(err)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallEksdManifestFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)

	err := errors.New("test")

	c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster)

	c.eksaInstaller.EXPECT().Install(
		c.ctx, logger.Get(), c.bootstrapCluster, c.managementComponents, c.clusterSpec)

	c.provider.EXPECT().InstallCustomProviderComponents(
		c.ctx, c.bootstrapCluster.KubeconfigFile)

	c.eksdInstaller.EXPECT().InstallEksdManifest(
		c.ctx, c.clusterSpec, c.bootstrapCluster).Return(err)

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, nil)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err = c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateSyncFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)

	test.clusterManager.EXPECT().CreateNamespace(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Namespace).Return(nil)
	test.clusterCreator.EXPECT().CreateSync(test.ctx, test.clusterSpec, test.bootstrapCluster).Return(nil, errors.New("test"))
	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateEKSANamespaceFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)

	test.clusterManager.EXPECT().CreateNamespace(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Namespace).Return(nil)
	test.clusterCreator.EXPECT().CreateSync(test.ctx, test.clusterSpec, test.bootstrapCluster).Return(test.workloadCluster, nil)
	test.clusterManager.EXPECT().CreateEKSANamespace(test.ctx, test.workloadCluster).Return(errors.New("test"))
	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)
	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateInstallCAPIWorkloadFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectPreflightValidationsToPass()

	test.clusterManager.EXPECT().CreateNamespace(test.ctx, test.bootstrapCluster, test.clusterSpec.Cluster.Namespace).Return(nil)
	test.clusterCreator.EXPECT().CreateSync(
		test.ctx, test.clusterSpec, test.bootstrapCluster).Return(test.workloadCluster, nil)

	test.clusterManager.EXPECT().CreateEKSANamespace(
		test.ctx, test.workloadCluster)

	test.clusterManager.EXPECT().InstallCAPI(
		test.ctx, test.managementComponents, test.clusterSpec, test.workloadCluster, test.provider).Return(errors.New("test"))

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
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectPreflightValidationsToPass()

	test.expectCreateWorkload(nil, nil, nil, nil, nil, errors.New("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("expected error from task")
	}
}

func TestCreatePostWorkloadInitFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)
	c.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	c.expectCreateWorkload(nil, nil, nil, nil, nil, nil)

	c.expectInstallResourcesOnManagementTask(fmt.Errorf("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)

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
	c.expectCAPIInstall(nil, nil, nil)
	c.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	c.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	c.expectInstallResourcesOnManagementTask(nil)
	c.expectPauseReconcile(nil)
	c.expectMoveManagement(errors.New("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestPauseReconcilerFailure(t *testing.T) {
	c := newCreateTest(t)
	c.expectSetup()
	c.expectCreateBootstrap()
	c.expectPreflightValidationsToPass()
	c.expectCAPIInstall(nil, nil, nil)
	c.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	c.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	c.expectInstallResourcesOnManagementTask(nil)
	c.expectPauseReconcile(errors.New("test"))

	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)

	c.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", c.clusterSpec.Cluster.Name), gomock.Any())

	err := c.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateEKSAWorkloadComponentsFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)

	test.eksdInstaller.EXPECT().InstallEksdCRDs(test.ctx, test.clusterSpec, test.workloadCluster).Return(fmt.Errorf("test"))

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)
	test.clusterManager.EXPECT().SaveLogsWorkloadCluster(test.ctx, test.provider, test.clusterSpec, test.workloadCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateEKSAWorkloadFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	test.expectInstallEksaComponentsWorkload(nil, fmt.Errorf("test"), nil)

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateEKSAWorkloadNamespaceFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	gomock.InOrder(

		test.eksdInstaller.EXPECT().InstallEksdCRDs(test.ctx, test.clusterSpec, test.workloadCluster),

		test.eksaInstaller.EXPECT().Install(
			test.ctx, logger.Get(), test.workloadCluster, test.managementComponents, test.clusterSpec),

		test.provider.EXPECT().InstallCustomProviderComponents(
			test.ctx, test.workloadCluster.KubeconfigFile),

		test.eksdInstaller.EXPECT().InstallEksdManifest(
			test.ctx, test.clusterSpec, test.workloadCluster),

		test.clusterManager.EXPECT().CreateNamespace(test.ctx, test.workloadCluster, test.clusterSpec.Cluster.Namespace).Return(fmt.Errorf("")),
	)

	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.bootstrapCluster)

	test.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", test.clusterSpec.Cluster.Name), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() expected to return an error %v", err)
	}
}

func TestCreateGitOPsFailure(t *testing.T) {
	test := newCreateTest(t)
	test.expectSetup()
	test.expectPreflightValidationsToPass()
	test.expectCreateBootstrap()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	test.expectInstallEksaComponentsWorkload(nil, nil, nil)

	gomock.InOrder(
		test.provider.EXPECT().DatacenterConfig(
			test.clusterSpec).Return(test.datacenterConfig),

		test.provider.EXPECT().MachineConfigs(
			test.clusterSpec).Return(test.machineConfigs),

		test.gitOpsManager.EXPECT().InstallGitOps(
			test.ctx, test.workloadCluster, test.managementComponents, test.clusterSpec, test.datacenterConfig, test.machineConfigs).Return(errors.New("test")),
	)
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap(nil)
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
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	test.expectInstallEksaComponentsWorkload(nil, nil, nil)
	test.expectInstallGitOpsManager()
	test.expectPreflightValidationsToPass()

	gomock.InOrder(
		test.provider.EXPECT().DatacenterConfig(
			test.clusterSpec).Return(test.datacenterConfig),

		test.provider.EXPECT().MachineConfigs(
			test.clusterSpec).Return(test.machineConfigs),

		test.writer.EXPECT().Write(
			"test-cluster-eks-a-cluster.yaml", gomock.Any(), gomock.Any()).Return("", errors.New("test")),
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
	test.expectPreflightValidationsToPass()
	test.expectCAPIInstall(nil, nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil)
	test.expectCreateWorkload(nil, nil, nil, nil, nil, nil)
	test.expectInstallResourcesOnManagementTask(nil)
	test.expectPauseReconcile(nil)
	test.expectMoveManagement(nil)
	test.expectInstallEksaComponentsWorkload(nil, nil, nil)
	test.expectInstallGitOpsManager()
	test.expectWriteClusterConfig()
	test.expectDeleteBootstrap(fmt.Errorf("test"))
	test.expectCuratedPackagesInstallation()

	test.writer.EXPECT().Write("test-cluster-checkpoint.yaml", gomock.Any(), gomock.Any())

	err := test.run()
	if err == nil {
		t.Fatalf("Create.Run() err = %v, want err = nil", err)
	}
}
