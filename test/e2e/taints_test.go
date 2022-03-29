// +build e2e

package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTaintsUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	test.StopIfFailed()
	test.DeleteCluster()
}

// this test covers the following scenarios:
// create a node with a taint
// remove a taint from a node
// add a taint to a node which already has another taint
// add a taint to a node which had no taints
func TestVSphereKubernetes122Taints(t *testing.T) {
	provider := ubuntu122ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube122),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes122TaintsBottlerocket(t *testing.T) {
	provider := bottlerocket122ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube122),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func ubuntu122ProviderWithTaints(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithUbuntu122(),
	)
}

func bottlerocket122ProviderWithTaints(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithBottleRocket122(),
	)
}
