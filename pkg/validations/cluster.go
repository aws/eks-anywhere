package validations

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
)

func ValidateTaintsSupport(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.TaintsSupport()) {
		workerNodeGroupTaintsPresent := false
		for _, nodeGroup := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
			if len(nodeGroup.Taints) > 0 {
				workerNodeGroupTaintsPresent = true
				break
			}
		}

		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 ||
			workerNodeGroupTaintsPresent {
			return fmt.Errorf("Taints feature is not enabled. Please set the env variable %v.", features.TaintsSupportEnvVar)
		}
	}
	return nil
}

func ValidateNodeLabelsSupport(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.NodeLabelsSupport()) {
		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Labels) > 0 {
			return fmt.Errorf("Node labels feature is not enabled. Please set the env variable %v.", features.NodeLabelsSupportEnvVar)
		}
		for _, workerNodeGroup := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
			if len(workerNodeGroup.Labels) > 0 {
				return fmt.Errorf("Node labels feature is not enabled. Please set the env variable %v.", features.NodeLabelsSupportEnvVar)
			}
		}
	}
	return nil
}

func ValidateK8s122Support(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.K8s122Support()) {
		if clusterSpec.Cluster.Spec.KubernetesVersion == "1.22" {
			return fmt.Errorf("Kubernetes version 1.22 is not enabled. Please set the env variable %v.", features.K8s122SupportEnvVar)
		}
	}
	return nil
}
