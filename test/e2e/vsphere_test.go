//go:build e2e && (vsphere || all_providers)
// +build e2e
// +build vsphere all_providers

package e2e

import (
	"fmt"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

// Autoimport
func TestVSphereKubernetes123BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes124BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes125BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes126BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes127BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAutoImportFlow(test, provider)
}

// AWS IAM Auth
func TestVSphereKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes124AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes125AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu125()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes126AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu126()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes123BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes124BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket124()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes125BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket125()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes126BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket126()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes127BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes126To127AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

// Curated packages
func kubeVersionVSphereOptUbuntu(version v1alpha1.KubernetesVersion) framework.VSphereOpt {
	switch version {
	case v1alpha1.Kube123:
		return framework.WithUbuntu123()
	case v1alpha1.Kube124:
		return framework.WithUbuntu124()
	case v1alpha1.Kube125:
		return framework.WithUbuntu125()
	case v1alpha1.Kube126:
		return framework.WithUbuntu126()
	case v1alpha1.Kube127:
		return framework.WithUbuntu127()
	default:
		panic(fmt.Sprintf("unsupported version: %v", version))
	}
}

func kubeVersionVSphereOptBottleRocket(version v1alpha1.KubernetesVersion) framework.VSphereOpt {
	switch version {
	case v1alpha1.Kube123:
		return framework.WithBottleRocket123()
	case v1alpha1.Kube124:
		return framework.WithBottleRocket124()
	case v1alpha1.Kube125:
		return framework.WithBottleRocket125()
	case v1alpha1.Kube126:
		return framework.WithBottleRocket126()
	case v1alpha1.Kube127:
		return framework.WithBottleRocket127()
	default:
		panic(fmt.Sprintf("unsupported version: %v", version))
	}
}

func TestVSphereCuratedPackagesSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageInstallSimpleFlow(test)
	}
}

func TestVSphereCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageEmissaryInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageEmissaryInstallSimpleFlow(test)
	}
}

func TestVSphereCuratedPackagesHarborSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
	}
}

func TestVSphereCuratedPackagesAdotUpdateFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesAdotInstallUpdateFlow(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesAdotInstallUpdateFlow(test)
	}
}

func TestVSphereUbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runAutoscalerWithMetricsServerSimpleFlow(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runAutoscalerWithMetricsServerSimpleFlow(test)
	}
}

func TestVSphereUbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesPrometheusInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		test := framework.NewClusterE2ETest(t,
			framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version)),
			framework.WithClusterFiller(api.WithKubernetesVersion(version)),
			framework.WithPackageConfig(t, packageBundleURI(version),
				EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		)
		runCuratedPackagesPrometheusInstallSimpleFlow(test)
	}
}

func TestVSphereUbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCuratedPackageRemoteClusterInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCuratedPackageRemoteClusterInstallSimpleFlow(test)
	}
}

func TestVSphereUbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
	}
}

func TestVSphereUbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		framework.CheckCertManagerCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptUbuntu(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCertManagerRemoteClusterInstallSimpleFlow(test)
	}
}

func TestVSphereBottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	for i, version := range KubeVersions {
		framework.CheckCuratedPackagesCredentials(t)
		framework.CheckCertManagerCredentials(t)
		os.Setenv(framework.ClusterPrefixVar, fmt.Sprintf("%s-%d", EksaPackagesNamespace, i))
		provider := framework.NewVSphere(t, kubeVersionVSphereOptBottleRocket(version))
		test := SetupSimpleMultiCluster(t, provider, version)
		runCertManagerRemoteClusterInstallSimpleFlow(test)
	}
}

// Disable CSI
func TestVSphereKubernetes127DisableCSIUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithUbuntu127(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runVSphereDisableCSIUpgradeFlow(
		test,
		v1alpha1.Kube127,
		provider,
	)
}

// Download artifacts
func TestVSphereDownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestVSphereKubernetes127FluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127BottleRocketFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127ThreeReplicasThreeWorkersFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFluxLegacy(),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes127GitopsOptionsFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithFluxLegacy(
			api.WithFluxBranch(fluxUserProvidedBranch),
			api.WithFluxNamespace(fluxUserProvidedNamespace),
			api.WithFluxConfigurationPath(fluxUserProvidedPath),
		),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes126To127FluxUpgradeLegacy(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes126To127GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereInstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu127())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

