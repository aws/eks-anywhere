//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func TestCPackagesClusterAutoscalerCloudStackRedhatKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121), api.WithWorkerNodeAutoScalingConfig(1, 2)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-autoscaler-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runAutoscalerCloudStackSimpleFlow(test)
}

func runAutoscalerCloudStackSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(e *framework.ClusterE2ETest) {
		autoscalerName := "cluster-autoscaler"
		metricServerName := "metrics-server"
		installNs := "eksa-packages"
		test.InstallAutoScalerWithMetricServer()
		test.CombinedAutoscalerMetricServerTest(autoscalerName, metricServerName, withMgmtCluster(test))
	})

}
