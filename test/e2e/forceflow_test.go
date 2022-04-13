//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTinkerbellForceFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithForce())
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestTinkerbellKubernetes121ForceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell()),
		framework.WithEnvVar("TINKERBELL_PROVIDER", "true"),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithHardware(api.HardwareVendorUnspecified, 2),
	)
	runTinkerbellForceFlow(test)
}

func TestTinkerbellKubernetes122ForceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithEnvVar("TINKERBELL_PROVIDER", "true"),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithHardware(api.HardwareVendorUnspecified, 2),
	)
	runTinkerbellForceFlow(test)
}

func TestTinkerbellKubernetes121ThreeReplicasTwoWorkersForceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell()),
		framework.WithEnvVar("TINKERBELL_PROVIDER", "true"),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithHardware(api.HardwareVendorUnspecified, 5),
	)
	runTinkerbellForceFlow(test)
}
