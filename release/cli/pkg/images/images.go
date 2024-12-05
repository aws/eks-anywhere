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

package images

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ecrsdk "github.com/aws/aws-sdk-go/service/ecr"
	ecrpublicsdk "github.com/aws/aws-sdk-go/service/ecrpublic"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecr"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/retrier"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
	commandutils "github.com/aws/eks-anywhere/release/cli/pkg/util/command"
	packagesutils "github.com/aws/eks-anywhere/release/cli/pkg/util/packages"
)

func PollForExistence(devRelease bool, authConfig *docker.AuthConfiguration, imageUri, imageContainerRegistry, releaseEnvironment, branchName string) error {
	repository, tag := artifactutils.SplitImageUri(imageUri, imageContainerRegistry)

	var requestUrl string
	if devRelease || releaseEnvironment == "development" {
		requestUrl = fmt.Sprintf("https://%s:%s@%s/v2/%s/manifests/%s", authConfig.Username, authConfig.Password, imageContainerRegistry, repository, tag)
	} else {
		requestUrl = fmt.Sprintf("https://%s:%s@public.ecr.aws/v2/%s/%s/manifests/%s", authConfig.Username, authConfig.Password, filepath.Base(imageContainerRegistry), repository, tag)
	}

	// Creating new GET request
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return errors.Cause(err)
	}

	// Retrier for downloading source ECR images. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ImageNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	retrier := retrier.NewRetrier(60*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if branchName == "main" && artifactutils.IsImageNotFoundError(err) && totalRetries < 60 {
			return true, 30 * time.Second
		}
		return false, 0
	}))

	err = retrier.Retry(func() error {
		var err error
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		bodyStr := string(body)
		if strings.Contains(bodyStr, "MANIFEST_UNKNOWN") {
			return fmt.Errorf("requested image not found: %v", imageUri)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("retries exhausted waiting for source image [%s] to be available for copy: %v", imageUri, err)
	}

	return nil
}

func CopyToDestination(sourceAuthConfig, releaseAuthConfig *docker.AuthConfiguration, sourceImageUri, releaseImageUri string) error {
	retrier := retrier.NewRetrier(60*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if err != nil && totalRetries < 10 {
			return true, 30 * time.Second
		}
		return false, 0
	}))

	sourceRegistryUsername := sourceAuthConfig.Username
	sourceRegistryPassword := sourceAuthConfig.Password
	releaseRegistryUsername := releaseAuthConfig.Username
	releaseRegistryPassword := releaseAuthConfig.Password
	err := retrier.Retry(func() error {
		cmd := exec.Command("skopeo", "copy", "--src-creds", fmt.Sprintf("%s:%s", sourceRegistryUsername, sourceRegistryPassword), "--dest-creds", fmt.Sprintf("%s:%s", releaseRegistryUsername, releaseRegistryPassword), fmt.Sprintf("docker://%s", sourceImageUri), fmt.Sprintf("docker://%s", releaseImageUri), "-f", "oci", "--all")
		out, err := commandutils.ExecCommand(cmd)
		fmt.Println(out)
		if err != nil {
			return fmt.Errorf("executing skopeo copy command: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("retries exhausted performing image copy from source [%s] to destination [%s]: %v", sourceImageUri, releaseImageUri, err)
	}

	return nil
}

func GetSourceImageURI(r *releasetypes.ReleaseConfig, name, repoName string, tagOptions map[string]string, imageTagConfiguration assettypes.ImageTagConfiguration, trimVersionSignifier, hasSeparateTagPerReleaseBranch bool) (string, string, error) {
	var sourceImageUri string
	var latestTag string
	sourcedFromBranch := r.BuildRepoBranchName
	sourceContainerRegistry := r.SourceContainerRegistry
	if packagesutils.NeedsPackagesAccountArtifacts(r) && (repoName == "eks-anywhere-packages" || repoName == "ecr-token-refresher" || repoName == "credential-provider-package") {
		sourceContainerRegistry = r.PackagesSourceContainerRegistry
	}
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		latestTag = artifactutils.GetLatestUploadDestination(r.BuildRepoBranchName)
		if imageTagConfiguration.NonProdSourceImageTagFormat != "" {
			sourceImageTagPrefix := generateFormattedTagPrefix(imageTagConfiguration.NonProdSourceImageTagFormat, tagOptions)
			sourceImageUri = fmt.Sprintf("%s/%s:%s-%s",
				sourceContainerRegistry,
				repoName,
				sourceImageTagPrefix,
				latestTag,
			)
		} else {
			sourceImageUri = fmt.Sprintf("%s/%s:%s",
				sourceContainerRegistry,
				repoName,
				latestTag,
			)
		}
		if strings.HasSuffix(name, "-helm") || strings.HasSuffix(name, "-chart") {
			sourceImageUri += "-helm"
		}
		if trimVersionSignifier {
			sourceImageUri = strings.ReplaceAll(sourceImageUri, ":v", ":")
		}
		if !r.DryRun {
			sourceEcrAuthConfig := r.SourceClients.ECR.AuthConfig
			if packagesutils.NeedsPackagesAccountArtifacts(r) && (repoName == "eks-anywhere-packages" || repoName == "ecr-token-refresher" || repoName == "credential-provider-package") {
				sourceEcrAuthConfig = r.SourceClients.Packages.AuthConfig
			}
			err := PollForExistence(r.DevRelease, sourceEcrAuthConfig, sourceImageUri, sourceContainerRegistry, r.ReleaseEnvironment, r.BuildRepoBranchName)
			if err != nil {
				if r.BuildRepoBranchName != "main" {
					fmt.Printf("Tag corresponding to %s branch not found for %s image. Using image artifact from main\n", r.BuildRepoBranchName, repoName)
					var gitTagFromMain string
					if strings.Contains(name, "bottlerocket-bootstrap") {
						gitTagFromMain = "non-existent"
					} else {
						gitTagPath := tagOptions["projectPath"]
						if hasSeparateTagPerReleaseBranch {
							gitTagPath = filepath.Join(tagOptions["projectPath"], tagOptions["eksDReleaseChannel"])
						}
						gitTagFromMain, err = filereader.ReadGitTag(gitTagPath, r.BuildRepoSource, "main")
						if err != nil {
							return "", "", errors.Cause(err)
						}
					}
					sourceImageUri = strings.NewReplacer(r.BuildRepoBranchName, "latest", tagOptions["gitTag"], gitTagFromMain).Replace(sourceImageUri)
					sourcedFromBranch = "main"
				} else {
					return "", "", errors.Cause(err)
				}
			}
		}
	} else if r.ReleaseEnvironment == "production" {
		if imageTagConfiguration.ProdSourceImageTagFormat != "" {
			sourceImageTagPrefix := generateFormattedTagPrefix(imageTagConfiguration.ProdSourceImageTagFormat, tagOptions)
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				sourceImageTagPrefix,
				r.BundleNumber,
			)
		} else {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["gitTag"],
				r.BundleNumber,
			)
		}
		if trimVersionSignifier {
			sourceImageUri = strings.ReplaceAll(sourceImageUri, ":v", ":")
		}
	}

	return sourceImageUri, sourcedFromBranch, nil
}

