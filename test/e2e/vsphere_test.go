//go:build e2e && (vsphere || all_providers)
// +build e2e
// +build vsphere all_providers

package e2e

import (
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
func TestVSphereKubernetes135BottlerocketAPIServerExtraArgsSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithEnvVar(features.APIServerExtraArgsEnabledEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithControlPlaneAPIServerExtraArgs(),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

// TODO: Investigate why this test takes long time to pass with service-account-issuer flag
func TestVSphereKubernetes135BottlerocketAPIServerExtraArgsUpgradeFlow(t *testing.T) {
	var addAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	var removeAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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
func TestVSphereKubernetes129BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes135BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runAutoImportFlow(test, provider)
}

// AWS IAM Auth
func TestVSphereKubernetes129AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes130AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes131AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes130AWSIamAuthWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
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

func TestVSphereKubernetes135AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes135BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes132To133AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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

func TestVSphereKubernetes134To135AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes135Template()),
	)
}

// AWS IAM Auth Add/Remove Tests
func TestVSphereKubernetes134UbuntuAddAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithAwsIamEnvVarCheck(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runUpgradeFlowAddAWSIamAuth(test, v1alpha1.Kube134)
}

func TestVSphereKubernetes135UbuntuAddAWSIamAuthUpgrade(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithAwsIamEnvVarCheck(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runUpgradeFlowAddAWSIamAuth(test, v1alpha1.Kube135)
}

// Curated Packages
func TestVSphereKubernetes129CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes135CuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes129CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes135CuratedPackagesWithProxyConfigFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
func TestVSphereKubernetes129CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes135CuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// Harbor
// Harbor
func TestVSphereKubernetes129CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes130CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes131CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes132CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes133CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes134CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes135CuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// ADOT
// ADOT
func TestVSphereKubernetes129CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes130CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes131CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes132CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes133CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes134CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes135CuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

// Cluster Autoscaler
// Cluster Autoscaler
func TestVSphereKubernetes129UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
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
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
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
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes135UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
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
		),
	)
	runAutoscalerUpgradeFlow(test)
}

// Prometheus
func TestVSphereKubernetes129UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes129BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes130BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes131BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes132BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes133BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes134UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes135UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Workload Cluster Curated Packages

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes135UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes134BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes135UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes130UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes131UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes132UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes133UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
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

func TestVSphereKubernetes135UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket135())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube135)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

// Download Artifacts
func TestVSphereDownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestVSphereKubernetes129GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes135GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes135GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135BottleRocketGithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes135BottleRocketGitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestVSphereKubernetes129To130GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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

func TestVSphereKubernetes134To135GitFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes135Template()),
	)
}

func TestVSphereInstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
func TestVSphereKubernetes129UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := ubuntu129ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes135UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := ubuntu135ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes129BottlerocketLabelsUpgradeFlow(t *testing.T) {
	provider := bottlerocket129ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes135BottlerocketLabelsUpgradeFlow(t *testing.T) {
	provider := bottlerocket135ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// Multicluster
func TestVSphereKubernetes129MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube129),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube129),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestVSphereKubernetes135MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204135())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube135),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube135),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

// OIDC
func TestVSphereKubernetes129OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes135OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestVSphereKubernetes132To133OIDCUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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

func TestVSphereKubernetes134To135OIDCUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(provider.Ubuntu2204Kubernetes135Template()),
	)
}

// Proxy Config
func TestVSphereKubernetes129UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130(),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131(),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132(),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestVSphereKubernetes135UbuntuProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135BottlerocketProxyConfigFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135(),
			framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Registry Mirror
