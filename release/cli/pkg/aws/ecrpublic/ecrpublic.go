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

package ecrpublic

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
)

func GetImageDigest(imageUri, imageContainerRegistry string, ecrPublicClient *ecrpublic.ECRPublic) (string, error) {
	repository, tag := artifactutils.SplitImageUri(imageUri, imageContainerRegistry)
	describeImagesOutput, err := ecrPublicClient.DescribeImages(
		&ecrpublic.DescribeImagesInput{
			ImageIds: []*ecrpublic.ImageIdentifier{
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

	imageDigest := describeImagesOutput.ImageDetails[0].ImageDigest
	imageDigestStr := *imageDigest
	return imageDigestStr, nil
}

func GetAuthToken(ecrPublicClient *ecrpublic.ECRPublic) (string, error) {
	authTokenOutput, err := ecrPublicClient.GetAuthorizationToken(&ecrpublic.GetAuthorizationTokenInput{})
	if err != nil {
		return "", errors.Cause(err)
	}
	authToken := *authTokenOutput.AuthorizationData.AuthorizationToken

	return authToken, nil
}

func GetAuthConfig(ecrPublicClient *ecrpublic.ECRPublic) (*docker.AuthConfiguration, error) {
	// Get ECR Public authorization token
	authToken, err := GetAuthToken(ecrPublicClient)
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

func GetAllImagesCount(imageRepository string, ecrPublicClient *ecrpublic.ECRPublic) (int, error) {
	allImages := []*ecrpublic.ImageDetail{}

	describeImagesOutput, err := ecrPublicClient.DescribeImages(
		&ecrpublic.DescribeImagesInput{
			RepositoryName: aws.String(imageRepository),
			MaxResults:     aws.Int64(1000),
		},
	)
	if err != nil {
		return 0, errors.Cause(err)
	}
	allImages = append(allImages, describeImagesOutput.ImageDetails...)
	nextToken := describeImagesOutput.NextToken

	for nextToken != nil {
		describeImagesOutput, err = ecrPublicClient.DescribeImages(
			&ecrpublic.DescribeImagesInput{
				RepositoryName: aws.String(imageRepository),
				MaxResults:     aws.Int64(1000),
				NextToken:      nextToken,
			},
		)
		if err != nil {
			return 0, errors.Cause(err)
		}
		allImages = append(allImages, describeImagesOutput.ImageDetails...)
		nextToken = describeImagesOutput.NextToken
	}

	return len(allImages), nil
}

func GetTagsCountForImage(imageRepository, imageDigest string, ecrPublicClient *ecrpublic.ECRPublic) (int, error) {
	describeImagesOutput, err := ecrPublicClient.DescribeImages(
		&ecrpublic.DescribeImagesInput{
			RepositoryName: aws.String(imageRepository),
			ImageIds: []*ecrpublic.ImageIdentifier{
				{
					ImageDigest: aws.String(imageDigest),
				},
			},
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), ecrpublic.ErrCodeImageNotFoundException) {
			return 0, nil
		} else {
			return 0, errors.Cause(err)
		}
	}

	if len(describeImagesOutput.ImageDetails) == 0 {
		return 0, nil
	}

	return len(describeImagesOutput.ImageDetails[0].ImageTags), nil
}

func CheckImageExistence(imageUri, imageContainerRegistry string, ecrPublicClient *ecrpublic.ECRPublic) (bool, error) {
	repository, tag := artifactutils.SplitImageUri(imageUri, imageContainerRegistry)
	_, err := ecrPublicClient.DescribeImages(
		&ecrpublic.DescribeImagesInput{
			ImageIds: []*ecrpublic.ImageIdentifier{
				{
					ImageTag: aws.String(tag),
				},
			},
			RepositoryName: aws.String(repository),
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), ecrpublic.ErrCodeImageNotFoundException) {
			return false, nil
		} else {
			return false, errors.Cause(err)
		}
	}

	return true, nil
}
