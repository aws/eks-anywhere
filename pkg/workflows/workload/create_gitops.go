package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type installGitOpsManagerTask struct{}

func (s *installGitOpsManagerTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing GitOps Toolkit on workload cluster")

	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	err := commandContext.GitOpsManager.InstallGitOps(ctx, commandContext.WorkloadCluster, managementComponents, commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec))
	if err != nil {
		logger.MarkFail("Error when installing GitOps toolkits on workload cluster; EKS-A will continue with cluster creation, but GitOps will not be enabled", "error", err)
	}
	return &writeClusterConfig{}
}

func (s *installGitOpsManagerTask) Name() string {
	return "gitops-manager-install"
}

func (s *installGitOpsManagerTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installGitOpsManagerTask) Checkpoint() *task.CompletedTask {
	return nil
}
