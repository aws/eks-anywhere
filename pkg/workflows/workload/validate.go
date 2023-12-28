package workload

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type setAndValidateWorkloadTask struct{}

// Run setAndValidateWorkloadTask performs actions needed to validate the workload cluster.
func (s *setAndValidateWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	currentSpec, err := commandContext.ClusterManager.GetCurrentClusterSpec(ctx, commandContext.ClusterSpec.ManagementCluster, commandContext.ClusterSpec.Cluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.CurrentClusterSpec = currentSpec
	runner := validations.NewRunner()
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.GitOpsManager.Validations(ctx, commandContext.ClusterSpec)...)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err = runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &upgradeCluster{}
}

func (s *setAndValidateWorkloadTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("workload cluster's %s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec),
			}
		},
	}
}

func (s *setAndValidateWorkloadTask) Name() string {
	return "setup-validate"
}

func (s *setAndValidateWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setAndValidateWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}
