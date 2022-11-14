package kubeconfig

import "fmt"

// FormatWorkloadClusterKubeconfigFilename returns a filename for the Kubeconfig of workload
// clusters. The filename does not include a basepath.
func FormatWorkloadClusterKubeconfigFilename(clusterName string) string {
	return fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName)
}
