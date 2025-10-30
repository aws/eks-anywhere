//go:build e2e && (vsphere || all_providers)
// +build e2e
// +build vsphere all_providers

package e2e

import (
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/test/framework"
)

// APIServerExtraArgs
func TestVSphereKubernetes134BottlerocketAPIServerExtraArgsSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithEnvVar(features.APIServerExtraArgsEnabledEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithControlPlaneAPIServerExtraArgs(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

// TODO: Investigate why this test takes long time to pass with service-account-issuer flag
func TestVSphereKubernetes134BottlerocketAPIServerExtraArgsUpgradeFlow(t *testing.T) {
	var addAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	var removeAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithEnvVar(features.APIServerExtraArgsEnabledEnvVar, "true"),
	)
	addAPIServerExtraArgsclusterOpts = append(
		addAPIServerExtraArgsclusterOpts,
		framework.WithClusterUpgrade(
			api.WithControlPlaneAPIServerExtraArgs(),
		),
	)
	removeAPIServerExtraArgsclusterOpts = append(
		removeAPIServerExtraArgsclusterOpts,
		framework.WithClusterUpgrade(
			api.RemoveAllAPIServerExtraArgs(),
		),
	)
	runAPIServerExtraArgsUpgradeFlow(
		test,
		addAPIServerExtraArgsclusterOpts,
		removeAPIServerExtraArgsclusterOpts,
	)
}

// Autoimport
func TestVSphereKubernetes128BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes134BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runAutoImportFlow(test, provider)
}

// AWS IAM Auth
func TestVSphereKubernetes128AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes129AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes130AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes131AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes130AWSIamAuthWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube130),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithAWSIam(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube130),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runAWSIamAuthFlowWorkload(test)
}

func TestVSphereKubernetes128BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes129BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes130BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes131BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes132BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes133BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes134AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes134BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes132To133AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes133To134AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

// AWS IAM Auth Add/Remove Tests
func TestVSphereKubernetes128UbuntuAddAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithAwsIamEnvVarCheck(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runUpgradeFlowAddAWSIamAuth(test, v1alpha1.Kube128)
}

func TestVSphereKubernetes134UbuntuAddAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithAwsIamEnvVarCheck(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runUpgradeFlowAddAWSIamAuth(test, v1alpha1.Kube134)
}

func TestVSphereKubernetes128UbuntuRemoveAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runUpgradeFlowRemoveAWSIamAuth(test, v1alpha1.Kube128)
}

func TestVSphereKubernetes134UbuntuRemoveAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runUpgradeFlowRemoveAWSIamAuth(test, v1alpha1.Kube134)
}

// Curated Packages
func TestVSphereKubernetes128CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes129CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes129CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
func TestVSphereKubernetes128CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes129CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// Harbor
func TestVSphereKubernetes128CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes129CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes130CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes131CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes132CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes133CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes134CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// ADOT
func TestVSphereKubernetes128CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes129CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

// Cluster Autoscaler
func TestVSphereKubernetes128UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesClusterAutoscalerUpgradeFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithExternalEtcdTopology(1),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithExternalEtcdTopology(1),
				api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes),
			),
			framework.WithPackageConfig(
				t,
				packageBundleURI(v1alpha1.Kube133),
				EksaPackageControllerHelmChartName,
				EksaPackageControllerHelmURI,
				EksaPackageControllerHelmVersion,
				EksaPackageControllerHelmValues,
				nil,
			),
		),
	)
	runAutoscalerUpgradeFlow(test)
}

// Prometheus
func TestVSphereKubernetes128UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Workload Cluster Curated Packages
func TestVSphereKubernetes128UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes128UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes128UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

// Download Artifacts
func TestVSphereDownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestVSphereKubernetes128GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes130GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes131GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes132GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes133GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes128GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes130GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes131GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes132GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes133GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes134GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes134GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes128BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes130BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes131BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes132BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes133BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes128BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes130BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes131BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes132BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes133BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes134BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes134BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes128To129GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestVSphereKubernetes129To130GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
}

