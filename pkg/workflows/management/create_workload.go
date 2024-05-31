package management

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

// createWorkloadClusterTask implementation.
type createWorkloadClusterTask struct{}

func (s *createWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	clusterType := "workload"
	if commandContext.ClusterSpec.Config.Cluster.IsSelfManaged() {
		clusterType = "management"
	}
	logger.Info(fmt.Sprintf("Creating new %s cluster", clusterType))

	commandContext.ClusterSpec.Cluster.AddManagedByCLIAnnotation()
	commandContext.ClusterSpec.Cluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)

	client, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.BootstrapCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.Cluster.Namespace != "" {
		if err := workflows.CreateNamespaceIfNotPresent(ctx, commandContext.ClusterSpec.Cluster.Namespace, client); err != nil {
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

	logger.Info("Creating EKS-A namespace")
	err = commandContext.ClusterManager.CreateEKSANamespace(ctx, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info(fmt.Sprintf("Installing cluster-api providers on %s cluster", clusterType))
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	err = commandContext.ClusterManager.InstallCAPI(ctx, managementComponents, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info(fmt.Sprintf("Installing EKS-A secrets on %s cluster", clusterType))
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
