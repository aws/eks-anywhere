package validations

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
)

func ValidateTaintsSupport(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.TaintsSupport()) {
		wngcTaintsPresent := false
		for _, wngc := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
			if len(wngc.Taints) > 0 {
				wngcTaintsPresent = true
				break
			}
		}

		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 ||
			wngcTaintsPresent {
			return fmt.Errorf("Taints feature is not enabled. Please set the env variable TAINTS_SUPPORT.")
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
