package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	cmPackageName     = "cert-manager"
	cmPackagePrefix   = "generated"
	cmTargetNamespace = "cert-manager"
)

func runCertManagerRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfig()
		e.CreateCluster()
		e.VerifyPackageControllerNotInstalled()
		test.ManagementCluster.SetPackageBundleActive()
		packageName := "cert-manager"
		packagePrefix := "test"
		packageFile := e.BuildPackageConfigFile(packageName, packagePrefix, EksaPackagesNamespace)
		test.ManagementCluster.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ManagementCluster.ClusterName))
		e.VerifyCertManagerPackageInstalled(cmPackagePrefix, cmTargetNamespace, cmPackageName, withMgmtCluster(test.ManagementCluster))
		e.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}
