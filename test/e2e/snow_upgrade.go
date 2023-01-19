//go:build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/test/framework"
)

// runSnowUpgradeTest creates a Snow cluster using a base configuration (multiple workers, labels, taints, etc.) plus the provided changes,
// upgrades it to the given changes and deletes it after validating its state. This is meant to be used for Snow full upgrade flow tests.
// It tries to test as many possible changes as possible: control plane scaling, worker node group scaling, wortker node group addition and
// removal, taints, labels, etc. It allows for extra customization through the baseAPIChanges and upgradeAPIChanges but the only required
// changes for those arguments are the osFamily and kubernetes version, since they are not set by default. These can all be provided
// using the provide methods `WithUbuntu124`, `WithBottlerocket124, etc.
func runSnowUpgradeTest(test *framework.ClusterE2ETest, snow *framework.Snow, baseAPIChanges, upgradeAPIChanges api.ClusterConfigFiller) {
	test.WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
			api.WithStackedEtcdTopology(),
		),
		snow.WithWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(
				worker0,
				api.WithCount(1),
				api.WithLabel(key1, val2),
			),
			api.WithDHCP(),
		),
		snow.WithWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(
				worker1,
				api.WithCount(1),
				api.WithTaint(framework.NoScheduleTaint())),
			api.WithDHCP(),
		),
		baseAPIChanges,
	)

	runUpgradeFlow(test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.RemoveWorkerNodeGroup(worker1),
			api.WithWorkerNodeGroup(worker0, api.WithCount(2)),
		),
		snow.WithWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1)),
			api.WithDHCP(),
		),
		upgradeAPIChanges,
	)
}

// runUpgradeFlow creates a cluster, upgrades it with the given changes with the CLI, validates it and finally deletes it if
// all previous steps are successfull. This is represents a basic user workflow and is meant for standalone cluster's CLI tests.
func runUpgradeFlow(test *framework.ClusterE2ETest, upgradeChanges ...api.ClusterConfigFiller) {
	test.CreateCluster()
	validateCluster(test)
	test.StopIfFailed()
	test.UpdateClusterConfig(upgradeChanges...)
	test.UpgradeCluster()
	validateCluster(test)
	test.StopIfFailed()
	test.DeleteCluster()
}

// validateCluster performs a set of validations comparing the cluster config definition with
// the current state of the cluster. This is meant to be used after a create or upgrade operation
// to make sure the cluster has reached the desired state. This should eventually be replaced by
// the cluster validator.
func validateCluster(test *framework.ClusterE2ETest) {
	test.ValidateCluster(test.ClusterConfig.Cluster.Spec.KubernetesVersion)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints, framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints, framework.ValidateControlPlaneLabels)
}
