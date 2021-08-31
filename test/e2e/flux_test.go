// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	fluxUserProvidedBranch    = "testbranch"
	fluxUserProvidedNamespace = "testns"
	fluxUserProvidedPath      = "test/testerson"
)

func runFluxFlow(test *framework.E2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes119Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes119ThreeReplicasFlux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes119Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu119()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes119ThreeReplicasFlux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu119()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes118Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes12OFlux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes121Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes118Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu118()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes120Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121Flux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes119GitopsOptionsFlux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(api.WithFluxBranch(fluxUserProvidedBranch)),
		framework.WithFlux(api.WithFluxNamespace(fluxUserProvidedNamespace)),
		framework.WithFlux(api.WithFluxConfigurationPath(fluxUserProvidedPath)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes119GitopsOptionsFlux(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu119()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube119)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFlux(api.WithFluxBranch(fluxUserProvidedBranch)),
		framework.WithFlux(api.WithFluxNamespace(fluxUserProvidedNamespace)),
		framework.WithFlux(api.WithFluxConfigurationPath(fluxUserProvidedPath)),
		framework.WithFlux(),
	)
	runFluxFlow(test)
}
