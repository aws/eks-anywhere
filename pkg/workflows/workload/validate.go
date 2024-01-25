package workload

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type (
	setAndValidateUpgradeWorkloadTask struct{}
	setAndValidateCreateWorkloadTask  struct{}
)

// Run setAndValidateCreateWorkloadTask performs actions needed to validate creating the workload cluster.
func (s *setAndValidateCreateWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	runner := validations.NewRunner()
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.GitOpsManager.Validations(ctx, commandContext.ClusterSpec)...)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err := runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &createCluster{}
}

func (s *setAndValidateCreateWorkloadTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("workload cluster's %s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateCreateCluster(ctx, commandContext.ClusterSpec),
			}
		},
	}
}

func (s *setAndValidateCreateWorkloadTask) Name() string {
	return "setup-validate-create"
}

func (s *setAndValidateCreateWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setAndValidateCreateWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}

// Run setAndValidateWorkloadTask performs actions needed to validate the workload cluster.
func (s *setAndValidateUpgradeWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
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
	return &preClusterUpgrade{}
}

func (s *setAndValidateUpgradeWorkloadTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("workload cluster's %s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec),
			}
		},
	}
}

func (s *setAndValidateUpgradeWorkloadTask) Name() string {
	return "setup-validate-upgrade"
}

func (s *setAndValidateUpgradeWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setAndValidateUpgradeWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}
