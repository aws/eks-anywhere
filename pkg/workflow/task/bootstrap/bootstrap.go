package bootstrap

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/contextutil"
)

// BootstrapOptionsRetriever supplies bootstrap cluster options. This is typically satisfied
// by a provider.
type BootstrapOptionsRetriever interface {
	BootstrapClusterOpts(*cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error)
}

// Bootstrapper creates and destroys bootstrap clusters. It is satisfied by the bootstrap package
// and exists predominently for testability.
type Bootstrapper interface {
	// CreateBootstrapCluster creates a new local cluster. It does not contain any EKS-A components.
	CreateBootstrapCluster(
		context.Context,
		*cluster.Spec,
		...bootstrapper.BootstrapClusterOption,
	) (*types.Cluster, error)

	// DeleteBootstrapCluster deletes a local cluster created with CreateBootstrapCluster.
	DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster, isUpgrade bool) error
}

// CreateBootstrapClusterName is the unique name for the CreateBootstrapCluster task.
const CreateBootstrapClusterName = "CreateBootstrapCluster"

// CreateBootstrapClusters creates a functional Kubernetes cluster that can be used to faciliate
// EKS-A operations. The bootstrap cluster is populated in the context using
// workflow.WithBootstrapCluster for subsequent tasks.
type CreateBootstrapCluster struct {
	// Spec is the spec to be used for bootstrapping the cluster.
	Spec *cluster.Spec

	// Options supplies bootstrap cluster creation options.
	Options BootstrapOptionsRetriever

	// Bootstrapper is used to create the cluster.
	Bootstrapper Bootstrapper
}

// RunTask satisfies workflow.Task.
func (t CreateBootstrapCluster) RunTask(ctx context.Context) (context.Context, error) {
	opts, err := t.Options.BootstrapClusterOpts(t.Spec)
	if err != nil {
		return ctx, err
	}

	cluster, err := t.Bootstrapper.CreateBootstrapCluster(ctx, t.Spec, opts...)
	if err != nil {
		return ctx, err
	}

	return contextutil.WithBootstrapCluster(ctx, *cluster), nil
}

// GetName satisfies workflow.Task.
func (CreateBootstrapCluster) GetName() workflow.TaskName {
	return CreateBootstrapClusterName
}

// DeleteBootstrapClusterName is the unique name for the DeleteBootstrapCluster task.
const DeleteBootstrapClusterName = "DeleteBootstrapCluster"

// DeleteBootstrapCluster deletes a bootstrap cluster. It expects the bootstrap cluster to be
// populated in the context using workflow.WithBootstrapCluster.
type DeleteBootstrapCluster struct {
	// Bootstrapper is used to delete the cluster.
	Bootstrapper Bootstrapper
}

// RunTask satisfies workflow.Task.
func (t DeleteBootstrapCluster) RunTask(ctx context.Context) (context.Context, error) {
	cluster := contextutil.BootstrapCluster(ctx)

	if err := t.Bootstrapper.DeleteBootstrapCluster(ctx, &cluster, false); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// GetName satisfies workflow.Task.
func (t DeleteBootstrapCluster) GetName() workflow.TaskName {
	return DeleteBootstrapClusterName
}
