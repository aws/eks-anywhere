//go:build e2e && (nutanix || all_providers)
// +build e2e
// +build nutanix all_providers

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

// Curated packages
func TestNutanixKubernetes122CuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes123CuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes124CuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes125CuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes126CuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes123CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes122CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes123CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes124CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes125CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes126CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes122UbuntuCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestNutanixKubernetes122CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
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

func TestNutanixKubernetes124CuratedPackagesClusterAutoscalerUbuntuSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runAutoscalerWitMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes125CuratedPackagesClusterAutoscalerUbuntuSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runAutoscalerWitMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes126CuratedPackagesClusterAutoscalerUbuntuSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runAutoscalerWitMetricsServerSimpleFlow(test)
}

// Simpleflow
func TestNutanixKubernetes122SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes123SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes124SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes125SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes126SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes122SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu122NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes123SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes124SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes125SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes126SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleFlow(test)
}

// Upgrade
func TestNutanixKubernetes122To123UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate123Var()),
	)
}

func TestNutanixKubernetes123To124UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate124Var()),
	)
}

func TestNutanixKubernetes124To125UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate125Var()),
	)
}

func TestNutanixKubernetes125To126UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu125Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate126Var()),
	)
}

func TestNutanixKubernetes122UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes123UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes124UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes125UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu125Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes126UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu126Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar("features.NutanixProviderEnvVar", "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar("features.NutanixProviderEnvVar", "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar("features.NutanixProviderEnvVar", "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes125UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu125Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes126UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu126Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes122UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestNutanixKubernetes123UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestNutanixKubernetes124UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes125UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu125Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes126UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu126Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes125UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu125Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes126UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu126Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// OIDC Tests
func TestNutanixKubernetes122OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes123OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes124OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes125OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes126OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s126SupportEnvVar, "true"),
	)
	runOIDCFlow(test)
}
