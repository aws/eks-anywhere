package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

type EC2Client interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	CreateKeyPair(ctx context.Context, params *ec2.CreateKeyPairInput, optFns ...func(*ec2.Options)) (*ec2.CreateKeyPairOutput, error)
}

func NewEC2Client(config aws.Config) *ec2.Client {
	return ec2.NewFromConfig(config)
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

// EC2CreateKeyPair calls aws sdk ec2.CreateKeyPair to create a key pair with
// name specified from the arg.
func (c *Client) EC2CreateKeyPair(ctx context.Context, keyName string) (keyVal string, err error) {
	params := &ec2.CreateKeyPairInput{
		KeyName: &keyName,
	}
	out, err := c.ec2.CreateKeyPair(ctx, params)
	if err != nil {
		return "", fmt.Errorf("creating key pairs in ec2: %v", err)
	}
	return *out.KeyMaterial, nil
}
