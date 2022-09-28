//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/test/framework"
)

func runWorkloadClusterFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementCluster()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfig()
		w.CreateCluster()
		w.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runWorkloadClusterFlowWithGitOps(test *framework.MulticlusterE2ETest, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.CreateManagementCluster()
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

func runWorkloadClusterUpgradeFlowCheckWorkloadRollingUpgrade(test *framework.MulticlusterE2ETest, clusterOpts ...framework.ClusterE2ETestOpt) {
	latest := latestMinorRelease(test.T)
	test.CreateManagementClusterForVersion(latest.Version, framework.ExecuteWithEksaRelease(latest))
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfigForVersion(latest.Version)
		w.CreateCluster(framework.ExecuteWithEksaRelease(latest))
	})
	preUpgradeMachines := make(map[string]map[string]types.Machine, 0)
	for key, workloadCluster := range test.WorkloadClusters {
		test.T.Logf("Capturing CAPI machines for cluster %v", workloadCluster)
		mdName := fmt.Sprintf("%s-%s", workloadCluster.ClusterName, "md-0")
		test.ManagementCluster.WaitForMachineDeploymentReady(mdName)
		preUpgradeMachines[key] = test.ManagementCluster.GetCapiMachinesForCluster(workloadCluster.ClusterName)
	}
	test.ManagementCluster.UpgradeClusterWithNewConfig(clusterOpts)
	test.T.Logf("Waiting for EKS-A controller to reconcile clusters")
	time.Sleep(2 * time.Minute) // Time for new eks-a controller to kick in and potentially trigger rolling upgrade
	for key, workloadCluster := range test.WorkloadClusters {
		test.T.Logf("Capturing CAPI machines for cluster %v", workloadCluster)
		postUpgradeMachines := test.ManagementCluster.GetCapiMachinesForCluster(workloadCluster.ClusterName)
		if anyMachinesChanged(preUpgradeMachines[key], postUpgradeMachines) {
			test.T.Fatalf("Found CAPI machines of workload cluster were changed after upgrading management cluster")
		}
	}
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.DeleteCluster()
	})
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

func anyMachinesChanged(machineMap1 map[string]types.Machine, machineMap2 map[string]types.Machine) bool {
	if len(machineMap1) != len(machineMap2) {
		return true
	}
	for machineName := range machineMap1 {
		if _, found := machineMap2[machineName]; !found {
			return true
		}
	}
	return false
}
