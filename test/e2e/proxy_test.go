// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runProxyConfigFlow(test *framework.E2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.DeleteCluster()
}

func TestVSphereKubernetes121UbuntuProxyConfig(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121(),
		framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(2)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(3)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes121BottlerocketProxyConfig(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket121(),
		framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		// enable external etcd when proxyconfig for etcd is fixed
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}
