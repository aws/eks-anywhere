//go:build e2e && (docker || all_providers)
// +build e2e
// +build docker all_providers

package e2e

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/test/framework"
)

// Labels
func TestDockerKubernetesLabels(t *testing.T) {
	provider := framework.NewDocker(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
			api.WithWorkerNodeGroup(worker1, api.WithCount(1)),
			api.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// Flux
func TestDockerKubernetes132GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes132GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runFluxFlow(test)
}

func TestDockerInstallGitFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube132,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestDockerInstallGithubFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube132,
		framework.WithFluxGithub(api.WithFluxConfigName(framework.DefaultFluxConfigName)),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestDockerKubernetes132UpgradeWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(2),
			api.WithWorkerNodeCount(2),
		),
		// Needed in order to replace the DockerDatacenterConfig namespace field with the value specified
		// compared to when it was initially created without it.
		provider.WithProviderUpgradeGit(),
	)
}

// Curated Packages
func TestDockerKubernetes128CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes129CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes130CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes131CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes132CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
func TestDockerKubernetes128CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes129CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes130CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes131CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes132CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// Harbor
func TestDockerKubernetes128CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes129CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes130CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes131CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes132CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// ADOT
func TestDockerKubernetes128CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes129CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes130CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes131CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes132CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

// Prometheus
func TestDockerKubernetes128CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes129CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes130CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes131CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes132CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Curated Packages Disabled
func TestDockerKubernetes128CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes129CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes130CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes131CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes132CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

// MetalLB
func TestDockerKubernetes128CuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTestsForKubeVersion(t, v1alpha1.Kube128)
}

func TestDockerKubernetes129CuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTestsForKubeVersion(t, v1alpha1.Kube129)
}

func TestDockerKubernetes130CuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTestsForKubeVersion(t, v1alpha1.Kube130)
}

func TestDockerKubernetes131CuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTestsForKubeVersion(t, v1alpha1.Kube131)
}

func TestDockerKubernetes132CuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTestsForKubeVersion(t, v1alpha1.Kube132)
}

// AWS IAM Auth
func TestDockerKubernetes128AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes129AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes130AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes131AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes132AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runAWSIamAuthFlow(test)
}

// OIDC
func TestDockerKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes129OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes130OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes131OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes132OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runOIDCFlow(test)
}

// Registry Mirror
func TestDockerKubernetes131RegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.DockerProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestDockerKubernetes131AirgappedRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.DockerProviderName),
	)
	runDockerAirgapConfigFlow(test)
}

func TestDockerKubernetes131AirgappedUpgradeFromLatestRegistryMirrorAndCert(t *testing.T) {
	release := latestMinorRelease(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.DockerProviderName),
	)
	runDockerAirgapUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube131,
	)
}

func TestDockerKubernetes131RegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.DockerProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestDockerKubernetes132RegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.DockerProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simple Flow
func TestDockerKubernetes128SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes129SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes130SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes131SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes132SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

// Stacked Etcd
func TestDockerKubernetesStackedEtcd(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestDockerKubernetes132Taints(t *testing.T) {
	provider := framework.NewDocker(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoScheduleTaint()), api.WithCount(2)),
			api.WithWorkerNodeGroup(worker1, api.WithCount(1)),
			api.WithWorkerNodeGroup(worker2, api.WithTaint(framework.PreferNoScheduleTaint()), api.WithCount(1)),
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestDockerKubernetes132WorkloadClusterTaints(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)

	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithTaint(framework.NoScheduleTaint())),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1), api.WithTaint(framework.NoExecuteTaint())),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)

	runWorkloadClusterExistingConfigFlow(test)
}

// Upgrade
func TestDockerKubernetes132UpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithLicenseToken(licenseToken),
		),
		provider.WithNewWorkerNodeGroup("", framework.WithWorkerNodeGroup("worker-1", api.WithCount(2))),
		provider.WithNewWorkerNodeGroup("", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1))),
		provider.WithNewWorkerNodeGroup(
			"", framework.WithWorkerNodeGroup("worker-3", api.WithCount(1), api.WithLabel("tier", "frontend")),
		),
	)

	runUpgradeFlowWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.RemoveWorkerNodeGroup("worker-2"),
			api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
			api.RemoveWorkerNodeGroup("worker-3"),
			api.WithWorkerNodeGroup("worker-3", api.WithCount(1), api.WithTaint(framework.NoScheduleTaint())),
		),
		provider.WithNewWorkerNodeGroup("", framework.WithWorkerNodeGroup("worker-4", api.WithCount(1))),
	)
}

