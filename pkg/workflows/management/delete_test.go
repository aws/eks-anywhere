package management_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	clientmocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
)

type deleteTestSetup struct {
	t                    *testing.T
	provider             *providermocks.MockProvider
	clusterManager       *mocks.MockClusterManager
	datacenterConfig     providers.DatacenterConfig
	machineConfigs       []providers.MachineConfig
	ctx                  context.Context
	clusterSpec          *cluster.Spec
	workloadCluster      *types.Cluster
	workload             *management.Delete
	writer               *writermocks.MockFileWriter
	bootstrapper         *mocks.MockBootstrapper
	gitopsManager        *mocks.MockGitOpsManager
	bootstrapCluster     *types.Cluster
	clusterDeleter       *mocks.MockClusterDeleter
	eksdInstaller        *mocks.MockEksdInstaller
	eksaInstaller        *mocks.MockEksaInstaller
	clientFactory        *mocks.MockClientFactory
	managementComponents *cluster.ManagementComponents
	client               *clientmocks.MockClient
	mover                *mocks.MockClusterMover
}

func newDeleteTest(t *testing.T) *deleteTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	manager := mocks.NewMockClusterManager(mockCtrl)
	client := clientmocks.NewMockClient(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	bootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	gitopsManager := mocks.NewMockGitOpsManager(mockCtrl)
	clusterDeleter := mocks.NewMockClusterDeleter(mockCtrl)
	eksdInstaller := mocks.NewMockEksdInstaller(mockCtrl)
	eksaInstaller := mocks.NewMockEksaInstaller(mockCtrl)
	clientFactory := mocks.NewMockClientFactory(mockCtrl)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "workload"
		s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
		s.ManagementCluster = &types.Cluster{Name: "management"}
		s.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	})
	managementComponents := cluster.ManagementComponentsFromBundles(clusterSpec.Bundles)
	mover := mocks.NewMockClusterMover(mockCtrl)

	workload := management.NewDelete(
		bootstrapper,
		provider,
		writer,
		manager,
		gitopsManager,
		clusterDeleter,
		eksdInstaller,
		eksaInstaller,
		clientFactory,
		mover,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &deleteTestSetup{
		t:                    t,
		provider:             provider,
		datacenterConfig:     datacenterConfig,
		machineConfigs:       machineConfigs,
		workload:             workload,
		ctx:                  context.Background(),
		clusterSpec:          clusterSpec,
		workloadCluster:      &types.Cluster{Name: "workload"},
		clusterManager:       manager,
		writer:               writer,
		bootstrapper:         bootstrapper,
		gitopsManager:        gitopsManager,
		bootstrapCluster:     &types.Cluster{},
		clusterDeleter:       clusterDeleter,
		eksdInstaller:        eksdInstaller,
		eksaInstaller:        eksaInstaller,
		clientFactory:        clientFactory,
		managementComponents: managementComponents,
		client:               client,
		mover:                mover,
	}
}

func (c *deleteTestSetup) expectSetup(err error) {
	c.provider.EXPECT().SetupAndValidateDeleteCluster(c.ctx, c.workloadCluster, c.clusterSpec).Return(err)
}

func (c *deleteTestSetup) expectCleanupGitRepo(err error) {
	c.gitopsManager.EXPECT().CleanupGitRepo(c.ctx, c.clusterSpec).Return(err)
}

func (c *deleteTestSetup) expectBootstrapOpts(err error) {
	c.provider.EXPECT().BootstrapClusterOpts(c.clusterSpec).Return([]bootstrapper.BootstrapClusterOption{}, err)
}

func (c *deleteTestSetup) expectCreateBootstrap(err error) {
	c.bootstrapper.EXPECT().CreateBootstrapCluster(c.ctx, c.clusterSpec, gomock.Any()).Return(&types.Cluster{}, err)
}

func (c *deleteTestSetup) run() error {
	return c.workload.Run(c.ctx, c.workloadCluster, c.clusterSpec)
}

func (c *deleteTestSetup) expectWrite() {
	c.writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
}

func (c *deleteTestSetup) expectSaveLogsWorkload() {
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)
}

func (c *deleteTestSetup) expectSaveLogsManagement() {
	c.clusterManager.EXPECT().SaveLogsManagementCluster(c.ctx, c.clusterSpec, c.bootstrapCluster)
	c.expectWrite()
}

func (c *deleteTestSetup) expectDeleteBootstrap(err error) {
	c.bootstrapper.EXPECT().DeleteBootstrapCluster(c.ctx, c.bootstrapCluster, constants.Delete, false).Return(err)
}

func (c *deleteTestSetup) expectPreCAPI(err error) {
	c.provider.EXPECT().PreCAPIInstallOnBootstrap(c.ctx, c.bootstrapCluster, c.clusterSpec).Return(err)
}

func (c *deleteTestSetup) expectInstallCAPI(err error) {
	c.clusterManager.EXPECT().InstallCAPI(c.ctx, gomock.Any(), c.clusterSpec, c.bootstrapCluster, c.provider).Return(err)
}

func (c *deleteTestSetup) expectMoveCAPI(err1, err2 error) {
	c.clusterManager.EXPECT().PauseEKSAControllerReconcile(c.ctx, c.workloadCluster, c.clusterSpec, c.provider).Return(err1)
	if err1 != nil {
		return
	}
	c.clusterManager.EXPECT().MoveCAPI(c.ctx, c.workloadCluster, c.bootstrapCluster, c.workloadCluster.Name, c.clusterSpec, gomock.Any()).Return(err2)
}

