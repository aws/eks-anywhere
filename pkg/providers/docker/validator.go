package docker

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

// ValidateControlPlaneEndpoint - checks to see if endpoint host configuration is specified for docker cluster and returns an error if true.
func ValidateControlPlaneEndpoint(clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint != nil {
		return fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")
	}
	return nil
}
