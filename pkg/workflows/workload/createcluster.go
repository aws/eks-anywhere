package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createCluster struct{}

// Run createCluster performs actions needed to create the management cluster.
func (c *createCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating workload cluster")

	if commandContext.ClusterSpec.Cluster.Namespace != "" {
		if err := workflows.CreateNamespaceIfNotPresent(ctx, commandContext.ClusterSpec.Cluster.Namespace, commandContext.ManagementCluster.KubeconfigFile, commandContext.ClientFactory); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectMgmtClusterDiagnosticsTask{}
		}
	}

	workloadCluster, err := commandContext.ClusterCreator.CreateSync(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	resourcesSpec, err := clustermarshaller.MarshalClusterSpec(commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
	}
	logger.V(6).Info(string(resourcesSpec))

	return &installGitOpsManagerTask{}
}

func (c *createCluster) Name() string {
	return "create-workload-cluster"
}

func (c *createCluster) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (c *createCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &installGitOpsManagerTask{}, nil
}
