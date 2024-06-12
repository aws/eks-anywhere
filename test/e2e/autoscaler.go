//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runAutoscalerWithMetricsServerSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(e *framework.ClusterE2ETest) {
		autoscalerName := "cluster-autoscaler"
		metricServerName := "metrics-server"
		targetNamespace := "eksa-packages"
		test.InstallAutoScalerWithMetricServer(targetNamespace)
		test.CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace, withCluster(test))
	})
}

func runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	autoscalerName := "cluster-autoscaler"
	metricServerName := "metrics-server"
	targetNamespace := "eksa-packages"
	test.InstallAutoScalerWithMetricServer(targetNamespace)
	test.CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace, withCluster(test))
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func runAutoscalerUpgradeFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfig()
		e.CreateCluster()
		autoscalerName := "cluster-autoscaler"
		targetNamespace := "eksa-system"
		mgmtCluster := withCluster(test.ManagementCluster)
		workloadCluster := withCluster(e.ClusterE2ETest)
		test.ManagementCluster.InstallAutoScaler(e.ClusterName, targetNamespace)
		test.ManagementCluster.VerifyAutoScalerPackageInstalled(autoscalerName, targetNamespace, mgmtCluster)
		e.T.Log("Cluster Autoscaler ready")
		e.DeployTestWorkload(workloadCluster)
		test.ManagementCluster.RestartClusterAutoscaler(targetNamespace)
		e.VerifyWorkerNodesScaleUp(mgmtCluster)
		e.DeleteCluster()
	})
	test.DeleteManagementCluster()
}
