//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runRegistryMirrorConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ImportImages()
	test.CreateCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.DeleteCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
}

func runTinkerbellRegistryMirrorFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ImportImages()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.StopIfFailed()
	test.DeleteCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.ValidateHardwareDecommissioned()
	test.CleanupDownloadedArtifactsAndImages()
}
