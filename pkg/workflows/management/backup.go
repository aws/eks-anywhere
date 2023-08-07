package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

// TODO(g-gaston): make this into a actual task and run it before the upgrade.
func TaskForBackup(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// Take best effort CAPI backup of workload cluster without filter.
	// If that errors, then take CAPI backup filtering on only workload cluster.
	logger.Info("Backing up workload cluster's management resources before moving to bootstrap cluster")
	err := commandContext.ClusterManager.BackupCAPI(ctx, commandContext.ManagementCluster, commandContext.ManagementClusterStateDir, "")
	if err != nil {
		err = commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.ManagementCluster, commandContext.ManagementClusterStateDir, commandContext.ManagementCluster.Name)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	logger.V(3).Info("Pausing workload clusters before moving management cluster resources to bootstrap cluster")
	err = commandContext.ClusterManager.PauseCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &upgradeCluster{}
}
