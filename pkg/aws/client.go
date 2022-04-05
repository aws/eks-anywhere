package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Client provides the single API client to make operations call to aws services
type Client struct {
	ec2 *ec2.Client
}

// Clients are a map between aws profile and its aws client
type Clients map[string]*Client

type ServiceEndpoint struct {
	ServiceID     string
	URL           string
	SigningRegion string
}

type awsConfigOpts []func(*config.LoadOptions) error

func LoadConfig(ctx context.Context, credsFilePath string, opts ...awsConfigOpts) (aws.Config, error) {
	optFns := []func(*config.LoadOptions) error{
		config.WithSharedCredentialsFiles([]string{credsFilePath}),
	}

	for _, opt := range opts {
		optFns = append(optFns, opt...)
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("setting aws config: %v", err)
	}
	return cfg, nil
}

func NewClient(ctx context.Context, cfg aws.Config) *Client {
	return &Client{
		ec2: NewEC2Client(cfg),
	}
}
