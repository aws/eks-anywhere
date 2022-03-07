package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
)

// FromClusterFormat defines the format of the kubeconfig of the
const FromClusterFormat = "%s-eks-a-cluster.kubeconfig"

// EnvName is the standard KubeConfig environment variable name.
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#set-the-kubeconfig-environment-variable
const EnvName = "KUBECONFIG"

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters. This includes a subdirecftory
// named after the cluster name. For example, if the clusterName is 'sandbox' the generated path would be
// sandbox/sandbox-eks-a-cluster.kubeconfig
func FromClusterName(clusterName string) string {
	return filepath.Join(clusterName, fmt.Sprintf(FromClusterFormat, clusterName))
}

// FromEnvironment reads the KubeConfig path from the standard KUBECONFIG environment variable.
func FromEnvironment() string {
	return os.Getenv(EnvName)
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