// Labels
func TestVSphereKubernetes127UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := ubuntu127ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

func TestVSphereKubernetes127BottlerocketLabelsUpgradeFlow(t *testing.T) {
	provider := bottlerocket127ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

// Multicluster
func TestVSphereKubernetes127MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu127())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube127),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube127),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestVSphereUpgradeMulticlusterWorkloadClusterWithFluxLegacy(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxLegacy(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxLegacy(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Ubuntu126Template(),
		),
	)
}

func TestVSphereUpgradeMulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Ubuntu127Template(),
		),
	)
}

// OIDC
func TestVSphereKubernetes123OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes124OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes125OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu125()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes126OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu126()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes126To127OIDCUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

// Proxy config
func TestVSphereKubernetes127UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes127BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Registry mirror
func TestVSphereKubernetes127UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes127UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes127BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes127UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes127BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Clone mode
func TestVSphereKubernetes127FullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu127(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes127LinkedClone(t *testing.T) {
	diskSize := 20
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu127(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes127BottlerocketFullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket127(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes127BottlerocketLinkedClone(t *testing.T) {
	diskSize := 22
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket127(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

// Simpleflow
func TestVSphereKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes124SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes125SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes126SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes123RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat123VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes124RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat124VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes125RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat125VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes126RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat126VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat127VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes123ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes123DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket127(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes127CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

// NTP Servers test
func TestVSphereKubernetes127BottleRocketWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket127(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runNTPFlow(test, v1alpha1.Bottlerocket)
}

func TestVSphereKubernetes127UbuntuWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithUbuntu127(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runNTPFlow(test, v1alpha1.Ubuntu)
}

// Bottlerocket Configuration test
func TestVSphereKubernetes127BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket127(),
			framework.WithBottlerocketKubernetesSettingsForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runBottlerocketConfigurationFlow(test)
}

// Stacked etcd
func TestVSphereKubernetes123StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes124StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes125StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes126StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes127StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestVSphereKubernetes127UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := ubuntu127ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

func TestVSphereKubernetes127BottlerocketTaintsUpgradeFlow(t *testing.T) {
	provider := bottlerocket127ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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
func TestVSphereKubernetes123UbuntuTo124Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Ubuntu124Template()),
	)
}

func TestVSphereKubernetes124UbuntuTo125Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.Ubuntu125Template()),
	)
}

func TestVSphereKubernetes125UbuntuTo126Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.Ubuntu126Template()),
	)
}

func TestVSphereKubernetes126UbuntuTo127Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes126UbuntuTo127UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes126UbuntuTo127MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(
			provider.Ubuntu127Template(),
			api.WithNumCPUsForAllMachines(vsphereCpVmNumCpuUpdateVar),
			api.WithMemoryMiBForAllMachines(vsphereCpVmMemoryUpdate),
			api.WithDiskGiBForAllMachines(vsphereCpDiskGiBUpdateVar),
			api.WithFolderForAllMachines(vsphereFolderUpdateVar),
			// Uncomment once we support tests with multiple machine configs
			/*api.WithWorkloadVMsNumCPUs(vsphereWlVmNumCpuUpdateVar),
			api.WithWorkloadVMsMemoryMiB(vsphereWlVmMemoryUpdate),
			api.WithWorkloadDiskGiB(vsphereWlDiskGiBUpdate),*/
			// Uncomment the network field once upgrade starts working with it
			// api.WithNetwork(vsphereNetwork2UpdateVar),
		),
	)
}

func TestVSphereKubernetes126UbuntuTo127WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes126UbuntuTo127DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126(),
		framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)),
	)
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes127UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu127())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes127UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu127())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes126BottlerocketTo127Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Bottlerocket127Template()),
	)
}

