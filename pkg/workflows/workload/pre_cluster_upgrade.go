package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type preClusterUpgrade struct{}

// Run preClusterUpgrade implements steps to be performed before workload cluster's upgrade.
func (s *preClusterUpgrade) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// Take CAPI backup filtering on only current workload cluster.
	logger.Info("Backing up workload cluster's resources before upgrading")
	err := commandContext.ClusterManager.BackupCAPI(ctx, commandContext.ManagementCluster, commandContext.BackupClusterStateDir, commandContext.WorkloadCluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &upgradeCluster{}
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
	return &upgradeCluster{}, nil
}
