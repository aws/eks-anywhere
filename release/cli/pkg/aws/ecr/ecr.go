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

package ecr

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
)

func GetImageDigest(imageUri, imageContainerRegistry string, ecrClient *ecr.ECR) (string, error) {
	repository, tag := artifactutils.SplitImageUri(imageUri, imageContainerRegistry)
	imageDetails, err := DescribeImagesPaginated(ecrClient,
		&ecr.DescribeImagesInput{
			ImageIds: []*ecr.ImageIdentifier{
				{
					ImageTag: aws.String(tag),
				},
			},
			RepositoryName: aws.String(repository),
		},
	)
	if err != nil {
		return "", errors.Cause(err)
	}

	imageDigest := imageDetails[0].ImageDigest
	imageDigestStr := *imageDigest
	return imageDigestStr, nil
}

func GetAuthToken(ecrClient *ecr.ECR) (string, error) {
	authTokenOutput, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", errors.Cause(err)
	}
	authToken := *authTokenOutput.AuthorizationData[0].AuthorizationToken

	return authToken, nil
}

func GetAuthConfig(ecrClient *ecr.ECR) (*docker.AuthConfiguration, error) {
	// Get ECR authorization token
	authToken, err := GetAuthToken(ecrClient)
	if err != nil {
		return nil, errors.Cause(err)
	}

	// Decode authorization token to get credential pair
	creds, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, errors.Cause(err)
	}

	// Get password from credential pair
	credsSplit := strings.Split(string(creds), ":")
	password := credsSplit[1]

	// Construct docker auth configuration
	authConfig := &docker.AuthConfiguration{
		Username: "AWS",
		Password: password,
	}

	return authConfig, nil
}

func DescribeImagesPaginated(ecrClient *ecr.ECR, describeInput *ecr.DescribeImagesInput) ([]*ecr.ImageDetail, error) {
	var images []*ecr.ImageDetail
	describeImagesOutput, err := ecrClient.DescribeImages(describeInput)
	if err != nil {
		return nil, errors.Cause(err)
	}
	images = append(images, describeImagesOutput.ImageDetails...)
	if describeImagesOutput.NextToken != nil {
		nextInput := describeInput
		nextInput.NextToken = describeImagesOutput.NextToken
		imageDetails, _ := DescribeImagesPaginated(ecrClient, nextInput)
		images = append(images, imageDetails...)
	}
	return images, nil
}

// FilterECRRepoByTagPrefix will take a substring, and a repository as input and find the latest pushed image matching that substring.
func FilterECRRepoByTagPrefix(ecrClient *ecr.ECR, repoName, prefix string, hasTag bool) (string, string, error) {
	imageDetails, err := DescribeImagesPaginated(ecrClient, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if len(imageDetails) == 0 {
		return "", "", fmt.Errorf("no image details obtained: %v", err)
	}
	if err != nil {
		return "", "", errors.Cause(err)
	}
	var filteredImageDetails []*ecr.ImageDetail
	if hasTag {
		filteredImageDetails = imageTagFilter(imageDetails, prefix)
	} else {
		filteredImageDetails = imageTagFilterWithout(imageDetails, prefix)
	}

	// Filter out any tags that don't match our prefix for doubletagged scenarios
	for _, detail := range filteredImageDetails {
		for _, tag := range detail.ImageTags {
			if tag != nil && !strings.HasPrefix(*tag, prefix) {
				detail.ImageTags = removeStringSlice(detail.ImageTags, *tag)
			}
		}
	}

	// In case we don't find any tag substring matches, we still want to populate the bundle with the latest version.
	if len(filteredImageDetails) < 1 {
		filteredImageDetails = imageDetails
	}
	version, sha, err := getLastestOCIShaTag(filteredImageDetails)
	if err != nil {
		return "", "", err
	}
	return version, sha, nil
}

// imageTagFilter is used when filtering a list of ECR images for a specific tag or tag substring
func imageTagFilter(details []*ecr.ImageDetail, substring string) []*ecr.ImageDetail {
	var filteredDetails []*ecr.ImageDetail
	for _, detail := range details {
		for _, tag := range detail.ImageTags {
			if strings.HasPrefix(*tag, substring) {
				filteredDetails = append(filteredDetails, detail)
			}
		}
	}
	return filteredDetails
}

// imageTagFilterWithout is used when filtering a list of ECR images for images without a specific tag or tag substring
func imageTagFilterWithout(details []*ecr.ImageDetail, substring string) []*ecr.ImageDetail {
	var filteredDetails []*ecr.ImageDetail
	for _, detail := range details {
		for _, tag := range detail.ImageTags {
			if !strings.HasPrefix(*tag, substring) {
				filteredDetails = append(filteredDetails, detail)
			}
		}
	}
	return filteredDetails
}

// getLastestOCIShaTag is used to find the tag/sha of the latest pushed OCI image from a list.
func getLastestOCIShaTag(details []*ecr.ImageDetail) (string, string, error) {
	latest := &ecr.ImageDetail{}
	latest.ImagePushedAt = &time.Time{}
	for _, detail := range details {
		if len(details) < 1 || detail.ImagePushedAt == nil || detail.ImageDigest == nil || detail.ImageTags == nil || len(detail.ImageTags) == 0 {
			continue
		}
		if detail.ImagePushedAt != nil && latest.ImagePushedAt.Before(*detail.ImagePushedAt) {
			latest = detail
		}
	}
	if reflect.DeepEqual(latest, ecr.ImageDetail{}) {
		return "", "", fmt.Errorf("error no images found")
	}
	return *latest.ImageTags[0], *latest.ImageDigest, nil
}

func GetLatestImageSha(ecrClient *ecr.ECR, repoName string) (string, error) {
	imageDetails, err := DescribeImagesPaginated(ecrClient, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
	})
	if len(imageDetails) == 0 {
		return "", fmt.Errorf("no image details obtained: %v", err)
	}
	if err != nil {
		return "", errors.Cause(err)
	}

	latest := &ecr.ImageDetail{}
	latest.ImagePushedAt = &time.Time{}
	for _, detail := range imageDetails {
		if detail.ImagePushedAt == nil || detail.ImageDigest == nil || detail.ImageTags == nil || len(detail.ImageTags) == 0 || *detail.ImageManifestMediaType != "application/vnd.oci.image.manifest.v1+json" {
			continue
		}
		if detail.ImagePushedAt != nil && latest.ImagePushedAt.Before(*detail.ImagePushedAt) {
			latest = detail
		}
	}
	// Check if latest is empty, and return error if that's the case.
	if *latest.ImageTags[0] == "" {
		return "", fmt.Errorf("error no images found")
	}
	return *latest.ImageTags[0], nil
}

// removeStringSlice removes a named string from a slice, without knowing it's index or it being ordered.
func removeStringSlice(l []*string, item string) []*string {
	for i, other := range l {
		if *other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}
