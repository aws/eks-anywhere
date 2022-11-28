//go:build e2e
// +build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
}

func runCuratedPackageEmissaryIngressTest(test *framework.ClusterE2ETest) {
	test.TestEmissaryPackageRouting(emissaryPackagePrefix+"-"+emissaryPackageName, withMgmtCluster(test))
}

func runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementCluster()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfig()
		e.CreateCluster()
		e.VerifyPackageControllerNotInstalled()
		test.ManagementCluster.SetPackageBundleActive()
		packageFile := e.BuildPackageConfigFile(emissaryPackageName, emissaryPackagePrefix, EksaPackagesNamespace)
		test.ManagementCluster.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ManagementCluster.ClusterName))
		e.VerifyEmissaryPackageInstalled(emissaryPackagePrefix+"-"+emissaryPackageName, withMgmtCluster(test.ManagementCluster))
		e.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackageEmissaryInstall(test)
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}

func runCuratedPackageEmissaryInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackageEmissaryInstall)
}

func runCuratedPackageEmissaryInstallAdvancedFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackageEmissaryInstall)
	test.WithCluster(runCuratedPackageEmissaryIngressTest)
}

func TestCPackagesEmissaryDockerUbuntuKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallAdvancedFlow(test)
}

func TestCPackagesEmissaryVSphereKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCPackagesEmissaryVSphereKubernetes122BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCPackagesEmissaryCloudStackRedhatKubernetes121WorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121()) // Only 1.20/1.21 is supported at the moment for Redhat.
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesEmissaryCloudStackRedhatKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()), // Only 1.20/1.21 is supported at the moment for Redhat.
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCPackagesEmissaryVSphereKubernetes122UbuntuWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesEmissaryTinkerbellUbuntuKubernetes122SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube122),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesEmissaryTinkerbellBottleRocketKubernetes122SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube122),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesEmissaryVSphereKubernetes122BottleRocketWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}
