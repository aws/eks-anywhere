// +build e2e

package e2e

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/test/framework"
)

func init() {
	if err := logger.InitZap(4, logger.WithName("e2e")); err != nil {
		log.Fatal(fmt.Errorf("failed init zap logger for e2e tests: %v", err))
	}
}

func runSimpleFlow(test *framework.E2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes120SimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes120SimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes121ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes120BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes120BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}
