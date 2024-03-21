package management

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeCuratedPackagesTask struct{}

func (s *upgradeCuratedPackagesTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	commandContext.PackageInstaller.InstallCuratedPackages(ctx)

	client, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.ManagementCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
	}

	packagesManager := &appsv1.Deployment{}

	err = client.Get(ctx, "eks-anywhere-packages", constants.EksaPackagesName, packagesManager)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	replicas := packagesManager.Spec.Replicas
	packagesManager.Spec.Replicas = ptr.Int32(0)
	err = client.Update(ctx, packagesManager)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	packagesManager.Spec.Replicas = replicas

	packagesManager.SetUID("")
	packagesManager.SetResourceVersion("")

	err = client.Update(ctx, packagesManager)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return nil
}

func (s *upgradeCuratedPackagesTask) Name() string {
	return "install-curated-packages"
}

func (s *upgradeCuratedPackagesTask) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *upgradeCuratedPackagesTask) Checkpoint() *task.CompletedTask {
	return nil
}
