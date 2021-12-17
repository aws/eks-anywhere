package ecrpublic

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/utils"
)

func GetImageDigest(imageUri, imageContainerRegistry string, ecrPublicClient *ecrpublic.ECRPublic) (string, error) {
	repository, tag := utils.SplitImageUri(imageUri, imageContainerRegistry)
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
