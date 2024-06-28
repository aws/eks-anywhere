package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type installEksaComponentsOnBootstrapTask struct{}

func (s *installEksaComponentsOnBootstrapTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components on bootstrap cluster")
	err := installEKSAComponents(ctx, commandContext, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &createWorkloadClusterTask{}
}

func (s *installEksaComponentsOnBootstrapTask) Name() string {
	return "eksa-components-bootstrap-install"
}

func (s *installEksaComponentsOnBootstrapTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsOnBootstrapTask) Checkpoint() *task.CompletedTask {
	return nil
}

type installEksaComponentsOnWorkloadTask struct{}

func (s *installEksaComponentsOnWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// Removes the tinkerbell IP annotation added to the spec for controller based cluster creation. This
	// helps use the right admin machine URL for hegel URLS.
	if commandContext.ClusterSpec.Cluster.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		logger.Info("Removing Tinkerbell IP annotation")
		commandContext.ClusterSpec.Cluster.ClearTinkerbellIPAnnotation()
	}

	logger.Info("Installing EKS-A custom components on the management cluster")

	err := installEKSAComponents(ctx, commandContext, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	commandContext.ClusterSpec.Cluster.AddManagedByCLIAnnotation()
	commandContext.ClusterSpec.Cluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)

	srcClient, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.BootstrapCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	dstClient, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.WorkloadCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.Cluster.Namespace != "" {
		if err := workflows.CreateNamespaceIfNotPresent(ctx, commandContext.ClusterSpec.Cluster.Namespace, dstClient); err != nil {
			commandContext.SetError(err)
			return &workflows.CollectMgmtClusterDiagnosticsTask{}
		}
	}

	logger.Info("Moving cluster spec to workload cluster")
	if err = commandContext.ClusterMover.Move(ctx, commandContext.ClusterSpec, srcClient, dstClient); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if err = commandContext.ClusterManager.ResumeEKSAControllerReconcile(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	resourcesSpec, err := clustermarshaller.MarshalClusterSpec(commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
	}
	logger.V(6).Info(string(resourcesSpec))

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
	logger.Info("Installing EKS-D components")
	if err := commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, targetCluster); err != nil {
		commandContext.SetError(err)
		return err
	}

	logger.Info("Installing EKS-A custom components (CRD and controller)")
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)
	if err := commandContext.EksaInstaller.Install(ctx, logger.Get(), targetCluster, managementComponents, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return err
	}

	if err := commandContext.Provider.InstallCustomProviderComponents(ctx, targetCluster.KubeconfigFile); err != nil {
		commandContext.SetError(err)
		return err
	}

	if err := commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, targetCluster); err != nil {
		commandContext.SetError(err)
		return err
	}

	return nil
}
