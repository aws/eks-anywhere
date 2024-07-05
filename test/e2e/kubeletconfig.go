//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runKubeletConfigurationFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateKubeletConfig()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runKubeletConfigurationTinkerbellFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster()
	test.ValidateKubeletConfig()
	test.StopIfFailed()
	test.DeleteCluster()
}
