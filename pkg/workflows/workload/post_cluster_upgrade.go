package workload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type postClusterUpgrade struct{}

// Run postClusterUpgrade implements steps to be performed after the upgrade process.
func (s *postClusterUpgrade) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Cleaning up backup resources")
	capiObjectFile := filepath.Join(commandContext.ManagementCluster.Name, commandContext.BackupClusterStateDir)
	if err := os.RemoveAll(capiObjectFile); err != nil {
		logger.Info(fmt.Sprintf("workload cluster CAPI backup file not found: %v", err))
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
