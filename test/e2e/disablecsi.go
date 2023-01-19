//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runVSphereDisableCSIUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, provider *framework.VSphere) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateVSphereCSI(true)
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{provider.WithProviderUpgrade(api.WithDisableCSI(true))})
	test.DeleteVSphereCSI()
	test.ValidateCluster(updateVersion)
	test.ValidateVSphereCSI(false)
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{provider.WithProviderUpgrade(api.WithDisableCSI(false))})
	test.ValidateCluster(updateVersion)
	test.ValidateVSphereCSI(true)
	test.StopIfFailed()
	test.DeleteCluster()
}
