package workload_test

import (
	"context"
	"fmt"
	"os"
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

type deleteTestSetup struct {
	t                *testing.T
	provider         *providermocks.MockProvider
	clusterDeleter   *mocks.MockClusterDeleter
	clusterManager   *mocks.MockClusterManager
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
	ctx              context.Context
	clusterSpec      *cluster.Spec
	workloadCluster  *types.Cluster
	workload         *workload.Delete
	writer           *writermocks.MockFileWriter
	gitopsManager    *mocks.MockGitOpsManager
}

func newDeleteTest(t *testing.T) *deleteTestSetup {
	featureEnvVars := []string{}
	mockCtrl := gomock.NewController(t)
	provider := providermocks.NewMockProvider(mockCtrl)
	writer := writermocks.NewMockFileWriter(mockCtrl)
	manager := mocks.NewMockClusterManager(mockCtrl)

	datacenterConfig := &v1alpha1.VSphereDatacenterConfig{}
	machineConfigs := []providers.MachineConfig{&v1alpha1.VSphereMachineConfig{}}
	clusterDeleter := mocks.NewMockClusterDeleter(mockCtrl)
	gitopsManager := mocks.NewMockGitOpsManager(mockCtrl)

	workload := workload.NewDelete(
		provider,
		writer,
		manager,
		clusterDeleter,
		gitopsManager,
	)

	for _, e := range featureEnvVars {
		t.Setenv(e, "true")
	}

	return &deleteTestSetup{
		t:                t,
		provider:         provider,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
		workload:         workload,
		ctx:              context.Background(),
		clusterDeleter:   clusterDeleter,
		clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = "workload"
			s.Cluster.Spec.DatacenterRef.Kind = v1alpha1.VSphereDatacenterKind
			s.ManagementCluster = &types.Cluster{Name: "management"}
		}),
		workloadCluster: &types.Cluster{Name: "workload"},
		clusterManager:  manager,
		writer:          writer,
		gitopsManager:   gitopsManager,
	}
}

func (c *deleteTestSetup) expectSetup(err error) {
	c.provider.EXPECT().SetupAndValidateDeleteCluster(c.ctx, c.workloadCluster, c.clusterSpec).Return(err)
}

func (c *deleteTestSetup) expectDeleteWorkloadCluster(err error) {
	c.clusterDeleter.EXPECT().Run(c.ctx, c.clusterSpec, *c.clusterSpec.ManagementCluster).Return(err)
}

func (c *deleteTestSetup) run() error {
	return c.workload.Run(c.ctx, c.workloadCluster, c.clusterSpec)
}

func (c *deleteTestSetup) expectWrite() {
	c.writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
}

func (c *deleteTestSetup) expectSaveLogsWorkload() {
	c.clusterManager.EXPECT().SaveLogsWorkloadCluster(c.ctx, c.provider, c.clusterSpec, c.workloadCluster)
	c.expectWrite()
}

func (c *deleteTestSetup) expectCleanup(err error) {
	c.gitopsManager.EXPECT().CleanupGitRepo(c.ctx, c.clusterSpec).Return(err)
	if err == nil {
		c.writer.EXPECT().CleanUp()
	}
}

func TestDeleteRunSuccess(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectDeleteWorkloadCluster(nil)
	test.expectCleanup(nil)

	err := test.run()
	if err != nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}

func TestDeleteRunFail(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectDeleteWorkloadCluster(fmt.Errorf("Failure"))
	test.expectSaveLogsWorkload()

	err := test.run()
	if err == nil {
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

func TestDeleteRunFailCleanup(t *testing.T) {
	features.ClearCache()
	os.Setenv(features.UseControllerForCli, "true")
	test := newDeleteTest(t)
	test.expectSetup(nil)
	test.expectDeleteWorkloadCluster(nil)
	test.expectCleanup(fmt.Errorf(""))
	test.clusterManager.EXPECT().SaveLogsManagementCluster(test.ctx, test.clusterSpec, test.clusterSpec.ManagementCluster)
	test.expectSaveLogsWorkload()

	err := test.run()
	if err == nil {
		t.Fatalf("Delete.Run() err = %v, want err = nil", err)
	}
}
