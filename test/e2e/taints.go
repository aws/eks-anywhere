//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTaintsUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	test.StopIfFailed()
	test.DeleteCluster()
}
