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

func TestCPackagesAdotDockerUbuntuKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runAdotInstallSimpleFlow(test) // other args as necessary
}

func runAdotInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		packageName := "generated-adot"
		targetNamespace := "observability"
		test.CreateNamespace(targetNamespace)
		test.InstallCuratedPackage("adot", packageName, kubeconfig.FromClusterName(test.ClusterName), targetNamespace,
			"--set mode=deployment",
		)
		test.VerifyAdotPackageInstalled(packageName, targetNamespace)
	})
}
