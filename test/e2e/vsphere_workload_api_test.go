//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/test/framework"
)

// TODO(g-gaston): this is a WIP. It serves as an example on how tests
// exercising the full cluster lifecycle can be written. It needs to0 be cleaned up
// and a more robust validation set.
func TestVSphereMulticlusterWorkloadClusterAPI(t *testing.T) {
	ctx := context.Background()
	vsphere := framework.NewVSphere(t)
	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithBottleRocket123(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithBottleRocket121(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithBottleRocket122(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithBottleRocket123(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithBottleRocket124(),
		),
	)
	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.T.Logf("Applying workload cluster %s spec for creation", wc.ClusterName)
		if err := wc.KubectlClient.ApplyManifest(ctx, wc.ManagementClusterKubeconfigFile(), wc.ClusterConfigLocation); err != nil {
			wc.T.Errorf("Failed applying workload cluster config: %s", err)
		}

		// ------------------------
		// This is where the proper validation should start, these are a quick place holder
		// ------------------------

		wc.T.Logf("Waiting for CAPI cluster %s to be created", wc.ClusterName)
		g := gomega.NewWithT(t)
		kc := kubernetes.NewUnAuthClient(wc.KubectlClient)
		g.Expect(kc.Init()).To(gomega.Succeed())
		client := kubernetes.NewKubeconfigClient(kc, wc.ManagementClusterKubeconfigFile())
		g.Eventually(func(gg gomega.Gomega) {
			gg.Expect(
				client.Get(ctx, clusterapi.ClusterName(wc.ClusterConfig.Cluster), constants.EksaSystemNamespace, &clusterv1.Cluster{}),
			).To(gomega.Succeed())
		}, 1*time.Minute).Should(gomega.Succeed())

		wc.T.Logf("Waiting for workload cluster %s to be ready", wc.ClusterName)
		c := &types.Cluster{KubeconfigFile: wc.ManagementClusterKubeconfigFile(), Name: wc.ClusterName}
		if err := wc.KubectlClient.WaitForClusterReady(ctx, c, "20m", clusterapi.ClusterName(wc.ClusterConfig.Cluster)); err != nil {
			wc.T.Errorf("Failed waiting for cluster read: %s", err)
		}

		// ------------------------
		// This is where validations end
		// ------------------------

		wc.T.Logf("Deleting workload cluster %s through API", wc.ClusterName)
		g.Expect(wc.KubectlClient.DeleteEKSACluster(ctx, c, wc.ClusterName, wc.ClusterConfig.Cluster.Namespace)).To(gomega.Succeed())
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}