func TestVSphereKubernetes130To131GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
}

func TestVSphereKubernetes131To132GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
}

func TestVSphereKubernetes132To133GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes133To134GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestVSphereInstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube132,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

// Labels
func TestVSphereKubernetes128UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := ubuntu128ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes134UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := ubuntu134ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes128BottlerocketLabelsUpgradeFlow(t *testing.T) {
	provider := bottlerocket128ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes134BottlerocketLabelsUpgradeFlow(t *testing.T) {
	provider := bottlerocket134ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// Multicluster
func TestVSphereKubernetes128MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestVSphereKubernetes134MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube134),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube134),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

// OIDC
func TestVSphereKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes129OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes130OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes131OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes132OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes133OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes134OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes132To133OIDCUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes133To134OIDCUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

// Proxy Config
func TestVSphereKubernetes128UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes130UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes131UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes132UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes133UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes128BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes129BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes130BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes131BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes132BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes133BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes134UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes134BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Registry Mirror
func TestVSphereKubernetes133UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes128UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes130UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes131UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes132UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes133UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes128BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes130BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes131BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes132BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes133BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes128UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes130UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes131UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes132UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes133UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes128BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes130BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes131BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes132BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes133BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes134BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes130BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes131BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes132BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes133BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes134UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes130UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes131UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes132UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes133UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName,
			v1alpha1.OCINamespace{
				Registry:  EksaPackagesRegistry,
				Namespace: EksaPackagesRegistryMirrorAlias,
			}),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

// Clone mode
func TestVSphereKubernetes128FullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu128(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes134FullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu134(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes128LinkedClone(t *testing.T) {
	diskSize := 20
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu128(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes134LinkedClone(t *testing.T) {
	diskSize := 20
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu134(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes128BottlerocketFullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket128(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes134BottlerocketFullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket134(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes128BottlerocketLinkedClone(t *testing.T) {
	diskSize := 22
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket128(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes134BottlerocketLinkedClone(t *testing.T) {
	diskSize := 22
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket134(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

// Simple Flow
func TestVSphereKubernetes128Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes129Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes130Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes131Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes132Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes133Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(api.WithLicenseToken(licenseToken)),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes133Ubuntu2204NetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			"worker-networks",
			framework.WithWorkerNodeGroup("worker-networks", api.WithCount(1)),
			api.WithNetworks([]string{
				os.Getenv("T_VSPHERE_NETWORK"),
				"/SDDC-Datacenter/network/sddc-cgw-network-1",
			}),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	test.CreateCluster()

	// Wait for cluster to be ready
	test.WaitForControlPlaneReady()

	test.DeleteCluster()

	//runSimpleFlowWithoutClusterConfigGenerationWithNetworkValidation(test)
}

func TestVSphereKubernetes128Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes129Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes130Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes131Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes132Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes133Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes134Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(api.WithLicenseToken(licenseToken)),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes134Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes128RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat128VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat129VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat130VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat131VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9128VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9129VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9130VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9131VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9132VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9133VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9134VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket130(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket131(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket132(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket133(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes128CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

// Cilium Helm Values Tests - Ubuntu
func TestVSphereKubernetes128CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
			},
		})),
	)
	runSimpleFlow(test)
}

// Cilium Helm Values Precedence Tests - Validate helmValues overrides deprecated fields
func TestVSphereKubernetes128CiliumHelmValuesPrecedenceUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeDefault)), // This should be ignored
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"policyEnforcementMode": "always", // This should take precedence
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133CiliumHelmValuesPrecedenceUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeDefault)), // This should be ignored
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"policyEnforcementMode": "always", // This should take precedence
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		})),
	)
	runSimpleFlow(test)
}

