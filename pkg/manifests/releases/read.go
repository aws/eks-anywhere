package releases

import (
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/semver"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// manifestURL holds the url to the eksa releases manifest
// this is injected at build time, this is just a sane default for development.
var manifestURL = "https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/eks-a-release.yaml"

func ManifestURL() string {
	return manifestURL
}

type Reader interface {
	ReadFile(url string) ([]byte, error)
}

func ReadReleases(reader Reader) (*releasev1.Release, error) {
	return ReadReleasesFromURL(reader, ManifestURL())
}

func ReadReleasesFromURL(reader Reader, url string) (*releasev1.Release, error) {
	logger.V(4).Info("Reading release manifest", "url", url)
	content, err := reader.ReadFile(url)
	if err != nil {
		return nil, errors.Wrapf(err, "reading Releases file")
	}

	release := &releasev1.Release{}
	if err = yaml.Unmarshal(content, release); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release manifest from [%s]: %v", url, err)
	}

	return release, nil
}

// GetBundleManifestURL fetches the bundle manifest URL pertaining to this
// release version of EKS Anywhere.
func GetBundleManifestURL(reader Reader, version string) (string, error) {
	eksAReleases, err := ReadReleases(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read releases: %v", err)
	}

	return BundleManifestURL(eksAReleases, version)
}

// BundleManifestURL returns the  Bundles manifest URL for the release matched by the provided
// version. If no release is found for the version, an error is returned.
func BundleManifestURL(releases *releasev1.Release, version string) (string, error) {
	eksAReleaseForVersion, err := ReleaseForVersion(releases, version)
	if err != nil {
		return "", fmt.Errorf("failed to get EKS-A release for version %s: %v", version, err)
	}

	if eksAReleaseForVersion == nil {
		return "", fmt.Errorf("no matching release found for version %s to get Bundles URL. Latest available version is %s", version, releases.Spec.LatestVersion)
	}

	return eksAReleaseForVersion.BundleManifestUrl, nil
}

func ReadBundlesForRelease(reader Reader, release *releasev1.EksARelease) (*releasev1.Bundles, error) {
	return bundles.Read(reader, release.BundleManifestUrl)
}

func ReleaseForVersion(releases *releasev1.Release, version string) (*releasev1.EksARelease, error) {
	semVer, err := semver.New(version)
	if err != nil {
		return nil, fmt.Errorf("invalid eksa version: %v", err)
	}

	// We treat "latest" as a special case to be able to get the latest build without requiring an exact match.
	// We will look for an exact match at the pre-release level and then compare the build metadata.
	// This allows a locally built CLI to get the latest dev build without needing to know the exact build number.
	wantLatestPreRelease := semVer.Buildmetadata == "latest"

	var latestPreRelease *releasev1.EksARelease
	var latestPreReleaseVersion *semver.Version
	for _, r := range releases.Spec.Releases {
		release := r
		if release.Version == version {
			return &release, nil
		}

		// If we are looking for the latest pre-release, we need to compare the build metadata
		// Else we continue to look for an exact match.
		if !wantLatestPreRelease {
			continue
		}

		releaseVersion, err := semver.New(release.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version for release %d: %v", release.Number, err)
		}

		if semVer.SamePrerelease(releaseVersion) &&
			(latestPreReleaseVersion == nil || releaseVersion.CompareBuildMetadata(latestPreReleaseVersion) > 0) {
			// If we are looking for the latest pre-release, we need to compare the build metadata
			// to find the latest one. CompareBuildMetadata will compare the build identifiers
			// in order. For example: v0.19.0-dev+build.10 > v0.19.0-dev+build.9
			latestPreRelease = &release
			latestPreReleaseVersion = releaseVersion
		}
	}

	return latestPreRelease, nil
}
