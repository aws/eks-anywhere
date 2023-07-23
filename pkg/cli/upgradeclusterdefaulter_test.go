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

	baseCluster := baseCluster()

	mhcDefaulter := cluster.NewMachineHealthCheckDefaulter(constants.DefaultNodeStartupTimeout.String(), constants.DefaultUnhealthyMachineTimeout.String())

	upgradeClusterDefaulter := cli.NewUpgradeClusterDefaulter(mhcDefaulter)
	updatedCluster, err := upgradeClusterDefaulter.Run(context.Background(), baseCluster)

	g.Expect(err).To(BeNil())
	g.Expect(updatedCluster.Spec.MachineHealthCheck).ToNot(BeNil())
}
