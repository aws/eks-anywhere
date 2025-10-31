//go:build e2e
// +build e2e

package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAWSIamAuthFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeFlowWithAWSIamAuth(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateAWSIamAuth()
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runTinkerbellAWSIamAuthFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func runAWSIamAuthFlowWorkload(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		licenseToken := framework.GetLicenseToken2()
		w.GenerateClusterConfigWithLicenseToken(licenseToken)
		w.CreateCluster()
		w.ValidateAWSIamAuth()
		w.StopIfFailed()
		w.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runUpgradeFlowAddAWSIamAuth(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateCluster(updateVersion)
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{
		framework.WithAWSIam(),
		framework.WithClusterUpgrade(api.WithAWSIamIdentityProviderRef("eksa-test")),
	})
	test.ValidateCluster(updateVersion)
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}
