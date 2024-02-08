package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/internal/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
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

// GetLatestMinorReleaseFromTestBranch inspects the T_BRANCH_NAME environment variable for a
// branch to retrieve the latest released CLI version. If T_BRANCH_NAME is main, it returns
// the latest minor release.
//
// If T_BRANCH_NAME is not main, it expects it to be of the format release-<major>.<minor>
// and will use the <major>.<minor> to retrieve the previous minor release. For example, if the
// release branch is release-0.2 it will retrieve the latest 0.1 release.
func GetLatestMinorReleaseFromTestBranch() (*releasev1alpha1.EksARelease, error) {
	testBranch := testBranch()
	if testBranch == "main" {
		return GetLatestMinorReleaseFromMain()
	}

	testBranchFirstSemver, err := semverForReleaseBranch(testBranch)
	if err != nil {
		return nil, err
	}

	return GetPreviousMinorReleaseFromVersion(testBranchFirstSemver)
}

// GetPreviousMinorReleaseFromTestBranch inspects the T_BRANCH_NAME environment variable for a
// branch to retrieve the previous latest minor released CLI version. If T_BRANCH_NAME is main, it returns
// the previous minor version which is one minor version below the latest minor version that is released.
//
// If T_BRANCH_NAME is not main, it expects it to be of the format release-<major>.<minor>
// and will use the <major>.<minor> to retrieve the previous minor release. For example, if the
// release branch is release-0.2 it will retrieve the latest 0.1 release.
func GetPreviousMinorReleaseFromTestBranch() (*releasev1alpha1.EksARelease, error) {
	testBranch := testBranch()
	var prevLatestMinorRelease *releasev1alpha1.EksARelease
	latestMinorRelease, err := GetLatestMinorReleaseFromTestBranch()
	if err != nil {
		return nil, fmt.Errorf("getting latest minor release: %v", err)
	}
	// For release branch just return the latest minor release
	if testBranch != "main" {
		return latestMinorRelease, nil
	}
	semLatestMinorRelease, err := semver.New(latestMinorRelease.Version)
	if err != nil {
		return nil, err
	}
	prevLatestMinorRelease, err = GetPreviousMinorReleaseFromVersion(semLatestMinorRelease)
	if err != nil {
		return nil, err
	}

	return prevLatestMinorRelease, nil
}

// EKSAVersionForTestBinary returns the "future" EKS-A version for the tested binary based on the TEST_BRANCH name.
// For main, it returns the next minor version.
// For a release branch, it returns the next path version for that release minor version.
func EKSAVersionForTestBinary() (string, error) {
	if testBranch := testBranch(); testBranch != "main" {
		return eksaVersionForReleaseBranch(testBranch)
	}
	return eksaVersionForMain()
}

func eksaVersionForMain() (string, error) {
	latestRelease, err := GetLatestMinorReleaseFromMain()
	if err != nil {
		return "", err
	}

	latestReleaseSemVer, err := semver.New(latestRelease.Version)
	if err != nil {
		return "", errors.Wrapf(err, "parsing version for release %s", latestRelease.Version)
	}

	localVersion := *latestReleaseSemVer
	localVersion.Patch = 0
	localVersion.Minor++

	return localVersion.String(), nil
}

func eksaVersionForReleaseBranch(branch string) (string, error) {
	semVer, err := semverForReleaseBranch(branch)
	if err != nil {
		return "", err
	}

	releases, err := prodReleases()
	if err != nil {
		return "", err
	}

	var latestReleaseSemVer *semver.Version

	latestRelease := GetLatestPatchRelease(releases, semVer)
	if latestRelease != nil {
		latestReleaseSemVer, err = semver.New(latestRelease.Version)
		if err != nil {
			return "", errors.Wrapf(err, "parsing version for release %s", latestRelease.Version)
		}

		localVersion := *latestReleaseSemVer
		localVersion.Patch++
	} else {
		// if no patch version for the release branch, this is an unreleased minor version
		// so the next version will be x.x.0
		latestReleaseSemVer = semVer
	}

	return latestReleaseSemVer.String(), nil
}

func GetLatestMinorReleaseBinaryFromMain() (binaryPath string, err error) {
	return getBinaryFromRelease(GetLatestMinorReleaseFromMain())
}

func GetLatestMinorReleaseFromMain() (*releasev1alpha1.EksARelease, error) {
	releases, err := prodReleases()
	if err != nil {
		return nil, err
	}

	return latestRelease(releases)
}

