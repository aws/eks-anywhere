package curatedpackages

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

const (
	devRegionalECR     string = "067575901363.dkr.ecr.us-west-2.amazonaws.com"
	stagingRegionalECR string = "TODO.dkr.ecr.us-west-2.amazonaws.com"
)

var prodRegionalECRMap = map[string]string{
	"us-west-2": "TODO.dkr.ecr.us-west-2.amazonaws.com",
	"us-east-2": "TODO.dkr.ecr.us-east-2.amazonaws.com",
}

// TestRegistryAccess test if the packageControllerClient has valid credential to access registry.
func TestRegistryAccess(ctx context.Context, accessKey, secret, registry, region string) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secret, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return err
	}

	ecrClient := ecr.NewFromConfig(cfg)
	out, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}
	authToken := out.AuthorizationData[0].AuthorizationToken

	return TestRegistryWithAuthToken(*authToken, registry, http.DefaultClient.Do)
}

// TestRegistryWithAuthToken test if the registry can be acccessed with auth token.
func TestRegistryWithAuthToken(authToken, registry string, getResponse func(req *http.Request) (*http.Response, error)) error {
	manifestPath := "/v2/eks-anywhere-packages/manifests/latest"

	req, err := http.NewRequest("GET", "https://"+registry+manifestPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Basic "+authToken)

	resp2, err := getResponse(req)
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
