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

const (
	packageName            = "prometheus"
	packagePrefix          = "generated"
	packageTargetNamespace = "observability"
)

func TestCPackagesPrometheusDockerUbuntuKubernetes124SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCPackagesPrometheusVSphereKubernetes123BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCPackagesPrometheusVSphereKubernetes122UbuntuSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func runCuratedPackagesPrometheusInstallSimpleFlow(test *framework.ClusterE2ETest) {
	packageFullName := packagePrefix + "-" + packageName
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.CreateNamespace(packageTargetNamespace)
		test.InstallCuratedPackage(packageName, packageFullName,
			kubeconfig.FromClusterName(test.ClusterName), packageTargetNamespace)
		test.VerifyPrometheusPackageInstalled(packageFullName, packageTargetNamespace)
		test.VerifyPrometheusNodeExporterStates(packageFullName, packageTargetNamespace)
		test.VerifyPrometheusPrometheusServerStates(packageFullName, packageTargetNamespace)
	})
}
