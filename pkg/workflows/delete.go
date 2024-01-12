package workflows

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Delete struct {
	bootstrapper   interfaces.Bootstrapper
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	gitOpsManager  interfaces.GitOpsManager
	writer         filewriter.FileWriter
}

func NewDelete(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
) *Delete {
	return &Delete{
		bootstrapper:   bootstrapper,
		provider:       provider,
		clusterManager: clusterManager,
		gitOpsManager:  gitOpsManager,
		writer:         writer,
	}
}

func (c *Delete) Run(ctx context.Context, workloadCluster *types.Cluster, clusterSpec *cluster.Spec, forceCleanup bool, kubeconfig string) error {
	if forceCleanup {
		if err := c.bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: workloadCluster.Name,
		}, constants.Delete, forceCleanup); err != nil {
			return err
		}
	}

	commandContext := &task.CommandContext{
		Bootstrapper:    c.bootstrapper,
		Provider:        c.provider,
		ClusterManager:  c.clusterManager,
		GitOpsManager:   c.gitOpsManager,
		WorkloadCluster: workloadCluster,
		ClusterSpec:     clusterSpec,
	}

	if clusterSpec.ManagementCluster != nil {
		commandContext.BootstrapCluster = clusterSpec.ManagementCluster
	}

	return task.NewTaskRunner(&setupAndValidate{}, c.writer).RunTask(ctx, commandContext)
}

type setupAndValidate struct{}

type createManagementCluster struct{}

type installCAPI struct{}

type moveClusterManagement struct{}

type deleteWorkloadCluster struct{}

type cleanupGitRepo struct{}

type deletePackageResources struct{}

type deleteManagementCluster struct{}

func (s *setupAndValidate) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Performing provider setup and validations")
	err := commandContext.Provider.SetupAndValidateDeleteCluster(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &createManagementCluster{}
}

func (s *setupAndValidate) Name() string {
	return "setup-and-validate"
}

func (s *setupAndValidate) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setupAndValidate) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *createManagementCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsManaged() {
		return &deleteWorkloadCluster{}
	}
	logger.Info("Creating management cluster")
	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts(commandContext.ClusterSpec)
	if err != nil {
		logger.Error(err, "Error getting management options from provider")
		commandContext.SetError(err)
		return nil
	}

	bootstrapCluster, err := commandContext.Bootstrapper.CreateBootstrapCluster(ctx, commandContext.ClusterSpec, bootstrapOptions...)
	if err != nil {
		commandContext.SetError(err)
		return &deleteManagementCluster{}
	}
	commandContext.BootstrapCluster = bootstrapCluster

	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err = commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, bootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &installCAPI{}
}

func (s *createManagementCluster) Name() string {
	return "management-cluster-init"
}

func (s *createManagementCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *createManagementCluster) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *installCAPI) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing cluster-api providers on management cluster")
	err := commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &deleteManagementCluster{}
	}
	return &moveClusterManagement{}
}

func (s *installCAPI) Name() string {
	return "install-capi"
}

func (s *installCAPI) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installCAPI) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *moveClusterManagement) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Moving cluster management from workload cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.WorkloadCluster, commandContext.BootstrapCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &deleteWorkloadCluster{}
}

func (s *moveClusterManagement) Name() string {
	return "cluster-management-move"
}

func (s *moveClusterManagement) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *moveClusterManagement) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *deleteWorkloadCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Deleting workload cluster")
	err := commandContext.ClusterManager.DeleteCluster(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.Provider, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &cleanupGitRepo{}
}

func (s *deleteWorkloadCluster) Name() string {
	return "delete-workload-cluster"
}

func (s *deleteWorkloadCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteWorkloadCluster) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *cleanupGitRepo) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Clean up Git Repo")
	err := commandContext.GitOpsManager.CleanupGitRepo(ctx, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &deletePackageResources{}
}

func (s *cleanupGitRepo) Name() string {
	return "clean-up-git-repo"
}

func (s *cleanupGitRepo) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *cleanupGitRepo) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *deletePackageResources) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		return &deleteManagementCluster{}
	}

	logger.Info("Delete package resources", "clusterName", commandContext.WorkloadCluster.Name)
	cluster := commandContext.ManagementCluster
	if cluster == nil {
		cluster = commandContext.BootstrapCluster
	}
	err := commandContext.ClusterManager.DeletePackageResources(ctx, cluster, commandContext.WorkloadCluster.Name)
	if err != nil {
		logger.Info("Problem delete package resources", "error", err)
	}

	// A bit odd to traverse to this state here, but it is the terminal state
	return &deleteManagementCluster{}
}

func (s *deletePackageResources) Name() string {
	return "package-resource-delete"
}

func (s *deletePackageResources) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deletePackageResources) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *deleteManagementCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.OriginalError != nil {
		collector := &CollectMgmtClusterDiagnosticsTask{}
		collector.Run(ctx, commandContext)
	}
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		if err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Delete, false); err != nil {
			commandContext.SetError(err)
		}
		return nil
	}
	logger.Info("Bootstrap cluster information missing - skipping delete kind cluster")
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster deleted!")
	}
	return nil
}

func (s *deleteManagementCluster) Name() string {
	return "kind-cluster-delete"
}

func (s *deleteManagementCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteManagementCluster) Checkpoint() *task.CompletedTask {
	return nil
}
