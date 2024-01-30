package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installProviderSpecificResources struct{}

func (s *installProviderSpecificResources) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := commandContext.Provider.PostWorkloadInit(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	return &moveClusterManagementTask{}
}

func (s *installProviderSpecificResources) Name() string {
	return "install-resources-on-management-cluster"
}

func (s *installProviderSpecificResources) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installProviderSpecificResources) Checkpoint() *task.CompletedTask {
	return nil
}
