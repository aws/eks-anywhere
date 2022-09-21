package docker_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
)

func TestControlPlaneSpec(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.VersionsBundle = versionsBundle
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
	})

	cp, err := docker.ControlPlaneSpec(logger, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
}
