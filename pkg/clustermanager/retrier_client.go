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

func NewRetrierClient(client *client, retrier *retrier.Retrier) *retrierClient {
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
		return fmt.Errorf("applying eks-a components spec: %v", err)
	}

	// inject proxy env vars the eksa-controller-manager deployment if proxy is configured
	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		envMap := clusterSpec.Cluster.ProxyConfiguration()
		err = c.Retrier.Retry(
			func() error {
				return c.UpdateEnvironmentVariablesInNamespace(ctx, "deployment", "eksa-controller-manager", envMap, cluster, "eksa-system")
			},
		)
		if err != nil {
			return fmt.Errorf("applying eks-a components spec: %v", err)
		}
	}
	return c.waitForDeployments(ctx, internal.EksaDeployments, cluster)
}