// Cilium Helm Values Complex Configuration Tests
func TestVSphereKubernetes128CiliumHelmValuesComplexConfigUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
				"replicas": 2,
			},
			"policyEnforcementMode": "always",
			"routingMode":           "native",
			"ipv4NativeRoutingCIDR": "192.168.0.0/16",
			"autoDirectNodeRoutes":  "true",
		})),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133CiliumHelmValuesComplexConfigUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithCiliumHelmValues(map[string]interface{}{
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"operator": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": false,
				},
				"replicas": 2,
			},
			"policyEnforcementMode": "always",
			"routingMode":           "native",
			"ipv4NativeRoutingCIDR": "192.168.0.0/16",
			"autoDirectNodeRoutes":  "true",
		})),
	)
	runSimpleFlow(test)
}

// NTP Servers test
func TestVSphereKubernetes128BottleRocketWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket128(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runNTPFlow(test, v1alpha1.Bottlerocket)
}

func TestVSphereKubernetes134BottleRocketWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket134(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runNTPFlow(test, v1alpha1.Bottlerocket)
}

func TestVSphereKubernetes128UbuntuWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithUbuntu128(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runNTPFlow(test, v1alpha1.Ubuntu)
}

func TestVSphereKubernetes134UbuntuWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithUbuntu134(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runNTPFlow(test, v1alpha1.Ubuntu)
}

// Bottlerocket Configuration tests
func TestVSphereKubernetes128BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket128(),
			framework.WithBottlerocketKubernetesSettingsForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runBottlerocketConfigurationFlow(test)
}

func TestVSphereKubernetes134BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket134(),
			framework.WithBottlerocketKubernetesSettingsForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runBottlerocketConfigurationFlow(test)
}

// Stacked Etcd
func TestVSphereKubernetes128StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes134StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestVSphereKubernetes128UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := ubuntu128ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes134UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := ubuntu134ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes128BottlerocketTaintsUpgradeFlow(t *testing.T) {
	provider := bottlerocket128ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes134BottlerocketTaintsUpgradeFlow(t *testing.T) {
	provider := bottlerocket134ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes128UbuntuWorkloadClusterTaintsFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewVSphere(t, framework.WithUbuntu128())

	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
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
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			provider.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			provider.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoExecuteTaint()))),
		),
	)

	runWorkloadClusterExistingConfigFlow(test)
}

// Upgrade
func TestVSphereKubernetes128UbuntuTo129Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestVSphereKubernetes129UbuntuTo130Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
}

func TestVSphereKubernetes130UbuntuTo131Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
}

func TestVSphereKubernetes131UbuntuTo132Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
}

func TestVSphereKubernetes132UbuntuTo133Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes133UbuntuTo134Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestVSphereKubernetes128To129Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes129Template()),
	)
}

func TestVSphereKubernetes129To130Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes130Template()),
	)
}

func TestVSphereKubernetes130To131Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes131Template()),
	)
}

func TestVSphereKubernetes131To132Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes132Template()),
	)
}

func TestVSphereKubernetes132To133Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes133Template()),
	)
}

func TestVSphereKubernetes133To134Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes134Template()),
	)
}

func TestVSphereKubernetes128To129Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes129Template()),
	)
}

func TestVSphereKubernetes129To130Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes130Template()),
	)
}

func TestVSphereKubernetes130To131Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes131Template()),
	)
}

func TestVSphereKubernetes131To132Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes132Template()),
	)
}

func TestVSphereKubernetes132To133Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes133Template()),
	)
}

func TestVSphereKubernetes133To134Ubuntu2204StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes134Template()),
	)
}

func TestVSphereKubernetes128To129Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes129Template()),
	)
}

func TestVSphereKubernetes129To130Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes130Template()),
	)
}

func TestVSphereKubernetes130To131Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes131Template()),
	)
}

func TestVSphereKubernetes131To132Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes132Template()),
	)
}

func TestVSphereKubernetes132To133Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes133Template()),
	)
}

func TestVSphereKubernetes133To134Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes134Template()),
	)
}

func TestVSphereKubernetes128To129Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes129Template()),
	)
}

func TestVSphereKubernetes129To130Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes130Template()),
	)
}

