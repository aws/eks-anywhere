package kubeconfig

import (
	"fmt"
	"path/filepath"
)

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters.
func FromClusterName(clusterName string) string {
	return filepath.Join(clusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName))
}

type missingFileError struct {
	ClusterName string
	Path        string
}

func NewMissingFileError(cluster, path string) error {
	return missingFileError{
		ClusterName: cluster,
		Path:        path,
	}
}

func (m missingFileError) Error() string {
	return fmt.Sprintf("kubeconfig missing for cluster '%v': kubeconfig path=%v", m.ClusterName, m.Path)
}
