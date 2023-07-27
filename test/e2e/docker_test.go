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
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/test/framework"
)

// Labels
func TestDockerKubernetesLabels(t *testing.T) {
	provider := framework.NewDocker(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
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
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// Flux
func TestDockerKubernetes127GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runFluxFlow(test)
}

func TestDockerKubernetes127GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runFluxFlow(test)
}

func TestDockerInstallGitFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestDockerInstallGithubFluxDuringUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube126,
		framework.WithFluxGithub(api.WithFluxConfigName(framework.DefaultFluxConfigName)),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

// Curated packages
func TestDockerKubernetes123CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes124CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes125CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes126CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes127CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestDockerKubernetes123CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes124CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes125CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes126CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes127CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestDockerKubernetes123CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes124CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes125CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes126CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes127CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestDockerKubernetes123CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes124CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes125CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes126CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes127CuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes123CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes124CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes125CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes126CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes127CuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestDockerKubernetes123CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes124CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes125CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes126CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerKubernetes127CuratedPackagesDisabled(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues,
			&v1alpha1.PackageConfiguration{Disable: true}),
	)
	runDisabledCuratedPackageInstallSimpleFlow(test) // other args as necessary
}

func TestDockerCuratedPackagesMetalLB(t *testing.T) {
	RunMetalLBDockerTests(t)
}

// AWS IAM Auth
func TestDockerKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes124AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes125AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes126AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

// Flux
func TestDockerUpgradeWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneCount(2),
			api.WithWorkerNodeCount(2),
		),
		// Needed in order to replace the DockerDatacenterConfig namespace field with the value specified
		// compared to when it was initially created without it.
		provider.WithProviderUpgradeGit(),
	)
}

// OIDC
func TestDockerKubernetes123OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes124OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes125OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes126OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runOIDCFlow(test)
}

func TestDockerKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runOIDCFlow(test)
}

// RegistryMirror
func TestDockerKubernetes127RegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.DockerProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestDockerKubernetes127AirgappedRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.DockerProviderName),
	)
	runDockerAirgapConfigFlow(test)
}

func TestDockerKubernetes127RegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.DockerProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simple flow
func TestDockerKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes124SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes125SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes126SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestDockerKubernetes127SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

// Stacked etcd
func TestDockerKubernetesStackedEtcd(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestDockerKubernetes127Taints(t *testing.T) {
	provider := framework.NewDocker(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
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
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

// Upgrade
func TestDockerKubernetes126To127StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

func TestDockerKubernetes126To127ExternalEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

func TestDockerKubernetes124UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
	)
}

func TestDockerKubernetes123to124UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
}

func TestDockerKubernetes124to125UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
}

func TestDockerKubernetes125to126UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
}

func TestDockerKubernetes126to127UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

func TestDockerKubernetes126to127GithubFluxEnabledUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewDocker(t),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKindnetd()),
	)
	runUpgradeWithFluxFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

func TestDockerKubernetes126WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	provider := framework.NewDocker(t)
	runTestManagementClusterUpgradeSideEffects(t, provider, framework.DockerOS, v1alpha1.Kube126, api.ClusterToConfigFiller(
		api.WithKindnetd(),
	))
}

func TestDockerKubernetes124UpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

// Workload Cluster API
func TestDockerUpgradeKubernetes123to124WorkloadClusterScaleupAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
		),
	)
}

func TestDockerUpgradeWorkloadClusterLabelsAndTaintsAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("tier", "frontend"), api.WithTaint(framework.NoScheduleTaint())),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithTaint(framework.PreferNoScheduleTaint())),
				api.WithStackedEtcdTopology(),
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
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1)),
				api.WithExternalEtcdTopology(1),
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

func TestDockerKubernetes123to124UpgradeFromLatestMinorReleaseAPI(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(api.ClusterToConfigFiller(
		api.WithKubernetesVersion(v1alpha1.Kube123),
	))

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	wc.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(api.ClusterToConfigFiller(
		api.WithKubernetesVersion(v1alpha1.Kube123),
		api.WithManagementCluster(managementCluster.ClusterName),
		api.WithControlPlaneCount(1),
		api.WithWorkerNodeCount(1),
		api.WithStackedEtcdTopology(),
	))
	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPI(
		test,
		release,
		api.ClusterToConfigFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
}

func TestDockerUpgradeKubernetes123to124WorkloadClusterScaleupGitHubFluxAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
		),
		framework.WithFluxGithubConfig(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1)),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterUpgradeFlowAPIWithFlux(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
		),
	)
}

func TestDockerUpgradeWorkloadClusterLabelsAndTaintsGitHubFluxAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
		),
		framework.WithFluxGithubConfig(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("tier", "frontend"), api.WithTaint(framework.NoScheduleTaint())),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithTaint(framework.PreferNoScheduleTaint())),
				api.WithStackedEtcdTopology(),
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

func TestDockerUpgradeWorkloadClusterScaleAddRemoveWorkerNodeGroupsGitHubFluxAPI(t *testing.T) {
	provider := framework.NewDocker(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithExternalEtcdTopology(1),
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
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
				api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
				api.WithWorkerNodeGroup("worker-2", api.WithCount(1)),
				api.WithExternalEtcdTopology(1),
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

func TestDockerCiliumSkipUpgrade_CLIUpgrade(t *testing.T) {
	release, err := framework.GetLatestMinorReleaseFromTestBranch()
	if err != nil {
		t.Fatal(err)
	}

	ver, _ := semver.New(release.Version)
	previousRelease, err := framework.GetPreviousMinorReleaseFromVersion(ver)
	if err != nil {
		t.Fatal(err)
	}

	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(t, provider,
		framework.WithClusterFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
	)

	test.ValidateCiliumCLIAvailable()

	test.GenerateClusterConfig(framework.ExecuteWithEksaRelease(previousRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(previousRelease))
	test.ReplaceCiliumWithOSSCilium()
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
	provider := framework.NewDocker(t)
	management := framework.NewClusterE2ETest(t, provider).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
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
	provider := framework.NewDocker(t)
	management := framework.NewClusterE2ETest(t, provider).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
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
