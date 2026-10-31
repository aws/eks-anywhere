package framework

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func loginToPackagesRegistry(e *ClusterE2ETest, registry string) {
	accessKey := os.Getenv(eksaPackagesAccessKey)
	secretKey := os.Getenv(eksaPackagesSecretKey)
	sessionToken := os.Getenv(eksaPackagesSessionTokenKey)
	loginToECRWithCredentials(e, registry, accessKey, secretKey, sessionToken)
}

// loginToECRWithCredentials performs Docker login with explicit credentials.
func loginToECRWithCredentials(e *ClusterE2ETest, registry, accessKey, secretKey, sessionToken string) {
	username, password, err := ecrBasicAuth(defaultRegion, accessKey, secretKey, sessionToken)
	if err != nil {
		e.T.Fatalf("getting ECR credentials for registry %s: %v", registry, err)
	}

	err = buildDocker(e.T).Login(context.Background(), registry, username, password)
	if err != nil {
		e.T.Fatalf("error logging into docker registry %s: %v", registry, err)
	}
}

// ecrBasicAuth exchanges AWS credentials for the username and password of the ECR registries in region.
func ecrBasicAuth(region, accessKey, secretKey, sessionToken string) (username, password string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
	)
	if err != nil {
		return "", "", fmt.Errorf("aws config error: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	authTokenOutput, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch authorization token from ECR: %v", err)
	}
	if len(authTokenOutput.AuthorizationData) == 0 || authTokenOutput.AuthorizationData[0].AuthorizationToken == nil {
		return "", "", fmt.Errorf("ECR returned no authorization token")
	}

	decoded, err := base64.StdEncoding.DecodeString(*authTokenOutput.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode authorization token from ECR: %v", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid authorization token format from ECR")
	}

	return parts[0], parts[1], nil
}
