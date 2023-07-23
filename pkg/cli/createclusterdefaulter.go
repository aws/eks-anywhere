package cli

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

// CreateClusterDefaulter defines the cluster defaulter for create cluster command defaults.
type CreateClusterDefaulter struct {
	runner *defaulting.Runner[*anywherev1.Cluster]
}

// NewCreateClusterDefaulter to instantiate and register defaults.
func NewCreateClusterDefaulter(skipIPCheck cluster.ControlPlaneIPCheckAnnotationDefaulter, mhcDefaulter cluster.MachineHealthCheckDefaulter) CreateClusterDefaulter {
	r := defaulting.NewRunner[*anywherev1.Cluster]()
	r.Register(
		skipIPCheck.ControlPlaneIPCheckDefault,
		mhcDefaulter.MachineHealthCheckDefault,
	)

	return CreateClusterDefaulter{
		runner: r,
	}
}

// Run will run all the defaults registered to the Create Cluster Defaulter.
func (v CreateClusterDefaulter) Run(ctx context.Context, cluster *anywherev1.Cluster) (*anywherev1.Cluster, error) {
	return v.runner.RunAll(ctx, cluster)
}
