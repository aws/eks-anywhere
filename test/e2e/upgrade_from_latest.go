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
	test.GenerateClusterConfigForVersion(latestRelease.Version, "", framework.ExecuteWithEksaRelease(latestRelease))
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
	test.GenerateClusterConfigForVersion(latestRelease.Version, "", framework.ExecuteWithEksaRelease(latestRelease))
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

func runInPlaceUpgradeFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateClusterState()
	test.StopIfFailed()
	test.DeleteCluster()
}

// runMulticlusterUpgradeFromReleaseFlowAPI tests the ability to create workload clusters with an old Bundle in a management cluster
// that has been updated to a new Bundle. It follows the following steps:
//  1. Create a management cluster with the old Bundle.
//  2. Create workload clusters with the old Bundle.
//  3. Upgrade the management cluster to the new Bundle and new Kubernetes version (newVersion).
//  4. Upgrade the workload clusters to the new Bundle and new Kubernetes version (newVersion).
//  5. Delete the workload clusters.
//  6. Re-create the workload clusters with the old Bundle and previous Kubernetes version (oldVersion). It's necessary to sometimes
//     use a different kube version because the old Bundle might not support the new kubernetes version.
//  7. Delete the workload clusters.
//  8. Delete the management cluster.
func runMulticlusterUpgradeFromReleaseFlowAPI(test *framework.MulticlusterE2ETest, release *releasev1.EksARelease, oldVersion, newVersion anywherev1.KubernetesVersion, os framework.OS) {
	provider := test.ManagementCluster.Provider
	// 1. Create management cluster
	test.CreateManagementCluster(framework.ExecuteWithEksaRelease(release))

	// 2. Create workload clusters with the old Bundle
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.CreateCluster(framework.ExecuteWithEksaRelease(release))
		wc.ValidateCluster(wc.ClusterConfig.Cluster.Spec.KubernetesVersion)
		wc.StopIfFailed()
	})

	oldCluster := test.ManagementCluster.GetEKSACluster()

	test.ManagementCluster.UpdateClusterConfig(
		provider.WithKubeVersionAndOS(newVersion, os, nil),
	)
	// 3. Upgrade management cluster to new Bundle and new Kubernetes version
	test.ManagementCluster.UpgradeCluster()
	test.ManagementCluster.ValidateCluster(test.ManagementCluster.ClusterConfig.Cluster.Spec.KubernetesVersion)
	test.ManagementCluster.StopIfFailed()

	cluster := test.ManagementCluster.GetEKSACluster()

	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		// 4. Upgrade the workload clusters to the new Bundle and new Kubernetes version (newVersion).
		wc.UpdateClusterConfig(
			provider.WithKubeVersionAndOS(newVersion, os, nil),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(cluster.Spec.EksaVersion),
			),
		)
		wc.ApplyClusterManifest()
		wc.ValidateClusterState()
		wc.StopIfFailed()
		// 5. Delete the workload clusters.
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
		wc.StopIfFailed()
	})

	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.UpdateClusterConfig(
			provider.WithKubeVersionAndOS(oldVersion, os, release),
			api.ClusterToConfigFiller(
				api.WithEksaVersion(oldCluster.Spec.EksaVersion),
			),
		)
		// 6. Re-create the workload clusters with the old Bundle and previous Kubernetes version (oldVersion).
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.StopIfFailed()
		// 7. Delete the workload clusters.
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	// It's necessary to call stop here because if any of the workload clusters failed,
	// their panic was thrown in a go routine, which doesn't stop the main test routine.
	test.RunInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.StopIfFailed()
	})

	// 8. Delete the management cluster.
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
	test.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
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
