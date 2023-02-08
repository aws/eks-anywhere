package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice"
	"github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice/types"
)

type SnowballDeviceClient interface {
	DescribeDevice(ctx context.Context, params *snowballdevice.DescribeDeviceInput, optFns ...func(*snowballdevice.Options)) (*snowballdevice.DescribeDeviceOutput, error)
	DescribeDeviceSoftware(ctx context.Context, params *snowballdevice.DescribeDeviceSoftwareInput, optFns ...func(*snowballdevice.Options)) (*snowballdevice.DescribeDeviceSoftwareOutput, error)
}

func NewSnowballClient(config aws.Config) *snowballdevice.Client {
	return snowballdevice.NewFromConfig(config)
}

func (c *Client) IsSnowballDeviceUnlocked(ctx context.Context) (bool, error) {
	out, err := c.snowballDevice.DescribeDevice(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("describing snowball device: %v", err)
	}
	return out.UnlockStatus.State == types.UnlockStatusStateUnlocked, nil
}

func (c *Client) SnowballDeviceSoftwareVersion(ctx context.Context) (string, error) {
	out, err := c.snowballDevice.DescribeDeviceSoftware(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("describing snowball device software: %v", err)
	}
	return *out.InstalledVersion, nil
}
