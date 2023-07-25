package cli_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestRunUpgradeClusterDefaulter(t *testing.T) {
	g := NewWithT(t)

	baseCluster := baseCluster()

	unhealthyTimeout := &metav1.Duration{
		Duration: constants.DefaultUnhealthyMachineTimeout,
	}
	nodeStartupTimeout := &metav1.Duration{
		Duration: constants.DefaultNodeStartupTimeout,
	}
	mhcDefaulter := cluster.NewMachineHealthCheckDefaulter(nodeStartupTimeout, unhealthyTimeout)

	upgradeClusterDefaulter := cli.NewUpgradeClusterDefaulter(mhcDefaulter)
	updatedCluster, err := upgradeClusterDefaulter.Run(context.Background(), baseCluster)

	g.Expect(err).To(BeNil())
	g.Expect(updatedCluster.Spec.MachineHealthCheck).ToNot(BeNil())
}
