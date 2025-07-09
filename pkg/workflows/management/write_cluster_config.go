package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type writeUpgradeClusterConfig struct{}

// Run writeClusterConfig writes new management cluster's cluster config file to the destination after the upgrade process.
func (s *writeUpgradeClusterConfig) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := writeClusterConfigToDisk(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
	}

	// Handle AWS IAM kubeconfig generation/cleanup during management cluster upgrade
	if commandContext.CurrentClusterSpec != nil {
		hadAWSIam := commandContext.CurrentClusterSpec.AWSIamConfig != nil
		hasAWSIam := commandContext.ClusterSpec.AWSIamConfig != nil

		if !hadAWSIam && hasAWSIam {
			// AWS IAM being added during upgrade
			logger.Info("Generating AWS IAM kubeconfig file for management cluster upgrade (AWS IAM added)")
			err = commandContext.IamAuth.GenerateManagementKubeconfig(ctx, commandContext.ManagementCluster)
			if err != nil {
				commandContext.SetError(err)
				logger.Error(err, "Generating AWS IAM kubeconfig file for management cluster upgrade")
			}
		} else if hadAWSIam && !hasAWSIam {
			// AWS IAM being removed during upgrade - cleanup existing kubeconfig
			logger.Info("Cleaning up AWS IAM kubeconfig file (AWS IAM removed during management cluster upgrade)")
			err = commandContext.IamAuth.CleanupKubeconfig(commandContext.ManagementCluster.Name)
			if err != nil {
				logger.Error(err, "Failed to cleanup AWS IAM kubeconfig file")
			}
		}
	}

	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster upgraded!")
	}
	return &postClusterUpgrade{}
}

func (s *writeUpgradeClusterConfig) Name() string {
	return "write-cluster-config"
}

func (s *writeUpgradeClusterConfig) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *writeUpgradeClusterConfig) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &postClusterUpgrade{}, nil
}

type writeCreateClusterConfig struct{}

func (s *writeCreateClusterConfig) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := writeClusterConfigToDisk(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.AWSIamConfig != nil {
		logger.Info("Generating the aws iam kubeconfig file")
		err = commandContext.IamAuth.GenerateManagementKubeconfig(ctx, commandContext.WorkloadCluster)
		if err != nil {
			commandContext.SetError(err)
			return &workflows.CollectDiagnosticsTask{}
		}
	}

	return &deleteBootstrapClusterTask{}
}

func (s *writeCreateClusterConfig) Name() string {
	return "write-cluster-config"
}

func (s *writeCreateClusterConfig) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *writeCreateClusterConfig) Checkpoint() *task.CompletedTask {
	return nil
}

func writeClusterConfigToDisk(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig, writer filewriter.FileWriter) error {
	logger.Info("Writing cluster config file")
	if err := clustermarshaller.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs, writer); err != nil {
		return err
	}

	return nil
}