func TestVSphereKubernetes130To131Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes131Template()),
	)
}

func TestVSphereKubernetes131To132Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes132Template()),
	)
}

func TestVSphereKubernetes132To133Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes133Template()),
	)
}

func TestVSphereKubernetes133To134Ubuntu2404StackedEtcdUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes134Template()),
	)
}

func TestVSphereKubernetes128To129RedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat128VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat129Template()),
	)
}

func TestVSphereKubernetes129To130RedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat129VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat130Template()),
	)
}

func TestVSphereKubernetes130To131RedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat130VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat131Template()),
	)
}

func TestVSphereKubernetes128To129RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9128VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat9129Template()),
	)
}

func TestVSphereKubernetes129To130RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9129VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat9130Template()),
	)
}

func TestVSphereKubernetes130To131RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9130VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat9131Template()),
	)
}

func TestVSphereKubernetes131To132RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9131VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Redhat9132Template()),
	)
}

func TestVSphereKubernetes132To133RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9132VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Redhat9133Template()),
	)
}

func TestVSphereKubernetes133To134RedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9133VSphere())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Redhat9134Template()),
	)
}

func TestVSphereKubernetes128To129StackedEtcdRedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat128VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat129Template()),
	)
}

func TestVSphereKubernetes129To130StackedEtcdRedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat129VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat130Template()),
	)
}

func TestVSphereKubernetes130To131StackedEtcdRedHatUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat130VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat131Template()),
	)
}

func TestVSphereKubernetes128To129StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9128VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9129Template()),
	)
}

func TestVSphereKubernetes129To130StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9129VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9130Template()),
	)
}

func TestVSphereKubernetes130To131StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9130VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9131Template()),
	)
}

func TestVSphereKubernetes131To132StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9131VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9132Template()),
	)
}

func TestVSphereKubernetes132To133StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9132VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9133Template()),
	)
}

func TestVSphereKubernetes133To134StackedEtcdRedHat9Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithRedHat9133VSphere())
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
		provider.WithProviderUpgrade(provider.Redhat9134Template()),
	)
}

func TestVSphereKubernetes128Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes128Template()),
	)
}

func TestVSphereKubernetes129Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes129Template()),
	)
}

func TestVSphereKubernetes130Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes130Template()),
	)
}

func TestVSphereKubernetes131Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes131Template()),
	)
}

func TestVSphereKubernetes132Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes132Template()),
	)
}

func TestVSphereKubernetes133Ubuntu2004To2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes133Template()),
	)
}

func TestVSphereKubernetes128UbuntuTo129InPlaceUpgradeCPOnly(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	kube128 := v1alpha1.Kube128
	kube129 := v1alpha1.Kube129
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(kube128),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube128),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129TemplateForMachineConfig(providers.GetControlPlaneNodeName(test.ClusterName))),
	)
}

func TestVSphereKubernetes132UbuntuTo133InPlaceUpgradeWorkerOnly(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	kube132 := v1alpha1.Kube132
	kube133 := v1alpha1.Kube133
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(kube133),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube132),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
	)
	test.UpdateClusterConfig(
		provider.WithKubeVersionAndOSMachineConfig(providers.GetControlPlaneNodeName(test.ClusterName), kube133, framework.Ubuntu2004),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()), // this will just set everything to 1.33 as expected
	)
}

func TestVSphereKubernetes128UbuntuTo129MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(
			provider.Ubuntu129Template(),
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

func TestVSphereKubernetes129UbuntuTo130MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(
			provider.Ubuntu130Template(),
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

func TestVSphereKubernetes130UbuntuTo131MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(
			provider.Ubuntu131Template(),
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

func TestVSphereKubernetes131UbuntuTo132MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(
			provider.Ubuntu132Template(),
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

func TestVSphereKubernetes132UbuntuTo133MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(
			provider.Ubuntu133Template(),
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

func TestVSphereKubernetes133UbuntuTo134MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(
			provider.Ubuntu134Template(),
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

func TestVSphereKubernetes128UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes129UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes130UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes131UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes132UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes133UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes128UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes129UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes130UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes131UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes132UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes133UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes134UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes134UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes128BottlerocketTo129Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Bottlerocket129Template()),
	)
}

func TestVSphereKubernetes129BottlerocketTo130Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Bottlerocket130Template()),
	)
}

