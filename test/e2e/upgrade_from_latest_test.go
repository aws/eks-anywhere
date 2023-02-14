//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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

func runMulticlusterUpgradeFromReleaseFlowAPI(test *framework.MulticlusterE2ETest, release *releasev1.EksARelease, opts ...framework.ClusterE2ETestOpt) {
	test.CreateManagementCluster(framework.ExecuteWithEksaRelease(release))
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
	})
	test.ManagementCluster.UpgradeCluster(opts)
	test.ManagementCluster.ValidateCluster(test.ManagementCluster.ClusterConfig.Cluster.Spec.KubernetesVersion)
	test.ManagementCluster.StopIfFailed()
	cluster := test.ManagementCluster.GetEKSACluster()
	// Upgrade bundle workload clusters now because they still have the old versions of the bundle.
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.UpdateClusterConfig(
			api.ClusterToConfigFiller(
				api.WithBundlesRef(cluster.Spec.BundlesRef.Name, cluster.Spec.BundlesRef.Namespace, cluster.Spec.BundlesRef.APIVersion),
			),
		)
		wc.ApplyClusterManifest()
		wc.ValidateClusterState()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})
	test.DeleteManagementCluster()
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

func TestCloudStackKubernetes121UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
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
		provider.WithProviderUpgrade(),
	)
}

func TestCloudStackKubernetes123UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube123,
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
			provider.Bottlerocket121Template(), // Set the template so it doesn't get autoimported
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
			provider.Ubuntu121Template(), // Set the template so it doesn't get autoimported
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
			provider.Ubuntu121Template(), // Set the template so it doesn't get autoimported
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
			provider.Ubuntu122Template(), // Set the template so it doesn't get autoimported
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
			provider.Ubuntu123Template(), // Set the template so it doesn't get autoimported
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
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube124,
		provider.WithProviderUpgrade(
			provider.Ubuntu124Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube124)),
	)
}

func TestVSphereKubernetes124To125UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, anywherev1.Kube124),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube125,
		provider.WithProviderUpgrade(
			provider.Ubuntu125Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube125)),
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

func TestDockerKubernetes123to124UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube124)),
	)
}

func TestDockerKubernetes124to125UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		anywherev1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube125)),
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
