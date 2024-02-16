package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type moveClusterManagementForDeleteTask struct{}

func (s *moveClusterManagementForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Moving cluster management from workload cluster to bootstrap")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.WorkloadCluster, commandContext.BootstrapCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef())
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &installEksaComponentsOnBootstrapForDeleteTask{}
}

func (s *moveClusterManagementForDeleteTask) Name() string {
	return "capi-management-move-for-delete"
}

func (s *moveClusterManagementForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *moveClusterManagementForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}
