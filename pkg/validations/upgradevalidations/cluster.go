package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateClusterObjectExists(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster) error {
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

func ValidateNodeLabelsSupport(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.NodeLabelsSupport()) {
		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Labels) > 0 {
			return fmt.Errorf("Node labels feature is not enabled. Please set the env variable NODE_LABELS_SUPPORT.")
		}
		for _, workerNodeGroup := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
			if len(workerNodeGroup.Labels) > 0 {
				return fmt.Errorf("Node labels feature is not enabled. Please set the env variable NODE_LABELS_SUPPORT.")
			}
		}
	}
	return nil
}
