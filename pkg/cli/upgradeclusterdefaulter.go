package cli

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

// UpgradeClusterDefaulter defines the cluster defaulter for Upgrade cluster command defaults.
type UpgradeClusterDefaulter struct {
	runner *defaulting.Runner[*cluster.Spec]
}

// NewUpgradeClusterDefaulter to instantiate and register defaults.
func NewUpgradeClusterDefaulter(mhcDefaulter cluster.MachineHealthCheckDefaulter) UpgradeClusterDefaulter {
	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		mhcDefaulter.MachineHealthCheckDefault,
	)

	return UpgradeClusterDefaulter{
		runner: r,
	}
}

// Run will run all the defaults registered to the Upgrade Cluster Defaulter.
func (v UpgradeClusterDefaulter) Run(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error) {
	return v.runner.RunAll(ctx, spec)
}
