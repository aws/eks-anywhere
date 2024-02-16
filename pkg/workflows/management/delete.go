package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Delete is the workflow that deletes a workload clusters.
type Delete struct {
	bootstrapper   interfaces.Bootstrapper
	provider       providers.Provider
	writer         filewriter.FileWriter
	clusterManager interfaces.ClusterManager
	gitopsManager  interfaces.GitOpsManager
	clusterDeleter interfaces.ClusterDeleter
	eksdInstaller  interfaces.EksdInstaller
	eksaInstaller  interfaces.EksaInstaller
	clientFactory  interfaces.ClientFactory
}

// NewDelete builds a new delete construct.
func NewDelete(bootstrapper interfaces.Bootstrapper,
	provider providers.Provider,
	writer filewriter.FileWriter,
	clusterManager interfaces.ClusterManager,
	gitopsManager interfaces.GitOpsManager,
	clusterDeleter interfaces.ClusterDeleter,
	eksdInstaller interfaces.EksdInstaller,
	eksaInstaller interfaces.EksaInstaller,
	clientFactory interfaces.ClientFactory,
) *Delete {
	return &Delete{
		bootstrapper:   bootstrapper,
		provider:       provider,
		writer:         writer,
		clusterManager: clusterManager,
		gitopsManager:  gitopsManager,
		clusterDeleter: clusterDeleter,
		eksdInstaller:  eksdInstaller,
		eksaInstaller:  eksaInstaller,
		clientFactory:  clientFactory,
	}
}

// Run executes the tasks to delete a management cluster.
func (c *Delete) Run(ctx context.Context, workload *types.Cluster, clusterSpec *cluster.Spec) error {
	commandContext := &task.CommandContext{
		Bootstrapper:    c.bootstrapper,
		Provider:        c.provider,
		Writer:          c.writer,
		ClusterManager:  c.clusterManager,
		ClusterSpec:     clusterSpec,
		WorkloadCluster: workload,
		GitOpsManager:   c.gitopsManager,
		ClusterDeleter:  c.clusterDeleter,
		EksdInstaller:   c.eksdInstaller,
		EksaInstaller:   c.eksaInstaller,
		ClientFactory:   c.clientFactory,
	}

	return task.NewTaskRunner(&setupAndValidateDelete{}, c.writer).RunTask(ctx, commandContext)
}
