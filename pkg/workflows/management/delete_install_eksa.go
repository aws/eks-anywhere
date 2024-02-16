package management

import (
	"context"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type installEksaComponentsOnBootstrapForDeleteTask struct{}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing EKS-A custom components on bootstrap cluster")
	err := installEKSAComponents(ctx, commandContext, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	commandContext.ClusterSpec.Cluster.PauseReconcile()
	commandContext.ClusterSpec.Cluster.SetFinalizers([]string{"clusters.anywhere.eks.amazonaws.com/finalizer"})
	commandContext.ClusterSpec.Cluster.AddManagedByCLIAnnotation()
	err = applyClusterSpecOnBootstrapForDeleteTask(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.ClientFactory)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &deleteManagementCluster{}
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Name() string {
	return "eksa-components-bootstrap-install-delete-task"
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *installEksaComponentsOnBootstrapForDeleteTask) Checkpoint() *task.CompletedTask {
	return nil
}

func applyClusterSpecOnBootstrapForDeleteTask(ctx context.Context, spec *cluster.Spec, cluster *types.Cluster, clientFactory interfaces.ClientFactory) error {
	if err := workflows.CreateNamespaceIfNotPresent(ctx, spec.Cluster.Namespace, cluster.KubeconfigFile, clientFactory); err != nil {
		return errors.Wrapf(err, "creating namespace on bootstrap")
	}

	client, err := clientFactory.BuildClientFromKubeconfig(cluster.KubeconfigFile)
	if err != nil {
		return errors.Wrap(err, "building client to apply cluster spec changes")
	}

	for _, obj := range spec.ClusterAndChildren() {
		if err := client.ApplyServerSide(ctx,
			"eks-a-cli",
			obj,
			kubernetes.ApplyServerSideOptions{ForceOwnership: true},
		); err != nil {
			return errors.Wrapf(err, "applying cluster spec")
		}
	}

	return nil
}
