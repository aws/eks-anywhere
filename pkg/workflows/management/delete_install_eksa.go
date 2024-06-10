package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installEksaComponentsOnBootstrapForDeleteTask struct{}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components on bootstrap cluster")
	err := installEKSAComponents(ctx, commandContext, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	srcClient, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.WorkloadCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	dstClient, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.BootstrapCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if err := workflows.CreateNamespaceIfNotPresent(ctx, commandContext.ClusterSpec.Cluster.Namespace, dstClient); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	err = commandContext.ClusterMover.Move(ctx, commandContext.ClusterSpec, srcClient, dstClient)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	if err = commandContext.ClusterManager.AllowDeleteWhilePaused(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &deleteManagementCluster{}
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Name() string {
	return "eksa-components-bootstrap-install-delete-task"
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}
