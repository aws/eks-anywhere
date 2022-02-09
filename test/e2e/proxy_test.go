//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/test/framework/cloudstack"
	"github.com/aws/eks-anywhere/test/framework/vsphere"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runProxyConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.DeleteCluster()
}

func TestVSphereKubernetes121UbuntuProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121(),
			vsphere.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes121BottlerocketProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket121(),
			vsphere.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes121UbuntuProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(),
	)
	runProxyConfigFlow(test)
}
