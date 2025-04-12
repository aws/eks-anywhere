//go:build e2e
// +build e2e

package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/test/framework"
)

const (
	cmPackageName = "cert-manager"
)

func runCertManagerRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	licenseToken := framework.GetLicenseToken2()
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfigWithLicenseToken(licenseToken)
		e.ApplyClusterManifest()
		e.WaitForKubeconfig()
		e.ValidateClusterState()
		e.VerifyPackageControllerNotInstalled()
		test.ManagementCluster.SetPackageBundleActive()
		packageName := "cert-manager"
		packagePrefix := "test"
		test.ManagementCluster.InstallCertManagerPackageWithAwsCredentials(packagePrefix, packageName, EksaPackagesNamespace, e.ClusterName)
		e.VerifyCertManagerPackageInstalled(packagePrefix, EksaPackagesNamespace, cmPackageName, withCluster(test.ManagementCluster))
		e.CleanupCerts(withCluster(test.ManagementCluster))
		e.DeleteClusterWithKubectl()
		e.ValidateClusterDelete()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}
