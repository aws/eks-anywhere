//go:build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/test/framework"
)

type cloudStackAPIUpgradeTest struct {
	name          string
	upgradeFiller api.ClusterConfigFiller
}

func cloudStackAPIUpgradeTests(cloudstack *framework.CloudStack) []cloudStackAPIUpgradeTest {
	return []cloudStackAPIUpgradeTest{
		{
			name: "add and remove labels and taints",
			upgradeFiller: api.ClusterToConfigFiller(
				// Add label
				api.WithWorkerNodeGroup("md-0", api.WithLabel(key1, val2)),
				// Remove label
				api.WithWorkerNodeGroup("md-1"),
				// Add taint
				api.WithWorkerNodeGroup("md-2", api.WithTaint(framework.NoExecuteTaint())),
				// Remove taint
				api.WithWorkerNodeGroup("md-3", api.WithNoTaints()),
			),
		},
		{
			name: "scale up cp and replace existing worker node groups",
			upgradeFiller: api.JoinClusterConfigFillers(
				api.ClusterToConfigFiller(
					// Scale up CP
					api.WithControlPlaneCount(3),
					api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				),
				// Add new WorkerNodeGroups and use md-1's machineConfig
				cloudstack.WithWorkerNodeGroupConfiguration("md-1", framework.WithWorkerNodeGroup("md-4", api.WithCount(2))),
				cloudstack.WithWorkerNodeGroupConfiguration("md-1", framework.WithWorkerNodeGroup("md-5", api.WithCount(1))),
			),
		},
		{
			name: "scale down cp and scaling worker node groups",
			upgradeFiller: api.ClusterToConfigFiller(
				// Scaling down cp
				api.WithControlPlaneCount(1),
				// Scaling down wng
				api.WithWorkerNodeGroup("md-4", api.WithCount(1)),
				// Scaling up wng
				api.WithWorkerNodeGroup("md-5", api.WithCount(2)),
			),
		},
	}

}

func cloudStackAPIWorkloadTestFillers(cloudstack *framework.CloudStack) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithWorkerNodeGroup("md-1", framework.WithWorkerNodeGroup("md-1", api.WithCount(1), api.WithLabel(key1, val2))),
		cloudstack.WithWorkerNodeGroup("md-2", framework.WithWorkerNodeGroup("md-2", api.WithCount(1))),
		cloudstack.WithWorkerNodeGroup("md-3", framework.WithWorkerNodeGroup("md-3", api.WithCount(1), api.WithTaint(framework.NoScheduleTaint()))),
		cloudstack.WithRedhat123(),
	)
}
