package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installNewComponents struct{}

func runInstallNewComponents(ctx context.Context, commandContext *task.CommandContext) error {
	if err := commandContext.ClusterManager.ApplyBundles(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return err
	}

	if err := commandContext.ClusterManager.ApplyReleases(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return err
	}

	err := commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	return nil
}

// Run installNewComponents performs actions needed to upgrade the management cluster.
func (s *installNewComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := runInstallNewComponents(ctx, commandContext); err != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	return &upgradeCluster{}
}

func (s *installNewComponents) Name() string {
	return "install-new-eksa-version-components"
}

func (s *installNewComponents) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *installNewComponents) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeCluster{}, nil
}
