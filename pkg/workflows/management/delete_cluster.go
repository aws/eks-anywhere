package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteManagementCluster struct{}

func (s *deleteManagementCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Deleting management cluster")

	err := commandContext.ClusterDeleter.Run(ctx, commandContext.ClusterSpec, *commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	err = commandContext.Provider.PostClusterDeleteValidate(ctx, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &cleanupGitRepo{}
}

func (s *deleteManagementCluster) Name() string {
	return "delete-management-cluster"
}

func (s *deleteManagementCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *deleteManagementCluster) Checkpoint() *task.CompletedTask {
	return nil
}

type cleanupGitRepo struct{}

func (s *cleanupGitRepo) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Clean up Git Repo")
	err := commandContext.GitOpsManager.CleanupGitRepo(ctx, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &deleteBootstrapClusterForDeleteTask{}
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
