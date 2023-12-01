package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installResourcesOnManagementTask struct{}

func (s *installResourcesOnManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster.ExistingManagement {
		return &moveClusterManagementTask{}
	}
	logger.Info("Installing resources on management cluster")

	if err := commandContext.Provider.PostWorkloadInit(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	return &moveClusterManagementTask{}
}

func (s *installResourcesOnManagementTask) Name() string {
	return "install-resources-on-management-cluster"
}

func (s *installResourcesOnManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installResourcesOnManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}