func TestDockerKubernetes131To132StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
}

func TestDockerKubernetes131To132ExternalEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
}

// Upgrade From Latest Minor Release
func TestDockerKubernetes128to129UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
}

func TestDockerKubernetes129to130UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
}

func TestDockerKubernetes130to131UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
}

func TestDockerKubernetes131to132UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
}

func TestDockerKubernetes131to132GithubFluxEnabledUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeWithFluxFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
}

func TestDockerKubernetes132WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	provider := framework.NewDocker(t)
	runTestManagementClusterUpgradeSideEffects(t, provider, framework.DockerOS, v1alpha1.Kube132)
}

// Workload Cluster API
func TestDockerUpgradeKubernetes131to132WorkloadClusterScaleupAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
		),
	)
}

func TestDockerUpgradeWorkloadClusterLabelsAndTaintsAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("tier", "frontend"), api.WithTaint(framework.NoScheduleTaint())),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithTaint(framework.PreferNoScheduleTaint())),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneLabel("cpKey1", "cpVal1"),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
			api.RemoveWorkerNodeGroup("worker-0"),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("key1", "val1"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-1", api.WithLabel("key2", "val2"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-2", api.WithNoTaints()),
		),
	)
}

func TestDockerUpgradeWorkloadClusterScaleAddRemoveWorkerNodeGroupsAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1)),
				api.WithExternalEtcdTopology(1),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
			api.WithWorkerNodeGroup("worker-1", api.WithCount(2)),
			api.RemoveWorkerNodeGroup("worker-2"),
			api.WithWorkerNodeGroup("worker-3", api.WithCount(1)),
		),
	)
}

func TestDockerKubernetes131to132UpgradeFromLatestMinorReleaseAPI(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(api.ClusterToConfigFiller(
		api.WithKubernetesVersion(v1alpha1.Kube131),
	))

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	wc.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(api.ClusterToConfigFiller(
		api.WithKubernetesVersion(v1alpha1.Kube131),
		api.WithManagementCluster(managementCluster.ClusterName),
		api.WithControlPlaneCount(1),
		api.WithWorkerNodeCount(1),
		api.WithStackedEtcdTopology(),
	))
	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPI(
		test,
		release,
		wc.ClusterConfig.Cluster.Spec.KubernetesVersion,
		v1alpha1.Kube132,
		"",
	)
}

func TestDockerUpgradeKubernetes131to132WorkloadClusterScaleupGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
		framework.WithFluxGithubConfig(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPIWithFlux(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
		),
	)
}

func TestDockerKubernetes132UpgradeWorkloadClusterLabelsAndTaintsGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
		framework.WithFluxGithubConfig(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("tier", "frontend"), api.WithTaint(framework.NoScheduleTaint())),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithTaint(framework.PreferNoScheduleTaint())),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPIWithFlux(
		test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneLabel("cpKey1", "cpVal1"),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
			api.RemoveWorkerNodeGroup("worker-0"),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("key1", "val1"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-1", api.WithLabel("key2", "val2"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-2", api.WithNoTaints()),
		),
	)
}

func TestDockerKubernetes132UpgradeWorkloadClusterScaleAddRemoveWorkerNodeGroupsGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
			api.WithLicenseToken(licenseToken),
		),
		framework.WithFluxGithubConfig(
			api.WithClusterConfigPath("test"),
			api.WithBranch("docker"),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1)),
				api.WithExternalEtcdTopology(1),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPIWithFlux(
		test,
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
			api.WithWorkerNodeGroup("worker-1", api.WithCount(2)),
			api.RemoveWorkerNodeGroup("worker-2"),
			api.WithWorkerNodeGroup("worker-3", api.WithCount(1)),
		),
	)
}

