//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func TestVSphereKubernetes124BottlerocketAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithBottleRocket124(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(anywherev1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		anywherev1.Kube124,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestVSphereKubernetes124UbuntuUpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithUbuntu124(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(anywherev1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runUpgradeFlowWithAPI(
		test,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes121RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat121(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(anywherev1.Kube121),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		anywherev1.Kube121,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestSnowKubernetes123UbuntuRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithSnowUbuntu123(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(anywherev1.Kube123),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithEnvVar(features.SnowProviderEnvVar, "true"),
		framework.WithEnvVar(features.FullLifecycleAPIEnvVar, "true"),
	)

	runSimpleUpgradeFlow(
		test,
		anywherev1.Kube123,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup(worker1),
			api.WithWorkerNodeGroup(worker0, api.WithCount(1)),
		),
		provider.WithNewSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(
				worker2,
				api.WithCount(1),
			),
		),
		framework.WithEnvVar(features.SnowProviderEnvVar, "true"),
		framework.WithEnvVar(features.FullLifecycleAPIEnvVar, "true"),
	)
}
