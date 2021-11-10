// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	prodReleasesManifest = "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"
	latestReleasePath    = "bin/latest-release"
	releaseBinaryName    = "eksctl-anywhere"
)

func getLatestMinorReleaseFromMain(test *framework.ClusterE2ETest) (binaryPath string) {
	reader := cluster.NewManifestReader()
	test.T.Logf("Reading prod release manifest %s", prodReleasesManifest)
	releases, err := reader.GetReleases(prodReleasesManifest)
	if err != nil {
		test.T.Fatal(err)
	}

	var latestRelease *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		if release.Version == releases.Spec.LatestVersion {
			latestRelease = &release
			break
		}
	}

	if latestRelease == nil {
		test.T.Fatalf("Releases manifest doesn't contain latest release %s", releases.Spec.LatestVersion)
	}

	latestReleaseBinaryFolder := filepath.Join(latestReleasePath, latestRelease.Version)
	latestReleaseBinaryPath := filepath.Join(latestReleaseBinaryFolder, releaseBinaryName)

	return latestReleaseBinaryPath
}

func getLatestMinorReleaseFromReleaseBranch(test *framework.ClusterE2ETest, releaseBranchVersion *semver.Version) (binaryPath string) {
	reader := cluster.NewManifestReader()
	test.T.Logf("Reading prod release manifest %s", prodReleasesManifest)
	releases, err := reader.GetReleases(prodReleasesManifest)
	if err != nil {
		test.T.Fatal(err)
	}

	var latestPrevMinorRelease *releasev1alpha1.EksARelease

	for _, release := range releases.Spec.Releases {
		releaseVersion := semver.New(release.Version)
		if releaseVersion.IsPrevMinorVersion(releaseBranchVersion) && releaseVersion.GreaterThan(latestPrevMinorRelease.Version) {
			latestPrevMinorRelease = &releasee
		}
	}

	if latestPrevMinorRelease == nil {
		test.T.Fatalf("Releases manifest doesn't contain a version of the previous minor release")
	}

	latestReleaseBinaryFolder := filepath.Join(latestReleasePath, latestPrevMinorRelease.Version)
	latestPrevMinorReleaseBinaryPath := filepath.Join(latestReleaseBinaryFolder, releaseBinaryName)

	return latestPrevMinorReleaseBinaryPath
}
