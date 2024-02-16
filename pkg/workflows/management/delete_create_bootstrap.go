package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type createBootStrapClusterForDeleteTask struct{}

func (s *createBootStrapClusterForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new bootstrap cluster")

	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts(commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	bootstrapCluster, err := commandContext.Bootstrapper.CreateBootstrapCluster(ctx, commandContext.ClusterSpec, bootstrapOptions...)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.BootstrapCluster = bootstrapCluster

	return &installCAPIComponentsForDeleteTask{}
}

func (s *createBootStrapClusterForDeleteTask) Name() string {
	return "bootstrap-cluster-for-delete-init"
}

func (s *createBootStrapClusterForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *createBootStrapClusterForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}
