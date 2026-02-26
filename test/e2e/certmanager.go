//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
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
		if err := WaitForPackageNamespace(test.ManagementCluster, context.Background(),
			kubeconfig.FromClusterName(test.ManagementCluster.ClusterName),
			e.ClusterName, 5*time.Minute); err != nil {
			e.T.Fatalf("package namespace not created on management cluster: %v", err)
		}
		packageName := "cert-manager"
		packagePrefix := "test"
		test.ManagementCluster.InstallCertManagerPackageWithAwsCredentials(packagePrefix, packageName, EksaPackagesNamespace, e.ClusterName)
		// Ensure cleanup happens even if the test fails
		e.T.Cleanup(func() {
			if err := e.CleanupCerts(withCluster(test.ManagementCluster)); err != nil {
				e.T.Logf("Warning: Failed to cleanup certificates: %v", err)
			}
		})
		e.VerifyCertManagerPackageInstalled(packagePrefix, EksaPackagesNamespace, cmPackageName, withCluster(test.ManagementCluster))
		e.DeleteClusterWithKubectl()
		e.ValidateClusterDelete()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}
