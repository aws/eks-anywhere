package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type ensureEtcdCAPIComponentsExist struct{}

// Run ensureEtcdCAPIComponentsExist ensures ETCD CAPI providers on the management cluster.
func (s *ensureEtcdCAPIComponentsExist) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Ensuring etcd CAPI providers exist on management cluster before upgrade")
	if err := commandContext.CAPIManager.EnsureEtcdProvidersInstallation(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	return &pauseGitOpsReconcile{}
}

func (s *ensureEtcdCAPIComponentsExist) Name() string {
	return "ensure-etcd-capi-components-exist"
}

func (s *ensureEtcdCAPIComponentsExist) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *ensureEtcdCAPIComponentsExist) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return &pauseGitOpsReconcile{}, nil
}

type upgradeCoreComponents struct {
	UpgradeChangeDiff *types.ChangeDiff
}

func runUpgradeCoreComponents(ctx context.Context, commandContext *task.CommandContext) error {
	logger.Info("Upgrading core components")

	err := commandContext.Provider.PreCoreComponentsUpgrade(
		ctx,
		commandContext.ManagementCluster,
		commandContext.ClusterSpec,
	)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	changeDiff, err := commandContext.CAPIManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	if err = commandContext.GitOpsManager.Install(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return err
	}

	changeDiff, err = commandContext.GitOpsManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.ClusterManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.EksdUpgrader.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	return nil
}

// Run upgradeCoreComponents upgrades pre cluster upgrade components.
func (s *upgradeCoreComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := runUpgradeCoreComponents(ctx, commandContext); err != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	s.UpgradeChangeDiff = commandContext.UpgradeChangeDiff

	return &preClusterUpgrade{}
}

func (s *upgradeCoreComponents) Name() string {
	return "upgrade-core-components"
}

func (s *upgradeCoreComponents) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: s.UpgradeChangeDiff,
	}
}

func (s *upgradeCoreComponents) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	s.UpgradeChangeDiff = &types.ChangeDiff{}
	if err := task.UnmarshalTaskCheckpoint(completedTask.Checkpoint, s.UpgradeChangeDiff); err != nil {
		return nil, err
	}
	commandContext.UpgradeChangeDiff = s.UpgradeChangeDiff
	return &preClusterUpgrade{}, nil
}
