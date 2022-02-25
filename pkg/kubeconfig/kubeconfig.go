package kubeconfig

import (
	"fmt"
	"path/filepath"
)

// FromClusterFormat defines the format of the kubeconfig of the
const FromClusterFormat = "%s-eks-a-cluster.kubeconfig"

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters. This includes a subdirecftory
// named after the cluster name. For example, if the clusterName is 'sandbox' the generated path would be
// sandbox/sandbox-eks-a-cluster.kubeconfig
func FromClusterName(clusterName string) string {
	return filepath.Join(clusterName, fmt.Sprintf(FromClusterFormat, clusterName))
}

type missingFileError struct {
	ClusterName string
	Path        string
}

// NewMissingFileError creates a missing kubeconfig file error.
func NewMissingFileError(cluster, path string) error {
	return missingFileError{Path: path}
}

func (m missingFileError) Error() string {
	return fmt.Sprintf("kubeconfig file missing: path=%v", m.Path)
}
