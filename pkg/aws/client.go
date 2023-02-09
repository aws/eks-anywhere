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
	imds           IMDSClient
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

// ClientOpt updates an aws.Client.
type ClientOpt func(*Client)

// WithEC2 returns a ClientOpt that sets the ec2 client.
func WithEC2(ec2 EC2Client) ClientOpt {
	return func(c *Client) {
		c.ec2 = ec2
	}
}

// WithIMDS returns a ClientOpt that sets the imds client.
func WithIMDS(imds IMDSClient) ClientOpt {
	return func(c *Client) {
		c.imds = imds
	}
}

// WithSnowballDevice returns a ClientOpt that sets the snowballdevice client.
func WithSnowballDevice(snowballdevice SnowballDeviceClient) ClientOpt {
	return func(c *Client) {
		c.snowballDevice = snowballdevice
	}
}

// NewClient builds an aws Client.
func NewClient(opts ...ClientOpt) *Client {
	c := &Client{}

	for _, o := range opts {
		o(c)
	}

	return c
}

// NewClientFromConfig builds an aws client with ec2 and snowballdevice apis from aws config.
func NewClientFromConfig(cfg aws.Config) *Client {
	return NewClient(
		WithEC2(NewEC2Client(cfg)),
		WithSnowballDevice(NewSnowballClient(cfg)),
	)
}
