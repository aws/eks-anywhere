package docker

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

func validateControlPlaneEndpoint(clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint != nil {
		return fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")
	}
	return nil
}
