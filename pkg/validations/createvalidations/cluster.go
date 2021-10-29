package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateClusterObjectExists(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterName string) error {
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
