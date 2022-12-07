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
	adotPackageName     = "adot"
	adotPackagePrefix   = "generated"
	adotTargetNamespace = "observability"
)

func TestCPackagesAdotDockerUbuntuKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCPackagesAdotVSphereKubernetes122BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCPackagesAdotVSphereKubernetes123UbuntuUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func runCuratedPackagesAdotInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.CreateNamespace(adotTargetNamespace)
		test.InstallCuratedPackage("adot", adotPackagePrefix+"-"+adotPackageName,
			kubeconfig.FromClusterName(test.ClusterName), adotTargetNamespace,
			"--set mode=deployment")
		test.VerifyAdotPackageInstalled(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
	})
}

func runCuratedPackagesAdotInstallUpdateFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.CreateNamespace(adotTargetNamespace)
		test.InstallCuratedPackage("adot", adotPackagePrefix+"-"+adotPackageName,
			kubeconfig.FromClusterName(test.ClusterName), adotTargetNamespace,
			"--set mode=deployment")
		test.VerifyAdotPackageInstalled(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
		test.VerifyAdotPackageDeploymentUpdated(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
		test.VerifyAdotPackageDaemonSetUpdated(adotPackagePrefix+"-"+adotPackageName, adotTargetNamespace)
	})
}
