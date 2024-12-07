package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Create is the workflow that creates a workload clusters.
type Create struct {
	clientFactory    interfaces.ClientFactory
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	clusterCreator   interfaces.ClusterCreator
	packageInstaller interfaces.PackageManager
	iamAuth          interfaces.AwsIamAuth
}

// NewCreate builds a new create construct.
func NewCreate(provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageManager,
	clusterCreator interfaces.ClusterCreator,
	clientFactory interfaces.ClientFactory,
	iamAuth interfaces.AwsIamAuth,
) *Create {
	createWorkflow := &Create{
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		clusterCreator:   clusterCreator,
		packageInstaller: packageInstaller,
		clientFactory:    clientFactory,
		iamAuth:          iamAuth,
	}

	return createWorkflow
}

// Run executes the tasks to create a workload cluster.
func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		ClientFactory:     c.clientFactory,
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
		ClusterCreator:    c.clusterCreator,
		IamAuth:           c.iamAuth,
	}

	return task.NewTaskRunner(&setAndValidateCreateWorkloadTask{}, c.writer).RunTask(ctx, commandContext)
}
