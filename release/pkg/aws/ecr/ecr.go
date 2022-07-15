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

	artifactutils "github.com/aws/eks-anywhere/release/pkg/util/artifacts"
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
