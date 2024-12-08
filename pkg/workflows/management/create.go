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
	bootstrapper   interfaces.Bootstrapper
	clientFactory  interfaces.ClientFactory
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	gitOpsManager  interfaces.GitOpsManager
	writer         filewriter.FileWriter
	eksdInstaller  interfaces.EksdInstaller
	packageManager interfaces.PackageManager
	clusterCreator interfaces.ClusterCreator
	eksaInstaller  interfaces.EksaInstaller
	clusterMover   interfaces.ClusterMover
	iamAuth        interfaces.AwsIamAuth
}

// NewCreate builds a new create construct.
func NewCreate(bootstrapper interfaces.Bootstrapper,
	clientFactory interfaces.ClientFactory, provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter, eksdInstaller interfaces.EksdInstaller,
	packageManager interfaces.PackageManager,
	clusterCreator interfaces.ClusterCreator,
	eksaInstaller interfaces.EksaInstaller,
	mover interfaces.ClusterMover,
	iamAuth interfaces.AwsIamAuth,
) *Create {
	createWorkflow := &Create{
		bootstrapper:   bootstrapper,
		clientFactory:  clientFactory,
		provider:       provider,
		clusterManager: clusterManager,
		gitOpsManager:  gitOpsManager,
		writer:         writer,
		eksdInstaller:  eksdInstaller,
		packageManager: packageManager,
		clusterCreator: clusterCreator,
		eksaInstaller:  eksaInstaller,
		clusterMover:   mover,
		iamAuth:        iamAuth,
	}

	return createWorkflow
}

// Run runs all the create management cluster tasks.
func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		Bootstrapper:   c.bootstrapper,
		ClientFactory:  c.clientFactory,
		Provider:       c.provider,
		ClusterManager: c.clusterManager,
		GitOpsManager:  c.gitOpsManager,
		ClusterSpec:    clusterSpec,
		Writer:         c.writer,
		Validations:    validator,
		EksdInstaller:  c.eksdInstaller,
		PackageManager: c.packageManager,
		ClusterCreator: c.clusterCreator,
		EksaInstaller:  c.eksaInstaller,
		ClusterMover:   c.clusterMover,
		IamAuth:        c.iamAuth,
	}

	return task.NewTaskRunner(&setupAndValidateCreate{}, c.writer).RunTask(ctx, commandContext)
}
