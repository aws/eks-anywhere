package snow

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type AwsClient interface {
	EC2KeyNameExists(ctx context.Context, imageID string) (bool, error)
	EC2ImageExists(ctx context.Context, keyName string) (bool, error)
	EC2CreateKeyPair(ctx context.Context, keyName string) (keyVal string, err error)
}

type AwsClientMap map[string]AwsClient

func NewAwsClientMap(awsClients aws.Clients) AwsClientMap {
	c := make(AwsClientMap, len(awsClients))
	for profile, client := range awsClients {
		c[profile] = client
	}
	return c
}
