package kubeconfig

import (
	"fmt"
	"path/filepath"
)

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters.
func FromClusterName(clusterName string) string {
	return filepath.Join(clusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName))
}
