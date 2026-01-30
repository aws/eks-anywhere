//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

// runStaticIPSimpleFlow runs a simple cluster creation flow with static IP configuration.
// This test validates:
// - Cluster creation with IP pool configuration
// - InClusterIPPool resource is created
// - IPAddressClaim and IPAddress resources are created by CAPV
// - Nodes receive IPs from the configured pool
// - Cluster deletion cleans up IP resources
func runStaticIPSimpleFlow(test *framework.ClusterE2ETest, ipConfig framework.StaticIPConfig) {
	test.GenerateClusterConfig()
	test.CreateCluster()

	// Validate CAPI IPAM resources
	test.ValidateInClusterIPPool(ipConfig.PoolName)
	test.ValidateIPAddressResources(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	test.DeleteCluster()
}

// runStaticIPUpgradeFlow runs a cluster upgrade flow with static IP configuration.
// This test validates:
// - Cluster creation with IP pool
// - Cluster upgrade preserves IP pool configuration
// - New nodes during rolling upgrade get IPs from the same pool
// - IPs are properly recycled during upgrade
func runStaticIPUpgradeFlow(test *framework.ClusterE2ETest, ipConfig framework.StaticIPConfig, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()

	// Validate initial static IP allocation
	test.ValidateInClusterIPPool(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	// Perform upgrade
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateClusterState()

	// Validate IPs after upgrade
	test.ValidateIPAddressResources(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	test.StopIfFailed()
	test.DeleteCluster()
}

// runStaticIPScaleFlow runs a cluster scaling flow with static IP configuration.
// This test validates:
// - Cluster creation with IP pool
// - Scale up allocates new IPs from the pool
// - Scale down releases IPs back to the pool
func runStaticIPScaleFlow(test *framework.ClusterE2ETest, ipConfig framework.StaticIPConfig, scaleUpOpts, scaleDownOpts []framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()

	// Validate initial state
	test.ValidateInClusterIPPool(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	// Scale up
	test.UpgradeClusterWithNewConfig(scaleUpOpts)
	test.ValidateClusterState()
	test.ValidateIPAddressResources(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	// Scale down
	test.UpgradeClusterWithNewConfig(scaleDownOpts)
	test.ValidateClusterState()
	test.ValidateIPReleasedAfterScaleDown(ipConfig.PoolName, 1)

	test.StopIfFailed()
	test.DeleteCluster()
}

// runStaticIPWithoutClusterConfigGeneration runs the static IP flow using
// pre-configured cluster config (set via WithClusterConfig).
func runStaticIPWithoutClusterConfigGeneration(test *framework.ClusterE2ETest, ipConfig framework.StaticIPConfig) {
	test.CreateCluster()

	// Validate CAPI IPAM resources
	test.ValidateInClusterIPPool(ipConfig.PoolName)
	test.ValidateIPAddressResources(ipConfig.PoolName)
	test.ValidateStaticIPAllocation(ipConfig)

	test.DeleteCluster()
}
