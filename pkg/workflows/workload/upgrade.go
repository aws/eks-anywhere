package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Upgrade is a schema for upgrade cluster.
type Upgrade struct {
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	ClusterUpgrader  interfaces.ClusterUpgrader
	packageInstaller interfaces.PackageInstaller
}

// NewUpgrade builds a new upgrade construct.
func NewUpgrade(provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	clusterUpgrade interfaces.ClusterUpgrader,
	eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
) *Upgrade {
	return &Upgrade{
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		ClusterUpgrader:  clusterUpgrade,
		packageInstaller: packageInstaller,
	}
}

// Run Upgrade implements upgrade functionality for workload cluster's upgrade operation.
func (c *Upgrade) Run(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
		WorkloadCluster:   cluster,
		ClusterUpgrader:   c.ClusterUpgrader,
	}

	return task.NewTaskRunner(&setAndValidateWorkloadTask{}, c.writer).RunTask(ctx, commandContext)
}
