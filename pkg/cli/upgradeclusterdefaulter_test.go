package cli_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestRunUpgradeClusterDefaulter(t *testing.T) {
	g := NewWithT(t)

	c := baseCluster()

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: c,
		},
	}
	mhcDefaulter := cluster.NewMachineHealthCheckDefaulter(constants.DefaultNodeStartupTimeout, constants.DefaultUnhealthyMachineTimeout)

	upgradeClusterDefaulter := cli.NewUpgradeClusterDefaulter(mhcDefaulter)
	clusterSpec, err := upgradeClusterDefaulter.Run(context.Background(), clusterSpec)

	g.Expect(err).To(BeNil())
	g.Expect(clusterSpec.Cluster.Spec.MachineHealthCheck).ToNot(BeNil())
}
