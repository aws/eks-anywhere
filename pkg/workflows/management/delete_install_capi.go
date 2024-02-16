package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installCAPIComponentsForDeleteTask struct{}

func (s *installCAPIComponentsForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err := commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on bootstrap cluster")
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	if err := commandContext.ClusterManager.InstallCAPI(ctx, managementComponents, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &deleteManagementCluster{}
	}

	return &moveClusterManagementForDeleteTask{}
}

func (s *installCAPIComponentsForDeleteTask) Name() string {
	return "install-capi-components-bootstrap-for-delete"
}

func (s *installCAPIComponentsForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installCAPIComponentsForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}
