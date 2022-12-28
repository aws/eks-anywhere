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

func TestCPackagesPrometheusDockerKubernetes124SimpleFlow(t *testing.T) {
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

func TestCPackagesPrometheusVSphereKubernetes122UbuntuUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusUpdateFlow(test)
}

func TestCPackagesPrometheusCloudStackRedhatKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCPackagesPrometheusNutanixKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCPackagesPrometheusTinkerbellUbuntuKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube122),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestCPackagesPrometheusTinkerbellBottleRocketKubernetes123SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube123),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func runCuratedPackagesPrometheusInstall(test *framework.ClusterE2ETest) {
	packageFullName := packagePrefix + "-" + packageName
	test.InstallLocalStorageProvisioner()
	test.CreateNamespace(packageTargetNamespace)
	test.SetPackageBundleActive()
	test.InstallCuratedPackage(packageName, packageFullName,
		kubeconfig.FromClusterName(test.ClusterName), packageTargetNamespace,
		"--set server.persistentVolume.storageClass=local-path")
	test.VerifyPrometheusPackageInstalled(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusNodeExporterStates(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, packageTargetNamespace, "deployment")
}

func runCuratedPackagesPrometheusUpdate(test *framework.ClusterE2ETest) {
	packageFullName := packagePrefix + "-" + packageName

	test.InstallLocalStorageProvisioner()
	test.CreateNamespace(packageTargetNamespace)
	test.SetPackageBundleActive()
	test.InstallCuratedPackage(packageName, packageFullName,
		kubeconfig.FromClusterName(test.ClusterName), packageTargetNamespace,
		"--set server.persistentVolume.storageClass=local-path")

	test.ApplyPrometheusPackageServerStatefulSetFile(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusPackageInstalled(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, packageTargetNamespace, "statefulset")
	test.VerifyPrometheusNodeExporterStates(packageFullName, packageTargetNamespace)

	test.ApplyPrometheusPackageServerDeploymentFile(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusPackageInstalled(packageFullName, packageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, packageTargetNamespace, "deployment")
	test.VerifyPrometheusNodeExporterStates(packageFullName, packageTargetNamespace)
}

func runCuratedPackagesPrometheusInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesPrometheusInstall)
}

func runCuratedPackagesPrometheusUpdateFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesPrometheusUpdate)
}

func runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackagesPrometheusInstall(test)
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}
