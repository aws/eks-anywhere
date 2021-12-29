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

func ValidateTaintsSupport(ctx context.Context, clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.TaintsSupport()) {
		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 ||
			len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints) > 0 {
			return fmt.Errorf("Taints feature is not enabled. Environment variable TAINTS_SUPPORT needs to be set to true.")
		}
	} else if len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints) > 0 {
		invalidWorkerNodeGroupTaints := false
		for _, slice := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints {
			if slice.Effect == "NoExecute" || slice.Effect == "NoSchedule" {
				invalidWorkerNodeGroupTaints = true
				break
			}
		}

		if invalidWorkerNodeGroupTaints {
			return fmt.Errorf("The first worker node group does not support NoExecute or NoSchedule taints.")
		}
	}
	return nil
}
