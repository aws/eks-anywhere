package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateClusterNameIsUnique(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterName string) error {
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

func ValidateManagementCluster(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster) error {
	if err := k.ValidateClustersCRD(ctx, cluster); err != nil {
		return err
	}
	return k.ValidateEKSAClustersCRD(ctx, cluster)
}

func ValidateTaintsSupport(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.TaintsSupport()) {
		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 {
			return fmt.Errorf("Taints feature is not enabled. Please set the env variable TAINTS_SUPPORT.")
		}
	}
	return nil
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
