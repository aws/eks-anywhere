// +build e2e

package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTaintsUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, preUpgradeNodeTaints, postUpgradeNodeTaints map[corev1.Taint]int, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateNodeTaints(preUpgradeNodeTaints)
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateNodeTaints(postUpgradeNodeTaints)
	test.StopIfFailed()
	test.DeleteCluster()
}

// this test covers the following scenarios:
// create a node with a taint
// remove a taint from a node
// add a taint to a node which already has another taint
// add a taint to a node which had no taints
func TestVSphereKubernetes121TaintsWorkerNodeGroups(t *testing.T) {
	worker0 := "worker-0"
	worker1 := "worker-1"
	worker2 := "worker-2"

	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			noScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			preferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithUbuntu121(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube121),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	preUpgradeTaints := map[corev1.Taint]int{
		framework.NoScheduleTaint(): 2,
		framework.PreferNoScheduleTaint(): 1,
	}

	postUpgradeTaints := map[corev1.Taint]int{
		framework.NoExecuteTaint(): 3,
		framework.NoScheduleTaint(): 2,
	}

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube121,
		preUpgradeTaints,
		postUpgradeTaints,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
		),
	)
}

func noScheduleWorkerNodeGroup(name string, count int) *framework.WorkerNodeGroup {
		return framework.WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(framework.NoScheduleTaint()))
}

func preferNoScheduleWorkerNodeGroup(name string, count int) *framework.WorkerNodeGroup {
	return framework.WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(framework.PreferNoScheduleTaint()))
}
