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

	err = c.ApplyKubeSpecFromBytes(ctx, cluster, componentsManifest.Content)
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

func (c *retrierClient) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

func (c *retrierClient) ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesForce(ctx, cluster, data)
		},
	)
}

func (c *retrierClient) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace)
		},
	)
}

func (c *retrierClient) UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.UpdateAnnotationInNamespace(ctx, resourceType, objectName, annotations, cluster, namespace)
		},
	)
}

func (c *retrierClient) RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error {
	return c.Retry(
		func() error {
			return c.ClusterClient.RemoveAnnotationInNamespace(ctx, resourceType, objectName, key, cluster, namespace)
		},
	)
}
