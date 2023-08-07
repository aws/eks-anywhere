package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type pauseGitOpsReconcile struct{}

func (s *pauseGitOpsReconcile) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Pausing GitOps cluster resources reconcile")
	err := commandContext.GitOpsManager.PauseClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	return &upgradeCoreComponents{}
}

func (s *pauseGitOpsReconcile) Name() string {
	return "pause-controllers-reconcile"
}

func (s *pauseGitOpsReconcile) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *pauseGitOpsReconcile) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeCoreComponents{}, nil
}

// resumeGitOps updates all the places that have a cluster definition to follow the cluster config provided to this workflow:
// the cluster config in the git repo if GitOps is enabled. It also resumes the GitOps reconciliations.
type resumeGitOps struct {
	eksaSpecDiff bool
}

func (s *resumeGitOps) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Updating Git Repo with new EKS-A cluster spec")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)
	err := commandContext.GitOpsManager.UpdateGitEksaSpec(ctx, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Forcing reconcile Git repo with latest commit")
	err = commandContext.GitOpsManager.ForceReconcileGitRepo(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	logger.Info("Resuming GitOps cluster resources kustomization")
	err = commandContext.GitOpsManager.ResumeClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &writeClusterConfig{}
	}
	if !s.eksaSpecDiff {
		return nil
	}
	return &writeClusterConfig{}
}

func (s *resumeGitOps) Name() string {
	return "resume-eksa-and-gitops-kustomization"
}

func (s *resumeGitOps) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *resumeGitOps) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &writeClusterConfig{}, nil
}
