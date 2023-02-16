package snow

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type AwsClient interface {
	EC2ImageExists(ctx context.Context, imageID string) (bool, error)
	EC2KeyNameExists(ctx context.Context, keyName string) (bool, error)
	EC2ImportKeyPair(ctx context.Context, keyName string, keyMaterial []byte) error
	EC2InstanceTypes(ctx context.Context) ([]aws.EC2InstanceType, error)
	IsSnowballDeviceUnlocked(ctx context.Context) (bool, error)
	SnowballDeviceSoftwareVersion(ctx context.Context) (string, error)
}

// LocalIMDSClient contains methods that fetch metadata from the local imds.
type LocalIMDSClient interface {
	EC2InstanceIP(ctx context.Context) (string, error)
}

type AwsClientMap map[string]AwsClient

func NewAwsClientMap(awsClients aws.Clients) AwsClientMap {
	c := make(AwsClientMap, len(awsClients))
	for profile, client := range awsClients {
		c[profile] = client
	}
	return c
}
