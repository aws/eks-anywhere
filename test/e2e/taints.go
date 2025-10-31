//go:build e2e
// +build e2e

package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/test/framework"
)

func runTaintsUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	// Add 1-minute wait for vSphere upgrade tests
	// where it fails during upgrade preflight validation
	// when packages controller installs credentials provider package on the node
	if test.Provider.Name() == constants.VSphereProviderName {
		test.T.Log("Waiting 2 minute before starting vSphere upgrade...")
		time.Sleep(1 * time.Minute)
	}
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints)
	test.StopIfFailed()
	test.DeleteCluster()
}
