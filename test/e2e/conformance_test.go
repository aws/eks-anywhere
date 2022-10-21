//go:build conformance_e2e
// +build conformance_e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func runConformanceFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.RunConformanceTests()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runTinkerbellConformanceFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.RunConformanceTests()
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestDockerKubernetes120ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes121ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes124ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes120ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes121ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes124ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes120BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes121BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes122BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes123BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestCloudStackKubernetes120ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestCloudStackKubernetes121ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestTinkerbellKubernetes121ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes122ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes121BottleRocketThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes122BottleRocketThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestSnowKubernetes121ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(SnowProviderEnvVar, "true"),
		framework.WithEnvVar(features.FullLifecycleAPIEnvVar, "true"),
	)
	runConformanceFlow(test)
}

func TestSnowKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(SnowProviderEnvVar, "true"),
		framework.WithEnvVar(features.FullLifecycleAPIEnvVar, "true"),
	)
	runConformanceFlow(test)
}

func TestSnowKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(SnowProviderEnvVar, "true"),
		framework.WithEnvVar(features.FullLifecycleAPIEnvVar, "true"),
	)
	runConformanceFlow(test)
}
