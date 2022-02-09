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

func runRegistryMirrorConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.ImportImages()
	test.CreateCluster()
	test.ImportImages()
	test.DeleteCluster()
}

func TestVSphereKubernetes121UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121(), vsphere.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithRegistryMirrorEndpointAndCert(),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes121BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket121(), vsphere.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithRegistryMirrorEndpointAndCert(),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes121RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithRegistryMirrorEndpointAndCert(),
	)
	runRegistryMirrorConfigFlow(test)
}
