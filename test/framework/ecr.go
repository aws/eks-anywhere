package framework

import (
	"context"
	"encoding/base64"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(defaultRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
	)
	if err != nil {
		e.T.Fatalf("aws config error: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	authTokenOutput, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		e.T.Fatalf("failed to fetch authorization token from ECR for registry %s : %w", registry, err)
	}

	decoded, err := base64.StdEncoding.DecodeString(*authTokenOutput.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		e.T.Fatalf("failed to decode authorization token from ECR for registry %s: %w", registry, err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		e.T.Fatalf("invalid authorization token format from ECR for registry %s", registry)
	}

	username, password := parts[0], parts[1]

	err = buildDocker(e.T).Login(context.Background(), registry, username, password)
	if err != nil {
		e.T.Fatalf("error logging into docker registry %s: %v", registry, err)
	}
}
