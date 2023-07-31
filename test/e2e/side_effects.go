//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/test/framework"
)

type eksaPackagedBinary interface {
	framework.PackagedBinary
	// Version returns the EKS-A version.
	Version() string
}

// runFlowUpgradeManagementClusterCheckForSideEffects creates management and workload cluster
// with a specific eks-a version then upgrades the management cluster with another CLI version
// and checks that this doesn't cause any side effects (machine rollout) in the workload clusters.
func runFlowUpgradeManagementClusterCheckForSideEffects(test *framework.MulticlusterE2ETest, currentEKSA, newEKSA eksaPackagedBinary, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.T.Logf("Creating management cluster with EKS-A version %s", currentEKSA.Version())
	test.CreateManagementCluster(framework.ExecuteWithBinary(currentEKSA))

	test.T.Logf("Creating workload clusters with EKS-A version %s", currentEKSA.Version())
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.CreateCluster(framework.ExecuteWithBinary(currentEKSA))
	})

	waitForWorkloadCustersMachineDeploymentsReady(test.ManagementCluster, test.WorkloadClusters)
	preUpgradeWorkloadClustersState := buildWorkloadClustersWithMachines(test.ManagementCluster, test.WorkloadClusters)

	test.T.Log("Machines state for workload clusters after first creation")
	printStateOfMachines(test.ManagementCluster.ClusterConfig.Cluster, preUpgradeWorkloadClustersState)

	test.T.Logf("Upgrading management cluster with EKS-A version %s", newEKSA.Version())
	test.ManagementCluster.UpgradeClusterWithNewConfig(clusterOpts, framework.ExecuteWithBinary(newEKSA))

	checker := machineSideEffectChecker{
		tb:                 test.T,
		checkDuration:      10 * time.Minute,
		waitInBetweenTries: 20 * time.Second,
	}
	if changed, reason := checker.haveMachinesChanged(test.ManagementCluster, preUpgradeWorkloadClustersState); changed {
		test.T.Fatalf("The new controller has had cascading changes: %s", reason)
	}

	test.T.Log("Your management cluster upgrade didn't create or delete machines in the workload clusters. Congrats!")

	if os.Getenv("MANUAL_TEST_PAUSE") == "true" {
		test.T.Log("Press enter to continue with the cleanup after you are done with your manual investigation: ")
		fmt.Scanln()
	}

	test.T.Log("Machines state for workload clusters after management cluster upgrade")
	printStateOfMachines(test.ManagementCluster.ClusterConfig.Cluster, buildWorkloadClustersWithMachines(test.ManagementCluster, test.WorkloadClusters))

	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.DeleteCluster()
	})
	test.DeleteManagementCluster()
}

type clusterWithMachines struct {
	name     string
	machines clusterMachines
}

type clusterMachines map[string]types.Machine

func anyMachinesChanged(original, current clusterMachines) (changed bool, reason string) {
	if len(original) != len(current) {
		return true, fmt.Sprintf("Different number of machines: before %d, after %d", len(original), len(current))
	}
	for machineName := range original {
		if m, found := current[machineName]; !found {
			return true, fmt.Sprintf("Machine %s not present in current cluster", m.Metadata.Name)
		}
	}
	return false, ""
}

func printStateOfMachines(managementCluster *anywherev1.Cluster, clusters []clusterWithMachines) {
	fmt.Println(managementCluster.Name)
	for _, cluster := range clusters {
		fmt.Printf("├── %s\n", cluster.name)
		for _, m := range cluster.machines {
			fmt.Printf("│   ├──  %s\n", m.Metadata.Name)

			fmt.Print("│   │    ├──  Labels\n")
			for k, v := range m.Metadata.Labels {
				fmt.Printf("│   │    │    ├──  %s: %s\n", k, v)
			}

			fmt.Print("│   │    ├──  Conditions\n")
			for _, c := range m.Status.Conditions {
				fmt.Printf("│   │    │    ├──  %s: %s\n", c.Type, c.Status)
			}
		}
	}
}

func buildClusterWithMachines(managementCluster *framework.ClusterE2ETest, clusterName string) clusterWithMachines {
	managementCluster.T.Logf("Reading CAPI machines for cluster %s", clusterName)
	return clusterWithMachines{
		name:     clusterName,
		machines: managementCluster.GetCapiMachinesForCluster(clusterName),
	}
}

func buildWorkloadClustersWithMachines(managementCluster *framework.ClusterE2ETest, workloadClusters framework.WorkloadClusters) []clusterWithMachines {
	cm := make([]clusterWithMachines, 0, len(workloadClusters))
	for _, w := range workloadClusters {
		cm = append(cm, buildClusterWithMachines(managementCluster, w.ClusterName))
	}

	return cm
}

func waitForMachineDeploymentReady(managementCluster *framework.ClusterE2ETest, cluster *anywherev1.Cluster, workerNodeGroup anywherev1.WorkerNodeGroupConfiguration) {
	machineDeploymetName := clusterapi.MachineDeploymentName(cluster, workerNodeGroup)
	managementCluster.WaitForMachineDeploymentReady(machineDeploymetName)
}

