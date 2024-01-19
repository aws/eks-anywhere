package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Upgrade is a schema for upgrade cluster.
type Upgrade struct {
	clientFactory     interfaces.ClientFactory
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

// NewUpgrade builds a new upgrade construct.
func NewUpgrade(clientFactory interfaces.ClientFactory, provider providers.Provider,
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
		clientFactory:     clientFactory,
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

// Run Upgrade implements upgrade functionality for management cluster's upgrade operation.
func (c *Upgrade) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		ClientFactory:     c.clientFactory,
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
		return task.NewTaskRunner(&setupAndValidateUpgrade{}, c.writer, task.WithCheckpointFile()).RunTask(ctx, commandContext)
	}

	return task.NewTaskRunner(&setupAndValidateUpgrade{}, c.writer).RunTask(ctx, commandContext)
}
