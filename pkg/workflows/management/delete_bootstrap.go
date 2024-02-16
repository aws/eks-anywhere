package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteBootstrapClusterForDeleteTask struct{}

func (s *deleteBootstrapClusterForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Deleting bootstrap cluster")
	if err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Delete, false); err != nil {
		commandContext.SetError(err)
	}

	if commandContext.OriginalError != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.MarkSuccess("Cluster deleted!")
	return nil
}

func (s *deleteBootstrapClusterForDeleteTask) Name() string {
	return "kind-cluster-delete"
}

func (s *deleteBootstrapClusterForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteBootstrapClusterForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}