func TestVSphereKubernetes130BottlerocketTo131Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Bottlerocket131Template()),
	)
}

func TestVSphereKubernetes131BottlerocketTo132Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Bottlerocket132Template()),
	)
}

func TestVSphereKubernetes132BottlerocketTo133Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Bottlerocket133Template()),
	)
}

func TestVSphereKubernetes133BottlerocketTo134Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Bottlerocket134Template()),
	)
}

func TestVSphereKubernetes128BottlerocketTo129MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket129Template(),
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

func TestVSphereKubernetes129BottlerocketTo130MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket130Template(),
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

func TestVSphereKubernetes130BottlerocketTo131MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket131Template(),
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

func TestVSphereKubernetes131BottlerocketTo132MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket132Template(),
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

func TestVSphereKubernetes132BottlerocketTo133MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket133Template(),
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

func TestVSphereKubernetes133BottlerocketTo134MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket134Template(),
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

func TestVSphereKubernetes128BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes129BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes130BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes131BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes132BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes133BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes128BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes129BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes130BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes131BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes132BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes133BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes134BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes134BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes128UbuntuTo129StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
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
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestVSphereKubernetes129UbuntuTo130StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
}

func TestVSphereKubernetes130UbuntuTo131StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
}

func TestVSphereKubernetes131UbuntuTo132StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu131())
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
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
}

func TestVSphereKubernetes128BottlerocketTo129StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
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
		provider.WithProviderUpgrade(provider.Bottlerocket129Template()),
	)
}

func TestVSphereKubernetes129BottlerocketTo130StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Bottlerocket130Template()),
	)
}

func TestVSphereKubernetes130BottlerocketTo131StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Bottlerocket131Template()),
	)
}

func TestVSphereKubernetes131BottlerocketTo132StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket131())
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
		provider.WithProviderUpgrade(provider.Bottlerocket132Template()),
	)
}

func TestVSphereKubernetes132BottlerocketTo133StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket132())
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
		provider.WithProviderUpgrade(provider.Bottlerocket133Template()),
	)
}

func TestVSphereKubernetes133BottlerocketTo134StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket133())
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
		provider.WithProviderUpgrade(provider.Bottlerocket134Template()),
	)
}

func TestVSphereKubernetes132Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.RedHat9, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube132,
		provider.WithProviderUpgrade(
			provider.Redhat9132Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes134WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	provider := framework.NewVSphere(t)
	runTestManagementClusterUpgradeSideEffects(t, provider, framework.Ubuntu2004, v1alpha1.Kube134)
}

func TestVSphereKubernetes128To129UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube128, framework.Ubuntu2004, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Ubuntu129Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes129To130UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube129, framework.Ubuntu2004, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Ubuntu130Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes130To131UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube130, framework.Ubuntu2004, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Ubuntu131Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes131To132UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube131, framework.Ubuntu2004, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Ubuntu132Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes132To133UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Ubuntu2004, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube133,
		provider.WithProviderUpgrade(
			provider.Ubuntu133Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes132To133UbuntuInPlaceUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(
		t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Ubuntu2004, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	)
	test.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
	test.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithStackedEtcdTopology(),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
	)
	runInPlaceUpgradeFromReleaseFlow(
		test,
		release,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes128BottlerocketAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithBottleRocket128(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
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

func TestVSphereKubernetes128To129RedhatUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube128, framework.RedHat8, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat129Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes129To130RedhatUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube129, framework.RedHat8, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat130Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes130To131RedhatUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube130, framework.RedHat8, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat131Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes128To129Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube128, framework.RedHat9, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat9129Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes129To130Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube129, framework.RedHat9, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat9130Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes130To131Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube130, framework.RedHat9, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat9131Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes131To132Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube131, framework.RedHat9, release, useBundlesOverride),
	)
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
		provider.WithProviderUpgrade(
			provider.Redhat9132Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes132To133Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.RedHat9, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube133,
		provider.WithProviderUpgrade(
			provider.Redhat9133Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes133To134UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube133, framework.Ubuntu2004, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube134,
		provider.WithProviderUpgrade(
			provider.Ubuntu134Template(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes133To134Redhat9UpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.RedHat),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube133, framework.RedHat9, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube134,
		provider.WithProviderUpgrade(
			provider.Redhat9134Template(), // Set the template so it doesn't get auto-imported
		),
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithLicenseToken(licenseToken),
		),
	)
}

