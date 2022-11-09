package bootstrap

import (
	"context"
	"errors"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// OptionsRetriever supplies bootstrap cluster options. This is typically satisfied
// by a provider.
type OptionsRetriever interface {
	BootstrapClusterOpts(*cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error)
}

// Bootstrapper creates and destroys bootstrap clusters. It is satisfied by the bootstrap package
// and exists predominently for testability.
type Bootstrapper interface {
	// CreateCluster creates a new local cluster. It does not contain any EKS-A components.
	CreateBootstrapCluster(
		context.Context,
		*cluster.Spec,
		...bootstrapper.BootstrapClusterOption,
	) (*types.Cluster, error)

	// DeleteBootstrapCluster deletes a local cluster created with CreateCluster.
	DeleteBootstrapCluster(
		ctx context.Context,
		cluster *types.Cluster,
		operationType constants.Operation,
		isForceCleanup bool,
	) error
}

// CreateCluster creates a functional Kubernetes cluster that can be used to faciliate
// EKS-A operations. The bootstrap cluster is populated in the context using
// workflow.WithBootstrapCluster for subsequent tasks.
type CreateCluster struct {
	// Spec is the spec to be used for bootstrapping the cluster.
	Spec *cluster.Spec

	// Options supplies bootstrap cluster creation options.
	Options OptionsRetriever

	// Bootstrapper is used to create the cluster.
	Bootstrapper Bootstrapper
}

// RunTask satisfies workflow.Task.
func (t CreateCluster) RunTask(ctx context.Context) (context.Context, error) {
	opts, err := t.Options.BootstrapClusterOpts(t.Spec)
	if err != nil {
		return ctx, err
	}

	cluster, err := t.Bootstrapper.CreateBootstrapCluster(ctx, t.Spec, opts...)
	if err != nil {
		return ctx, err
	}

	return workflowcontext.WithBootstrapAsManagementCluster(ctx, cluster), nil
}

// DeleteCluster deletes a bootstrap cluster. It expects the bootstrap cluster to be
// populated in the context using workflow.WithBootstrapCluster.
type DeleteCluster struct {
	// Bootstrapper is used to delete the cluster.
	Bootstrapper Bootstrapper
}

// RunTask satisfies workflow.Task.
func (t DeleteCluster) RunTask(ctx context.Context) (context.Context, error) {
	cluster := workflowcontext.BootstrapCluster(ctx)
	if cluster == nil {
		return ctx, errors.New("bootstrap cluster not found in context")
	}

	if err := t.Bootstrapper.DeleteBootstrapCluster(ctx, cluster, constants.Create, false); err != nil {
		return ctx, err
	}

	return ctx, nil
}
