package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Client provides the single API client to make operations call to aws services.
type Client struct {
	ec2            EC2Client
	snowballDevice SnowballDeviceClient
}

// Clients are a map between aws profile and its aws client.
type Clients map[string]*Client

type ServiceEndpoint struct {
	ServiceID     string
	URL           string
	SigningRegion string
}

type AwsConfigOpt = config.LoadOptionsFunc

func AwsConfigOptSet(opts ...AwsConfigOpt) AwsConfigOpt {
	return func(conf *config.LoadOptions) error {
		for _, opt := range opts {
			if err := opt(conf); err != nil {
				return err
			}
		}

		return nil
	}
}

// LoadConfig reads the optional aws configurations, and populates an AWS Config
// with the values from the configurations.
func LoadConfig(ctx context.Context, opts ...AwsConfigOpt) (aws.Config, error) {
	optFns := []func(*config.LoadOptions) error{}

	for _, opt := range opts {
		optFns = append(optFns, opt)
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("setting aws config: %v", err)
	}
	return cfg, nil
}

func NewClient(ctx context.Context, cfg aws.Config) *Client {
	return &Client{
		ec2:            NewEC2Client(cfg),
		snowballDevice: NewSnowballClient(cfg),
	}
}

// NewClientFromEC2 is mainly used for EC2 related unit tests.
func NewClientFromEC2(ec2 EC2Client) *Client {
	return &Client{
		ec2: ec2,
	}
}

// NewClientFromSnowball is mainly used for Snowballdevice related unit tests.
func NewClientFromSnowball(snowballdevice SnowballDeviceClient) *Client {
	return &Client{
		snowballDevice: snowballdevice,
	}
}