func TestVSphereKubernetes133To134UbuntuInPlaceUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(
		t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube133, framework.Ubuntu2004, release, useBundlesOverride),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	)
	test.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
	test.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithStackedEtcdTopology(),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
	)
	runInPlaceUpgradeFromReleaseFlow(
		test,
		release,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestVSphereKubernetes134MulticlusterWorkloadClusterAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)
	// add licensetoken for k8s version 1.28 and 1.29
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu134(),
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
			vsphere.WithUbuntu131(),
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
			vsphere.WithUbuntu132(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken),
			),
			vsphere.WithUbuntu129(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithUbuntu130(),
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

func TestVSphereKubernetes134UpgradeLabelsTaintsUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithUbuntu134(),
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

func TestVSphereKubernetes134UpgradeWorkerNodeGroupsUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu134(),
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

func TestVSphereKubernetes134MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
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
		vsphere.WithUbuntu134(),
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
			vsphere.WithUbuntu134(),
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
			vsphere.WithUbuntu134(),
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

func TestVSphereKubernetes134CiliumUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu134(),
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
		)
		wc.ApplyClusterManifest()
		wc.ValidateClusterState()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereKubernetes134UpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithBottleRocket134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithBottleRocket134(),
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

func TestVSphereKubernetes134UpgradeWorkerNodeGroupsUbuntuGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu134(),
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

func TestVSphereUpgradeKubernetes134CiliumUbuntuGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu134(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu134(),
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
			vsphere.WithUbuntu134(),
		)
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereKubernetes134UbuntuUpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithLicenseToken(licenseToken),
		),
		provider.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(2))),
		provider.WithNewWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1))),
		provider.WithNewWorkerNodeGroup("worker-3", framework.WithWorkerNodeGroup("worker-3", api.WithCount(1), api.WithLabel("tier", "frontend"))),
		provider.WithUbuntu134(),
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

func TestVSphereKubernetes132to133UpgradeFromLatestMinorReleaseBottleRocketAPI(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t)
	useBundlesOverride := false
	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
		),
		api.VSphereToConfigFiller(
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
		provider.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Bottlerocket1, release, useBundlesOverride),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t, provider, framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	wc.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithManagementCluster(managementCluster.ClusterName),
		),
		api.VSphereToConfigFiller(
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
		provider.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Bottlerocket1, release, useBundlesOverride),
	)

	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPI(
		test,
		release,
		wc.ClusterConfig.Cluster.Spec.KubernetesVersion,
		v1alpha1.Kube133,
		framework.Bottlerocket1,
	)
}

func TestVSphereKubernetes128UbuntuTo129InPlaceUpgrade_1CP_3Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(3),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(api.RemoveEtcdVsphereMachineConfig()),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)

	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestVSphereKubernetes132UbuntuTo133InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(api.RemoveEtcdVsphereMachineConfig()),
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
	)

	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

