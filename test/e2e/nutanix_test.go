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

func TestNutanixKubernetes132UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
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

func TestNutanixKubernetes132UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// ADOT
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

func TestNutanixKubernetes132UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

// Prometheus
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

func TestNutanixKubernetes132UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Cluster Autoscaler
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

func TestNutanixKubernetes132UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

// Harbor
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

func TestNutanixKubernetes132UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes133UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestNutanixKubernetes134UbuntuCuratedPackagesHarborSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// Simple Flow
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

func TestNutanixKubernetes132UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes129Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes130Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes131Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes132Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134Ubuntu2204SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
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

func TestNutanixKubernetes132RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes133Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
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

func TestNutanixKubernetes132UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu132NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu133NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes128Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes129Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes129NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes130Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes130NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes131Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes131NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes132Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes132NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes133NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134Ubuntu2204SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
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

func TestNutanixKubernetes132RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes132NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes133RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes133NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134RedHat9SimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134UbuntuSimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu134NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestNutanixKubernetes134RedHat9SimpleFlowWithUUID(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes134NutanixUUID(),
			framework.WithPrismElementClusterUUID(),
			framework.WithNutanixSubnetUUID()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

// Upgrade
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

func TestNutanixKubernetes131To132StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
}

func TestNutanixKubernetes128To129StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes129Template()),
	)
}

func TestNutanixKubernetes129To130StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes129Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes130Template()),
	)
}

func TestNutanixKubernetes130To131StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes130Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes131Template()),
	)
}

func TestNutanixKubernetes131To132StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes132Template()),
	)
}

func TestNutanixKubernetes132To133StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes133Template()),
	)
}

func TestNutanixKubernetes133To134StackedEtcdUbuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes134Template()),
	)
}

func TestNutanixKubernetes128To129Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes129Template()),
	)
}

func TestNutanixKubernetes129To130Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes129Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes130Template()),
	)
}

func TestNutanixKubernetes130To131Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes130Nutanix())
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
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes131Template()),
	)
}

func TestNutanixKubernetes131To132Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes132Template()),
	)
}

func TestNutanixKubernetes132To133Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes133Template()),
	)
}

func TestNutanixKubernetes133To134Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes134Template()),
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

func TestNutanixKubernetes131To132UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
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

func TestNutanixKubernetes131to132RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes132Template()),
	)
}

func TestNutanixKubernetes131to132StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes132Template()),
	)
}

func TestNutanixKubernetes132To133UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestNutanixKubernetes132To133StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestNutanixKubernetes132to133RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes133Template()),
	)
}

func TestNutanixKubernetes132to133StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes132Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes133Template()),
	)
}

func TestNutanixKubernetes133To134UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestNutanixKubernetes133To134StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestNutanixKubernetes133to134RedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes134Template()),
	)
}

func TestNutanixKubernetes133to134StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes133Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.RedHat9Kubernetes134Template()),
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

func TestNutanixKubernetes134UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
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
func TestNutanixKubernetes134UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes128Ubuntu2204ControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes134Ubuntu2204ControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
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
func TestNutanixKubernetes134UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes128Ubuntu2204WorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
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
func TestNutanixKubernetes134Ubuntu2204WorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
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
func TestNutanixKubernetes134UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes128Ubuntu2204ControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix())
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
func TestNutanixKubernetes134Ubuntu2204ControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// OIDC
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

func TestNutanixKubernetes134OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes128Ubuntu2204OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes134Ubuntu2204OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// AWS IAM Auth
func TestNutanixKubernetes128AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes134AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes128Ubuntu2204AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes134Ubuntu2204AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes134UbuntuManagementCPUpgradeAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewNutanix(t, framework.WithUbuntu134Nutanix())
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
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
func TestNutanixKubernetes128KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes134KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes128Ubuntu2204KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes134Ubuntu2204KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu2204Kubernetes134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

// RedHat 8 tests for K8s 1.28
func TestNutanixKubernetes128RedHat8OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat128Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes128RedHat8AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat128Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes128RedHat8KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes128RedHat8WorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
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

func TestNutanixKubernetes128RedHat8WorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
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

func TestNutanixKubernetes128RedHat8ControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes128RedHat8ControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat128Nutanix())
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

// RedHat 9 tests for K8s 1.28
func TestNutanixKubernetes128RedHat9OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes128RedHat9AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes128RedHat9KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes128RedHat9WorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
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

func TestNutanixKubernetes128RedHat9WorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
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

func TestNutanixKubernetes128RedHat9ControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes128RedHat9ControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes128Nutanix())
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

// RedHat 9 tests for K8s 1.34
func TestNutanixKubernetes134RedHat9OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestNutanixKubernetes134RedHat9AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runAWSIamAuthFlow(test)
}

func TestNutanixKubernetes134RedHat9KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestNutanixKubernetes134RedHat9WorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

func TestNutanixKubernetes134RedHat9WorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestNutanixKubernetes134RedHat9ControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

func TestNutanixKubernetes134RedHat9ControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithRedHat9Kubernetes134Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}
