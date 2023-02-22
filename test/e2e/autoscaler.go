//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runAutoscalerWitMetricsServerSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(e *framework.ClusterE2ETest) {
		autoscalerName := "cluster-autoscaler"
		metricServerName := "metrics-server"
		targetNamespace := "eksa-packages"
		test.InstallAutoScalerWithMetricServer(targetNamespace)
		test.CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace, withMgmtCluster(test))
	})
}
