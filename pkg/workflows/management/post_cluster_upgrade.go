package management

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type postClusterUpgrade struct{}

// Run postClusterUpgrade implements steps to be performed after the upgrade process.
func (s *postClusterUpgrade) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.V(3).Info("Resuming all workload clusters after management cluster upgrade")
	err := commandContext.ClusterManager.ResumeCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Cleaning up backup resources")
	capiObjectFile := filepath.Join(commandContext.ManagementCluster.Name, commandContext.BackupClusterStateDir)
	if err := os.RemoveAll(capiObjectFile); err != nil {
		logger.Info(fmt.Sprintf("management cluster CAPI backup file not found: %v", err))
	}

	return nil
}

func (s *postClusterUpgrade) Name() string {
	return "post-cluster-upgrade"
}

func (s *postClusterUpgrade) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *postClusterUpgrade) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}
