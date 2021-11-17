package framework

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/semver"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const prodReleasesManifest = "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"

func (e *ClusterE2ETest) GetLatestMinorReleaseFromMain() *releasev1alpha1.EksARelease {
	reader := cluster.NewManifestReader()
	e.T.Logf("Reading prod release manifest %s", prodReleasesManifest)
	releases, err := reader.GetReleases(prodReleasesManifest)
	if err != nil {
		e.T.Fatal(err)
	}

	var latestRelease *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		if release.Version == releases.Spec.LatestVersion {
			latestRelease = &release
			break
		}
	}

	if latestRelease == nil {
		e.T.Fatalf("Releases manifest doesn't contain latest release %s", releases.Spec.LatestVersion)
	}

	return latestRelease
}

func (e *ClusterE2ETest) GetLatestMinorReleaseFromReleaseBranch(releaseBranchVersion *semver.Version) *releasev1alpha1.EksARelease {
	reader := cluster.NewManifestReader()
	e.T.Logf("Reading prod release manifest %s", prodReleasesManifest)
	releases, err := reader.GetReleases(prodReleasesManifest)
	if err != nil {
		e.T.Fatal(err)
	}

	var latestPrevMinorRelease *releasev1alpha1.EksARelease
	latestPrevMinorReleaseVersion, err := semver.New("0.0.0")
	if err != nil {
		e.T.Fatal(err)
	}

	for _, release := range releases.Spec.Releases {
		releaseVersion, err := semver.New(release.Version)
		if err != nil {
			e.T.Fatal(err)
		}
		if releaseVersion.LessThan(releaseBranchVersion) && releaseVersion.Minor != releaseBranchVersion.Minor && releaseVersion.GreaterThan(latestPrevMinorReleaseVersion) {
			latestPrevMinorRelease = &release
			latestPrevMinorReleaseVersion = releaseVersion
		}
	}

	if latestPrevMinorRelease == nil {
		e.T.Fatalf("Releases manifest doesn't contain a version of the previous minor release")
	}

	return latestPrevMinorRelease
}
