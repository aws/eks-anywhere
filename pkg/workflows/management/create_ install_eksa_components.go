package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installEksaComponentsTask struct{}

func (s *installEksaComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components (CRD and controller) on bootstrap cluster")
	err := commandContext.ClusterManager.InstallCustomComponents(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	logger.Info("Installing EKS-D components on bootstrap cluster")
	err = commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Creating EKS-A CRDs instances on bootstrap cluster")
	err = commandContext.ClusterManager.CreateEKSAReleaseBundle(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}
	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
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

	logger.Info("Creating EKS-A CRDs instances on workload cluster")

	err = commandContext.ClusterManager.CreateEKSAReleaseBundle(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Applying cluster spec to workload cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.WorkloadCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	// datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	// machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	// err = commandContext.ClusterManager.ApplyEKSASpec(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	// if err != nil {
	// 	commandContext.SetError(err)
	// 	return &workflows.CollectDiagnosticsTask{}
	// }

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
