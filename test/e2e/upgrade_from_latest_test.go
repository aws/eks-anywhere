//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runUpgradeFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeWithFluxFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func latestMinorRelease(t testing.TB) *releasev1.EksARelease {
	t.Helper()
	latestRelease, err := framework.GetLatestMinorReleaseFromTestBranch()
	if err != nil {
		t.Fatal(err)
	}

	return latestRelease
}

func TestVSphereKubernetes120BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		),
		framework.WithBottlerocketFromRelease(release, anywherev1.Kube120),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube120,
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate120(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestCloudStackKubernetes120UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube120,
		provider.WithProviderUpgrade(),
	)
}

func TestVSphereKubernetes121BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		),
		framework.WithBottlerocketFromRelease(release, anywherev1.Kube121),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube121,
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate121(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes120UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube120),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube120,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate120Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes121UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube121),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube121,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate121Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes121UbuntuUpgradeFromLatestMinorReleaseAlwaysNetworkPolicy(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube121),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube121,
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(anywherev1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate121Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestDockerKubernetes121UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube121,
	)
}

func TestVSphereKubernetes121To122UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube121),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube122,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate122Var(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube122)),
	)
}

func TestVSphereKubernetes122To123UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube122),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube123,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate123Var(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube123)),
	)
}

func TestVSphereKubernetes123To124UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube123),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube124,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate124Var(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube124)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestDockerKubernetes121to122UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube122)),
	)
}

func TestDockerKubernetes122to123UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube123)),
	)
}

func TestDockerKubernetes122to123GithubFluxEnabledUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeWithFluxFromReleaseFlow(
		test,
		release,
		anywherev1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube123)),
	)
}
