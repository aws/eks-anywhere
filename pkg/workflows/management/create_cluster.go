package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createBootStrapClusterTask struct{}

type createWorkloadClusterTask struct{}

type installResourcesOnManagementTask struct{}

type installEksaComponentsTask struct{}

type installEksaComponentsOnWorkloadTask struct{}

type installGitOpsManagerTask struct{}

type moveClusterManagementTask struct{}

type writeClusterConfigTask struct{}

type deleteBootstrapClusterTask struct{}

type installCuratedPackagesTask struct{}

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

func (s *installEksaComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Installing EKS-A custom components (CRD and controller) on workload cluster")
		err := commandContext.ClusterManager.InstallCustomComponents(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
		logger.Info("Installing EKS-D components on workload cluster")
		err = commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	logger.Info("Creating EKS-A CRDs instances on workload cluster")
	err := commandContext.ClusterManager.CreateEKSAReleaseBundle(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Applying cluster spec to bootstrap cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.BootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &createWorkloadClusterTask{}
}

func (s *installEksaComponentsTask) Name() string {
	return "eksa-components-install"
}

func (s *installEksaComponentsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsTask) Checkpoint() *task.CompletedTask {
	return nil
}

// createWorkloadClusterTask implementation

func (s *createWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new workload cluster")
	workloadCluster, err := commandContext.ClusterManager.GetWorkloadCluster(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Creating EKS-A namespace")
		err := commandContext.ClusterManager.CreateEKSANamespace(ctx, commandContext.WorkloadCluster)
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

func (s *installResourcesOnManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster.ExistingManagement {
		return &moveClusterManagementTask{}
	}
	logger.Info("Installing resources on management cluster")

	if err := commandContext.Provider.PostWorkloadInit(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	return &moveClusterManagementTask{}
}

func (s *installResourcesOnManagementTask) Name() string {
	return "install-resources-on-management-cluster"
}

func (s *installResourcesOnManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installResourcesOnManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *moveClusterManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster.ExistingManagement {
		return &installEksaComponentsOnWorkloadTask{}
	}
	logger.Info("Moving cluster management from bootstrap to workload cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef())
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &installEksaComponentsOnWorkloadTask{}
}

func (s *moveClusterManagementTask) Name() string {
	return "capi-management-move"
}

func (s *moveClusterManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *moveClusterManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *installEksaComponentsOnWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Installing EKS-A custom components (CRD and controller) on workload cluster")
		err := commandContext.ClusterManager.InstallCustomComponents(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
		logger.Info("Installing EKS-D components on workload cluster")
		err = commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	logger.Info("Creating EKS-A CRDs instances on workload cluster")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	err := commandContext.ClusterManager.CreateEKSAReleaseBundle(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	err = commandContext.ClusterManager.ApplyEKSASpec(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &installGitOpsManagerTask{}
}

func (s *installEksaComponentsOnWorkloadTask) Name() string {
	return "eksa-components-workload-install"
}

func (s *installEksaComponentsOnWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsOnWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *installGitOpsManagerTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing GitOps Toolkit on workload cluster")

	err := commandContext.GitOpsManager.InstallGitOps(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec))
	if err != nil {
		logger.MarkFail("Error when installing GitOps toolkits on workload cluster; EKS-A will continue with cluster creation, but GitOps will not be enabled", "error", err)
		return &writeClusterConfigTask{}
	}
	return &writeClusterConfigTask{}
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

func (s *writeClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	return &deleteBootstrapClusterTask{}
}

func (s *writeClusterConfigTask) Name() string {
	return "write-cluster-config"
}

func (s *writeClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *writeClusterConfigTask) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *deleteBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Deleting bootstrap cluster")
		err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Create, false)
		if err != nil {
			commandContext.SetError(err)
		}
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster created!")
	}
	return &installCuratedPackagesTask{}
}

func (s *deleteBootstrapClusterTask) Name() string {
	return "delete-kind-cluster"
}

func (s *deleteBootstrapClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteBootstrapClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *installCuratedPackagesTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	commandContext.PackageInstaller.InstallCuratedPackages(ctx)
	return nil
}

func (s *installCuratedPackagesTask) Name() string {
	return "install-curated-packages"
}

func (s *installCuratedPackagesTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installCuratedPackagesTask) Checkpoint() *task.CompletedTask {
	return nil
}
