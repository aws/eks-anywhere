package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type deleteBootstrapClusterTask struct{}

func (s *deleteBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Deleting bootstrap cluster")
	err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Create, false)
	if err != nil {
		commandContext.SetError(err)
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster created!")
	}
	return &installCuratedPackagesTask{}
}

func (s *deleteBootstrapClusterTask) Name() string {
	return "delete-kind-cluster"
}

func (s *deleteBootstrapClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteBootstrapClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}
