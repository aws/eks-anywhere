package cli

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

// DeleteClusterDefaulter defines the cluster defaulter for delete cluster command defaults.
type DeleteClusterDefaulter struct {
	runner *defaulting.Runner[*cluster.Spec]
}

// NewDeleteClusterDefaulter to instantiate and register defaults.
func NewDeleteClusterDefaulter(nsDefaulter cluster.NamespaceDefaulter) DeleteClusterDefaulter {
	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		nsDefaulter.NamespaceDefault,
	)

	return DeleteClusterDefaulter{
		runner: r,
	}
}

// Run will run all the defaults registered to the Delete Cluster Defaulter.
func (v DeleteClusterDefaulter) Run(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error) {
	return v.runner.RunAll(ctx, spec)
}
