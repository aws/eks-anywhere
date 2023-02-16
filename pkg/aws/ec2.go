package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

// EC2Client is an ec2 client that wraps around the aws sdk ec2 client.
type EC2Client interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	DescribeInstanceTypes(ctx context.Context, params *ec2.DescribeInstanceTypesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error)
	ImportKeyPair(ctx context.Context, params *ec2.ImportKeyPairInput, optFns ...func(*ec2.Options)) (*ec2.ImportKeyPairOutput, error)
}

// NewEC2Client builds a new ec2 client.
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

// EC2InstanceType has the information of an ec2 instance type.
type EC2InstanceType struct {
	Name        string
	DefaultVCPU *int32
}

// EC2InstanceTypes calls aws sdk ec2.DescribeInstanceTypes to get a list of supported instance type for a device.
func (c *Client) EC2InstanceTypes(ctx context.Context) ([]EC2InstanceType, error) {
	out, err := c.ec2.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{})
	if err != nil {
		return nil, fmt.Errorf("describing ec2 instance type in device: %v", err)
	}

	instanceTypes := make([]EC2InstanceType, 0, len(out.InstanceTypes))
	for _, it := range out.InstanceTypes {
		instanceTypes = append(instanceTypes, EC2InstanceType{
			Name:        string(it.InstanceType),
			DefaultVCPU: it.VCpuInfo.DefaultVCpus,
		})
	}
	return instanceTypes, nil
}
