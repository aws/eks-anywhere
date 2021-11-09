package workflows

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type CollectDiagnosticsTask struct{}

func (s *CollectDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting cluster diagnostics")
	_ = commandContext.ClusterManager.SaveLogsManagementCluster(ctx, commandContext.BootstrapCluster)
	_ = commandContext.ClusterManager.SaveLogsWorkloadCluster(ctx, commandContext.Provider, commandContext.ClusterSpec, commandContext.WorkloadCluster)
	return nil
}

func (s *CollectDiagnosticsTask) Name() string {
	return "collect-cluster-diagnostics"
}
