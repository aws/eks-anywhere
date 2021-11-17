package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/internal/pkg/files"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/validations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	prodReleasesManifest = "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"
	releaseBinaryName    = "eksctl-anywhere"
)

func (e *ClusterE2ETest) GetLatestMinorReleaseBinaryFromMain() string {
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

	binaryPath, err := e.getBinary(latestRelease)
	if err != nil {
		e.T.Fatalf("Failed getting binary for latest release: %s", err)
	}

	return binaryPath
}

func (e *ClusterE2ETest) GetLatestMinorReleaseBinaryFromReleaseBranch(releaseBranchVersion *semver.Version) string {
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

	binaryPath, err := e.getBinary(latestPrevMinorRelease)
	if err != nil {
		e.T.Fatalf("Failed getting binary for latest previous minor release: %s", err)
	}

	return binaryPath
}

func (e *ClusterE2ETest) getBinary(release *releasev1alpha1.EksARelease) (string, error) {
	latestReleaseBinaryFolder := filepath.Join("bin", release.Version)
	latestReleaseBinaryPath := filepath.Join(latestReleaseBinaryFolder, releaseBinaryName)

	if !validations.FileExists(latestReleaseBinaryPath) {
		e.T.Logf("Downloading binary for EKS-A release [%s] to path ./%s", release.Version, latestReleaseBinaryPath)
		err := os.MkdirAll(latestReleaseBinaryFolder, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed creating directory ./%s: %s", latestReleaseBinaryFolder, err)
		}

		err = files.GzipFileDownloadExtract(release.EksABinary.LinuxBinary.URI, releaseBinaryName, latestReleaseBinaryFolder)
		if err != nil {
			return "", fmt.Errorf("failed extracting binary for EKS-A release [%s] to path ./%s: %s", release.Version, latestReleaseBinaryPath, err)
		}
	}
	return latestReleaseBinaryPath, nil
}
