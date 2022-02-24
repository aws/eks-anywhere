package validations

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
)

func ValidateK8s122Support(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.K8s122Support()) {
		if clusterSpec.Cluster.Spec.KubernetesVersion == "1.22" {
			return fmt.Errorf("Kubernetes version 1.22 is not enabled. Please set the env variable %v.", features.K8s122SupportEnvVar)
		}
	}
	return nil
}
