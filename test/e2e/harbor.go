//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

// installHarborWithLocalStorage installs the Harbor package with a self-signed certificate and
// local-path volumes. The local storage provisioner and namespace must already exist.
func installHarborWithLocalStorage(test *framework.ClusterE2ETest, packagePrefix, namespace string) {
	test.InstallCuratedPackage("harbor", packagePrefix, kubeconfig.FromClusterName(test.ClusterName),
		"--set secretKey=use-a-secret-key",
		"--set expose.tls.certSource=auto",
		"--set expose.tls.auto.commonName=localhost",
		"--set persistence.persistentVolumeClaim.registry.storageClass=local-path",
		"--set persistence.persistentVolumeClaim.jobservice.jobLog.storageClass=local-path",
		"--set persistence.persistentVolumeClaim.database.storageClass=local-path",
		"--set persistence.persistentVolumeClaim.redis.storageClass=local-path",
		"--set persistence.persistentVolumeClaim.trivy.storageClass=local-path",
	)
}

func runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.ValidatePackageBundleControllerRegistry()
		test.InstallLocalStorageProvisioner()

		packagePrefix := "test"
		installNs := "harbor"
		test.CreateNamespace(installNs)
		installHarborWithLocalStorage(test, packagePrefix, installNs)
		test.VerifyHarborPackageInstalled(packagePrefix, installNs)
	})
}

func runCuratedPackageHarborInstall(test *framework.ClusterE2ETest) {
	test.InstallLocalStorageProvisioner()
	packagePrefix := "test"
	installNs := "harbor"
	test.CreateNamespace(installNs)
	installHarborWithLocalStorage(test, packagePrefix, installNs)
	test.VerifyHarborPackageInstalled(packagePrefix, installNs)
}

func runCuratedPackageHarborInstallTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackageHarborInstall(test)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}
