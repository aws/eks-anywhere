//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runMultiTemplatesSimpleFlow(test *framework.ClusterE2ETest) {
	test.CreateCluster()
	test.ValidateVsphereMachine(
		framework.ControlPlaneMachineLabel,
		test.ClusterConfig.VSphereMachineConfigs[test.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name],
		framework.ValidateMachineTemplate)
	test.ValidateVsphereMachine(
		framework.EtcdMachineLabel,
		test.ClusterConfig.VSphereMachineConfigs[test.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name],
		framework.ValidateMachineTemplate)
	test.ValidateWorkerNodeVsphereMachine(framework.ValidateMachineTemplate)
	test.DeleteCluster()
}
