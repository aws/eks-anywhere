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

func latestMinorRelease(t testing.TB) *releasev1.EksARelease {
	t.Helper()
	latestRelease, err := framework.GetLatestMinorReleaseFromTestBranch()
	if err != nil {
		t.Fatal(err)
	}

	return latestRelease
}

func runUpgradeFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeWithFluxFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runMulticlusterUpgradeFromReleaseFlowAPI(test *framework.MulticlusterE2ETest, release *releasev1.EksARelease, filler ...api.ClusterConfigFiller) {
	test.CreateManagementCluster(framework.ExecuteWithEksaRelease(release))
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
	})
	test.ManagementCluster.UpdateClusterConfig(filler...)
	test.ManagementCluster.UpgradeCluster()
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
