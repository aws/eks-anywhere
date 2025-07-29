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
		logger.Error(err, "Writing cluster config file")

	}

	// Handle AWS IAM kubeconfig generation/cleanup during cluster operations
	if commandContext.CurrentClusterSpec == nil {
		// New cluster creation
		if commandContext.ClusterSpec.AWSIamConfig != nil {
			logger.Info("Generating AWS IAM kubeconfig file for new workload cluster")
			err = commandContext.IamAuth.GenerateWorkloadKubeconfig(ctx, commandContext.ManagementCluster, commandContext.WorkloadCluster, commandContext.ClusterSpec)
			if err != nil {
				commandContext.SetError(err)
				logger.Error(err, "Generating AWS IAM kubeconfig file for new workload cluster")
			}
		}
	} else {
		// Workload cluster upgrade scenarios
		hadAWSIam := commandContext.CurrentClusterSpec.AWSIamConfig != nil
		hasAWSIam := commandContext.ClusterSpec.AWSIamConfig != nil

		if !hadAWSIam && hasAWSIam {
			// AWS IAM being added during upgrade
			logger.Info("Generating AWS IAM kubeconfig file for workload cluster upgrade (AWS IAM added)")
			err = commandContext.IamAuth.GenerateWorkloadKubeconfig(ctx, commandContext.ManagementCluster, commandContext.WorkloadCluster, commandContext.ClusterSpec)
			if err != nil {
				commandContext.SetError(err)
				logger.Error(err, "Generating AWS IAM kubeconfig file for workload cluster upgrade")
			}
		} else if hadAWSIam && !hasAWSIam {
			// AWS IAM being removed during upgrade - cleanup existing kubeconfig
			logger.Info("Cleaning up AWS IAM kubeconfig file (AWS IAM removed during workload cluster upgrade)")
			err = commandContext.IamAuth.CleanupKubeconfig(commandContext.WorkloadCluster.Name)
			if err != nil {
				logger.Error(err, "Failed to cleanup AWS IAM kubeconfig file")
			}
		}
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
