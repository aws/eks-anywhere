package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
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

	return nil
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

func installEKSAComponents(ctx context.Context, commandContext *task.CommandContext, targetCluster *types.Cluster) error {
	logger.Info("Installing EKS-D components")
	if err := commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, targetCluster); err != nil {
		commandContext.SetError(err)
		return err
	}

	logger.Info("Installing EKS-A custom components (CRD and controller)")
	if err := commandContext.EksaInstaller.Install(ctx, logger.Get(), targetCluster, commandContext.ClusterSpec); err != nil {
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

	client, err := commandContext.ClientFactory.BuildClientFromKubeconfig(targetCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	commandContext.ClusterSpec.Cluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)
	if err := client.ApplyServerSide(ctx,
		constants.EKSACLIFieldManager,
		commandContext.ClusterSpec.Cluster,
		kubernetes.ApplyServerSideOptions{ForceOwnership: true},
	); err != nil {
		commandContext.SetError(err)
		return err
	}

	return nil
}
