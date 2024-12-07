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
	clientFactory    interfaces.ClientFactory
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	clusterUpgrader  interfaces.ClusterUpgrader
	packageInstaller interfaces.PackageManager
	iamAuth          interfaces.AwsIamAuth
}

// NewUpgrade builds a new upgrade construct.
func NewUpgrade(clientFactory interfaces.ClientFactory,
	provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	clusterUpgrader interfaces.ClusterUpgrader,
	eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageManager,
	iamAuth interfaces.AwsIamAuth,
) *Upgrade {
	upgradeWorkflow := &Upgrade{
		clientFactory:    clientFactory,
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		clusterUpgrader:  clusterUpgrader,
		packageInstaller: packageInstaller,
		iamAuth:          iamAuth,
	}

	return upgradeWorkflow
}

// Run Upgrade implements upgrade functionality for workload cluster's upgrade operation.
func (c *Upgrade) Run(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		ClientFactory:     c.clientFactory,
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
		WorkloadCluster:   cluster,
		ClusterUpgrader:   c.clusterUpgrader,
		IamAuth:           c.iamAuth,
	}

	return task.NewTaskRunner(&setAndValidateUpgradeWorkloadTask{}, c.writer).RunTask(ctx, commandContext)
}
