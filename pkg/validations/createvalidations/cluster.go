package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
)

func ValidateClusterObjectExists(ctx context.Context, k ValidationsKubectlClient, cluster *types.Cluster, clusterName string) error {
	c, err := k.GetClusters(ctx, cluster)
	if err != nil {
		return err
	}
	for _, capiCluster := range c {
		if capiCluster.Metadata.Name == clusterName {
			return fmt.Errorf("cluster name %s already exists", cluster.Name)
		}
	}
	return nil
}
