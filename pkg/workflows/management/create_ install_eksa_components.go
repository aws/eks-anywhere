package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installEksaComponentsTask struct{}

func (s *installEksaComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components on bootstrap cluster")
	err := installEKSAComponents(ctx, commandContext, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &createWorkloadClusterTask{}
}

func (s *installEksaComponentsTask) Name() string {
	return "eksa-components-bootstrap-install"
}

func (s *installEksaComponentsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsTask) Checkpoint() *task.CompletedTask {
	return nil
}

type installEksaComponentsOnWorkloadTask struct{}

func (s *installEksaComponentsOnWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components on workload cluster")

	err := installEKSAComponents(ctx, commandContext, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Applying cluster spec to workload cluster")
	if err = commandContext.ClusterCreate.ClusterCreator.Run(ctx, commandContext.ClusterSpec, *commandContext.WorkloadCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if err = commandContext.ClusterManager.RemoveManagedByCLIAnnotationForCluster(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
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

func installEKSAComponents(ctx context.Context, commandContext *task.CommandContext, targetCluster *types.Cluster) error {

	logger.Info("Installing EKS-A custom components (CRD and controller)")
	err := commandContext.ClusterManager.InstallCustomComponents(ctx, commandContext.ClusterSpec, targetCluster, commandContext.Provider)
	if err != nil {
		return err
	}
	logger.Info("Installing EKS-D components")
	err = commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, targetCluster)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	logger.Info("Creating EKS-A CRDs instances")

	err = commandContext.ClusterManager.CreateEKSAReleaseBundle(ctx, targetCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, targetCluster)
	if err != nil {
		return err
	}

	return nil
}
