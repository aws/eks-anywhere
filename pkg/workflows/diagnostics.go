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

func (s *CollectDiagnosticsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return s.Run(ctx, commandContext), nil
}

func (s *CollectDiagnosticsTask) Checkpoint() *task.CompletedTask {
	return nil
}

// CollectWorkloadClusterDiagnosticsTask implementation

// Run starts collecting the logs for workload cluster diagnostics.
func (s *CollectWorkloadClusterDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting workload cluster diagnostics")
	_ = commandContext.ClusterManager.SaveLogsWorkloadCluster(ctx, commandContext.Provider, commandContext.ClusterSpec, commandContext.WorkloadCluster)
	return nil
}

// Name returns the name of CollectWorkloadClusterDiagnosticsTask.
func (s *CollectWorkloadClusterDiagnosticsTask) Name() string {
	return "collect-workload-cluster-diagnostics"
}

// Restore restores from CollectWorkloadClusterDiagnosticsTask.
func (s *CollectWorkloadClusterDiagnosticsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

// Checkpoint sets a checkpoint at CollectWorkloadClusterDiagnosticsTask.
func (s *CollectWorkloadClusterDiagnosticsTask) Checkpoint() *task.CompletedTask {
	return nil
}

// CollectMgmtClusterDiagnosticsTask implementation

func (s *CollectMgmtClusterDiagnosticsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("collecting management cluster diagnostics")
	mgmt := commandContext.BootstrapCluster
	if mgmt == nil {
		mgmt = commandContext.ManagementCluster
	}

	_ = commandContext.ClusterManager.SaveLogsManagementCluster(ctx, commandContext.ClusterSpec, mgmt)
	return nil
}

func (s *CollectMgmtClusterDiagnosticsTask) Name() string {
	return "collect-management-cluster-diagnostics"
}

func (s *CollectMgmtClusterDiagnosticsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *CollectMgmtClusterDiagnosticsTask) Checkpoint() *task.CompletedTask {
	return nil
}
