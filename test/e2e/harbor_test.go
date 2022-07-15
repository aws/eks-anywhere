//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/test/framework"
)

func TestHarborInstallSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithPackageConfig(t, eksAnywherePackagesBundleUri,
			eksAnywherePackagesHelmChartName, eksAnywherePackagesHelmUri,
			eksAnywherePackagesHelmVersion, eksAnywherePackagesHelmValues),
	)
	runHarborInstallSimpleFlow(test) // other args as necessary
}

func runHarborInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.InstallCuratedPackagesController()
		if _, ok := test.Provider.(*framework.Docker); ok {
			test.InstallLocalStorageProvisioner()
		}
		packagePrefix := "test"
		test.InstallCuratedPackage("harbor", packagePrefix,
			"--set secretKey=use-a-secret-key",
			"--set expose.tls.certSource=auto",
			"--set expose.tls.auto.commonName=localhost",
			"--set persistence.persistentVolumeClaim.registry.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.jobservice.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.database.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.redis.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.trivy.storageClass=local-path",
		)
		test.VerifyHarborPackageInstalled(packagePrefix)
	})
}
