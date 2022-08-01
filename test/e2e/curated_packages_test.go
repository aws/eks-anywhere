//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	eksAnywherePackagesHelmChartName = "eks-anywhere-packages"
	eksAnywherePackagesHelmUri       = "oci://public.ecr.aws/l0g8r8j6/eks-anywhere-packages"
	eksAnywherePackagesHelmVersion   = "0.1.14-eks-a-v0.0.0-dev-build.3481"
	eksAnywherePackagesBundleUri     = "oci://public.ecr.aws/l0g8r8j6/eks-anywhere-packages-bundles:v1-21-latest"

	eksaPackageControllerHelmChartName = "eks-anywhere-packages"
	eksaPackageControllerHelmURI       = "oci://public.ecr.aws/eks-anywhere/eks-anywhere-packages"
	eksaPackageControllerHelmVersion   = "0.1.10-eks-a-10"
	eksaPackageBundleBaseURI           = "oci://public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles"
)

var (
	eksaPackageControllerHelmValues = []string{}
	eksAnywherePackagesHelmValues   = []string{"sourceRegistry=public.ecr.aws/l0g8r8j6"}
)

// packageBundleURI uses a KubernetesVersion argument to complete a package
// bundle URI by adding the approprate tag.
func packageBundleURI(version v1alpha1.KubernetesVersion) string {
	tag := "v" + strings.Replace(string(version), ".", "-", 1) + "-latest"
	return fmt.Sprintf("%s:%s", eksaPackageBundleBaseURI, tag)
}

func runCuratedPackageInstallSimpleFlow(test *framework.ClusterE2ETest) {
	os.Setenv(features.CuratedPackagesEnvVar, "false")
	test.WithCluster(func(e *framework.ClusterE2ETest) {
		os.Setenv(features.CuratedPackagesEnvVar, "true")
		defer os.Setenv(features.CuratedPackagesEnvVar, "false")

		e.InstallCuratedPackagesController()
		packageName := "hello-eks-anywhere"
		packagePrefix := "test"
		e.InstallCuratedPackage(packageName, packagePrefix)
		e.VerifyHelloPackageInstalled(packagePrefix + "-" + packageName)
	})
}

// There are many tests here, each covers a different combination described in
// the matrix found in
// https://github.com/aws/eks-anywhere-packages/issues/96. They're each named
// according to the columns of that matrix, that is,
// "TestCPackages<Provider><OS><K8s ver>SimpleFlow". Better organization,
// whether via test suites, testing tables, or other functionality is welcome,
// but this is a simple solution for now, without having to make any major
// decisions about test packages or methodologies, right now.

func TestCPackagesDockerUbuntuKubernetes120SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube120),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesDockerUbuntuKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesDockerUbuntuKubernetes122SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes120SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube120),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			eksaPackageControllerHelmChartName, eksaPackageControllerHelmURI,
			eksaPackageControllerHelmVersion, eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}
