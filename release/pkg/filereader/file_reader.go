// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filereader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/release/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/git"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/pkg/util/artifacts"
)

const (
	releasePath             = "release"
	imageBuilderProjectPath = "projects/kubernetes-sigs/image-builder"
)

type EksDLatestRelease struct {
	Branch      string `json:"branch"`
	KubeVersion string `json:"kubeVersion"`
	Number      int    `json:"number"`
	Dev         bool   `json:"dev,omitempty"`
}

type EksDLatestReleases struct {
	Releases []EksDLatestRelease `json:"releases"`
	Latest   string              `json:"latest"`
}

func ReadShaSums(filename string, r *releasetypes.ReleaseConfig) (string, string, error) {
	var sha256, sha512 string
	var err error
	if r.DryRun {
		sha256, err = artifactutils.GetFakeSHA(256)
		if err != nil {
			return "", "", errors.Cause(err)
		}

		sha512, err = artifactutils.GetFakeSHA(512)
		if err != nil {
			return "", "", errors.Cause(err)
		}
	} else {
		sha256Path := filename + ".sha256"
		sha256, err = readShaFile(sha256Path)
		if err != nil {
			return "", "", errors.Cause(err)
		}
		sha512Path := filename + ".sha512"
		sha512, err = readShaFile(sha512Path)
		if err != nil {
			return "", "", errors.Cause(err)
		}
	}
	return sha256, sha512, nil
}

func readShaFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Cause(err)
	}
	if parts := strings.Split(string(data), "  "); len(parts) == 2 {
		return parts[0], nil
	}
	return "", errors.Errorf("Error parsing shasum file %s", filename)
}

func ReadFileContentsTrimmed(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Cause(err)
	}
	return strings.TrimSpace(string(data)), nil
}

func ReadEksDReleases(r *releasetypes.ReleaseConfig) (*EksDLatestReleases, error) {
	// Read the eks-d latest release file to get all the releases
	eksDLatestReleases := &EksDLatestReleases{}
	eksDReleaseFilePath := filepath.Join(r.BuildRepoSource, "EKSD_LATEST_RELEASES")

	eksDReleaseFile, err := ioutil.ReadFile(eksDReleaseFilePath)
	if err != nil {
		return nil, errors.Cause(err)
	}
	err = yaml.Unmarshal(eksDReleaseFile, eksDLatestReleases)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return eksDLatestReleases, nil
}

func GetSupportedK8sVersions(r *releasetypes.ReleaseConfig) ([]string, error) {
	// Read the eks-d latest release file to get all the releases
	releaseFilePath := filepath.Join(r.BuildRepoSource, releasePath, "SUPPORTED_RELEASE_BRANCHES")

	releaseFile, err := ioutil.ReadFile(releaseFilePath)
	if err != nil {
		return nil, errors.Cause(err)
	}
	supportedK8sVersions := strings.Split(strings.TrimRight(string(releaseFile), "\n"), "\n")

	return supportedK8sVersions, nil
}

func GetBottlerocketSupportedK8sVersionsByFormat(r *releasetypes.ReleaseConfig, imageFormat string) ([]string, error) {
	if r.DryRun {
		if imageFormat == "ova" {
			return []string{"1-20", "1-21", "1-22", "1-23"}, nil
		} else if imageFormat == "raw" {
			return []string{"1-21", "1-22", "1-23"}, nil
		}
	}
	// Read the eks-d latest release file to get all the releases
	var bottlerocketReleaseMap map[string]interface{}
	var bottlerocketSupportedK8sVersions []string
	bottlerocketReleasesFilename := "BOTTLEROCKET_RELEASES"
	if r.BuildRepoBranchName == "release-0.9" {
		bottlerocketReleasesFilename = "BOTTLEROCKET_OVA_RELEASES"
	}
	bottlerocketReleasesFilePath := filepath.Join(r.BuildRepoSource, imageBuilderProjectPath, bottlerocketReleasesFilename)

	bottlerocketReleasesFileContents, err := ioutil.ReadFile(bottlerocketReleasesFilePath)
	if err != nil {
		return nil, errors.Cause(err)
	}
	err = yaml.Unmarshal(bottlerocketReleasesFileContents, &bottlerocketReleaseMap)
	if err != nil {
		return nil, errors.Cause(err)
	}

	for channel := range bottlerocketReleaseMap {
		// new format for BR releases file
		releaseVersionByFormat := bottlerocketReleaseMap[channel].(map[string]interface{})[fmt.Sprintf("%s-release-version", imageFormat)]
		if releaseVersionByFormat != nil {
			bottlerocketSupportedK8sVersions = append(bottlerocketSupportedK8sVersions, channel)
		} else if releaseVersionByFormat == nil && imageFormat == "ova" {
			// old format for BR releases file for backward compatibility
			releaseVersionByFormat = bottlerocketReleaseMap[channel].(map[string]interface{})["releaseVersion"]
			if releaseVersionByFormat != nil {
				bottlerocketSupportedK8sVersions = append(bottlerocketSupportedK8sVersions, channel)
			}
		}
	}

	return bottlerocketSupportedK8sVersions, nil
}

