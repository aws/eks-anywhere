package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

// createWorkloadClusterTask implementation.
type createWorkloadClusterTask struct{}

func (s *createWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new workload cluster")

	commandContext.ClusterSpec.Cluster.AddManagedByCLIAnnotation()
	commandContext.ClusterSpec.Cluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)

	if commandContext.ClusterSpec.Cluster.Namespace != "" {
		if err := commandContext.ClusterManager.CreateNamespace(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec.Cluster.Namespace); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectMgmtClusterDiagnosticsTask{}
		}
	}

	workloadCluster, err := commandContext.ClusterCreator.CreateSync(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	if commandContext.ClusterSpec.AWSIamConfig != nil {
		logger.Info("Generating the aws iam kubeconfig file")
		err = commandContext.ClusterManager.GenerateIamAuthKubeconfig(ctx, commandContext.BootstrapCluster, workloadCluster, commandContext.ClusterSpec)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	logger.Info("Creating EKS-A namespace")
	err = commandContext.ClusterManager.CreateEKSANamespace(ctx, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on workload cluster")
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	err = commandContext.ClusterManager.InstallCAPI(ctx, managementComponents, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Installing EKS-A secrets on workload cluster")
	err = commandContext.Provider.UpdateSecrets(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	err = commandContext.ClusterManager.CreateRegistryCredSecret(ctx, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &installProviderSpecificResources{}
}

func (s *createWorkloadClusterTask) Name() string {
	return "workload-cluster-init"
}

func (s *createWorkloadClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *createWorkloadClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}