func GetReleaseImageURI(r *releasetypes.ReleaseConfig, name, repoName string, tagOptions map[string]string, imageTagConfiguration assettypes.ImageTagConfiguration, trimVersionSignifier, hasSeparateTagPerReleaseBranch bool) (string, error) {
	var releaseImageUri string

	if imageTagConfiguration.ReleaseImageTagFormat != "" {
		releaseImageTagPrefix := generateFormattedTagPrefix(imageTagConfiguration.ReleaseImageTagFormat, tagOptions)
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			releaseImageTagPrefix,
		)
	} else {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["gitTag"],
		)
	}

	var semver string
	if r.DevRelease {
		if r.Weekly {
			semver = r.DevReleaseUriVersion
		} else {
			currentSourceImageUri, _, err := GetSourceImageURI(r, name, repoName, tagOptions, imageTagConfiguration, trimVersionSignifier, hasSeparateTagPerReleaseBranch)
			if err != nil {
				return "", errors.Cause(err)
			}

			previousReleaseImageSemver, err := GetPreviousReleaseImageSemver(r, releaseImageUri)
			if err != nil {
				return "", errors.Cause(err)
			}
			if previousReleaseImageSemver == "" {
				semver = r.DevReleaseUriVersion
			} else {
				fmt.Printf("Previous release image semver for %s image: %s\n", repoName, previousReleaseImageSemver)
				previousReleaseImageUri := fmt.Sprintf("%s-%s", releaseImageUri, previousReleaseImageSemver)

				sameDigest, err := CompareHashWithPreviousBundle(r, currentSourceImageUri, previousReleaseImageUri)
				if err != nil {
					return "", errors.Cause(err)
				}
				if sameDigest {
					semver = previousReleaseImageSemver
					fmt.Printf("Image digest for %s image has not changed, tagging with previous dev release semver: %s\n", repoName, semver)
				} else {
					buildNumber, err := filereader.NewBuildNumberFromLastVersion(previousReleaseImageSemver, "vDev", r.BuildRepoBranchName)
					if err != nil {
						return "", err
					}
					newSemver, err := filereader.GetCurrentEksADevReleaseVersion("vDev", r, buildNumber)
					if err != nil {
						return "", err
					}
					semver = strings.ReplaceAll(newSemver, "+", "-")
					fmt.Printf("Image digest for %s image has changed, tagging with new dev release semver: %s\n", repoName, semver)
				}
			}
		}
	} else {
		semver = fmt.Sprintf("%d", r.BundleNumber)
	}

	releaseImageUri = fmt.Sprintf("%s-%s", releaseImageUri, semver)
	if trimVersionSignifier {
		releaseImageUri = strings.ReplaceAll(releaseImageUri, ":v", ":")
	}

	return releaseImageUri, nil
}

