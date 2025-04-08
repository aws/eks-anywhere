package curatedpackages

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// RegistryAccessTestParams struct for testing resigtry access.
type RegistryAccessTestParams struct {
	AccessKey    string
	Secret       string
	SessionToken string
	Region       string
	AwsConfig    string
	Registry     string
}

// RegistryAccessTester test if AWS credentials has valid permission to access an ECR registry.
type RegistryAccessTester interface {
	Test(ctx context.Context, params RegistryAccessTestParams) error
}

// DefaultRegistryAccessTester the default implementation of RegistryAccessTester.
type DefaultRegistryAccessTester struct{}

// Test if the AWS static credential or sharedConfig has valid permission to access an ECR registry.
func (r *DefaultRegistryAccessTester) Test(ctx context.Context, params RegistryAccessTestParams) (err error) {
	authTokenProvider := &DefaultRegistryAuthTokenProvider{}

	var authToken string
	if len(params.AwsConfig) > 0 {
		authToken, err = authTokenProvider.GetTokenByAWSConfig(ctx, params.AwsConfig)
	} else {
		authToken, err = authTokenProvider.GetTokenByAWSKeySecret(ctx, params.AccessKey, params.Secret, params.SessionToken, params.Region)
	}
	if err != nil {
		return err
	}

	return TestRegistryWithAuthToken(authToken, params.Registry, http.DefaultClient.Do)
}

// TestRegistryWithAuthToken test if the registry can be acccessed with auth token.
func TestRegistryWithAuthToken(authToken, registry string, do Do) error {
	manifestPath := "/v2/eks-anywhere-packages/manifests/latest"

	req, err := http.NewRequest("GET", "https://"+registry+manifestPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Basic "+authToken)

	resp2, err := do(req)
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(resp2.Body)
	// 404 means the IAM policy is good, so 404 is good here
	if resp2.StatusCode != 200 && resp2.StatusCode != 404 {
		return fmt.Errorf("%s\n, %v", string(bodyBytes), err)
	}

	return nil
}

// GetRegionalRegistry get the regional registry corresponding to defaultRegistry in a specific region.
func GetRegionalRegistry(defaultRegistry, region string) string {
	if strings.Contains(defaultRegistry, devRegionalPublicRegistryAlias) {
		return devRegionalPrivateRegistryURI
	}
	if strings.Contains(defaultRegistry, stagingPublicRegistryAlias) {
		return stagingRegionalPrivateRegistryURI
	}
	return prodRegionalPrivateRegistryURIByRegion[region]
}

// RegistryAuthTokenProvider provides auth token for registry access.
type RegistryAuthTokenProvider interface {
	GetTokenByAWSConfig(ctx context.Context, awsConfig string) (string, error)
	GetTokenByAWSKeySecret(ctx context.Context, key, secret, region string) (string, error)
}

// DefaultRegistryAuthTokenProvider provides auth token for AWS ECR registry access.
type DefaultRegistryAuthTokenProvider struct{}

// GetTokenByAWSConfig get auth token by AWS config.
func (d *DefaultRegistryAuthTokenProvider) GetTokenByAWSConfig(ctx context.Context, awsConfig string) (string, error) {
	cfg, err := ParseAWSConfig(ctx, awsConfig)
	if err != nil {
		return "", err
	}
	return getAuthorizationToken(*cfg)
}

// ParseAWSConfig parse AWS config from string.
func ParseAWSConfig(ctx context.Context, awsConfig string) (*aws.Config, error) {
	file, err := os.CreateTemp("", "eksa-temp-aws-config-*")
	if err != nil {
		return nil, err
	}
	if _, err := file.Write([]byte(awsConfig)); err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigFiles([]string{file.Name()}),
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetAWSConfigFromKeySecret get AWS config from key, secret and region.
func GetAWSConfigFromKeySecret(ctx context.Context, key, secret, sessionToken, region string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, sessionToken)),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetTokenByAWSKeySecret get auth token by AWS key and secret.
func (d *DefaultRegistryAuthTokenProvider) GetTokenByAWSKeySecret(ctx context.Context, key, secret, sessionToken, region string) (string, error) {
	cfg, err := GetAWSConfigFromKeySecret(ctx, key, secret, sessionToken, region)
	if err != nil {
		return "", err
	}

	return getAuthorizationToken(*cfg)
}

func getAuthorizationToken(cfg aws.Config) (string, error) {
	ecrClient := ecr.NewFromConfig(cfg)
	out, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", fmt.Errorf("ecrClient cannot get authorization token: %w", err)
	}
	authToken := out.AuthorizationData[0].AuthorizationToken
	return *authToken, nil
}

// Do is a function type that takes a http request and returns a http response.
type Do func(req *http.Request) (*http.Response, error)

// TestRegistryAccessWithAWSConfig test if the AWS config has valid permission to access container registry.
func TestRegistryAccessWithAWSConfig(ctx context.Context, awsConfig, registry string, tokenProvider RegistryAuthTokenProvider, do Do) error {
	token, err := tokenProvider.GetTokenByAWSConfig(ctx, awsConfig)
	if err != nil {
		return err
	}
	return TestRegistryWithAuthToken(token, registry, do)
}

// TestRegistryAccessWithAWSKeySecret test if the AWS key and secret has valid permission to access container registry.
func TestRegistryAccessWithAWSKeySecret(ctx context.Context, key, secret, region, registry string, tokenProvider RegistryAuthTokenProvider, do Do) error {
	token, err := tokenProvider.GetTokenByAWSKeySecret(ctx, key, secret, region)
	if err != nil {
		return err
	}
	return TestRegistryWithAuthToken(token, registry, do)
}
