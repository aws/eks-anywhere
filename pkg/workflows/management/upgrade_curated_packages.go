package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/task"
)

type upgradeCuratedPackagesTask struct{}

func (s *upgradeCuratedPackagesTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.CurrentClusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Equal(commandContext.ClusterSpec.Cluster.Spec.RegistryMirrorConfiguration) {
		return nil
	}

	commandContext.PackageManager.UpgradeCuratedPackages(ctx)

	return nil
}

func (s *upgradeCuratedPackagesTask) Name() string {
	return "upgrade-curated-packages"
}

func (s *upgradeCuratedPackagesTask) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *upgradeCuratedPackagesTask) Checkpoint() *task.CompletedTask {
	return nil
}