func TestVSphereKubernetes128UbuntuTo133InPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	var kube129clusterOpts []framework.ClusterE2ETestOpt
	var kube130clusterOpts []framework.ClusterE2ETestOpt
	var kube131clusterOpts []framework.ClusterE2ETestOpt
	var kube132clusterOpts []framework.ClusterE2ETestOpt
	var kube133clusterOpts []framework.ClusterE2ETestOpt
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)
	kube129clusterOpts = append(
		kube129clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
	kube130clusterOpts = append(
		kube130clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu130Template()),
	)
	kube131clusterOpts = append(
		kube131clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu131Template()),
	)
	kube132clusterOpts = append(
		kube132clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
	kube133clusterOpts = append(
		kube133clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
	runInPlaceMultipleUpgradesFlow(
		test,
		kube129clusterOpts,
		kube130clusterOpts,
		kube131clusterOpts,
		kube132clusterOpts,
		kube133clusterOpts,
	)
}

func TestVSphereKubernetes133UbuntuInPlaceCPScaleUp1To3(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithControlPlaneCount(3),
			api.WithInPlaceUpgradeStrategy(),
		),
	)
}

func TestVSphereKubernetes133UbuntuInPlaceCPScaleDown3To1(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithControlPlaneCount(1),
			api.WithInPlaceUpgradeStrategy(),
		),
	)
}

func TestVSphereKubernetes133UbuntuInPlaceWorkerScaleUp1To2(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeCount(2),
			api.WithInPlaceUpgradeStrategy(),
		),
	)
}

func TestVSphereKubernetes133UbuntuInPlaceWorkerScaleDown2To1(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(2),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeCount(1),
			api.WithInPlaceUpgradeStrategy(),
		),
	)
}

func TestVSphereKubernetes128UpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	runUpgradeManagementComponentsFlow(t, release, provider, v1alpha1.Kube128, framework.Ubuntu2004)
}

func TestVSphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade128To129(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewVSphere(t, framework.WithUbuntu128())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithStackedEtcdTopology(),
				api.WithInPlaceUpgradeStrategy(),
				api.WithLicenseToken(licenseToken2),
			),
			api.VSphereToConfigFiller(
				api.RemoveEtcdVsphereMachineConfig(),
			),
			provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
		),
	)
	runInPlaceWorkloadUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu129Template()),
	)
}

func TestVSphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade132To133(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewVSphere(t, framework.WithUbuntu132())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithStackedEtcdTopology(),
			api.WithInPlaceUpgradeStrategy(),
			api.WithLicenseToken(licenseToken),
		),
		api.VSphereToConfigFiller(
			api.RemoveEtcdVsphereMachineConfig(),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithEnvVar(features.VSphereInPlaceEnvVar, "true"),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithStackedEtcdTopology(),
				api.WithInPlaceUpgradeStrategy(),
				api.WithLicenseToken(licenseToken2),
			),
			api.VSphereToConfigFiller(
				api.RemoveEtcdVsphereMachineConfig(),
			),
			provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
		),
	)
	runInPlaceWorkloadUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithInPlaceUpgradeStrategy(),
		),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()),
	)
}

// Workload API
func TestVSphereKubernetes133MulticlusterWorkloadClusterAPI(t *testing.T) {
	vsphere := framework.NewVSphere(t)
	// add licensetoken for k8s version 1.28 and 1.29
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		vsphere.WithUbuntu133(),
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
			vsphere.WithUbuntu131(),
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
			vsphere.WithUbuntu132(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken),
			),
			vsphere.WithUbuntu129(),
		),
		framework.NewClusterE2ETest(
			t, vsphere, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithUbuntu130(),
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

func TestVSphereKubernetes133UpgradeLabelsTaintsUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithUbuntu133(),
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

func TestVSphereKubernetes133UpgradeWorkerNodeGroupsUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu133(),
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

func TestVSphereKubernetes133MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
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
		vsphere.WithUbuntu133(),
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
			vsphere.WithUbuntu133(),
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
			vsphere.WithUbuntu133(),
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

func TestVSphereKubernetes133CiliumUbuntuAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu133(),
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
		)
		wc.ApplyClusterManifest()
		wc.ValidateClusterState()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestVSphereKubernetes133UpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithBottleRocket133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithLabel("key1", "val2"), api.WithTaint(framework.NoScheduleTaint()))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-2", framework.WithWorkerNodeGroup("worker-2", api.WithCount(1), api.WithLabel("key2", "val2"), api.WithTaint(framework.PreferNoScheduleTaint()))),
			vsphere.WithBottleRocket133(),
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

