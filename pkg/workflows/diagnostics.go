package workflows

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type CollectDiagnosticsTask struct {
	*CollectWorkloadClusterDiagnosticsTask
	*CollectMgmtClusterDiagnosticsTask
}

type CollectWorkloadClusterDiagnosticsTask struct{}

type CollectMgmtClusterDiagnosticsTask struct{}

// CollectDiagnosticsTask implementation

func (s *CollectDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting cluster diagnostics")
	_ = s.CollectMgmtClusterDiagnosticsTask.Run(ctx, commandContext)
	_ = s.CollectWorkloadClusterDiagnosticsTask.Run(ctx, commandContext)
	return nil
}

func (s *CollectDiagnosticsTask) Name() string {
	return "collect-cluster-diagnostics"
}

// CollectWorkloadClusterDiagnosticsTask implementation

func (s *CollectWorkloadClusterDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting workload cluster diagnostics")
	_ = commandContext.ClusterManager.SaveLogsWorkloadCluster(ctx, commandContext.Provider, commandContext.ClusterSpec, commandContext.WorkloadCluster)
	return nil
}

func (s *CollectWorkloadClusterDiagnosticsTask) Name() string {
	return "collect-workload-cluster-diagnostics"
}

// CollectMgmtClusterDiagnosticsTask implementation

func (s *CollectMgmtClusterDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting management cluster diagnostics")
	_ = commandContext.ClusterManager.SaveLogsManagementCluster(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
	return nil
}

func (s *CollectMgmtClusterDiagnosticsTask) Name() string {
	return "collect-management-cluster-diagnostics"
}
