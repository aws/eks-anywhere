//go:build e2e && (nutanix || all_providers)
// +build e2e
// +build nutanix all_providers

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

// Curated Packages
func TestNutanixKubernetes127UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
func TestNutanixKubernetes127UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// ADOT
func TestNutanixKubernetes127UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

// Prometheus
func TestNutanixKubernetes127UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Cluster Autoscaler
func TestNutanixKubernetes127UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

// Harbor
func TestNutanixKubernetes127UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes128UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes129UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes130UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes131UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// Simple Flow
func TestNutanixKubernetes127UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes127RedHat8SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128RedHat8SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129RedHat8SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130RedHat8SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131RedHat8SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes127RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes127Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes127UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu131NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128RedHatSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat128NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129RedHatSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat129NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130RedHatSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat130NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131RedHatSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat131NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes128NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes129NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes130NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes131NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

// Upgrade
func TestNutanixKubernetes127To128UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Ubuntu128Template()),
	)
}

func TestNutanixKubernetes128To129StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestNutanixKubernetes129To130StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
}

func TestNutanixKubernetes130To131StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
}

func TestNutanixKubernetes128To129UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestNutanixKubernetes129To130UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
}

func TestNutanixKubernetes130To131UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
}

func TestNutanixKubernetes127to128RedHatUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.RedHat128Template()),
	)
}

func TestNutanixKubernetes128to129RedHatUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.RedHat129Template()),
	)
}

func TestNutanixKubernetes129to130RedHatUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.RedHat130Template()),
	)
}

func TestNutanixKubernetes130to131RedHatUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.RedHat131Template()),
	)
}

func TestNutanixKubernetes128to129StackedEtcdRedHat8Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.RedHat129Template()),
	)
}

func TestNutanixKubernetes127to128RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes127Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes128Template()),
	)
}

func TestNutanixKubernetes128to129RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes129Template()),
	)
}

func TestNutanixKubernetes129to130RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes130Template()),
	)
}

func TestNutanixKubernetes130to131RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes131Template()),
	)
}

func TestNutanixKubernetes128to129StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes129Template()),
	)
}

func TestNutanixKubernetes128UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes129UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes130UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes131UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
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
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
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
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes128UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar("features.NutanixProviderEnvVar", "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes129UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes130UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes131UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
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
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes128UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes129UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes130UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes131UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
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
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes128UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes129UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu129Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes130UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu130Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes131UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// OIDC
func TestNutanixKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes129OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes130OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes131OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// AWS IAM Auth
func TestNutanixKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu127Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes129AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes130AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes131AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes131UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(1),
			api.WithWorkerNodeCount(1),
		),
	)
	runUpgradeFlowWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
		),
	)
}

// Kubelet Configuration tests
func TestNutanixKubernetes129KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes130KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes131KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}
