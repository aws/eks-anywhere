package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeCluster struct{}

// Run upgradeCluster performs actions needed to upgrade the management cluster.
func (s *upgradeCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Upgrading management cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &reconcileGitOps{}
}

func (s *upgradeCluster) Name() string {
	return "upgrade-workload-cluster"
}

func (s *upgradeCluster) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &reconcileGitOps{}, nil
}
