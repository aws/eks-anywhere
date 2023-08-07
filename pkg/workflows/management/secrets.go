package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type updateSecrets struct{}

// Run updateSecrets updates management cluster's secrets.
func (s *updateSecrets) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := commandContext.Provider.UpdateSecrets(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	return &ensureEtcdCAPIComponentsExist{}
}

func (s *updateSecrets) Name() string {
	return "update-secrets"
}

func (s *updateSecrets) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateSecrets) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &ensureEtcdCAPIComponentsExist{}, nil
}
