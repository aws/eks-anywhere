//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runAPIServerExtraArgsUpgradeFlow(test *framework.ClusterE2ETest, clusterOpts ...[]framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.GenerateSupportBundleOnCleanupIfTestFailed()
	test.CreateCluster()
	for _, opts := range clusterOpts {
		test.UpgradeClusterWithNewConfig(opts)
		test.ValidateClusterState()
		test.StopIfFailed()
	}
	test.DeleteCluster()
}
