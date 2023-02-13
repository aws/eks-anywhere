package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// IMDSClient is an imds client that wraps around the aws sdk imds client.
type IMDSClient interface {
	GetMetadata(ctx context.Context, params *imds.GetMetadataInput, optFns ...func(*imds.Options)) (*imds.GetMetadataOutput, error)
}

// NewIMDSClient builds a new imds client.
func NewIMDSClient(config aws.Config) *imds.Client {
	return imds.NewFromConfig(config)
}

// BuildIMDS builds or overrides the imds client in the Client with default aws config.
func (c *Client) BuildIMDS(ctx context.Context) error {
	cfg, err := LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading default aws config: %v", err)
	}

	c.imds = NewIMDSClient(cfg)

	return nil
}

// EC2InstanceIP calls aws sdk imds.GetMetadata with public-ipv4 path to fetch the instance ip from metadata service.
func (c *Client) EC2InstanceIP(ctx context.Context) (string, error) {
	if c.imds == nil || reflect.ValueOf(c.imds).IsNil() {
		return "", errors.New("imds client is not initialized")
	}

	params := &imds.GetMetadataInput{
		Path: "public-ipv4",
	}
	out, err := c.imds.GetMetadata(ctx, params)
	if err != nil {
		return "", fmt.Errorf("fetching instance IP from IMDSv2: %v", err)
	}

	defer out.Content.Close()

	b, err := io.ReadAll(out.Content)
	if err != nil {
		return "", fmt.Errorf("reading output content from IMDSv2 instance ip endpoint: %v", err)
	}
	return string(b), nil
}