// Cilium Skip Upgrade
func TestDockerCiliumSkipUpgrade_CLICreate(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(t, provider,
		framework.WithClusterFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithCiliumSkipUpgrade(),
		),
	)

	test.ValidateCiliumCLIAvailable()

	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateClusterState()
	test.ValidateEKSACiliumInstalled()
	test.DeleteCluster()
}

func TestDockerUpgradeFromLatestMinorReleaseCiliumSkipUpgrade_CLIUpgrade(t *testing.T) {
	release := latestMinorRelease(t)

	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(t, provider,
		framework.WithClusterFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
	)

	test.ValidateCiliumCLIAvailable()

	test.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	test.CreateCluster(framework.ExecuteWithEksaRelease(release))
	test.ReplaceCiliumWithOSSCilium()

	t.Log("Waiting for cilium replacement to complete")
	// Wait two minutes before validating cluster state and attempting the upgrade
	// After replacing cilium, the nodes can temporarily go into a not ready state
	// and we want to give them time to recover before validating the cluster state
	time.Sleep(5 * time.Minute)

	test.ValidateClusterState()
	test.UpgradeClusterWithNewConfig(
		[]framework.ClusterE2ETestOpt{
			framework.WithClusterUpgrade(api.WithCiliumSkipUpgrade()),
		},
	)
	test.ValidateClusterState()
	test.ValidateEKSACiliumNotInstalled()
	test.DeleteCluster()
}

func TestDockerCiliumSkipUpgrade_ControllerCreate(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	management := framework.NewClusterE2ETest(t, provider).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)

	management.ValidateCiliumCLIAvailable()

	test := framework.NewMulticlusterE2ETest(t, management)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(management.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithCiliumSkipUpgrade(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)

	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()

		client, err := wc.BuildWorkloadClusterClient()
		if err != nil {
			wc.T.Fatalf("Error creating workload cluster client: %v", err)
		}

		framework.AwaitCiliumDaemonSetReady(context.Background(), client, 20, 5*time.Second)

		wc.DeleteClusterWithKubectl()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestDockerCiliumSkipUpgrade_ControllerUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewDocker(t)
	management := framework.NewClusterE2ETest(t, provider).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)

	management.ValidateCiliumCLIAvailable()

	test := framework.NewMulticlusterE2ETest(t, management)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(management.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
		),
	)

	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()

		client, err := wc.BuildWorkloadClusterClient()
		if err != nil {
			wc.T.Fatalf("Error creating workload cluster client: %v", err)
		}

		// Wait for Cilium to come up.
		framework.AwaitCiliumDaemonSetReady(context.Background(), client, 20, 5*time.Second)

		// Skip Cilium upgrades and reapply the kubeconfig.
		wc.UpdateClusterConfig(api.ClusterToConfigFiller(api.WithCiliumSkipUpgrade()))
		wc.ApplyClusterManifest()

		// Give some time for reconciliation to happen.
		time.Sleep(15 * time.Second)

		// Validate EKSA Cilium is still installed because we haven't done anything to it yet
		// and the controller shouldn't have removed it.
		framework.AwaitCiliumDaemonSetReady(context.Background(), client, 20, 5*time.Second)

		// Introduce a different OSS Cillium, wait for it to come up and validate the controller
		// doesn't try to override the new Cilium.
		wc.ReplaceCiliumWithOSSCilium()
		wc.ValidateClusterState()
		wc.ValidateEKSACiliumNotInstalled()

		wc.DeleteClusterWithKubectl()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestDockerKubernetesNonRegionalCuratedPackages(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)

	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetesUpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	runUpgradeManagementComponentsFlow(t, release, provider, v1alpha1.Kube132, "")
}

// Etcd Scale tests
func TestDockerKubernetes132EtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestDockerKubernetes132EtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestDockerKubernetes131to132EtcdScaleUp(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestDockerKubernetes131to132EtcdScaleDown(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
		),
	)
}

// Kubelet Configuration tests
func TestDockerKubernetes129KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestDockerKubernetes130KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestDockerKubernetes131KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestDockerKubernetes132KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}