func TestVSphereKubernetes126BottlerocketTo127MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket127Template(),
			api.WithNumCPUsForAllMachines(vsphereCpVmNumCpuUpdateVar),
			api.WithMemoryMiBForAllMachines(vsphereCpVmMemoryUpdate),
			api.WithDiskGiBForAllMachines(vsphereCpDiskGiBUpdateVar),
			api.WithFolderForAllMachines(vsphereFolderUpdateVar),
			// Uncomment once we support tests with multiple machine configs
			/*api.WithWorkloadVMsNumCPUs(vsphereWlVmNumCpuUpdateVar),
			api.WithWorkloadVMsMemoryMiB(vsphereWlVmMemoryUpdate),
			api.WithWorkloadDiskGiB(vsphereWlDiskGiBUpdate),*/
			// Uncomment the network field once upgrade starts working with it
			// api.WithNetwork(vsphereNetwork2UpdateVar),
		),
	)
}

func TestVSphereKubernetes126BottlerocketTo127WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket126())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Bottlerocket127Template()),
	)
}

func TestVSphereKubernetes126BottlerocketTo127DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket126(),
		framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Bottlerocket127Template()),
	)
}

func TestVSphereKubernetes127BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket127())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes127BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket127())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes126UbuntuTo127StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Ubuntu127Template()),
	)
}

func TestVSphereKubernetes126BottlerocketTo127StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Bottlerocket127Template()),
	)
}

func TestVSphereKubernetes126UbuntuTo127UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewVSphere(t, framework.WithUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Ubuntu127Template(), api.WithResourcePoolForAllMachines(vsphereInvalidResourcePoolUpdateVar)), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Ubuntu127Template(), api.WithResourcePoolForAllMachines(os.Getenv(vsphereResourcePoolVar))), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube127,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestVSphereKubernetes124BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
		framework.WithBottlerocketFromRelease(release, v1alpha1.Kube124),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
		provider.WithProviderUpgrade(
			provider.Bottlerocket124Template(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes124UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube124),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
		provider.WithProviderUpgrade(
			provider.Ubuntu124Template(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes124UbuntuUpgradeFromLatestMinorReleaseAlwaysNetworkPolicy(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube124),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(
			provider.Ubuntu124Template(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes123To124UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube123),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube124,
		provider.WithProviderUpgrade(
			provider.Ubuntu124Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
}

func TestVSphereKubernetes124To125UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube124),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube125,
		provider.WithProviderUpgrade(
			provider.Ubuntu125Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
}

func TestVSphereKubernetes125To126UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube125),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube126,
		provider.WithProviderUpgrade(
			provider.Ubuntu126Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
}

func TestVSphereKubernetes126To127UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithUbuntuForRelease(release, v1alpha1.Kube126),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube127,
		provider.WithProviderUpgrade(
			provider.Ubuntu127Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

func TestVSphereKubernetes127BottlerocketAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithBottleRocket127(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestVSphereKubernetes124UbuntuUpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T) {
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
		provider.WithWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(2))),
		provider.WithWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1))),
		provider.WithWorkerNodeGroup("worker-3", framework.WithWorkerNodeGroup("worker-3", api.WithCount(1), api.WithLabel("tier", "frontend"))),
		provider.WithUbuntu124(),
	)

	runUpgradeFlowWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.RemoveWorkerNodeGroup("worker-2"),
			api.WithWorkerNodeGroup("worker-1", api.WithCount(1)),
			api.RemoveWorkerNodeGroup("worker-3"),
		),
		// Re-adding with no labels and a taint
		provider.WithWorkerNodeGroupConfiguration("worker-3", framework.WithWorkerNodeGroup("worker-3", api.WithCount(1), api.WithTaint(framework.NoScheduleTaint()))),
		provider.WithWorkerNodeGroupConfiguration("worker-1", framework.WithWorkerNodeGroup("worker-4", api.WithCount(1))),
	)
}

func TestVSphereKubernetes123to124UpgradeFromLatestMinorReleaseBottleRocketAPI(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t)
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
		),
		api.VSphereToConfigFiller(
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
		provider.WithBottleRocketForRelease(release, v1alpha1.Kube123),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	wc.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithManagementCluster(managementCluster.ClusterName),
		),
		api.VSphereToConfigFiller(
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
		provider.WithBottleRocketForRelease(release, v1alpha1.Kube123),
	)

	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPI(
		test,
		release,
		api.JoinClusterConfigFillers(
			provider.WithBottleRocket124(),
			api.VSphereToConfigFiller(
				provider.Bottlerocket124Template(), // Set the template so it doesn't get autoimported
			),
		),
	)
}

