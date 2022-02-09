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

const (
	fluxUserProvidedBranch    = "testbranch"
	fluxUserProvidedNamespace = "testns"
	fluxUserProvidedPath      = "test/testerson"
)

func runUpgradeFlowWithFlux(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runFluxFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes120Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes121Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes120Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu120()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes120BottleRocketFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket120()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121BottleRocketFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket121()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121ThreeReplicasThreeWorkersFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes121GitopsOptionsFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121GitopsOptionsFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes120Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat120()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes121Flux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes121ThreeReplicasThreeWorkersFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes121GitopsOptionsFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}
