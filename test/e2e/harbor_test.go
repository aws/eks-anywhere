//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

func TestCuratedPackagesHarborInstallSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithPackageConfig(t, EksaPackageBundleURI,
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runHarborInstallSimpleFlowLocalStorageProvisioner(test) // other args as necessary
}

func TestCuratedPackagesHarborNutanixKubernetes123SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func runHarborInstallSimpleFlowLocalStorageProvisioner(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.InstallLocalStorageProvisioner()

		packagePrefix := "test"
		installNs := "harbor"
		test.CreateNamespace(installNs)
		test.InstallCuratedPackage("harbor", packagePrefix, kubeconfig.FromClusterName(test.ClusterName),
			"--set secretKey=use-a-secret-key",
			"--set expose.tls.certSource=auto",
			"--set expose.tls.auto.commonName=localhost",
			"--set persistence.persistentVolumeClaim.registry.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.jobservice.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.database.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.redis.storageClass=local-path",
			"--set persistence.persistentVolumeClaim.trivy.storageClass=local-path",
		)
		test.VerifyHarborPackageInstalled(packagePrefix, installNs)
	})
}
