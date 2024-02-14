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

func prevLatestMinorRelease(t testing.TB) *releasev1.EksARelease {
	t.Helper()
	// Fetch the previous latest minor release for workload creation For ex. curr latest release 15.x prev latest minor release: 14.x
	prevLatestRel, err := framework.GetPreviousMinorReleaseFromTestBranch()
	if err != nil {
		t.Fatal(err)
	}

	return prevLatestRel
}

func runUpgradeFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeClusterWithNewConfig(clusterOpts)
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
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.ValidateFlux()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runMulticlusterUpgradeFromReleaseFlowAPI(test *framework.MulticlusterE2ETest, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion, os framework.OS) {
	provider := test.ManagementCluster.Provider
	test.CreateManagementCluster(framework.ExecuteWithEksaRelease(release))

	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.CreateCluster(framework.ExecuteWithEksaRelease(release))
		wc.ValidateCluster(wc.ClusterConfig.Cluster.Spec.KubernetesVersion)
		wc.StopIfFailed()
	})

	oldCluster := test.ManagementCluster.GetEKSACluster()

	test.ManagementCluster.UpdateClusterConfig(
		provider.WithKubeVersionAndOS(kubeVersion, os, nil),
	)
	test.ManagementCluster.UpgradeCluster()
	test.ManagementCluster.ValidateCluster(test.ManagementCluster.ClusterConfig.Cluster.Spec.KubernetesVersion)
	test.ManagementCluster.StopIfFailed()

	cluster := test.ManagementCluster.GetEKSACluster()

	// Upgrade bundle workload clusters now because they still have the old versions of the bundle.
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.UpdateClusterConfig(
			provider.WithKubeVersionAndOS(kubeVersion, os, nil),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(cluster.Spec.EksaVersion),
			),
		)
		wc.ApplyClusterManifest()
		wc.ValidateClusterState()
		wc.StopIfFailed()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
		wc.StopIfFailed()
	})

	// Create workload cluster with old bundle
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.UpdateClusterConfig(
			provider.WithKubeVersionAndOS(kubeVersion, os, release),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(oldCluster.Spec.EksaVersion),
			),
		)
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.StopIfFailed()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
		wc.StopIfFailed()
	})

	test.DeleteManagementCluster()
}

func runMulticlusterUpgradeFromReleaseFlowAPIWithFlux(test *framework.MulticlusterE2ETest, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion, os framework.OS) {
	provider := test.ManagementCluster.Provider
	test.CreateManagementCluster(framework.ExecuteWithEksaRelease(release))

	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.CreateCluster(framework.ExecuteWithEksaRelease(release))
		wc.ValidateCluster(wc.ClusterConfig.Cluster.Spec.KubernetesVersion)
		wc.StopIfFailed()
	})

	oldCluster := test.ManagementCluster.GetEKSACluster()

	test.ManagementCluster.UpdateClusterConfig(
		provider.WithKubeVersionAndOS(kubeVersion, os, nil),
	)
	test.ManagementCluster.UpgradeCluster()
	test.ManagementCluster.ValidateCluster(test.ManagementCluster.ClusterConfig.Cluster.Spec.KubernetesVersion)
	test.ManagementCluster.StopIfFailed()

	cluster := test.ManagementCluster.GetEKSACluster()

	// Upgrade bundle workload clusters now using the new EksaVersion
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc,
			provider.WithKubeVersionAndOS(kubeVersion, os, nil),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(cluster.Spec.EksaVersion),
			),
		)

		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	// Create workload cluster with the old EksaVersion
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc,
			provider.WithKubeVersionAndOS(kubeVersion, os, release),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(oldCluster.Spec.EksaVersion),
			),
		)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.DeleteManagementCluster()
}

func runUpgradeManagementComponentsFlow(t *testing.T, release *releasev1.EksARelease, provider framework.Provider, kubeVersion anywherev1.KubernetesVersion, os framework.OS) {
	test := framework.NewClusterE2ETest(t, provider)
	// create cluster with old eksa
	test.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	test.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(kubeVersion),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		provider.WithKubeVersionAndOS(kubeVersion, os, release),
	)

	test.CreateCluster(framework.ExecuteWithEksaRelease(release))
	// upgrade management-components with new eksa
	test.RunEKSA([]string{"upgrade", "management-components", "-f", test.ClusterConfigLocation, "-v", "99"})
	test.DeleteCluster()
}