func GetBottlerocketAdminContainerMetadata(r *releasetypes.ReleaseConfig) (string, string, error) {
	var bottlerocketAdminContainerMetadataMap map[string]interface{}
	bottlerocketAdminContainerMetadataFilePath := filepath.Join(r.BuildRepoSource, imageBuilderProjectPath, "BOTTLEROCKET_ADMIN_CONTAINER_METADATA")
	metadata, err := ioutil.ReadFile(bottlerocketAdminContainerMetadataFilePath)
	if err != nil {
		return "", "", errors.Cause(err)
	}
	err = yaml.Unmarshal(metadata, &bottlerocketAdminContainerMetadataMap)
	if err != nil {
		return "", "", errors.Cause(err)
	}

	tag, imageDigest := bottlerocketAdminContainerMetadataMap["tag"].(string), bottlerocketAdminContainerMetadataMap["imageDigest"].(string)

	return tag, imageDigest, nil
}

func GetEksDReleaseManifestUrl(releaseChannel, releaseNumber string, dev bool) string {
	if dev {
		return fmt.Sprintf("https://eks-d-postsubmit-artifacts.s3.us-west-2.amazonaws.com/kubernetes-%[1]s/kubernetes-%[1]s-eks-%s.yaml", releaseChannel, releaseNumber)
	}
	return fmt.Sprintf("https://distro.eks.amazonaws.com/kubernetes-%[1]s/kubernetes-%[1]s-eks-%s.yaml", releaseChannel, releaseNumber)
}

func GetCurrentEksADevReleaseVersion(releaseVersion string, r *releasetypes.ReleaseConfig) (string, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("              Dev Release Version Computation")
	fmt.Println("==========================================================")

	var latestBuildVersion string
	var newDevReleaseVersion string
	tempFileName := "latest-dev-release-version"

	var latestReleaseKey string

	if r.BuildRepoBranchName == "main" {
		latestReleaseKey = "LATEST_RELEASE_VERSION"
	} else {
		latestReleaseKey = fmt.Sprintf("%s/LATEST_RELEASE_VERSION", r.BuildRepoBranchName)
	}

	if s3.KeyExists(r.ReleaseBucket, latestReleaseKey) {
		err := s3.DownloadFile(tempFileName, r.ReleaseBucket, latestReleaseKey)
		if err != nil {
			return "", errors.Cause(err)
		}

		// Check if current version and latest version are the same semver
		latestBuildS3, err := ioutil.ReadFile(tempFileName)
		if err != nil {
			return "", errors.Cause(err)
		}
		latestBuildVersion = string(latestBuildS3)
		newDevReleaseVersion, err = GenerateNewDevReleaseVersion(latestBuildVersion, releaseVersion, r.BuildRepoBranchName)
		if err != nil {
			return "", errors.Cause(err)
		}
		fmt.Printf("Previous release version: %s\n", latestBuildVersion)
		fmt.Printf("Current release version provided: %s\n", releaseVersion)
	} else {
		newDevReleaseVersion = "v0.0.0-dev+build.0"
		if r.BuildRepoBranchName != "main" {
			newDevReleaseVersion = fmt.Sprintf("v0.0.0-dev-%s+build.0", r.BuildRepoBranchName)
		}
	}
	fmt.Printf("New dev release release version: %s\n", newDevReleaseVersion)

	fmt.Printf("%s Successfully computed current dev release version\n", constants.SuccessIcon)
	return newDevReleaseVersion, nil
}