// Workload API
func TestVSphereMulticlusterWorkloadClusterAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)
	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu123(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithUbuntu123(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithUbuntu124(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithUbuntu125(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithUbuntu126(),
		),
	)
	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereUpgradeLabelsTaintsUbuntuAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithUbuntu124(),
		),
	)

	runWorkloadClusterUpgradeFlowAPI(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeGroup("worker-0", api.WithLabel("key1", "val1"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-1", api.WithLabel("key2", "val2"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-2", api.WithNoTaints()),
			api.WithControlPlaneLabel("cpKey1", "cpVal1"),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereUpgradeWorkerNodeGroupsUbuntuAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu124(),
		),
	)

	runWorkloadClusterUpgradeFlowAPI(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
			api.RemoveWorkerNodeGroup("worker-1"),
		),
		vsphere.WithWorkerNodeGroupConfiguration("worker-1", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1))),
	)
}

func TestVSphereMulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)
	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		framework.WithFluxGithubConfig(),
		vsphere.WithUbuntu124(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			vsphere.WithUbuntu123(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithExternalEtcdTopology(1),
			),
			vsphere.WithUbuntu124(),
		),
	)

	test.CreateManagementCluster()
	test.RunInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereUpgradeKubernetes123to124UbuntuWorkloadClusterAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu123(),
		),
	)

	runWorkloadClusterUpgradeFlowAPI(test,
		vsphere.WithUbuntu124(),
	)
}

func TestVSphereCiliumDisableCSIUbuntuAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu124(),
		),
	)

	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.UpdateClusterConfig(
			api.ClusterToConfigFiller(
				api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways),
			),
			api.VSphereToConfigFiller(api.WithDisableCSI(true)),
		)
		wc.ApplyClusterManifest()
		wc.DeleteWorkloadVsphereCSI()
		wc.ValidateClusterState()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereUpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithBottleRocket124(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithBottleRocket124(),
		),
	)

	runWorkloadClusterUpgradeFlowAPIWithFlux(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeGroup("worker-0", api.WithLabel("key1", "val1"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-1", api.WithLabel("key2", "val2"), api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup("worker-2", api.WithNoTaints()),
			api.WithControlPlaneLabel("cpKey1", "cpVal1"),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereUpgradeWorkerNodeGroupsUbuntuGitHubFluxAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu124(),
		),
	)

	runWorkloadClusterUpgradeFlowAPIWithFlux(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeGroup("worker-0", api.WithCount(2)),
			api.RemoveWorkerNodeGroup("worker-1"),
		),
		vsphere.WithWorkerNodeGroupConfiguration("worker-1", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1))),
	)
}

func TestVSphereUpgradeKubernetesCiliumDisableCSIUbuntuGitHubFluxAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu124(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			),
			vsphere.WithWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu123(),
		),
	)

	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		test.PushWorkloadClusterToGit(wc,
			api.ClusterToConfigFiller(
				api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways),
			),
			api.VSphereToConfigFiller(api.WithDisableCSI(true)),
			vsphere.WithUbuntu124(),
		)
		wc.DeleteWorkloadVsphereCSI()
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereKubernetes127UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu127(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func ubuntu127ProviderWithLabels(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithUbuntu127(),
	)
}

func bottlerocket127ProviderWithLabels(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithBottleRocket127(),
	)
}

func ubuntu127ProviderWithTaints(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithUbuntu127(),
	)
}

func bottlerocket127ProviderWithTaints(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithBottleRocket127(),
	)
}

func runVSphereCloneModeFlow(test *framework.ClusterE2ETest, vsphere *framework.VSphere, diskSize int) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	vsphere.ValidateNodesDiskGiB(test.GetCapiMachinesForCluster(test.ClusterName), diskSize)
	test.DeleteCluster()
}
