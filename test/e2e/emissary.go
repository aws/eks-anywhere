//go:build e2e
// +build e2e

package e2e

import (
	"time"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	emissaryPackageName   = "emissary"
	emissaryPackagePrefix = "test"
)

func runCuratedPackageEmissaryInstall(test *framework.ClusterE2ETest) {
	test.SetPackageBundleActive()
	packageFile := test.BuildPackageConfigFile(emissaryPackageName, emissaryPackagePrefix, EksaPackagesNamespace)
	test.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ClusterName))
	test.VerifyEmissaryPackageInstalled(emissaryPackagePrefix+"-"+emissaryPackageName, withMgmtCluster(test))
	if test.Provider.Name() == constants.DockerProviderName {
		test.TestEmissaryPackageRouting(emissaryPackagePrefix+"-"+emissaryPackageName, "hello", withMgmtCluster(test))
	}
}

func runCuratedPackageEmissaryInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackageEmissaryInstall)
}

func runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfig()
		e.ApplyClusterManifest()
		e.WaitForKubeconfig()
		e.ValidateClusterState()
		e.VerifyPackageControllerNotInstalled()
		test.ManagementCluster.SetPackageBundleActive()
		packageFile := e.BuildPackageConfigFile(emissaryPackageName, emissaryPackagePrefix, EksaPackagesNamespace)
		test.ManagementCluster.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ManagementCluster.ClusterName))
		e.VerifyEmissaryPackageInstalled(emissaryPackagePrefix+"-"+emissaryPackageName, withMgmtCluster(test.ManagementCluster))
		e.DeleteClusterWithKubectl()
		e.ValidateClusterDelete()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackageEmissaryInstall(test)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}