func waitForClusterMachineDeploymentsReady(managementCluster *framework.ClusterE2ETest, cluster *anywherev1.Cluster) {
	for _, w := range cluster.Spec.WorkerNodeGroupConfigurations {
		waitForMachineDeploymentReady(managementCluster, cluster, w)
	}
}

func waitForWorkloadCustersMachineDeploymentsReady(managementCluster *framework.ClusterE2ETest, workloadClusters framework.WorkloadClusters) {
	for _, w := range workloadClusters {
		waitForClusterMachineDeploymentsReady(managementCluster, w.ClusterConfig.Cluster)
	}
}

type machineSideEffectChecker struct {
	tb                                testing.TB
	checkDuration, waitInBetweenTries time.Duration
}

func (m machineSideEffectChecker) haveMachinesChanged(managementCluster *framework.ClusterE2ETest, preUpgradeWorkloadClustersState []clusterWithMachines) (changed bool, changeReason string) {
	m.tb.Logf("Checking for changes in machines for %s", m.checkDuration)
	start := time.Now()
retry:
	for now := time.Now(); now.Sub(start) <= m.checkDuration; now = time.Now() {
		for _, workloadCluster := range preUpgradeWorkloadClustersState {
			m.tb.Logf("Reading CAPI machines for cluster %s", workloadCluster.name)
			postUpgradeMachines, err := managementCluster.CapiMachinesForCluster(workloadCluster.name)
			if err != nil {
				m.tb.Logf("Failed getting machines for cluster %s, omitting error in case it's transient: %s", workloadCluster.name, err)
				continue
			}

			if changed, changeReason = anyMachinesChanged(workloadCluster.machines, postUpgradeMachines); changed {
				changeReason = fmt.Sprintf("cluster %s has chaged: %s", workloadCluster.name, changeReason)
				break retry
			}

			m.tb.Logf("Machines for workload cluster %s haven't changed after upgrading the management cluster", workloadCluster.name)
		}
		m.tb.Logf("Waiting for %s until next check", m.waitInBetweenTries)
		time.Sleep(m.waitInBetweenTries)
	}

	return changed, changeReason
}

// eksaLocalPackagedBinary implements eksaPackagedBinary using the local eks-a binary
// being tested by this suite.
type eksaLocalPackagedBinary struct {
	path, version string
}

func (b eksaLocalPackagedBinary) BinaryPath() (string, error) {
	return b.path, nil
}

func (b eksaLocalPackagedBinary) Version() string {
	return b.version
}

func newEKSAPackagedBinaryForLocalBinary(tb testing.TB) eksaPackagedBinary {
	tb.Helper()
	version, err := framework.EKSAVersionForTestBinary()
	if err != nil {
		tb.Fatal(err)
	}

	path, err := framework.DefaultLocalEKSABinaryPath()
	if err != nil {
		tb.Fatal(err)
	}

	return eksaLocalPackagedBinary{
		path:    path,
		version: version,
	}
}

func runTestManagementClusterUpgradeSideEffects(t *testing.T, provider framework.Provider, os framework.OS, kubeVersion anywherev1.KubernetesVersion, configFillers ...api.ClusterConfigFiller) {
	latestRelease := latestMinorRelease(t)

	managementCluster := framework.NewClusterE2ETest(t, provider, framework.PersistentCluster())
	managementCluster.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	managementCluster.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithEtcdCountIfExternal(1),
		),
		provider.WithKubeVersionAndOS(kubeVersion, os, latestRelease),
		api.JoinClusterConfigFillers(
			configFillers...,
		),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	workloadCluster := framework.NewClusterE2ETest(t, provider,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	workloadCluster.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	workloadCluster.UpdateClusterConfig(api.ClusterToConfigFiller(
		api.WithManagementCluster(managementCluster.ClusterName),
		api.WithControlPlaneCount(2),
		api.WithControlPlaneLabel("cluster.x-k8s.io/failure-domain", "ds.meta_data.failuredomain"),
		api.RemoveAllWorkerNodeGroups(),
		api.WithWorkerNodeGroup("workers-0",
			api.WithCount(3),
			api.WithLabel("cluster.x-k8s.io/failure-domain", "ds.meta_data.failuredomain"),
		),
		api.WithEtcdCountIfExternal(3),
		api.WithCiliumPolicyEnforcementMode(anywherev1.CiliumPolicyModeAlways)),
		provider.WithNewWorkerNodeGroup("workers-0",
			framework.WithWorkerNodeGroup("workers-0",
				api.WithCount(2),
				api.WithLabel("cluster.x-k8s.io/failure-domain", "ds.meta_data.failuredomain"))),
		framework.WithOIDCClusterConfig(t),
		provider.WithKubeVersionAndOS(kubeVersion, os, latestRelease),
		api.JoinClusterConfigFillers(
			configFillers...,
		),
	)
	test.WithWorkloadClusters(workloadCluster)

	runFlowUpgradeManagementClusterCheckForSideEffects(test,
		framework.NewEKSAReleasePackagedBinary(latestRelease),
		newEKSAPackagedBinaryForLocalBinary(t),
		framework.WithUpgradeClusterConfig(provider.WithKubeVersionAndOS(kubeVersion, os, nil)),
	)
}
