package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aws/eks-anywhere/internal/pkg/files"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/validations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	prodReleasesManifest = "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"
	releaseBinaryName    = "eksctl-anywhere"
)

func GetLatestMinorReleaseBinaryFromMain() (binaryPath string, err error) {
	releases, err := prodReleases()
	if err != nil {
		return "", err
	}
	var latestRelease *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		if release.Version == releases.Spec.LatestVersion {
			latestRelease = &release
			break
		}
	}

	if latestRelease == nil {
		return "", fmt.Errorf("releases manifest doesn't contain latest release %s", releases.Spec.LatestVersion)
	}

	binaryPath, err = getBinary(latestRelease)
	if err != nil {
		return "", fmt.Errorf("failed getting binary for latest release: %s", err)
	}

	return binaryPath, nil
}

func GetLatestMinorReleaseBinaryFromVersion(releaseBranchVersion *semver.Version) (binaryPath string, err error) {
	releases, err := prodReleases()
	if err != nil {
		return "", err
	}
	var latestPrevMinorRelease *releasev1alpha1.EksARelease
	latestPrevMinorReleaseVersion, err := semver.New("0.0.0")
	if err != nil {
		return "", err
	}

	for _, release := range releases.Spec.Releases {
		releaseVersion, err := semver.New(release.Version)
		if err != nil {
			return "", err
		}
		if releaseVersion.LessThan(releaseBranchVersion) && releaseVersion.Minor != releaseBranchVersion.Minor && releaseVersion.GreaterThan(latestPrevMinorReleaseVersion) {
			latestPrevMinorRelease = &release
			latestPrevMinorReleaseVersion = releaseVersion
		}
	}

	if latestPrevMinorRelease == nil {
		return "", fmt.Errorf("releases manifest doesn't contain a version of the previous minor release")
	}

	binaryPath, err = getBinary(latestPrevMinorRelease)
	if err != nil {
		return "", fmt.Errorf("failed getting binary for latest previous minor release: %s", err)
	}

	return binaryPath, nil
}

func GetReleaseBinaryFromVersion(version *semver.Version) (binaryPath string, err error) {
	releases, err := prodReleases()
	if err != nil {
		return "", err
	}

	var targetVersion *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		releaseVersion, err := semver.New(release.Version)
		if err != nil {
			return "", err
		}
		if releaseVersion == version {
			targetVersion = &release
		}
	}

	binaryPath, err = getBinary(targetVersion)
	if err != nil {
		return "", fmt.Errorf("failed getting binary for specified version %s: %s", version.String(), err)
	}

	return binaryPath, nil
}

func getBinary(release *releasev1alpha1.EksARelease) (string, error) {
	r := platformAwareRelease{release}

	latestReleaseBinaryFolder := filepath.Join("bin", r.Version)
	latestReleaseBinaryPath := filepath.Join(latestReleaseBinaryFolder, releaseBinaryName)

	if !validations.FileExists(latestReleaseBinaryPath) {
		logger.Info("Downloading binary for EKS-A release [%s] to path ./%s", r.Version, latestReleaseBinaryPath)
		err := os.MkdirAll(latestReleaseBinaryFolder, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed creating directory ./%s: %s", latestReleaseBinaryFolder, err)
		}

		binaryUri, err := r.binaryUri()
		if err != nil {
			return "", fmt.Errorf("error determining URI for EKS-A binary: %v", err)
		}
		err = files.GzipFileDownloadExtract(binaryUri, releaseBinaryName, latestReleaseBinaryFolder)
		if err != nil {
			return "", fmt.Errorf("failed extracting binary for EKS-A release [%s] to path ./%s: %s", r.Version, latestReleaseBinaryPath, err)
		}
	}
	return latestReleaseBinaryPath, nil
}

type platformAwareRelease struct {
	*releasev1alpha1.EksARelease
}

func (p *platformAwareRelease) binaryUri () (binaryUri string, err error) {
	r := runtime.GOOS
	switch r {
	case "darwin":
		return p.EksABinary.DarwinBinary.URI, nil
	case "linux":
		return p.EksABinary.LinuxBinary.URI, nil
	default:
		return "", fmt.Errorf("unsupported runtime %s", r)
	}
}

func prodReleases() (release *releasev1alpha1.Release, err error) {
	reader := cluster.NewManifestReader()
	logger.Info("Reading prod release manifest", "manifest", prodReleasesManifest)
	releases, err := reader.GetReleases(prodReleasesManifest)
	if err != nil {
		return nil, err
	}
	return releases, nil
}
