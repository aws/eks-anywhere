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
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes125ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
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
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes125ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
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

func TestVSphereKubernetes124BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
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

func TestCloudStackKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestCloudStackKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
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

func TestTinkerbellKubernetes123ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes124ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu124Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
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

func TestTinkerbellKubernetes123BottleRocketThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes124BottleRocketThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestSnowKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestSnowKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes122ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes123ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes124ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes125ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes126ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runConformanceFlow(test)
}
