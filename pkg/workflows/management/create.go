package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Create is a schema for create cluster.
type Create struct {
	bootstrapper     interfaces.Bootstrapper
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	packageInstaller interfaces.PackageInstaller
	clusterUpgrader  interfaces.ClusterUpgrader
}

// NewCreate builds a new create construct.
func NewCreate(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter, eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
	clusterUpgrade interfaces.ClusterUpgrader,
) *Create {
	return &Create{
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
func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator, forceCleanup bool) error {
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

	return task.NewTaskRunner(&setupAndValidateCreate{}, c.writer).RunTask(ctx, commandContext)
}
