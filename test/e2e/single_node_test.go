//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTinkerbellSingleNodeFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}

func TestTinkerbellKubernetes123BottleRocketSingleNode(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes123UbuntuSingleNode(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}
