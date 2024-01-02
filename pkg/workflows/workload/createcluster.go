package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createCluster struct{}

// Run createCluster performs actions needed to create the management cluster.
func (c *createCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating workload cluster")
	if err := commandContext.ClusterCreator.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &writeClusterConfig{}
}

func (c *createCluster) Name() string {
	return "create-workload-cluster"
}

func (c *createCluster) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (c *createCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &writeClusterConfig{}, nil
}
