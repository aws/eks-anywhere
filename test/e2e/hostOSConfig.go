//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runNTPFlow(test *framework.ClusterE2ETest, osFamily v1alpha1.OSFamily) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateNTPConfig(osFamily)
	test.DeleteCluster()
}

func runBottlerocketConfigurationFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateBottlerocketKubernetesSettings()
	test.DeleteCluster()
}
