package management

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeCluster struct{}

func (s *upgradeCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// TODO(g-gaston): move this to eks-a installer and eks-d installer
	err := commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	if commandContext.UpgradeChangeDiff.Changed() {
		if err := commandContext.ClusterManager.ApplyBundles(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}

		if err := commandContext.ClusterManager.ApplyReleases(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	logger.Info("Upgrading workload cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		// Take backup of bootstrap cluster capi components
		if commandContext.BootstrapCluster != nil {
			logger.Info("Backing up management components from bootstrap cluster")
			err := commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.BootstrapCluster, fmt.Sprintf("bootstrap-%s", commandContext.ManagementClusterStateDir), commandContext.ManagementCluster.Name)
			if err != nil {
				logger.Info("Bootstrap management component backup failed, use existing workload cluster backup", "error", err)
			}
		}
		return &workflows.CollectDiagnosticsTask{}
	}

	// TODO: move to its own task
	logger.V(3).Info("Resuming all workload clusters after moving management cluster resources from bootstrap to management clusters")
	err = commandContext.ClusterManager.ResumeCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &resumeGitOps{}
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
	return &resumeGitOps{}, nil
}