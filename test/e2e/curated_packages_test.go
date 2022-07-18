//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	eksAnywherePackagesHelmChartName = "eks-anywhere-packages"
	eksAnywherePackagesHelmUri       = "oci://public.ecr.aws/l0g8r8j6/eks-anywhere-packages"
	eksAnywherePackagesHelmVersion   = "0.1.6-eks-a-v0.0.0-dev-build.2404"

	eksaPackageControllerHelmChartName = "eks-anywhere-packages"
	eksaPackageControllerHelmURI       = "oci://public.ecr.aws/eks-anywhere/eks-anywhere-packages"
	eksaPackageControllerHelmVersion   = "0.1.10-eks-a-10"
	eksaPackageBundleURI               = "oci://public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles:v1-21-latest"
)

var eksaPackageControllerHelmValues = []string{}

func TestPackagesInstallSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithPackageConfig(t, eksaPackageBundleURI, eksaPackageControllerHelmChartName,
			eksaPackageControllerHelmURI, eksaPackageControllerHelmVersion,
			eksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func runCuratedPackageInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		test.InstallCuratedPackagesController()
		packageName := "hello-eks-anywhere"
		packagePrefix := "test"
		test.InstallCuratedPackage(packageName, packagePrefix)
		test.VerifyHelloPackageInstalled(packagePrefix + "-" + "hello-eks-anywhere")
	})
}

var eksAnywherePackagesHelmValues = []string{"sourceRegistry=public.ecr.aws/l0g8r8j6"}

func TestVSphereKubernetes122BottleRocketPackagesInstallSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithHelmInstallConfig(t, eksAnywherePackagesHelmChartName, eksAnywherePackagesHelmUri, eksAnywherePackagesHelmVersion, eksAnywherePackagesHelmValues),
	)
	runHelmInstallSimpleFlow(test)
}
