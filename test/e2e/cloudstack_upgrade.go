//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
	corev1 "k8s.io/api/core/v1"
)

type cloudStackAPIUpgradeTestStep struct {
	name         string
	configFiller api.ClusterConfigFiller
}

type cloudStackAPIUpgradeTest struct {
	name string
	// steps is a list are grouped updates to be applied during a test synchronously.
	steps []cloudStackAPIUpgradeTestStep
}

func clusterPrefix(value string, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, value)
}

func cloudStackAPIUpdateTestBaseStep(e *framework.ClusterE2ETest, cloudstack *framework.CloudStack) cloudStackAPIUpgradeTestStep {
	clusterName := e.ClusterName
	return cloudStackAPIUpgradeTestStep{
		name: "setting base state",
		configFiller: api.JoinClusterConfigFillers(
			api.ClusterToConfigFiller(
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			// Add new WorkerNodeGroups
			cloudstack.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(1))),
			cloudstack.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), api.WithCount(1))),
			cloudstack.WithRedhatVersion(e.ClusterConfig.Cluster.Spec.KubernetesVersion),
		),
	}
}

func cloudstackAPIManagementClusterUpgradeTests(e *framework.ClusterE2ETest, cloudstack *framework.CloudStack) []cloudStackAPIUpgradeTest {
	clusterName := e.ClusterName
	return []cloudStackAPIUpgradeTest{
		{
			name: "add and remove labels and taints",
			steps: []cloudStackAPIUpgradeTestStep{
				cloudStackAPIUpdateTestBaseStep(e, cloudstack),
				{
					name: "adding label and taint to worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithLabel(key1, val2)),
						api.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), api.WithTaint(framework.NoExecuteTaint())),
					),
				},
				{
					name: "removing label and taint from worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), api.WithCount(1))),
						cloudstack.WithRedhatVersion(e.ClusterConfig.Cluster.Spec.KubernetesVersion),
					),
				},
			},
		},
		{
			name: "scale up and down worker node group",
			steps: []cloudStackAPIUpgradeTestStep{
				cloudStackAPIUpdateTestBaseStep(e, cloudstack),
				{
					name: "scaling up worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(2)),
					),
				},
				{
					name: "scaling down worker node group",
					configFiller: api.ClusterToConfigFiller(
						api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(1)),
					),
				},
			},
		},
		{
			name: "replace existing worker node groups",
			steps: []cloudStackAPIUpgradeTestStep{
				cloudStackAPIUpdateTestBaseStep(e, cloudstack),
				{
					name: "replacing existing worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						// Add new WorkerNodeGroups
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-2", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-2", clusterName), api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-3", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-3", clusterName), api.WithCount(1))),
						cloudstack.WithRedhatVersion(e.ClusterConfig.Cluster.Spec.KubernetesVersion),
					),
				},
			},
		},
	}
}

func cloudStackAPIWorkloadUpgradeTests(wc *framework.WorkloadCluster, cloudstack *framework.CloudStack) []cloudStackAPIUpgradeTest {
	clusterName := wc.ClusterName

	return []cloudStackAPIUpgradeTest{
		{
			name: "add and remove labels and taints",
			steps: []cloudStackAPIUpgradeTestStep{
				cloudStackAPIUpdateTestBaseStep(wc.ClusterE2ETest, cloudstack),
				{
					name: "adding label and taint to worker node groups",
					configFiller: api.ClusterToConfigFiller(
						api.WithControlPlaneLabel(cpKey1, cpVal1),
						api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithLabel(key1, val2)),
						api.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), api.WithTaint(framework.NoExecuteTaint())),
					),
				},
				{
					name: "removing label and taint from worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-1", clusterName), api.WithCount(1))),
						cloudstack.WithRedhatVersion(wc.ClusterConfig.Cluster.Spec.KubernetesVersion),
					),
				},
			},
		},
		{
			name: "scale up and down cp and worker node group and availability zones ",
			steps: []cloudStackAPIUpgradeTestStep{
				cloudStackAPIUpdateTestBaseStep(wc.ClusterE2ETest, cloudstack),
				{
					name: "scaling up cp and worker node group",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.WithControlPlaneCount(3),
							api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(2)),
						),
					),
				},
				{
					name: "scaling down cp and worker node group + remove availability zone",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.WithControlPlaneCount(1),
							api.WithWorkerNodeGroup(clusterPrefix("md-0", clusterName), api.WithCount(1)),
						),
					),
				},
			},
		},
		{
			name: "replace existing worker node groups",
			steps: []cloudStackAPIUpgradeTestStep{
				{
					name: "replacing existing worker node groups",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
						),
						// Add new WorkerNodeGroups
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-2", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-2", clusterName), api.WithCount(1))),
						cloudstack.WithWorkerNodeGroup(clusterPrefix("md-3", clusterName), framework.WithWorkerNodeGroup(clusterPrefix("md-3", clusterName), api.WithCount(1))),
						cloudstack.WithRedhatVersion(wc.ClusterConfig.Cluster.Spec.KubernetesVersion),
					),
				},
			},
		},
		{
			name: "availability zones and cilium policy enforcement mode",
			steps: []cloudStackAPIUpgradeTestStep{
				{
					name: "add availability zone + update cilium policy enforcement mode always",
					configFiller: api.JoinClusterConfigFillers(
						api.ClusterToConfigFiller(
							api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways),
						),
						api.CloudStackToConfigFiller(
							framework.UpdateAddCloudStackAz2(),
						),
					),
				},
				{
					name: "remove cloudstack availability zone",
					configFiller: api.JoinClusterConfigFillers(
						api.CloudStackToConfigFiller(
							framework.RemoveAllCloudStackAzs(),
							framework.UpdateAddCloudStackAz1(),
						),
					),
				},
			},
		},
	}

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
		t.Logf("Running API workload upgrade test: %s", step.name)
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
