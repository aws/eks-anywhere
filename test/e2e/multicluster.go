//go:build e2e
// +build e2e

package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runWorkloadClusterFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.CreateCluster()
		w.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runWorkloadClusterFlowWithGitOps(test *framework.MulticlusterE2ETest, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.CreateCluster()
		w.UpgradeWithGitOps(clusterOpts...)
		time.Sleep(5 * time.Minute)
		w.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runTinkerbellWorkloadClusterFlow(test *framework.MulticlusterE2ETest) {
	test.CreateTinkerbellManagementCluster()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.PowerOffHardware()
		w.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
		w.StopIfFailed()
		w.DeleteCluster()
		w.ValidateHardwareDecommissioned()
	})
	test.DeleteTinkerbellManagementCluster()
}

func runSimpleWorkloadUpgradeFlowForBareMetal(test *framework.MulticlusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.CreateTinkerbellManagementCluster()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.PowerOffHardware()
		w.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
		time.Sleep(2 * time.Minute)
		w.UpgradeCluster(clusterOpts)
		time.Sleep(2 * time.Minute)
		w.ValidateCluster(updateVersion)
		w.StopIfFailed()
		w.DeleteCluster()
		w.ValidateHardwareDecommissioned()
	})
	test.DeleteManagementCluster()
}

func runTinkerbellWorkloadClusterFlowSkipPowerActions(test *framework.MulticlusterE2ETest) {
	test.CreateTinkerbellManagementCluster()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.PowerOffHardware()
		w.PXEBootHardware()
		w.PowerOnHardware()
		w.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
		w.StopIfFailed()
		w.DeleteCluster()
		w.PowerOffHardware()
		w.ValidateHardwareDecommissioned()
	})
	test.ManagementCluster.StopIfFailed()
	test.ManagementCluster.DeleteCluster()
	test.ManagementCluster.PowerOffHardware()
	test.ManagementCluster.ValidateHardwareDecommissioned()
}
