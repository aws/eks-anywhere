package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	BranchNameEnvVar     = "T_BRANCH_NAME"
	defaultTestBranch    = "main"
)

func GetLatestMinorReleaseBinaryFromTestBranch() (binaryPath string, err error) {
	return getBinaryFromRelease(GetLatestMinorReleaseFromTestBranch())
}

func GetLatestMinorReleaseFromTestBranch() (*releasev1alpha1.EksARelease, error) {
	testBranch := getEnvWithDefault(BranchNameEnvVar, defaultTestBranch)
	if testBranch == "main" {
		return GetLatestMinorReleaseFromMain()
	}

	majorMinor := getMajorMinorFromTestBranch(testBranch)
	testBranchFirstVersion := fmt.Sprintf("%s.0", majorMinor)
	testBranchFirstSemver, err := semver.New(testBranchFirstVersion)
	if err != nil {
		return nil, fmt.Errorf("can't extract semver from release branch [%s]: %v", testBranch, err)
	}

	return GetLatestMinorReleaseFromVersion(testBranchFirstSemver)
}

func GetLatestMinorReleaseBinaryFromMain() (binaryPath string, err error) {
	return getBinaryFromRelease(GetLatestMinorReleaseFromMain())
}

func GetLatestMinorReleaseFromMain() (*releasev1alpha1.EksARelease, error) {
	releases, err := prodReleases()
	if err != nil {
		return nil, err
	}
	var latestRelease *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		if release.Version == releases.Spec.LatestVersion {
			latestRelease = &release
			break
		}
	}

	if latestRelease == nil {
		return nil, fmt.Errorf("releases manifest doesn't contain latest release %s", releases.Spec.LatestVersion)
	}

	return latestRelease, nil
}

func GetLatestMinorReleaseBinaryFromVersion(releaseBranchVersion *semver.Version) (binaryPath string, err error) {
	return getBinaryFromRelease(GetLatestMinorReleaseFromVersion(releaseBranchVersion))
}

func GetLatestMinorReleaseFromVersion(releaseBranchVersion *semver.Version) (*releasev1alpha1.EksARelease, error) {
	releases, err := prodReleases()
	if err != nil {
		return nil, err
	}

	release, err := getLatestPrevMinorRelease(releases, releaseBranchVersion)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func GetReleaseBinaryFromVersion(version *semver.Version) (binaryPath string, err error) {
	releases, err := prodReleases()
	if err != nil {
		return "", err
	}

	var targetVersion *releasev1alpha1.EksARelease
	for _, release := range releases.Spec.Releases {
		releaseVersion := newVersion(release.Version)
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
		logger.Info("Downloading binary for EKS-A release", r.Version, latestReleaseBinaryPath)
		err := os.MkdirAll(latestReleaseBinaryFolder, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed creating directory ./%s: %s", latestReleaseBinaryFolder, err)
		}

		binaryUri, err := r.binaryUri()
		if err != nil {
			return "", fmt.Errorf("determining URI for EKS-A binary: %v", err)
		}
		err = files.GzipFileDownloadExtract(binaryUri, releaseBinaryName, latestReleaseBinaryFolder)
		if err != nil {
			return "", fmt.Errorf("failed extracting binary for EKS-A release [%s] to path ./%s: %s", r.Version, latestReleaseBinaryPath, err)
		}
	}
	return latestReleaseBinaryPath, nil
}

func getBinaryFromRelease(release *releasev1alpha1.EksARelease, chainedErr error) (binaryPath string, err error) {
	if chainedErr != nil {
		return "", err
	}

	binaryPath, err = getBinary(release)
	if err != nil {
		return "", fmt.Errorf("failed getting binary for release [%s]: %v", release.Version, err)
	}

	return binaryPath, nil
}

type platformAwareRelease struct {
	*releasev1alpha1.EksARelease
}

func (p *platformAwareRelease) binaryUri() (binaryUri string, err error) {
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

func getLatestPrevMinorRelease(releases *releasev1alpha1.Release, releaseBranchVersion *semver.Version) (*releasev1alpha1.EksARelease, error) {
	targetRelease := &releasev1alpha1.EksARelease{
		Version:           "",
		BundleManifestUrl: "",
	}
	latestPrevMinorReleaseVersion := newVersion("0.0.0")

	for _, release := range releases.Spec.Releases {
		releaseVersion := newVersion(release.Version)
		if releaseVersion.LessThan(releaseBranchVersion) && releaseVersion.Minor != releaseBranchVersion.Minor && releaseVersion.GreaterThan(latestPrevMinorReleaseVersion) {
			*targetRelease = release
			latestPrevMinorReleaseVersion = releaseVersion
		}
	}

	if targetRelease.Version == "" {
		return nil, fmt.Errorf("releases manifest doesn't contain a version of the previous minor release")
	}

	return targetRelease, nil
}

func getMajorMinorFromTestBranch(testBranch string) string {
	return strings.TrimPrefix(testBranch, "release-")
}