func TestVSphereKubernetes133UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes135UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes135UbuntuRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135BottlerocketRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithRegistryMirrorEndpointAndCert(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes135UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135BottlerocketRegistryMirrorOciNamespaces(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithRegistryMirrorOciNamespaces(constants.VSphereProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestVSphereKubernetes129UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes134UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes130UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes131UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes132UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes133UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

func TestVSphereKubernetes135UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithAuthenticatedRegistryMirror(constants.VSphereProviderName),
	)
	runCuratedPackageInstallSimpleFlowRegistryMirror(test)
}

// Clone mode
func TestVSphereKubernetes129FullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu2204129(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes135FullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu2204135(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes129LinkedClone(t *testing.T) {
	diskSize := 20
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu2204129(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes135LinkedClone(t *testing.T) {
	diskSize := 20
	vsphere := framework.NewVSphere(t,
		framework.WithUbuntu2204135(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes129BottlerocketFullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket129(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes135BottlerocketFullClone(t *testing.T) {
	diskSize := 30
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket135(),
		framework.WithFullCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes129BottlerocketLinkedClone(t *testing.T) {
	diskSize := 22
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket129(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

func TestVSphereKubernetes135BottlerocketLinkedClone(t *testing.T) {
	diskSize := 22
	vsphere := framework.NewVSphere(t,
		framework.WithBottleRocket135(),
		framework.WithLinkedCloneMode(),
		framework.WithDiskGiBForAllMachines(diskSize),
	)

	test := framework.NewClusterE2ETest(
		t,
		vsphere,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runVSphereCloneModeFlow(test, vsphere, diskSize)
}

// Simple Flow
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
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
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

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
}

func TestVSphereKubernetes134BottlerocketNetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Bottlerocket1, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
}

func TestVSphereKubernetes134Redhat9NetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.RedHat9, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
}

func TestVSphereKubernetes135Ubuntu2204NetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
}

func TestVSphereKubernetes135BottlerocketNetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Bottlerocket1, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
}

func TestVSphereKubernetes135Redhat9NetworksSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t,
		// Add worker node group with second network config
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
			framework.WithSecondNetwork(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.RedHat9, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Create cluster and validate the network is up
	runSimpleFlowWithSecondNetworkValidation(test, worker0)
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

func TestVSphereKubernetes135Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(api.WithLicenseToken(licenseToken)),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes135Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
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

func TestVSphereKubernetes135RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9135VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes135Ubuntu2204ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(5),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes135Ubuntu2404ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(5),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes135RedHat9ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithRedHat9135VSphere()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes135BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes130DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes131DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes132DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes133DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes134DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
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

func TestVSphereKubernetes135BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
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

func TestVSphereKubernetes135Ubuntu2204DifferentNamespaceSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithClusterNamespace(clusterNamespace),
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestVSphereKubernetes135BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135(),
			framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestVSphereKubernetes129CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

// Cilium Helm Values Tests - Ubuntu
func TestVSphereKubernetes129CiliumHelmValuesUbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
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
func TestVSphereKubernetes129CiliumHelmValuesPrecedenceUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
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
func TestVSphereKubernetes129CiliumHelmValuesComplexConfigUbuntuFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133()),
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
func TestVSphereKubernetes129BottleRocketWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket129(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runNTPFlow(test, v1alpha1.Bottlerocket)
}

func TestVSphereKubernetes135BottleRocketWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket135(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runNTPFlow(test, v1alpha1.Bottlerocket)
}

func TestVSphereKubernetes129UbuntuWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithUbuntu2204129(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runNTPFlow(test, v1alpha1.Ubuntu)
}

func TestVSphereKubernetes135UbuntuWithNTP(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithUbuntu2204135(),
			framework.WithNTPServersForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runNTPFlow(test, v1alpha1.Ubuntu)
}

// Bottlerocket Configuration tests
func TestVSphereKubernetes129BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket129(),
			framework.WithBottlerocketKubernetesSettingsForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runBottlerocketConfigurationFlow(test)
}

func TestVSphereKubernetes135BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(
			t, framework.WithBottleRocket135(),
			framework.WithBottlerocketKubernetesSettingsForAllMachines(),
			framework.WithSSHAuthorizedKeyForAllMachines(""), // set SSH key to empty
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
	)
	runBottlerocketConfigurationFlow(test)
}

// Stacked Etcd
func TestVSphereKubernetes129StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestVSphereKubernetes135StackedEtcdUbuntu(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestVSphereKubernetes129UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := ubuntu129ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes135UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := ubuntu135ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes129BottlerocketTaintsUpgradeFlow(t *testing.T) {
	provider := bottlerocket129ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes135BottlerocketTaintsUpgradeFlow(t *testing.T) {
	provider := bottlerocket135ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestVSphereKubernetes129UbuntuWorkloadClusterTaintsFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())

	managementCluster := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
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
				api.WithKubernetesVersion(v1alpha1.Kube129),
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
func TestVSphereKubernetes129UbuntuTo130Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes130Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes131Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes132Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes133Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes130Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes131Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes132Template()),
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes133Template()),
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

func TestVSphereKubernetes129Ubuntu2204To2404Upgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes129Template()),
	)
}

func TestVSphereKubernetes130Ubuntu2204To2404Upgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes130Template()),
	)
}

func TestVSphereKubernetes131Ubuntu2204To2404Upgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes131Template()),
	)
}

func TestVSphereKubernetes132Ubuntu2204To2404Upgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes132Template()),
	)
}