func (c *deleteTestSetup) expectInstallEksaComponentsBootstrap(err1, err2, err3, err4, err5, err6, err7, err8, err9 error) {
	gomock.InOrder(
		c.eksdInstaller.EXPECT().InstallEksdCRDs(c.ctx, c.clusterSpec, c.bootstrapCluster).Return(err1).AnyTimes(),

		c.eksaInstaller.EXPECT().Install(
			c.ctx, logger.Get(), c.bootstrapCluster, c.managementComponents, c.clusterSpec).Return(err2).AnyTimes(),

		c.provider.EXPECT().InstallCustomProviderComponents(
			c.ctx, c.bootstrapCluster.KubeconfigFile).Return(err3).AnyTimes(),

		c.eksdInstaller.EXPECT().InstallEksdManifest(
			c.ctx, c.clusterSpec, c.bootstrapCluster).Return(err4).AnyTimes(),

		c.clientFactory.EXPECT().BuildClientFromKubeconfig(c.workloadCluster.KubeconfigFile).Return(c.client, err5).MaxTimes(1),

		c.clientFactory.EXPECT().BuildClientFromKubeconfig(c.bootstrapCluster.KubeconfigFile).Return(c.client, err6).MaxTimes(1),

		c.client.EXPECT().Create(c.ctx, gomock.AssignableToTypeOf(&corev1.Namespace{})).Return(err7).AnyTimes(),

		c.mover.EXPECT().Move(c.ctx, c.clusterSpec, c.client, c.client).Return(err8).AnyTimes(),

		c.clusterManager.EXPECT().AllowDeleteWhilePaused(c.ctx, c.bootstrapCluster, c.clusterSpec).Return(err9).AnyTimes(),
	)
}

func (c *deleteTestSetup) expectDeleteCluster(err1, err2 error) {
	gomock.InOrder(
		c.clusterDeleter.EXPECT().Run(c.ctx, c.clusterSpec, *c.bootstrapCluster).Return(err1).AnyTimes(),

		c.provider.EXPECT().PostClusterDeleteValidate(c.ctx, c.bootstrapCluster).Return(err2).AnyTimes(),
	)
}

func (c *deleteTestSetup) expectApplyOnBootstrap(err error) {
	c.client.EXPECT().ApplyServerSide(c.ctx, "eks-a-cli", gomock.Any(), gomock.Any()).Return(err).AnyTimes()
}

func TestDeleteRunSuccess(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	test.expectApplyOnBootstrap(nil)
	test.expectDeleteCluster(nil, nil)
	test.expectCleanupGitRepo(nil)
	test.expectDeleteBootstrap(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailSetup(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(fmt.Errorf("Failure"))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailBootstrapOpts(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(fmt.Errorf(""))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailCreateBootstrap(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(fmt.Errorf(""))
	test.expectWrite()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailPreCAPI(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(fmt.Errorf(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailInstallCAPI(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(fmt.Errorf(""))
	test.expectDeleteCluster(fmt.Errorf(""), nil)
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailPauseEksa(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(fmt.Errorf(""), nil)
	test.expectSaveLogsWorkload()
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailMoveCAPI(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, fmt.Errorf(""))
	test.expectSaveLogsWorkload()
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailResumeReconcile(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(fmt.Errorf(""), nil, nil, nil, nil, nil, nil, nil, nil)
	test.expectSaveLogsManagement()
	test.expectSaveLogsWorkload()
	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailAddAnnotation(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, fmt.Errorf(""), nil, nil, nil, nil, nil, nil, nil)
	test.expectSaveLogsManagement()
	test.expectSaveLogsWorkload()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailProviderInstall(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, fmt.Errorf(""), nil, nil, nil, nil, nil, nil)
	test.expectSaveLogsManagement()
	test.expectSaveLogsWorkload()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailEksdInstall(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, fmt.Errorf(""), nil, nil, nil, nil, nil)
	test.expectSaveLogsManagement()
	test.expectSaveLogsWorkload()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailBuildSrcClient(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, fmt.Errorf(""), nil, nil, nil, nil)
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailBuildDstClient(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, fmt.Errorf(""), nil, nil, nil)
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailCreateNamespace(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, fmt.Errorf(""), nil, nil)
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailAllowDeleteWhilePaused(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailPostDelete(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	test.expectApplyOnBootstrap(nil)
	test.expectDeleteCluster(nil, fmt.Errorf(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailCleanupGit(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	test.expectApplyOnBootstrap(nil)
	test.expectDeleteCluster(nil, nil)
	test.expectCleanupGitRepo(fmt.Errorf(""))
	test.expectSaveLogsWorkload()
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFailDeleteBootstrap(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectBootstrapOpts(nil)
	test.expectCreateBootstrap(nil)
	test.expectPreCAPI(nil)
	test.expectInstallCAPI(nil)
	test.expectMoveCAPI(nil, nil)
	test.expectInstallEksaComponentsBootstrap(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	test.expectApplyOnBootstrap(nil)
	test.expectDeleteCluster(nil, nil)
	test.expectCleanupGitRepo(nil)
	test.expectDeleteBootstrap(fmt.Errorf(""))
	test.expectSaveLogsManagement()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}
