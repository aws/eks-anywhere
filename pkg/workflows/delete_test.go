package workflows_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type deleteTestSetup struct {
	t                *testing.T
	bootstrapper     *mocks.MockBootstrapper
	clusterManager   *mocks.MockClusterManager
	addonManager     *mocks.MockAddonManager
	provider         *providermocks.MockProvider
	workflow         *workflows.Delete
	ctx              context.Context
	clusterSpec      *cluster.Spec
	forceCleanup     bool
	bootstrapCluster *types.Cluster
	workloadCluster  *types.Cluster
}

func newDeleteTest(t *testing.T) *deleteTestSetup {
	mockCtrl := gomock.NewController(t)
	mockBootstrapper := mocks.NewMockBootstrapper(mockCtrl)
	clusterManager := mocks.NewMockClusterManager(mockCtrl)
	addonManager := mocks.NewMockAddonManager(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	workflow := workflows.NewDelete(mockBootstrapper, provider, clusterManager, addonManager)

	return &deleteTestSetup{
		t:                t,
		bootstrapper:     mockBootstrapper,
		clusterManager:   clusterManager,
		addonManager:     addonManager,
		provider:         provider,
		workflow:         workflow,
		ctx:              context.Background(),
		clusterSpec:      test.NewClusterSpec(func(s *cluster.Spec) { s.Name = "cluster-name" }),
		bootstrapCluster: &types.Cluster{Name: "bootstrap"},
		workloadCluster:  &types.Cluster{Name: "workload"},
	}
}

func (c *deleteTestSetup) expectSetup() {
	c.provider.EXPECT().SetupAndValidateDeleteCluster(c.ctx)
}

func (c *deleteTestSetup) expectCreateBootstrap() {
	opts := []bootstrapper.BootstrapClusterOption{
		bootstrapper.WithDefaultCNIDisabled(), bootstrapper.WithExtraDockerMounts(),
	}

	gomock.InOrder(
		c.provider.EXPECT().BootstrapClusterOpts().Return(opts, nil),
		c.bootstrapper.EXPECT().CreateBootstrapCluster(
			c.ctx, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil()),
		).Return(c.bootstrapCluster, nil),

		c.clusterManager.EXPECT().InstallCAPI(c.ctx, gomock.Not(gomock.Nil()), c.bootstrapCluster, c.provider),
	)
}

func (c *deleteTestSetup) expectDeleteBootstrap() {
	gomock.InOrder(
		c.bootstrapper.EXPECT().DeleteBootstrapCluster(
			c.ctx, c.bootstrapCluster,
			gomock.Any()).Return(nil),
	)
}

func (c *deleteTestSetup) expectDeleteWorkload() {
	gomock.InOrder(
		c.clusterManager.EXPECT().DeleteCluster(
			c.ctx, c.bootstrapCluster, c.workloadCluster,
		).Return(nil),
	)
}

func (c *deleteTestSetup) expectCleanupGitRepo() {
	gomock.InOrder(
		c.addonManager.EXPECT().CleanupGitRepo(
			c.ctx, c.clusterSpec,
		).Return(nil),
	)
}

func (c *deleteTestSetup) expectMoveManagement() {
	gomock.InOrder(
		c.clusterManager.EXPECT().MoveCAPI(
			c.ctx, c.workloadCluster, c.bootstrapCluster, gomock.Any(),
		),
	)
}

func (c *deleteTestSetup) expectCleanupProvider() {
	gomock.InOrder(
		c.provider.EXPECT().CleanupProviderInfrastructure(
			c.ctx,
		).Return(nil))
}

func (c *deleteTestSetup) run() error {
	// ctx context.Context, workloadCluster *types.Cluster, forceCleanup bool
	return c.workflow.Run(c.ctx, c.workloadCluster, c.clusterSpec, c.forceCleanup)
}

func TestDeleteRunSuccess(t *testing.T) {
	test := newDeleteTest(t)
	test.expectSetup()
	test.expectCreateBootstrap()
	test.expectDeleteWorkload()
	test.expectCleanupGitRepo()
	test.expectMoveManagement()
	test.expectDeleteBootstrap()
	test.expectCleanupProvider()

	err := test.run()
	if err != nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}
