package bootstrapper

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierClient struct {
	ClusterClient
	*retrier.Retrier
}

func NewRetrierClient(client *ClusterClient, retrier *retrier.Retrier) *retrierClient {
	return &retrierClient{
		ClusterClient: *client,
		Retrier:       retrier,
	}
}

func (c *retrierClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}
