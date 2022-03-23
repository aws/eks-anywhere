package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

type ClusterClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
}

func ApplyExtraObjects(ctx context.Context, clusterClient ClusterClient, cluster *types.Cluster, clusterSpec *Spec) error {
	extraObjects := BuildExtraObjects(clusterSpec)
	if len(extraObjects) <= 0 {
		return nil
	}

	resourcesSpec := templater.AppendYamlResources(extraObjects.Values()...)

	logger.V(4).Info("Applying extra objects", "cluster", clusterSpec.Cluster.Name, "resources", extraObjects.Names())
	err := clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, resourcesSpec)
	if err != nil {
		return fmt.Errorf("error applying spec for extra resources to cluster %s: %v", cluster.Name, err)
	}

	return nil
}
