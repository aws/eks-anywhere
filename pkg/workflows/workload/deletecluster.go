package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteWorkloadCluster struct{}

func (s *deleteWorkloadCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Deleting workload cluster")
	err := commandContext.ClusterDeleter.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectWorkloadClusterDiagnosticsTask{}
	}

	return &cleanupGitRepo{}
}

func (s *deleteWorkloadCluster) Name() string {
	return "delete-workload-cluster"
}

func (s *deleteWorkloadCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteWorkloadCluster) Checkpoint() *task.CompletedTask {
	return nil
}
