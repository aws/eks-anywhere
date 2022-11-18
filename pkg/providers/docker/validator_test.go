package docker_test

import (
	"fmt"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestValidateControlplaneEndpoint(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "test-ip"}
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
	})
	wantErr := fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")

	err := docker.ValidateControlPlaneEndpoint(clusterSpec)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("Got err %v, wanted %v", err, wantErr)
	}
}
