package releases

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/semver"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// manifestURL holds the url to the eksa releases manifest
// this is injected at build time, this is just a sane default for development
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
	content, err := reader.ReadFile(url)
	if err != nil {
		return nil, err
	}

	release := &releasev1.Release{}
	if err = yaml.Unmarshal(content, release); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release manifest from [%s]: %v", url, err)
	}

	return release, nil
}

func ReadBundlesForRelease(reader Reader, release *releasev1.EksARelease) (*releasev1.Bundles, error) {
	return bundles.Read(reader, release.BundleManifestUrl)
}

func ReleaseForVersion(releases *releasev1.Release, version string) (*releasev1.EksARelease, error) {
	semVer, err := semver.New(version)
	if err != nil {
		return nil, fmt.Errorf("invalid eksa version: %v", err)
	}

	for _, r := range releases.Spec.Releases {
		if r.Version == version {
			return &r, nil
		}

		releaseVersion, err := semver.New(r.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version for release %d: %v", r.Number, err)
		}

		if semVer.SamePrerelease(releaseVersion) {
			return &r, nil
		}
	}

	return nil, nil
}
