//go:build conformance_e2e
// +build conformance_e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.RunConformanceTests()
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestDockerKubernetes126ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes127ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes128ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes129ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestDockerKubernetes130ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes126ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes127ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes128ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes129ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes130ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes126BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes127BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes128BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes129BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestVSphereKubernetes130BottleRocketThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestTinkerbellKubernetes126ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes127ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes128ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes129ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes130ThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestTinkerbellKubernetes126BottleRocketThreeReplicasTwoWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(2),
	)
	runTinkerbellConformanceFlow(test)
}

func TestNutanixKubernetes126ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes127ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes128ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes129ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}

func TestNutanixKubernetes130ThreeWorkersConformanceFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runConformanceFlow(test)
}
