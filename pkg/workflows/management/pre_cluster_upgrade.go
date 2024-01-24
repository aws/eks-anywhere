package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type preClusterUpgrade struct{}

// Run preClusterUpgrade implements steps to be performed before management cluster's upgrade.
func (s *preClusterUpgrade) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// Take best effort CAPI backup of management cluster without filter.
	// If that errors, then take CAPI backup filtering on only management cluster.
	logger.Info("Backing up management cluster's resources before upgrading")
	err := commandContext.ClusterManager.BackupCAPI(ctx, commandContext.ManagementCluster, commandContext.BackupClusterStateDir, "")
	if err != nil {
		err = commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.ManagementCluster, commandContext.BackupClusterStateDir, commandContext.ManagementCluster.Name)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectMgmtClusterDiagnosticsTask{}
		}
	}

	logger.V(3).Info("Pausing workload clusters before upgrading management cluster")
	err = commandContext.ClusterManager.PauseCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &installNewComponents{}
}

func (s *preClusterUpgrade) Name() string {
	return "pre-cluster-upgrade"
}

func (s *preClusterUpgrade) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *preClusterUpgrade) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &installNewComponents{}, nil
}
