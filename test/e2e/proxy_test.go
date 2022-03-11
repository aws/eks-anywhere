//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func runProxyConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.DeleteCluster()
}

func TestVSphereKubernetes122UbuntuProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu122(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithProxy(),
		framework.WithEnvVar(features.K8s122SupportEnvVar, "true"),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes121BottlerocketProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket121(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes121UbuntuProxyConfig(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithRedhat121()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}
