package cli

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

// CreateClusterDefaulter defines the cluster defaulter for create cluster command defaults.
type CreateClusterDefaulter struct {
	runner *defaulting.Runner[*cluster.Spec]
}

// Default is the type used for registering various default methods in the defaulting runner.
type Default func(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error)

// NewCreateClusterDefaulter to instantiate and register defaults.
func NewCreateClusterDefaulter(skipIPCheck cluster.ControlPlaneIPCheckAnnotationDefaulter) CreateClusterDefaulter {
	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		skipIPCheck.ControlPlaneIPCheckDefault,
	)

	return CreateClusterDefaulter{
		runner: r,
	}
}

// Run will run all the defaults registered to the Create Cluster Defaulter.
func (v CreateClusterDefaulter) Run(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error) {
	return v.runner.RunAll(ctx, spec)
}
