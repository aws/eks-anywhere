//go:build e2e
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

func runUpgradeFlowWithFlux(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeClusterWithNewConfig(clusterOpts)
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

func TestDockerKubernetes124GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes124GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes121FluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121FluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes124FluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes124GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes124GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121BottleRocketFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes121BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes124ThreeReplicasThreeWorkersFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFluxLegacy(),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes124GitopsOptionsFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFluxLegacy(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes124GitopsOptionsFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFluxLegacy(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes123To124FluxUpgradeLegacy(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Ubuntu124Template()),
	)
}

func TestVSphereKubernetes123To124GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Ubuntu124Template()),
	)
}

func TestCloudStackKubernetes123GitopsOptionsFluxLegacy(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
	)

	test.RunClusterFlowWithGitOps(
		framework.WithClusterUpgradeGit(
			api.WithWorkerNodeCount(3),
		),
		// Needed in order to replace the CloudStackDatacenterConfig namespace field with the value specified
		// compared to when it was initially created without it.
		provider.WithProviderUpgradeGit(),
	)
}

func TestDockerInstallGitFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestDockerInstallGithubFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithFluxGithub(api.WithFluxConfigName(framework.DefaultFluxConfigName)),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestVSphereInstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}
