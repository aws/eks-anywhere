package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type ClientRegistry interface {
	Get(ctx context.Context) (AwsClientMap, error)
}

type AwsClientRegistry struct {
	deviceClientMap AwsClientMap
}

func NewAwsClientRegistry() *AwsClientRegistry {
	return &AwsClientRegistry{}
}

// Build creates the device client map based on the filepaths specified.
// This method must be called before any Get operations.
func (b *AwsClientRegistry) Build(ctx context.Context) error {
	clients, err := aws.BuildClients(ctx)
	if err != nil {
		return err
	}
	b.deviceClientMap = NewAwsClientMap(clients)
	return nil
}

func (b *AwsClientRegistry) Get(ctx context.Context) (AwsClientMap, error) {
	if b.deviceClientMap == nil {
		return nil, fmt.Errorf("aws clients for snow not initialized")
	}
	return b.deviceClientMap, nil
}
