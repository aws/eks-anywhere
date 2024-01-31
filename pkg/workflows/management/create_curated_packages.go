package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/task"
)

type installCuratedPackagesTask struct{}

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
