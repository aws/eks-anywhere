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

func runOIDCFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateOIDC()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes120OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes121OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes120OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu120()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes121OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes120OIDC(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat120()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes121OIDC(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}