func TestVSphereKubernetes133Ubuntu2204To2404Upgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(provider.Ubuntu2404Kubernetes133Template()),
	)
}

func TestVSphereKubernetes132UbuntuTo133InPlaceUpgradeWorkerOnly(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
		provider.WithKubeVersionAndOSMachineConfig(providers.GetControlPlaneNodeName(test.ClusterName), kube133, framework.Ubuntu2204),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube133)),
		provider.WithProviderUpgrade(provider.Ubuntu133Template()), // this will just set everything to 1.33 as expected
	)
}

func TestVSphereKubernetes129UbuntuTo130MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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

func TestVSphereKubernetes129UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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

func TestVSphereKubernetes129UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204134())
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

func TestVSphereKubernetes129UbuntuTo130StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204130())
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204131())
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

func TestVSphereKubernetes133WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	provider := framework.NewVSphere(t)
	runTestManagementClusterUpgradeSideEffects(t, provider, framework.Ubuntu2204, v1alpha1.Kube133)
}

func TestVSphereKubernetes129To130UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	release := latestMinorRelease(t)
	useBundlesOverride := false
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube129, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube130, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube131, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube132, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube133, framework.Ubuntu2204, release, useBundlesOverride),
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
		framework.WithKubeVersionAndOSForRelease(v1alpha1.Kube133, framework.Ubuntu2204, release, useBundlesOverride),
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

func TestVSphereKubernetes135UpgradeLabelsTaintsUbuntuAPI(t *testing.T) {
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
		vsphere.WithUbuntu2204135(),
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
			vsphere.WithUbuntu2204135(),
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

func TestVSphereKubernetes135UpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T) {
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
		vsphere.WithBottleRocket135(),
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
			vsphere.WithBottleRocket135(),
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

func TestVSphereKubernetes133UbuntuTo134InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
	)

	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Ubuntu134Template()),
	)
}

func TestVSphereKubernetes133RedHat9To134InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithRedHat9133VSphere())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.RedHat9, nil),
	)

	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Redhat9134Template()),
	)
}

func TestVSphereKubernetes133UbuntuInPlaceCPScaleUp1To3(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
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
	provider := framework.NewVSphere(t, framework.WithUbuntu2204133())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
	)
	runInPlaceUpgradeFlow(
		test,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeCount(1),
			api.WithInPlaceUpgradeStrategy(),
		),
	)
}

func TestVSphereKubernetes129UpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu2204129())
	runUpgradeManagementComponentsFlow(t, release, provider, v1alpha1.Kube129, framework.Ubuntu2204)
}

func TestVSphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade132To133(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewVSphere(t, framework.WithUbuntu2204132())
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
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
			provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
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
func TestVSphereKubernetes129UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204129(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204130(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204131(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204132(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204133(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
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
		framework.NewVSphere(t, framework.WithUbuntu2204134(), framework.WithPrivateNetwork()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "195.18.0.1/16,196.18.0.1/16")
}

// Etcd Encryption
func TestVSphereKubernetesUbuntu134EtcdEncryption(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
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

func ubuntu129ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu2204129(),
	)
}

func bottlerocket129ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket129(),
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
		framework.WithUbuntu2204134(),
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

func ubuntu129ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu2204129(),
	)
}

func bottlerocket129ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket129(),
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
		framework.WithUbuntu2204134(),
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

func ubuntu135ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu2204135(),
	)
}

func bottlerocket135ProviderWithLabels(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket135(),
	)
}

func ubuntu135ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithUbuntu2204135(),
	)
}

func bottlerocket135ProviderWithTaints(t *testing.T) *framework.VSphere {
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
		framework.WithBottleRocket135(),
	)
}

func runVSphereCloneModeFlow(test *framework.ClusterE2ETest, vsphere *framework.VSphere, diskSize int) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	vsphere.ValidateNodesDiskGiB(test.GetCapiMachinesForCluster(test.ClusterName), diskSize)
	test.DeleteCluster()
}

// Bottlerocket Etcd Scale tests
func TestVSphereKubernetes129BottlerocketEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes135BottlerocketEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes129BottlerocketEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket129()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes135BottlerocketEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket135()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

// Ubuntu Etcd Scale tests
func TestVSphereKubernetes129UbuntuEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes135UbuntuEtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestVSphereKubernetes129UbuntuEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

func TestVSphereKubernetes135UbuntuEtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204135()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube135),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

// Kubelet Configuration tests
func TestVSphereKubernetes129UbuntuKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestVSphereKubernetes134UbuntuKubeletConfiguration(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu2204134()),
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