func semverForReleaseBranch(branch string) (*semver.Version, error) {
	majorMinor := getMajorMinorFromTestBranch(branch)
	testBranchFirstVersion := fmt.Sprintf("%s.0", majorMinor)
	testBranchFirstSemver, err := semver.New(testBranchFirstVersion)
	if err != nil {
		return nil, fmt.Errorf("can't extract semver from release branch [%s]: %v", branch, err)
	}

	return testBranchFirstSemver, nil
}

func latestRelease(releases *releasev1alpha1.Release) (*releasev1alpha1.EksARelease, error) {
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

// localEksaCLIDevVersionRelease returns the EKS-A release for the local eks-a CLI version.
// It reads the version from the local eks-a CLI by running `eksctl anywhere version` command
// and follows the same logic as the CLI to extract the release.
func localEksaCLIDevVersionRelease() (*releasev1alpha1.EksARelease, error) {
	version, err := localEKSAVersionCommand()
	if err != nil {
		return nil, err
	}

	r, err := getReleases(version.ReleaseManifestURL)
	if err != nil {
		return nil, err
	}

	eksaRelease, err := releases.ReleaseForVersion(r, version.GitVersion)
	if err != nil {
		return nil, err
	}

	if eksaRelease == nil {
		return nil, fmt.Errorf("no matching release found for version %s in manifest %s. Latest available version is %s", version.GitVersion, version.ReleaseManifestURL, r.Spec.LatestVersion)
	}

	return eksaRelease, nil
}

// GetPreviousMinorReleaseFromVersion calculates the previous minor release by decrementing the
// version minor number, then retrieves the latest <major>.<minor>.<patch>  for the calculated
// version.
func GetPreviousMinorReleaseFromVersion(version *semver.Version) (*releasev1alpha1.EksARelease, error) {
	releases, err := prodReleases()
	if err != nil {
		return nil, err
	}

	release, err := getLatestPrevMinorRelease(releases, version)
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

// NewEKSAReleasePackagedBinary builds a new EKSAReleasePackagedBinary.
func NewEKSAReleasePackagedBinary(release *releasev1alpha1.EksARelease) *EKSAReleasePackagedBinary {
	return &EKSAReleasePackagedBinary{release}
}

// EKSAReleasePackagedBinary decorates an EKSA release with extra functionality.
type EKSAReleasePackagedBinary struct {
	*releasev1alpha1.EksARelease
}

// BinaryPath implements EKSAPackagedBinary.
func (b *EKSAReleasePackagedBinary) BinaryPath() (string, error) {
	return getBinary(b.EksARelease)
}

// Version returns the eks-a release version.
func (b *EKSAReleasePackagedBinary) Version() string {
	return b.EksARelease.Version
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
	return getReleases(prodReleasesManifest)
}

func getReleases(url string) (release *releasev1alpha1.Release, err error) {
	reader := newFileReader()
	logger.Info("Reading release manifest", "manifest", url)
	releases, err := releases.ReadReleasesFromURL(reader, url)
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

// GetLatestPatchRelease return the latest patch version for the major.minor release version.
// If releases doesn't contain a major.minor for version, it returns nil.
func GetLatestPatchRelease(releases *releasev1alpha1.Release, version *semver.Version) *releasev1alpha1.EksARelease {
	var release *releasev1alpha1.EksARelease

	current := newVersion("0.0.0")
	for _, r := range releases.Spec.Releases {
		r := r
		rv := newVersion(r.Version)
		if rv.SameMajor(version) && rv.SameMinor(version) && rv.GreaterThan(current) {
			release = &r
			current = version
		}
	}

	return release
}

// GetLatestProductionPatchRelease retrieves the latest patch release for version from the
// production release manifest. If the production release manifest does not contain a release for
// the major.minor of version it errors.
func GetLatestProductionPatchRelease(version *semver.Version) (*releasev1alpha1.EksARelease, error) {
	releases, err := prodReleases()
	if err != nil {
		return nil, err
	}

	release := GetLatestPatchRelease(releases, version)
	if release == nil {
		return nil, fmt.Errorf("no release found in the production release bundle for %v", version)
	}

	return release, nil
}

func getMajorMinorFromTestBranch(testBranch string) string {
	return strings.TrimPrefix(testBranch, "release-")
}

func devReleaseURL() string {
	testBranch := testBranch()
	if testBranch == "main" {
		return "https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/eks-a-release.yaml"
	}
	return fmt.Sprintf("https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/%s/eks-a-release.yaml", testBranch)
}

func testBranch() string {
	return getEnvWithDefault(BranchNameEnvVar, defaultTestBranch)
}
