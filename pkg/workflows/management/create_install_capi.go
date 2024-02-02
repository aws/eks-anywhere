package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installCAPIComponentsTask struct{}

func (s *installCAPIComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err := commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on bootstrap cluster")
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	if err := commandContext.ClusterManager.InstallCAPI(ctx, managementComponents, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Provider specific post-setup")
	if err := commandContext.Provider.PostBootstrapSetup(ctx, commandContext.ClusterSpec.Cluster, commandContext.BootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &installEksaComponentsOnBootstrapTask{}
}

func (s *installCAPIComponentsTask) Name() string {
	return "install-capi-components-bootstrap"
}

func (s *installCAPIComponentsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installCAPIComponentsTask) Checkpoint() *task.CompletedTask {
	return nil
}