func generateFormattedTagPrefix(imageTagFormat string, tagOptions map[string]string) string {
	formattedTag := imageTagFormat
	re := regexp.MustCompile(`<(\w+)>`)
	searchResults := re.FindAllString(imageTagFormat, -1)
	for _, result := range searchResults {
		trimmedResult := strings.Trim(result, "<>")
		formattedTag = strings.ReplaceAll(formattedTag, result, tagOptions[trimmedResult])
	}
	return formattedTag
}

func CompareHashWithPreviousBundle(r *releasetypes.ReleaseConfig, currentSourceImageUri, previousReleaseImageUri string) (bool, error) {
	if r.DryRun {
		return false, nil
	}
	fmt.Printf("Comparing digests for [%s] and [%s]\n", currentSourceImageUri, previousReleaseImageUri)
	currentSourceImageUriDigest, err := ecr.GetImageDigest(currentSourceImageUri, r.SourceContainerRegistry, r.SourceClients.ECR.EcrClient)
	if err != nil {
		return false, errors.Cause(err)
	}

	previousReleaseImageUriDigest, err := ecrpublic.GetImageDigest(previousReleaseImageUri, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
	if err != nil {
		return false, errors.Cause(err)
	}

	return currentSourceImageUriDigest == previousReleaseImageUriDigest, nil
}

func GetPreviousReleaseImageSemver(r *releasetypes.ReleaseConfig, releaseImageUri string) (string, error) {
	var semver string
	if r.DryRun {
		semver = "v0.0.0-dev-build.0"
	} else {
		bundles := &anywherev1alpha1.Bundles{}
		bundleReleaseManifestKey := r.BundlesManifestFilepath()
		bundleManifestUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", r.ReleaseBucket, bundleReleaseManifestKey)

		keyExists, err := s3.KeyExists(r.ReleaseClients.S3.Client, r.ReleaseBucket, bundleReleaseManifestKey, false)
		if err != nil {
			return "", fmt.Errorf("checking if object [%s] is present in S3 bucket: %v", bundleReleaseManifestKey, err)
		}
		if keyExists {
			contents, err := filereader.ReadHttpFile(bundleManifestUrl)
			if err != nil {
				return "", fmt.Errorf("Error reading bundle manifest from S3: %v", err)
			}

			if err = yaml.Unmarshal(contents, bundles); err != nil {
				return "", fmt.Errorf("Error unmarshaling bundles manifest from [%s]: %v", bundleManifestUrl, err)
			}

			for _, versionedBundle := range bundles.Spec.VersionsBundles {
				vbImages := versionedBundle.Images()
				for _, image := range vbImages {
					if strings.Contains(image.URI, releaseImageUri) {
						imageUri := image.URI
						var differential int
						if r.BuildRepoBranchName == "main" {
							differential = 1
						} else {
							differential = 2
						}
						numDashes := strings.Count(imageUri, "-")
						splitIndex := numDashes - strings.Count(r.BuildRepoBranchName, "-") - differential
						imageUriSplit := strings.SplitAfterN(imageUri, "-", splitIndex)
						semver = imageUriSplit[len(imageUriSplit)-1]
					}
				}
			}
		}
	}
	return semver, nil
}

func ComputeImageDigestFromManifest(ecrPublicClient *ecrpublicsdk.ECRPublic, registry, repository, tag string) (string, error) {
	authToken, err := ecrpublic.GetAuthToken(ecrPublicClient)
	if err != nil {
		return "", errors.Cause(err)
	}
	imageRequestURL := fmt.Sprintf("https://public.ecr.aws/v2/%s/%s/manifests/%s", filepath.Base(registry), repository, tag)
	req, err := http.NewRequest("GET", imageRequestURL, nil)
	if err != nil {
		return "", errors.Cause(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Cause(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Cause(err)
	}

	h := sha256.New()
	h.Write(body)
	digest := h.Sum(nil)

	return fmt.Sprintf("%x", digest), nil
}

func CheckRepositoryImagesAndTagsCountLimit(sourceImageUri, releaseImageUri, sourceContainerRegistry, releaseContainerRegistry string, ecrClient interface{}, ecrPublicClient *ecrpublicsdk.ECRPublic) error {
	repository, _ := artifactutils.SplitImageUri(releaseImageUri, releaseContainerRegistry)

	fmt.Printf("Checking if image %s can be pushed to repository %s\n", releaseImageUri, repository)

	var sourceImageDigest string
	var err error
	switch ecrClient.(type) {
	case *ecrsdk.ECR:
		sourceImageDigest, err = ecr.GetImageDigest(sourceImageUri, sourceContainerRegistry, ecrClient.(*ecrsdk.ECR))
		if err != nil {
			return fmt.Errorf("getting image digest for source image [%s] from ECR: %v", sourceImageUri, err)
		}
	case *ecrpublicsdk.ECRPublic:
		sourceImageDigest, err = ecrpublic.GetImageDigest(sourceImageUri, sourceContainerRegistry, ecrClient.(*ecrpublicsdk.ECRPublic))
		if err != nil {
			return fmt.Errorf("getting image digest for source image [%s] from ECR Public: %v", sourceImageUri, err)
		}
	}

	allImagesCount, err := ecrpublic.GetAllImagesCount(repository, ecrPublicClient)
	if err != nil {
		return fmt.Errorf("getting images count in destination repository [%s]: %v", repository, err)
	}
	if allImagesCount >= constants.MAX_IMAGES_PER_REPOSITORY {
		return fmt.Errorf("cannot push image [%s] since the repository %s already has the maximum allowed number of images which is '%d'", releaseImageUri, repository, constants.MAX_IMAGES_PER_REPOSITORY)
	}

	tagsForImageCount, err := ecrpublic.GetTagsCountForImage(repository, sourceImageDigest, ecrPublicClient)
	if err != nil {
		return fmt.Errorf("getting tags count for image digest [%s] in repository [%s]: %v", sourceImageDigest, repository, err)
	}
	if tagsForImageCount >= constants.MAX_TAGS_PER_IMAGE {
		return fmt.Errorf("cannot push image [%s] since the image with digest [%s] in repository %s already has the maximum allowed number of tags per image which is '%d'", releaseImageUri, sourceImageDigest, repository, constants.MAX_TAGS_PER_IMAGE)
	}

	return nil
}
