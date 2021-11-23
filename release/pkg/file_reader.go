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

package pkg

import (
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

var imageBuilderProjectSource = "projects/kubernetes-sigs/image-builder"

func (r *ReleaseConfig) readShaSums(filename string) (string, string, error) {
	var sha256, sha512 string
	var err error
	if r.DryRun {
		sha256, err = GenerateRandomSha(256)
		if err != nil {
			return "", "", errors.Cause(err)
		}

		sha512, err = GenerateRandomSha(512)
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

func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Cause(err)
	}
	return strings.TrimSpace(string(data)), nil
}

func readEksDReleases(r *ReleaseConfig) (map[string]interface{}, error) {
	// Read the eks-d latest release file to get all the releases
	var eksDReleaseMap map[string]interface{}
	eksDReleaseFilePath := filepath.Join(r.BuildRepoSource, "EKSD_LATEST_RELEASES")

	eksDReleaseFile, err := ioutil.ReadFile(eksDReleaseFilePath)
	if err != nil {
		return nil, errors.Cause(err)
	}
	err = yaml.Unmarshal(eksDReleaseFile, &eksDReleaseMap)
	if err != nil {
		return nil, errors.Cause(err)
	}
	return eksDReleaseMap, nil
}

func getBottlerocketSupportedK8sVersions(r *ReleaseConfig) ([]string, error) {
	// Read the eks-d latest release file to get all the releases
	var bottlerocketOvaReleaseMap map[string]interface{}
	var bottlerocketSupportedK8sVersions []string
	bottlerocketOvaReleaseFilePath := filepath.Join(r.BuildRepoSource, imageBuilderProjectSource, "BOTTLEROCKET_OVA_RELEASES")

	bottlerocketOvaReleaseFile, err := ioutil.ReadFile(bottlerocketOvaReleaseFilePath)
	if err != nil {
		return nil, errors.Cause(err)
	}
	err = yaml.Unmarshal(bottlerocketOvaReleaseFile, &bottlerocketOvaReleaseMap)
	if err != nil {
		return nil, errors.Cause(err)
	}

	for channel := range bottlerocketOvaReleaseMap {
		bottlerocketSupportedK8sVersions = append(bottlerocketSupportedK8sVersions, channel)
	}

	return bottlerocketSupportedK8sVersions, nil
}

func (r *ReleaseConfig) getBottlerocketAdminContainerMetadata() (string, string, error) {
	var bottlerocketAdminContainerMetadataMap map[string]interface{}
	bottlerocketAdminContainerMetadataFilePath := filepath.Join(r.BuildRepoSource, imageBuilderProjectSource, "BOTTLEROCKET_ADMIN_CONTAINER_METADATA")
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

func GetEksDReleaseManifestUrl(releaseChannel, releaseNumber string) string {
	manifestUrl := fmt.Sprintf("https://distro.eks.amazonaws.com/kubernetes-%s/kubernetes-%s-eks-%s.yaml", releaseChannel, releaseChannel, releaseNumber)
	return manifestUrl
}

func (r *ReleaseConfig) GetCurrentEksADevReleaseVersion(sourceClients *SourceClients, releaseVersion string) (string, error) {
	if r.DryRun {
		return "v0.0.0-dev+build.0", nil
	}
	s3Downloader := sourceClients.S3.Downloader
	tempFile, err := ioutil.TempFile("", "latest-dev-release-version")
	if err != nil {
		return "", errors.Cause(err)
	}
	defer os.Remove(tempFile.Name())

	var latestReleaseKey string
	if r.BuildRepoBranchName == "main" {
		latestReleaseKey = "LATEST_RELEASE_VERSION"
	} else {
		latestReleaseKey = fmt.Sprintf("%s/LATEST_RELEASE_VERSION", r.BuildRepoBranchName)
	}

	exists, err := ExistsInS3(sourceClients.S3.Client, r.ReleaseBucket, latestReleaseKey)
	if err != nil {
		return "", errors.Cause(err)
	}

	if exists {
		_, err = s3Downloader.Download(tempFile, &s3.GetObjectInput{Bucket: &r.ReleaseBucket, Key: &latestReleaseKey})
		if err != nil {
			return "", errors.Cause(err)
		}

		// Check if current version and latest version are the same semver
		latestBuildS3, err := ioutil.ReadFile(tempFile.Name())
		if err != nil {
			return "", errors.Cause(err)
		}
		latestBuildVersion := string(latestBuildS3)
		fmt.Printf("Previous release version: %s\n", latestBuildVersion)
		fmt.Printf("Current release version provided: %s\n", releaseVersion)

		return generateNewDevReleaseVersion(latestBuildVersion, releaseVersion)
	} else {
		return "v0.0.0-dev+build.0", nil
	}
}

func generateNewDevReleaseVersion(latestBuildVersion, releaseVersion string) (string, error) {
	if releaseVersion == "vDev" { // TODO: remove when we update the pipeline
		releaseVersion = "v0.0.0"
	}

	if !strings.Contains(latestBuildVersion, releaseVersion) && !strings.Contains(latestBuildVersion, "vDev") { // TODO: adding vDev case temporally to support old run, remove later
		// different semver, reset build number suffix on release version
		newReleaseVersion := releaseVersion + "-dev+build.0"
		fmt.Printf("New release version will be: %s\n", newReleaseVersion)
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
	newReleaseVersion := fmt.Sprintf("%s-dev+build.%d", releaseVersion, newBuildNumber)

	return newReleaseVersion, nil
}

func (r *ReleaseConfig) PutEksAReleaseVersion(releaseClients *ReleaseClients, version string) error {
	s3Uploader := releaseClients.S3.Uploader

	var currentReleaseKey string
	if r.BuildRepoBranchName == "main" {
		currentReleaseKey = "LATEST_RELEASE_VERSION"
	} else {
		currentReleaseKey = fmt.Sprintf("%s/LATEST_RELEASE_VERSION", r.BuildRepoBranchName)
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
	err = UploadFileToS3(currentReleaseKey, aws.String(r.ReleaseBucket), aws.String(currentReleaseKey), s3Uploader)
	if err != nil {
		return errors.Cause(err)
	}
	return nil
}

func existsInList(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GenerateRandomSha(hashType int) (string, error) {
	if (hashType != 256) && (hashType != 512) {
		return "", fmt.Errorf("Unsupported hash algorithm: %d", hashType)
	}
	l := mrand.Intn(1000)
	buff := make([]byte, l)

	_, err := crand.Read(buff)
	if err != nil {
		return "", errors.Cause(err)
	}

	var shaSum string
	if hashType == 256 {
		shaSum = fmt.Sprintf("%x", sha256.Sum256(buff))
	} else {
		shaSum = fmt.Sprintf("%x", sha512.Sum512(buff))
	}
	return shaSum, nil
}
