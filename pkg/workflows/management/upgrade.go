package management

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Upgrade struct {
	provider          providers.Provider
	clusterManager    interfaces.ClusterManager
	gitOpsManager     interfaces.GitOpsManager
	writer            filewriter.FileWriter
	capiManager       interfaces.CAPIManager
	eksdInstaller     interfaces.EksdInstaller
	eksdUpgrader      interfaces.EksdUpgrader
	upgradeChangeDiff *types.ChangeDiff
	clusterUpgrader   interfaces.ClusterUpgrader
}

func NewUpgrade(provider providers.Provider,
	capiManager interfaces.CAPIManager,
	clusterManager interfaces.ClusterManager,
	gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	eksdUpgrader interfaces.EksdUpgrader,
	eksdInstaller interfaces.EksdInstaller,
	clusterUpgrade interfaces.ClusterUpgrader,
) *Upgrade {
	upgradeChangeDiff := types.NewChangeDiff()
	return &Upgrade{
		provider:          provider,
		clusterManager:    clusterManager,
		gitOpsManager:     gitOpsManager,
		writer:            writer,
		capiManager:       capiManager,
		eksdUpgrader:      eksdUpgrader,
		eksdInstaller:     eksdInstaller,
		upgradeChangeDiff: upgradeChangeDiff,
		clusterUpgrader:   clusterUpgrade,
	}
}

func (c *Upgrade) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ManagementCluster: managementCluster,
		ClusterSpec:       clusterSpec,
		Validations:       validator,
		Writer:            c.writer,
		CAPIManager:       c.capiManager,
		EksdInstaller:     c.eksdInstaller,
		EksdUpgrader:      c.eksdUpgrader,
		UpgradeChangeDiff: c.upgradeChangeDiff,
		ClusterUpgrader:   c.clusterUpgrader,
	}
	if features.IsActive(features.CheckpointEnabled()) {
		return task.NewTaskRunner(&setupAndValidate{}, c.writer, task.WithCheckpointFile()).RunTask(ctx, commandContext)
	}

	return task.NewTaskRunner(&setupAndValidate{}, c.writer).RunTask(ctx, commandContext)
}

type writeClusterConfig struct{}

func (s *writeClusterConfig) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
	}

	// TODO(g-gaston): move to another task
	// What is this for?
	capiObjectFile := filepath.Join(commandContext.BootstrapCluster.Name, commandContext.ManagementClusterStateDir)
	if err := os.RemoveAll(capiObjectFile); err != nil {
		logger.Info(fmt.Sprintf("management cluster CAPI backup file not found: %v", err))
	}

	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster upgraded!")
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
	return nil, nil
}
