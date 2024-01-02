package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createBootStrapClusterTask struct{}

type createWorkloadClusterTask struct{}

func (s *createBootStrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new bootstrap cluster")

	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts(commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	bootstrapCluster, err := commandContext.Bootstrapper.CreateBootstrapCluster(ctx, commandContext.ClusterSpec, bootstrapOptions...)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.BootstrapCluster = bootstrapCluster

	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err = commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, bootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on bootstrap cluster")
	if err = commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, bootstrapCluster, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.AWSIamConfig != nil {
		logger.Info("Creating aws-iam-authenticator certificate and key pair secret on bootstrap cluster")
		if err = commandContext.ClusterManager.CreateAwsIamAuthCaSecret(ctx, bootstrapCluster, commandContext.ClusterSpec.Cluster.Name); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectMgmtClusterDiagnosticsTask{}
		}
	}

	logger.Info("Provider specific post-setup")
	if err = commandContext.Provider.PostBootstrapSetup(ctx, commandContext.ClusterSpec.Cluster, bootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &installEksaComponentsTask{}
}

func (s *createBootStrapClusterTask) Name() string {
	return "bootstrap-cluster-init"
}

func (s *createBootStrapClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *createBootStrapClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}

// createWorkloadClusterTask implementation

func (s *createWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new workload cluster")
	commandContext.ClusterSpec.Cluster.AddManagedByCLIAnnotation()

	logger.Info("Applying cluster spec to bootstrap cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.BootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	workloadCluster, err := commandContext.ClusterManager.GetWorkloadCluster(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	logger.Info("Creating EKS-A namespace")
	err = commandContext.ClusterManager.CreateEKSANamespace(ctx, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on workload cluster")
	err = commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
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

	return &installResourcesOnManagementTask{}
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
