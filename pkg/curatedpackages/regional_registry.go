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

const (
	devRegionalECR       string = "067575901363.dkr.ecr.us-west-2.amazonaws.com"
	devRegionalPublicECR string = "public.ecr.aws/x3k6m8v0"
	stagingRegionalECR   string = "TODO.dkr.ecr.us-west-2.amazonaws.com"
)

var prodRegionalECRMap = map[string]string{
	"af-south-1":     "783635962247.dkr.ecr.af-south-1.amazonaws.com",
	"ap-east-1":      "804323328300.dkr.ecr.ap-east-1.amazonaws.com",
	"ap-northeast-1": "143143237519.dkr.ecr.ap-northeast-1.amazonaws.com",
	"ap-northeast-2": "447311122189.dkr.ecr.ap-northeast-2.amazonaws.com",
	"ap-northeast-3": "376465423944.dkr.ecr.ap-northeast-3.amazonaws.com",
	"ap-south-1":     "357015164304.dkr.ecr.ap-south-1.amazonaws.com",
	"ap-south-2":     "388483641499.dkr.ecr.ap-south-2.amazonaws.com",
	"ap-southeast-1": "654894141437.dkr.ecr.ap-southeast-1.amazonaws.com",
	"ap-southeast-2": "299286866837.dkr.ecr.ap-southeast-2.amazonaws.com",
	"ap-southeast-3": "703305448174.dkr.ecr.ap-southeast-3.amazonaws.com",
	"ap-southeast-4": "106475008004.dkr.ecr.ap-southeast-4.amazonaws.com",
	"ca-central-1":   "064352486547.dkr.ecr.ca-central-1.amazonaws.com",
	"eu-central-1":   "364992945014.dkr.ecr.eu-central-1.amazonaws.com",
	"eu-central-2":   "551422459769.dkr.ecr.eu-central-2.amazonaws.com",
	"eu-north-1":     "826441621985.dkr.ecr.eu-north-1.amazonaws.com",
	"eu-south-1":     "787863792200.dkr.ecr.eu-south-1.amazonaws.com",
	"eu-west-1":      "090204409458.dkr.ecr.eu-west-1.amazonaws.com",
	"eu-west-2":      "371148654473.dkr.ecr.eu-west-2.amazonaws.com",
	"eu-west-3":      "282646289008.dkr.ecr.eu-west-3.amazonaws.com",
	"il-central-1":   "131750224677.dkr.ecr.il-central-1.amazonaws.com",
	"me-central-1":   "454241080883.dkr.ecr.me-central-1.amazonaws.com",
	"me-south-1":     "158698011868.dkr.ecr.me-south-1.amazonaws.com",
	"sa-east-1":      "517745584577.dkr.ecr.sa-east-1.amazonaws.com",
	"us-east-1":      "331113665574.dkr.ecr.us-east-1.amazonaws.com",
	"us-east-2":      "297090588151.dkr.ecr.us-east-2.amazonaws.com",
	"us-west-1":      "440460740297.dkr.ecr.us-west-1.amazonaws.com",
	"us-west-2":      "346438352937.dkr.ecr.us-west-2.amazonaws.com",
}

// RegistryAccessTester test if AWS credentials has valid permission to access an ECR registry.
type RegistryAccessTester interface {
	Test(ctx context.Context, accessKey, secret, region, awsConfig, registry string) error
}

// DefaultRegistryAccessTester the default implementation of RegistryAccessTester.
type DefaultRegistryAccessTester struct{}

// Test if the AWS static credential or sharedConfig has valid permission to access an ECR registry.
func (r *DefaultRegistryAccessTester) Test(ctx context.Context, accessKey, secret, region, awsConfig, registry string) (err error) {
	authTokenProvider := &DefaultRegistryAuthTokenProvider{}

	var authToken string
	if len(awsConfig) > 0 {
		authToken, err = authTokenProvider.GetTokenByAWSConfig(ctx, awsConfig)
	} else {
		authToken, err = authTokenProvider.GetTokenByAWSKeySecret(ctx, accessKey, secret, region)
	}
	if err != nil {
		return err
	}

	return TestRegistryWithAuthToken(authToken, registry, http.DefaultClient.Do)
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
	if strings.Contains(defaultRegistry, devAccount) {
		return devRegionalECR
	}
	if strings.Contains(defaultRegistry, stagingAccount) {
		return stagingRegionalECR
	}
	return prodRegionalECRMap[region]
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
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigFiles([]string{file.Name()}),
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetAWSConfigFromKeySecret get AWS config from key, secret and region.
func GetAWSConfigFromKeySecret(ctx context.Context, key, secret, region string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetTokenByAWSKeySecret get auth token by AWS key and secret.
func (d *DefaultRegistryAuthTokenProvider) GetTokenByAWSKeySecret(ctx context.Context, key, secret, region string) (string, error) {
	cfg, err := GetAWSConfigFromKeySecret(ctx, key, secret, region)
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
