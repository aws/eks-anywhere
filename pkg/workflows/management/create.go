package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// CreateManagement is a schema for create cluster.
type CreateManagement struct {
	bootstrapper     interfaces.Bootstrapper
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	packageInstaller interfaces.PackageInstaller
	clusterUpgrader  interfaces.ClusterUpgrader
}

// NewCreateManagement builds a new create construct.
func NewCreateManagement(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter, eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
	clusterUpgrade interfaces.ClusterUpgrader,
) *CreateManagement {
	return &CreateManagement{
		bootstrapper:     bootstrapper,
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		packageInstaller: packageInstaller,
		clusterUpgrader:  clusterUpgrade,
	}
}

// Run runs all the create management cluster tasks.
func (c *CreateManagement) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator, forceCleanup bool) error {
	if forceCleanup {
		if err := c.bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: clusterSpec.Cluster.Name,
		}, constants.Create, forceCleanup); err != nil {
			return err
		}
	}
	commandContext := &task.CommandContext{
		Bootstrapper:     c.bootstrapper,
		Provider:         c.provider,
		ClusterManager:   c.clusterManager,
		GitOpsManager:    c.gitOpsManager,
		ClusterSpec:      clusterSpec,
		Writer:           c.writer,
		Validations:      validator,
		EksdInstaller:    c.eksdInstaller,
		PackageInstaller: c.packageInstaller,
		ClusterUpgrader:  c.clusterUpgrader,
	}

	if clusterSpec.ManagementCluster != nil {
		commandContext.BootstrapCluster = clusterSpec.ManagementCluster
	}

	err := task.NewTaskRunner(&setupAndValidateCreate{}, c.writer).RunTask(ctx, commandContext)

	return err
}
