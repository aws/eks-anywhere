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

// Delete is the workflow that deletes a workload clusters.
type Delete struct {
	provider       providers.Provider
	writer         filewriter.FileWriter
	clusterManager interfaces.ClusterManager
	clusterDeleter interfaces.ClusterDeleter
	gitopsManager  interfaces.GitOpsManager
}

// NewDelete builds a new delete construct.
func NewDelete(provider providers.Provider,
	writer filewriter.FileWriter,
	clusterManager interfaces.ClusterManager,
	clusterDeleter interfaces.ClusterDeleter,
	gitopsManager interfaces.GitOpsManager,
) *Delete {
	return &Delete{
		provider:       provider,
		writer:         writer,
		clusterManager: clusterManager,
		clusterDeleter: clusterDeleter,
		gitopsManager:  gitopsManager,
	}
}

// Run executes the tasks to delete a workload cluster.
func (c *Delete) Run(ctx context.Context, workload *types.Cluster, clusterSpec *cluster.Spec) error {
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		Writer:            c.writer,
		ClusterManager:    c.clusterManager,
		ClusterSpec:       clusterSpec,
		ManagementCluster: clusterSpec.ManagementCluster,
		WorkloadCluster:   workload,
		ClusterDeleter:    c.clusterDeleter,
		GitOpsManager:     c.gitopsManager,
	}

	return task.NewTaskRunner(&setupAndValidateDelete{}, c.writer).RunTask(ctx, commandContext)
}
