package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
	"github.com/aws/eks-anywhere/pkg/workflows/structs"
)

// Create is the workflow that creates a workload clusters.
type Create struct {
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	clusterCreate    structs.ClusterCreate
	packageInstaller interfaces.PackageInstaller
}

// NewCreate builds a new create construct.
func NewCreate(provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	clusterCreate structs.ClusterCreate,
	eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
) *Create {
	return &Create{
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		clusterCreate:    clusterCreate,
		packageInstaller: packageInstaller,
	}
}

// Run executes the tasks to create a workload cluster.
func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
		ClusterCreate:     c.clusterCreate,
	}

	return task.NewTaskRunner(&setAndValidateCreateWorkloadTask{}, c.writer).RunTask(ctx, commandContext)
}
