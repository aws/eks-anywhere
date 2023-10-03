package curatedpackages_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestTestRegistryWithAuthToken(t *testing.T) {
	cases := []struct {
		description string
		statusCode  int
		hasError    bool
	}{
		{"200 status code does not cause error", 200, false},
		{"404 status code does not cause error", 404, false},
		{"400 status code causes error", 400, true},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			err := curatedpackages.TestRegistryWithAuthToken("authToken", "registry_url", func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.statusCode,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil
			})
			if test.hasError && err == nil {
				t.Errorf("Error should have been returned")
			}
			if !test.hasError && err != nil {
				t.Errorf("Registry is good, but error has been returned %v\n", err)
			}
		})
	}
}

var (
	goodAwsConfig  = "good-aws-config"
	badAwsConfig   = "bad-aws-config"
	errorAwsConfig = "error-aws-config"

	goodAwsKey    = "good-aws-key"
	goodAwsSecret = "good-aws-secret"
	errorAwsKey   = "error-aws-key"
	badAwsKey     = "bad-aws-key"
	badAwsSecret  = "bad-aws-secret"

	goodAuthToken = "good-auth-token"
	badAuthToken  = "bad-auth-token"
)

type mockRegistryAuthTokenProvider struct{}

func (m *mockRegistryAuthTokenProvider) GetTokenByAWSConfig(ctx context.Context, awsConfig string) (string, error) {
	if awsConfig == goodAwsConfig {
		return goodAuthToken, nil
	}
	if awsConfig == errorAwsKey {
		return "", fmt.Errorf("something wrong with the aws config")
	}
	return badAuthToken, nil
}

func (m *mockRegistryAuthTokenProvider) GetTokenByAWSKeySecret(ctx context.Context, key, secret, region string) (string, error) {
	if key == goodAwsKey && secret == goodAwsSecret {
		return goodAuthToken, nil
	}
	if key == errorAwsKey {
		return "", fmt.Errorf("something wrong with the aws key")
	}
	return badAuthToken, nil
}

func TestTestAWSConfigRegistryAccessWithAWSConfig(t *testing.T) {
	registryServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.Contains(auth, goodAuthToken) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer registryServer.Close()

	registry := strings.TrimPrefix(registryServer.URL, "https://")
	mockProvider := &mockRegistryAuthTokenProvider{}

	t.Run("return error for bad auth token", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSConfig(context.Background(), badAwsConfig, registry, mockProvider, registryServer.Client().Do)
		if err == nil {
			t.Errorf("Error should have been returned")
		}
	})
	t.Run("return no error for good auth token", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSConfig(context.Background(), goodAwsConfig, registry, mockProvider, registryServer.Client().Do)
		if err != nil {
			t.Errorf("Error should not have been returned for good AWS Config: %s", err)
		}
	})

	t.Run("return error when auth token retrival failed", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSConfig(context.Background(), errorAwsConfig, registry, mockProvider, registryServer.Client().Do)
		if err == nil {
			t.Errorf("Error should have returned for error aws config")
		}
	})
}

func TestTestAWSConfigRegistryAccessWithAWSKeySecret(t *testing.T) {
	registryServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.Contains(auth, goodAuthToken) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer registryServer.Close()

	registry := strings.TrimPrefix(registryServer.URL, "https://")
	mockProvider := &mockRegistryAuthTokenProvider{}

	t.Run("return error for bad auth token", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSKeySecret(context.Background(), badAwsKey, badAwsSecret, "us-west-2", registry, mockProvider, registryServer.Client().Do)
		if err == nil {
			t.Errorf("Error should have been returned")
		}
	})
	t.Run("return no error for good auth token", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSKeySecret(context.Background(), goodAwsKey, goodAwsSecret, "us-west-2", registry, mockProvider, registryServer.Client().Do)
		if err != nil {
			t.Errorf("Error should not have been returned for good AWS Config: %s", err)
		}
	})

	t.Run("retur error if auth token generation failed", func(t *testing.T) {
		err := curatedpackages.TestRegistryAccessWithAWSKeySecret(context.Background(), errorAwsKey, goodAwsSecret, "us-west-2", registry, mockProvider, registryServer.Client().Do)
		if err == nil {
			t.Errorf("Error should have been returned for error AWS Config")
		}
	})
}

func TestParseAWSConfig(t *testing.T) {
	awsConfig := `
[default]
region=us-west-2
aws_access_key_id=keyid
aws_secret_access_key=secret
	`
	parsed, err := curatedpackages.ParseAWSConfig(context.Background(), awsConfig)
	if err != nil {
		t.Errorf("Error parsing AWS Config: %s", err)
	}
	if parsed.Region != "us-west-2" {
		t.Errorf("Region is not parsed correctly")
	}
}

func TestGetAWSConfigFromKeySecret(t *testing.T) {
	cfg, err := curatedpackages.GetAWSConfigFromKeySecret(context.Background(), "key", "secret", "us-west-2")
	if err != nil {
		t.Errorf("Error parsing AWS Config: %s", err)
	}
	cred, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Errorf("Error retrieving credentials: %s", err)
	}

	if cred.AccessKeyID != "key" {
		t.Errorf("AccessKeyId is not generated correctly")
	}

	if cred.SecretAccessKey != "secret" {
		t.Errorf("secret is not generated correctly")
	}

	if cfg.Region != "us-west-2" {
		t.Errorf("Region is not parsed correctly")
	}
}
