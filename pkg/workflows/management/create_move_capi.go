package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type moveClusterManagementTask struct{}

func (s *moveClusterManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := commandContext.ClusterManager.PauseEKSAControllerReconcile(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	// Install custom provider components (like IPAM provider) on the management cluster BEFORE the move
	// This is required because after clusterctl move, the management cluster needs the IPAM provider
	// to reconcile the cluster resources and update controlPlaneInitialized status
	logger.Info("Installing custom provider components on management cluster before move")
	if err := commandContext.Provider.InstallCustomProviderComponents(ctx, commandContext.WorkloadCluster.KubeconfigFile); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Moving the cluster management components from the bootstrap cluster to the management cluster")
	err = commandContext.ClusterManager.MoveCAPI(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef())
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &installEksaComponentsOnWorkloadTask{}
}

func (s *moveClusterManagementTask) Name() string {
	return "capi-management-move"
}

func (s *moveClusterManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *moveClusterManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}
