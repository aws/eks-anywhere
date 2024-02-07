//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	adotTargetNamespace = "observability"
	adotPackageName     = "adot"
	adotPackagePrefix   = "generated"
)

func runCuratedPackagesAdotInstall(test *framework.ClusterE2ETest) {
	test.SetPackageBundleActive()
	test.CreateNamespace(adotTargetNamespace)
	test.InstallCuratedPackage(adotPackageName, adotPackagePrefix+"-"+adotPackageName,
		kubeconfig.FromClusterName(test.ClusterName),
		"--set mode=deployment")
	test.VerifyAdotPackageInstalled(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
}

func runCuratedPackagesAdotInstallWithUpdate(test *framework.ClusterE2ETest) {
	test.SetPackageBundleActive()
	test.CreateNamespace(adotTargetNamespace)
	test.InstallCuratedPackage(adotPackageName, adotPackagePrefix+"-"+adotPackageName,
		kubeconfig.FromClusterName(test.ClusterName),
		"--set mode=deployment")
	test.VerifyAdotPackageInstalled(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
	test.VerifyAdotPackageDeploymentUpdated(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
	test.VerifyAdotPackageDaemonSetUpdated(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
}

func runCuratedPackagesAdotInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesAdotInstall)
}

func runCuratedPackagesAdotInstallUpdateFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesAdotInstallWithUpdate)
}

func runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackagesAdotInstall(test)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}