func TestVSphereKubernetes133UpgradeWorkerNodeGroupsUbuntuGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithNewWorkerNodeGroup("worker-1", framework.WithWorkerNodeGroup("worker-1", api.WithCount(1))),
			vsphere.WithUbuntu133(),
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

func TestVSphereUpgradeKubernetes133CiliumUbuntuGitHubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	vsphere := framework.NewVSphere(t)

	managementCluster := framework.NewClusterE2ETest(
		t, vsphere, framework.WithFluxGithubEnvVarCheck(), framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		vsphere.WithUbuntu133(),
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
				api.WithLicenseToken(licenseToken2),
			),
			vsphere.WithNewWorkerNodeGroup("worker-0", framework.WithWorkerNodeGroup("worker-0", api.WithCount(1))),
			vsphere.WithUbuntu133(),
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
			vsphere.WithUbuntu133(),
		)
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

// Airgapped tests
func TestVSphereKubernetes128UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes129UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes130UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes131UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes132UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes133UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes129UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes130UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes131UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes132UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes133UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes134UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)

	runAirgapConfigFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

func TestVSphereKubernetes134UbuntuAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

// Etcd Encryption
func TestVSphereKubernetesUbuntu128EtcdEncryption(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.Ubuntu
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.ValidateEtcdEncryption()
	test.DeleteCluster()
}

func TestVSphereKubernetesUbuntu134EtcdEncryption(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.Ubuntu
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.ValidateEtcdEncryption()
	test.DeleteCluster()
}

func TestVSphereKubernetesBottlerocket128EtcdEncryption(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.Bottlerocket
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestVSphereKubernetesBottlerocket134EtcdEncryption(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.Bottlerocket
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.DeleteCluster()
}

func ubuntu128ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu128(),
	)
}

func bottlerocket128ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket128(),
	)
}

func ubuntu134ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu134(),
	)
}

func bottlerocket134ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket134(),
	)
}

func ubuntu128ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu128(),
	)
}

func bottlerocket128ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket128(),
	)
}

func ubuntu134ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu134(),
	)
}

func bottlerocket134ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket134(),
	)
}

func runVSphereCloneModeFlow(test *framework.ClusterE2ETest, vsphere *framework.VSphere, diskSize int) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	vsphere.ValidateNodesDiskGiB(test.GetCapiMachinesForCluster(test.ClusterName), diskSize)
	test.DeleteCluster()
}

// Bottlerocket Etcd Scale tests
func TestVSphereKubernetes128BottlerocketEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes134BottlerocketEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes128BottlerocketEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes134BottlerocketEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes128to129BottlerocketEtcdScaleUp(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(3),
		),
		provider.WithProviderUpgrade(provider.Bottlerocket129Template()),
	)
}

func TestVSphereKubernetes128to129BottlerocketEtcdScaleDown(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
		),
		provider.WithProviderUpgrade(provider.Bottlerocket129Template()),
	)
}

// Ubuntu Etcd Scale tests
func TestVSphereKubernetes128UbuntuEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes134UbuntuEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes128UbuntuEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu128()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes134UbuntuEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes133to134UbuntuEtcdScaleUp(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(3),
		),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestVSphereKubernetes133to134UbuntuEtcdScaleDown(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
		),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

// Kubelet Configuration tests
func TestVSphereKubernetes129UbuntuKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestVSphereKubernetes134UbuntuKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestVSphereKubernetes129BottlerocketKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestVSphereKubernetes134BottlerocketKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}
