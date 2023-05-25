//go:build e2e && (nutanix || all_providers)
// +build e2e
// +build nutanix all_providers

package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

// Curated packages
func kubeVersionNutanixOpt(version v1alpha1.KubernetesVersion) framework.NutanixOpt {
	switch version {
	case v1alpha1.Kube123:
		return framework.WithUbuntu123Nutanix()
	case v1alpha1.Kube124:
		return framework.WithUbuntu124Nutanix()
	case v1alpha1.Kube125:
		return framework.WithUbuntu125Nutanix()
	case v1alpha1.Kube126:
		return framework.WithUbuntu126Nutanix()
	case v1alpha1.Kube127:
		return framework.WithUbuntu127Nutanix()
	default:
		panic(fmt.Sprintf("unsupported version: %v", version))
	}
}

func TestNutanixCuratedPackagesSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageInstallSimpleFlow(test)
	}
}

func TestNutanixCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageEmissaryInstallSimpleFlow(test)
	}
}

func TestNutanixCuratedPackagesHarborSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
	}
}

func TestNutanixCuratedPackagesAdotUpdateFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesAdotInstallUpdateFlow(test)
	}
}

func TestNutanixCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runAutoscalerWithMetricsServerSimpleFlow(test)
	}
}

func TestNutanixCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewNutanix(t, kubeVersionNutanixOpt(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesPrometheusInstallSimpleFlow(test)
	}
}

// Simpleflow
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
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes127SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes123SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes124SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes125SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes126SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes127SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

// Upgrade
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
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate126Var()),
	)
}

func TestNutanixKubernetes126To127UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu126Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate127Var()),
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
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes127UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
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
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes127UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
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
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes127UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
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
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes127UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// OIDC Tests
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
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// IAMAuthenticator Tests
func TestNutanixKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes124AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes125AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu125Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes126AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu126Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}