func GenerateNewDevReleaseVersion(latestBuildVersion, releaseVersion, branchName string) (string, error) {
	if releaseVersion == "vDev" { // TODO: remove when we update the pipeline
		releaseVersion = "v0.0.0"
	}

	releaseVersionIdentifier := "dev+build"
	if branchName != "main" {
		releaseVersionIdentifier = fmt.Sprintf("dev-%s+build", branchName)
	}

	if !strings.Contains(latestBuildVersion, releaseVersion) && !strings.Contains(latestBuildVersion, "vDev") { // TODO: adding vDev case temporally to support old run, remove later
		// different semver, reset build number suffix on release version

		newReleaseVersion := fmt.Sprintf("%s-%s.0", releaseVersion, releaseVersionIdentifier)

		return newReleaseVersion, nil
	}

	// Same semver, only bump build number suffix on release version
	i := strings.LastIndex(latestBuildVersion, ".")
	if i == -1 || i >= len(latestBuildVersion)-1 {
		return "", fmt.Errorf("invalid dev release version found for latest release: %s", latestBuildVersion)
	}

	lastBuildNumber, err := strconv.Atoi(latestBuildVersion[i+1:])
	if err != nil {
		return "", fmt.Errorf("invalid dev release version found for latest release [%s]: %v", latestBuildVersion, err)
	}

	newBuildNumber := lastBuildNumber + 1
	newReleaseVersion := fmt.Sprintf("%s-%s.%d", releaseVersion, releaseVersionIdentifier, newBuildNumber)

	return newReleaseVersion, nil
}

func PutEksAReleaseVersion(version string, r *releasetypes.ReleaseConfig) error {
	var currentReleaseKey string
	if r.BuildRepoBranchName == "main" {
		currentReleaseKey = "LATEST_RELEASE_VERSION"
	} else {
		currentReleaseKey = fmt.Sprintf("%s/LATEST_RELEASE_VERSION", r.BuildRepoBranchName)
	}

	err := os.MkdirAll(filepath.Dir(currentReleaseKey), 0o755)
	if err != nil {
		return errors.Cause(err)
	}

	f, err := os.Create(currentReleaseKey)
	if err != nil {
		return errors.Cause(err)
	}
	defer os.Remove(f.Name())
	versionByteArr := []byte(version)
	if _, err = f.Write(versionByteArr); err != nil {
		return errors.Cause(err)
	}

	// Upload the file to S3
	fmt.Println("Uploading latest release version file")
	err = s3.UploadFile(currentReleaseKey, aws.String(r.ReleaseBucket), aws.String(currentReleaseKey), r.ReleaseClients.S3.Uploader)
	if err != nil {
		return errors.Cause(err)
	}
	return nil
}

func GetEksdRelease(eksdReleaseURL string) (*eksdv1alpha1.Release, error) {
	content, err := ReadHttpFile(eksdReleaseURL)
	if err != nil {
		return nil, err
	}

	eksd := &eksdv1alpha1.Release{}
	if err = k8syaml.UnmarshalStrict(content, eksd); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal eksd manifest")
	}

	return eksd, nil
}

// Retrieve values from https://github.com/aws/eks-anywhere-build-tooling/blob/main/EKSD_LATEST_RELEASES
func GetEksdReleaseValues(release interface{}) (string, bool) {
	releaseNumber := release.(map[interface{}]interface{})["number"]
	releaseNumberInt := releaseNumber.(int)
	releaseNumberStr := strconv.Itoa(releaseNumberInt)

	dev := false
	devValue := release.(map[interface{}]interface{})["dev"]
	if devValue != nil && devValue.(bool) {
		dev = true
	}
	return releaseNumberStr, dev
}

func ReadHttpFile(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading file from url [%s]", uri)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading file from url [%s]", uri)
	}

	return data, nil
}

func ReadGitTag(projectPath, gitRootPath, branch string) (string, error) {
	_, err := git.CheckoutRepo(gitRootPath, branch)
	if err != nil {
		return "", fmt.Errorf("error reading git tag: %v", err)
	}

	tagFile := filepath.Join(gitRootPath, projectPath, "GIT_TAG")
	gitTag, err := ReadFileContentsTrimmed(tagFile)
	if err != nil {
		return "", fmt.Errorf("error reading git tag: %v", err)
	}

	return gitTag, nil
}
