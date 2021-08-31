package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
)

func ValidateClusterObjectExists(ctx context.Context, k ValidationsKubectlClient, cluster *types.Cluster) error {
	c, err := k.GetClusters(ctx, cluster)
	if err != nil {
		return err
	}
	if len(c) == 0 {
		return fmt.Errorf("no CAPI cluster objects present on workload cluster %s", cluster.Name)
	}
	for _, capiCluster := range c {
		if capiCluster.Metadata.Name == cluster.Name {
			return nil
		}
	}
	return fmt.Errorf("couldn't find CAPI cluster object for cluster with name %s", cluster.Name)
}
