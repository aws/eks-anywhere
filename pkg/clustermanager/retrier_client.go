package clustermanager

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierClient struct {
	*client
	*retrier.Retrier
}

func newRetrierClient(client *client, retrier *retrier.Retrier) *retrierClient {
	return &retrierClient{
		client:  client,
		Retrier: retrier,
	}
}

func (c *retrierClient) installCustomComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	componentsManifest, err := clusterSpec.LoadManifest(clusterSpec.VersionsBundle.Eksa.Components)
	if err != nil {
		return fmt.Errorf("failed loading manifest for eksa components: %v", err)
	}

	err = c.Retry(
		func() error {
			return c.ApplyKubeSpecFromBytes(ctx, cluster, componentsManifest.Content)
		},
	)
	if err != nil {
		return fmt.Errorf("error applying eks-a components spec: %v", err)
	}
	return c.waitForDeployments(ctx, internal.EksaDeployments, cluster)
}
