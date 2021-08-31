// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runExternalEtcdFlow(test *framework.E2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.DeleteCluster()
}

func TestVSphereKubernetes120ExternalEtcdBottlerocket(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(3)))
	runExternalEtcdFlow(test)
}

func TestVSphereKubernetes120ExternalEtcdUbuntu(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)))
	runExternalEtcdFlow(test)
}

func TestDockerKubernetes120To121ExternalEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
}

func TestVSphereKubernetes120ExternalEtcdBottlerocket3CP2Workers(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
	)
	runExternalEtcdFlow(test)
}
