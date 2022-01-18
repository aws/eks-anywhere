package ecr

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/utils"
)

func GetImageDigest(imageUri, imageContainerRegistry string, ecrClient *ecr.ECR) (string, error) {
	repository, tag := utils.SplitImageUri(imageUri, imageContainerRegistry)
	describeImagesOutput, err := ecrClient.DescribeImages(
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

	imageDigest := describeImagesOutput.ImageDetails[0].ImageDigest
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
