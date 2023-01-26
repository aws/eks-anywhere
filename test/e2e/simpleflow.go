//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.DeleteCluster()
}

func runTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}
