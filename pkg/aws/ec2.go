package aws

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

// EC2Client is an ec2 client that wraps around the aws sdk ec2 client.
type EC2Client interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	ImportKeyPair(ctx context.Context, params *ec2.ImportKeyPairInput, optFns ...func(*ec2.Options)) (*ec2.ImportKeyPairOutput, error)
}

// IMDSClient is an imds client that wraps around the aws sdk imds client.
type IMDSClient interface {
	GetMetadata(ctx context.Context, params *imds.GetMetadataInput, optFns ...func(*imds.Options)) (*imds.GetMetadataOutput, error)
}

// EC2 contains the ec2 and imds client.
type EC2 struct {
	EC2Client
	IMDSClient
}

// NewEC2Client builds a new ec2 client.
func NewEC2Client(config aws.Config) *EC2 {
	return &EC2{
		ec2.NewFromConfig(config),
		imds.NewFromConfig(config),
	}
}

// EC2ImageExists calls aws sdk ec2.DescribeImages with filter imageID to fetch a
// specified images (AMIs, AKIs, and ARIs) available.
// Returns (false, nil) if the image does not exist, (true, nil) if image exists,
// and (false, err) if there is an non 400 status code error from ec2.DescribeImages.
func (c *Client) EC2ImageExists(ctx context.Context, imageID string) (bool, error) {
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{imageID},
	}
	_, err := c.ec2.DescribeImages(ctx, params)
	if err == nil {
		return true, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "400" {
		return false, nil
	}
	return false, fmt.Errorf("aws describe image [imageID=%s]: %v", imageID, err)
}

// EC2KeyNameExists calls aws sdk ec2.DescribeKeyPairs with filter keyName to fetch a
// specified key pair available in aws.
// Returns (false, nil) if the key pair does not exist, (true, nil) if key pair exists,
// and (false, err) if there is an error from ec2.DescribeKeyPairs.
func (c *Client) EC2KeyNameExists(ctx context.Context, keyName string) (bool, error) {
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	}
	out, err := c.ec2.DescribeKeyPairs(ctx, params)
	if err != nil {
		return false, fmt.Errorf("aws describe key pair [keyName=%s]: %v", keyName, err)
	}
	if len(out.KeyPairs) <= 0 {
		return false, nil
	}
	return true, nil
}

// EC2ImportKeyPair calls aws sdk ec2.ImportKeyPair to import a key pair to ec2.
func (c *Client) EC2ImportKeyPair(ctx context.Context, keyName string, keyMaterial []byte) error {
	params := &ec2.ImportKeyPairInput{
		KeyName:           &keyName,
		PublicKeyMaterial: keyMaterial,
	}
	_, err := c.ec2.ImportKeyPair(ctx, params)
	if err != nil {
		return fmt.Errorf("importing key pairs in ec2: %v", err)
	}
	return nil
}

// EC2InstanceIP calls aws sdk imds.GetMetadata with public-ipv4 path to fetch the instance ip from metadata service.
func (c *Client) EC2InstanceIP(ctx context.Context) (string, error) {
	params := &imds.GetMetadataInput{
		Path: "public-ipv4",
	}
	out, err := c.ec2.GetMetadata(ctx, params)
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
