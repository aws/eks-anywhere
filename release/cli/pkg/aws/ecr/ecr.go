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

// GetLatestImage takes a repository as input and returns the latest pushed image along with it's sha256 digest.
func GetLatestImage(ecrClient *ecr.ECR, repoName, branchName string, isHelmChart bool) (string, string, error) {
	imageDetails, err := DescribeImagesPaginated(ecrClient, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
		RegistryId: aws.String("067575901363"),
	})
	if len(imageDetails) == 0 {
		return "", "", fmt.Errorf("no image details obtained with DescribeImages API for %s repo: %v", repoName, err)
	}
	if err != nil {
		return "", "", errors.Cause(err)
	}
	
	filteredImageDetails := filterImageDetailsWithBranchName(imageDetails, branchName, isHelmChart)
	if len(filteredImageDetails) == 0 {
		return "", "", fmt.Errorf("no images found with the required filters")
	}

	version, sha, err := getLatestOCIShaTag(filteredImageDetails)
	if err != nil {
		return "", "", err
	}
	return version, sha, nil
}

// filterImageDetailsWithBranchName filters ECR image details based on tag content and branch naming conventions.
func filterImageDetailsWithBranchName(imageDetails []*ecr.ImageDetail, branchName string, isHelmChart bool) []*ecr.ImageDetail {
	filteredImageDetails := []*ecr.ImageDetail{}
	for _, detail := range imageDetails {
		// Skip invalid image entries
		if detail.ImagePushedAt == nil || detail.ImageDigest == nil || detail.ImageTags == nil || len(detail.ImageTags) == 0 {
			continue
		}

		// Exclude image if any tag contains "cache"
		skip := false
		for _, tag := range detail.ImageTags {
			if strings.Contains(*tag, "cache") {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Helm chart filtering logic:
		// - If isHelmChart is false: exclude any image that has a tag containing "helm"
		// - If isHelmChart is true: exclude any image that does not have at least one tag containing "helm"
		if !isHelmChart {
			for _, tag := range detail.ImageTags {
				if strings.Contains(*tag, "helm") {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		} else {
			containsHelm := false
			for _, tag := range detail.ImageTags {
				if strings.Contains(*tag, "helm") {
					containsHelm = true
					break
				}
			}
			if !containsHelm {
				continue
			}
		}

		if branchName == "main" {
			// Exclude image if any tag contains "release"
			for _, tag := range detail.ImageTags {
				if strings.Contains(*tag, "release") {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		} else {
			// Exclude image if none of the tags contain the branchName
			containsBranch := false
			for _, tag := range detail.ImageTags {
				if strings.Contains(*tag, branchName) {
					containsBranch = true
					break
				}
			}
			if !containsBranch {
				continue
			}
		}
		filteredImageDetails = append(filteredImageDetails, detail)
	}
	return filteredImageDetails
}

// getLatestOCIShaTag is used to find the tag/sha of the latest pushed OCI image from a list.
func getLatestOCIShaTag(details []*ecr.ImageDetail) (string, string, error) {
	latest := &ecr.ImageDetail{}
	latest.ImagePushedAt = &time.Time{}
	for _, detail := range details {
		if latest.ImagePushedAt.Before(*detail.ImagePushedAt) {
			latest = detail
		}
	}
	return *latest.ImageTags[0], *latest.ImageDigest, nil
}
