//go:build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/test/framework"
)

type cloudStackAPIUpgradeTest struct {
	name string
	// steps is a list are grouped updates to be applied during a test synchronously.
	steps []cloudStackAPIUpgradeTestStep
}

type cloudStackAPIUpgradeTestStep struct {
	name         string
	configFiller api.ClusterConfigFiller
}

func defaultCloudStackAPIUpgradeTestFillers(cloudstack *framework.CloudStack) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
		// Add new WorkerNodeGroups
		cloudstack.WithWorkerNodeGroup("md-0", framework.WithWorkerNodeGroup("md-0", api.WithCount(1))),
		cloudstack.WithWorkerNodeGroup("md-1", framework.WithWorkerNodeGroup("md-1", api.WithCount(1))),
		cloudstack.WithRedhat123(),
	)
}

func defaultCloudStackAPIUpgradeTestStep(cloudstack *framework.CloudStack) cloudStackAPIUpgradeTestStep {
	return cloudStackAPIUpgradeTestStep{
		name:         "resetting to default state if needed",
		configFiller: defaultCloudStackAPIUpgradeTestFillers(cloudstack),
	}
}

func cloudstackAPIManagementClusterUpgradeTests(cloudstack *framework.CloudStack) []cloudStackAPIUpgradeTest {
	return []cloudStackAPIUpgradeTest{
		{
			name: "add and remove labels and taints",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "adding label and taint to worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0", api.WithLabel(key1, val2)),
						api.WithWorkerNodeGroup("md-1", api.WithTaint(framework.NoExecuteTaint())),
					),
				},
				{
					name: "removing label and taint from worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0"),
						api.WithWorkerNodeGroup("md-1", api.WithNoTaints()),
					),
				},
			},
		},
		{
			name: "scale up and down worker node group",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "scaling up worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0", api.WithCount(2)),
					),
				},
				{
					name: "scaling down worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0", api.WithCount(1)),
					),
				},
			},
		},
		{
			name: "replace existing worker node groups",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "replacing existing worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						// Add new WorkerNodeGroups
						cloudstack.WithWorkerNodeGroup("md-2", framework.WithWorkerNodeGroup("md-2", api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup("md-3", framework.WithWorkerNodeGroup("md-3", api.WithCount(1))),
						cloudstack.WithRedhat123(),
					),
				},
			},
		},
	}
}

func cloudStackAPIWorkloadUpgradeTests(cloudstack *framework.CloudStack) []cloudStackAPIUpgradeTest {
	return []cloudStackAPIUpgradeTest{
		{
			name: "add and remove labels and taints",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "adding label and taint to worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0", api.WithLabel(key1, val2)),
						api.WithWorkerNodeGroup("md-1", api.WithTaint(framework.NoExecuteTaint())),
					),
				},
				{
					name: "removing label and taint from worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup("md-0"),
						api.WithWorkerNodeGroup("md-1", api.WithNoTaints()),
					),
				},
			},
		},
		{
			name: "scale up and down cp and worker node group",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "scaling up cp and worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithControlPlaneCount(3),
						api.WithWorkerNodeGroup("md-0", api.WithCount(2)),
					),
				},
				{
					name: "scaling down cp and worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithControlPlaneCount(1),
						api.WithWorkerNodeGroup("md-0", api.WithCount(1)),
					),
				},
			},
		},
		{
			name: "replace existing worker node groups",
			steps: []cloudStackAPIUpgradeTestStep{
				defaultCloudStackAPIUpgradeTestStep(cloudstack),
				{
					name: "replacing existing worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						// Add new WorkerNodeGroups
						cloudstack.WithWorkerNodeGroup("md-2", framework.WithWorkerNodeGroup("md-2", api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup("md-3", framework.WithWorkerNodeGroup("md-3", api.WithCount(1))),
						cloudstack.WithRedhat123(),
					),
				},
			},
		},
	}

}

func cloudStackAPIClusterBaseChanges(cloudstack *framework.CloudStack) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		defaultCloudStackAPIUpgradeTestFillers(cloudstack),
	)
}

func runCloudStackAPIUpgradeTest(t *testing.T, test *framework.ClusterE2ETest, ut cloudStackAPIUpgradeTest) {
	for _, step := range ut.steps {
		t.Logf("Running API upgrade test: %s", step.name)
		test.UpgradeClusterWithKubectl(step.configFiller)
		test.ValidateClusterStateWithT(t)
	}
}

func runCloudStackAPIWorkloadUpgradeTest(t *testing.T, wc *framework.WorkloadCluster, ut cloudStackAPIUpgradeTest) {
	for _, step := range ut.steps {
		t.Logf("Running API upgrade test: %s", step.name)
		wc.UpdateClusterConfig(step.configFiller)
		wc.ApplyClusterManifest()
		wc.ValidateClusterStateWithT(t)
	}
}

func runCloudStackAPIWorkloadUpgradeTestWithFlux(t *testing.T, test *framework.MulticlusterE2ETest, wc *framework.WorkloadCluster, ut cloudStackAPIUpgradeTest) {
	for _, step := range ut.steps {
		t.Logf("Running API upgrade test: %s", step.name)
		test.PushWorkloadClusterToGit(wc, step.configFiller)
		wc.ValidateClusterStateWithT(t)
	}
}
