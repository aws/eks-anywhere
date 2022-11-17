package workflowcontext

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
)

// bootstrapCluster is used to store and retrieve a target cluster kubeconfig.
const bootstrapCluster contextKey = "bootstrap-cluster"

// WithBootstrapCluster returns a context based on ctx containing the target cluster kubeconfig.
func WithBootstrapCluster(ctx context.Context, cluster *types.Cluster) context.Context {
	return context.WithValue(ctx, bootstrapCluster, cluster)
}

// BootstrapCluster retrieves the bootstrap cluster configured in ctx or returns a nil pointer.
func BootstrapCluster(ctx context.Context) *types.Cluster {
	return ctx.Value(bootstrapCluster).(*types.Cluster)
}

const managementCluster contextKey = "management-cluster"

// WithManagementCluster returns a context based on ctx containing a management cluster.
func WithManagementCluster(ctx context.Context, cluster *types.Cluster) context.Context {
	return context.WithValue(ctx, managementCluster, cluster)
}

// ManagementCluster retrieves the management cluster configured in ctx or returns a nil pointer.
func ManagementCluster(ctx context.Context) *types.Cluster {
	return ctx.Value(managementCluster).(*types.Cluster)
}

// workloadCluster is used to store and retrieve a target cluster kubeconfig.
const workloadCluster contextKey = "workload-cluster"

// WithWorkloadCluster returns a context based on ctx containing the target cluster kubeconfig.
func WithWorkloadCluster(ctx context.Context, cluster *types.Cluster) context.Context {
	return context.WithValue(ctx, workloadCluster, cluster)
}

// WorkloadCluster retrieves the workload cluster configured in ctx or returns a nil pointer.
func WorkloadCluster(ctx context.Context) *types.Cluster {
	return ctx.Value(workloadCluster).(*types.Cluster)
}

// WithBootstrapAsManagementCluster is shorthand for WithBootstrapCluster followed by
// WithManagementCluster.
func WithBootstrapAsManagementCluster(ctx context.Context, cluster *types.Cluster) context.Context {
	ctx = WithBootstrapCluster(ctx, cluster)
	return WithManagementCluster(ctx, cluster)
}
