package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type (
	cleanupGitRepo     struct{}
	postDeleteWorkload struct{}
)

func (s *postDeleteWorkload) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	commandContext.Writer.CleanUp()

	if commandContext.OriginalError != nil {
		collector := &workflows.CollectMgmtClusterDiagnosticsTask{}
		collector.Run(ctx, commandContext)
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster deleted!")
	}
	return nil
}

func (s *postDeleteWorkload) Name() string {
	return "validate-delete-workload-success"
}

func (s *postDeleteWorkload) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *postDeleteWorkload) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *cleanupGitRepo) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Clean up Git Repo")
	err := commandContext.GitOpsManager.CleanupGitRepo(ctx, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &postDeleteWorkload{}
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
