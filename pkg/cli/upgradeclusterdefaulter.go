package cli

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

// UpgradeClusterDefaulter defines the cluster defaulter for Upgrade cluster command defaults.
type UpgradeClusterDefaulter struct {
	runner *defaulting.Runner[*anywherev1.Cluster]
}

// NewUpgradeClusterDefaulter to instantiate and register defaults.
func NewUpgradeClusterDefaulter(mhcDefaulter cluster.MachineHealthCheckDefaulter) UpgradeClusterDefaulter {
	r := defaulting.NewRunner[*anywherev1.Cluster]()
	r.Register(
		mhcDefaulter.MachineHealthCheckDefault,
	)

	return UpgradeClusterDefaulter{
		runner: r,
	}
}

// Run will run all the defaults registered to the Upgrade Cluster Defaulter.
func (v UpgradeClusterDefaulter) Run(ctx context.Context, cluster *anywherev1.Cluster) (*anywherev1.Cluster, error) {
	return v.runner.RunAll(ctx, cluster)
}
