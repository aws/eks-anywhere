package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
)

type writeClusterConfig struct{}

// Run writeClusterConfig writes new management cluster's cluster config file to the destination after the create/upgrade process.
func (s *writeClusterConfig) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
	}

	successMsg := ""
	if commandContext.CurrentClusterSpec != nil {
		successMsg = "Cluster upgraded!"
	} else {
		successMsg = "Cluster created!"
	}

	if commandContext.OriginalError == nil {
		logger.MarkSuccess(successMsg)
	}
	if commandContext.CurrentClusterSpec != nil {
		return &postClusterUpgrade{}
	}
	return nil
}

func (s *writeClusterConfig) Name() string {
	return "write-cluster-config"
}

func (s *writeClusterConfig) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *writeClusterConfig) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	if commandContext.CurrentClusterSpec == nil {
		return &postClusterUpgrade{}, nil
	}
	return nil, nil
}
