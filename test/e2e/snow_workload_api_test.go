//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/test/framework"
)

func TestSnowMulticlusterWorkloadClusterAPI(t *testing.T) {
	snow := framework.NewSnow(t)
	managementCluster := framework.NewClusterE2ETest(
		t, snow,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		snow.WithBottlerocket124(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket121(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket122(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket123(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket124(),
		),
	)
	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.DeleteClusterWithKubectl()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}
