//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

const (
		autoscalerName = "cluster-autoscaler"
		metricServerName = "metrics-server"
		targetNamespace = "eksa-packages"
)

func runAutoscalerWithMetricsServerSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(e *framework.ClusterE2ETest) {
		test.InstallAutoScalerWithMetricServer(targetNamespace)
		test.CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace, withMgmtCluster(test))
	})
}

func runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.GenerateClusterConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	test.InstallAutoScalerWithMetricServer(targetNamespace)
	test.CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace, withMgmtCluster(test))
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}
