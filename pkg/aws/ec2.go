package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

func NewEC2Client(config aws.Config) *ec2.Client {
	return ec2.NewFromConfig(config)
}

func (a *Client) ImageExists(ctx context.Context, imageID string) (bool, error) {
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{imageID},
	}
	_, err := a.ec2.DescribeImages(ctx, params)
	if err == nil {
		return true, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "400" {
		return false, nil
	}
	return false, fmt.Errorf("aws describe image [%s]: %v", imageID, err)
}

func (a *Client) KeyPairExists(ctx context.Context, keyName string) (bool, error) {
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	}
	out, err := a.ec2.DescribeKeyPairs(ctx, params)
	if err != nil {
		return false, fmt.Errorf("aws describe key pair [%s]: %v", keyName, err)
	}
	if len(out.KeyPairs) <= 0 {
		return false, nil
	}
	return true, nil
}

func (a *Client) CreateEC2KeyPairs(ctx context.Context, keyName string) (keyVal string, err error) {
	params := &ec2.CreateKeyPairInput{
		KeyName: &keyName,
	}
	out, err := a.ec2.CreateKeyPair(ctx, params)
	if err != nil {
		return "", fmt.Errorf("creating key pairs in ec2: %v", err)
	}
	return *out.KeyMaterial, nil
}
