//go:build e2e && (cloudstack || all_providers)
// +build e2e
// +build cloudstack all_providers

package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

// APIServerExtraArgs
func TestCloudStackKubernetes134RedHat9APIServerExtraArgsSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
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
func TestCloudStackKubernetes134Redhat9APIServerExtraArgsUpgradeFlow(t *testing.T) {
	var addAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	var removeAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
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

// AWS IAM Auth
func TestCloudStackKubernetes128AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes129AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes130AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes131AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes132AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes133AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes134AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes128to129AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()),
	)
}

func TestCloudStackKubernetes129to130AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()),
	)
}

func TestCloudStackKubernetes130to131AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

func TestCloudStackKubernetes131to132AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()),
	)
}

func TestCloudStackKubernetes132to133AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()),
	)
}

func TestCloudStackKubernetes133to134AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()),
	)
}

// Curated Packages
func TestCloudStackKubernetes128RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

// Emissary
func TestCloudStackKubernetes128RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

// Harbor
func TestCloudStackKubernetes128RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

// Workload Cluster Curated Packages
func TestCloudStackKubernetes128RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes128RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

// Cert Manager
func TestCloudStackKubernetes128RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube128)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube129)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube130)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube131)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube132)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube133)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube134)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

// ADOT
func TestCloudStackKubernetes128RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes128RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

// Cluster Autoscaler
func TestCloudStackKubernetes128RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes129RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes130RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes131RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes132RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes133RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes134RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

// Prometheus
func TestCloudStackKubernetes128RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes129RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube129),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes130RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube130),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes131RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube131),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes132RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes133RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes134RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube134),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Download Artifacts
func TestCloudStackDownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat131()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runDownloadArtifactsFlow(test)
}

func TestCloudStackRedhat9DownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestCloudStackKubernetes128GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes129GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes130GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes131GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes132GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes128GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes129GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes130GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes131GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes132GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes133GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes133GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes134GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes134GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes129To130GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()),
	)
}

func TestCloudStackKubernetes130To131GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

func TestCloudStackKubernetes131To132GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()),
	)
}

func TestCloudStackKubernetes132To133GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()),
	)
}

func TestCloudStackKubernetes133To134GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()),
	)
}

func TestCloudStackKubernetes128InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube128,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes129InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube129,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes130InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube130,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes131InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube131,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes132InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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

func TestCloudStackKubernetes133InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube133,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes134InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube134,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes134UpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	runUpgradeManagementComponentsFlow(t, release, provider, v1alpha1.Kube134, framework.RedHat9)
}

// Labels
func TestCloudStackKubernetes128LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes128(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes129LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes129(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes130LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes130(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes131LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes131(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes132LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes132(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes133LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes133(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes134LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes134(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithControlPlaneLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(constants.FailureDomainLabelName, constants.CloudstackFailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels, framework.ValidateControlPlaneNodeNameMatchCAPIMachineName)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels, framework.ValidateWorkerNodeNameMatchCAPIMachineName)
	test.DeleteCluster()
}

func TestCloudStackKubernetes128RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat128ProviderWithLabels(t)

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

func TestCloudStackKubernetes129RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat129ProviderWithLabels(t)

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

func TestCloudStackKubernetes130RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat130ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes131RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat131ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes132RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat132ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

func TestCloudStackKubernetes133RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat133ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes134RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat134ProviderWithLabels(t)

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

func redhat128ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes128(),
	)
}

func redhat129ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes129(),
	)
}

func redhat130ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes130(),
	)
}

func redhat131ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes131(),
	)
}

func redhat132ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes132(),
	)
}

func redhat133ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes133(),
	)
}

func redhat134ProviderWithLabels(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithCloudStackRedhat9Kubernetes134(),
	)
}

func redhat133ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes133(),
	)
}

func redhat134ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes134(),
	)
}

// Multicluster
func TestCloudStackKubernetes128MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
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

func TestCloudStackKubernetes129MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
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

func TestCloudStackKubernetes130MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube130),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes131MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes132MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes133MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes134MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
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

func TestCloudStackUpgradeKubernetes129MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
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
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube129),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes129Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes130MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
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
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube129),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes130Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes131MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
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
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube130),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes131Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes132MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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
				api.WithStackedEtcdTopology(),
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
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes132Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes133MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes133Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes134MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes134Template(),
		),
	)
}

// OIDC
func TestCloudStackKubernetes128WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube128)
}

func TestCloudStackKubernetes129WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube129)
}

func TestCloudStackKubernetes130WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube130)
}

func TestCloudStackKubernetes131WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube131)
}

func TestCloudStackKubernetes132WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube132)
}

func TestCloudStackKubernetes133WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube133)
}

func TestCloudStackKubernetes134WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube134)
}

func TestCloudStackKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes129OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes130OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes131OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes132OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes133OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes134OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes130To131OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

func TestCloudStackKubernetes131To132OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()),
	)
}

func TestCloudStackKubernetes132To133OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()),
	)
}

func TestCloudStackKubernetes133To134OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()),
	)
}

// Proxy Config
func TestCloudStackKubernetes128RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes129RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes130RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes131RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes132RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes133RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes134RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Proxy Config Multicluster
func TestCloudStackKubernetes128RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes128(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes128(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes129RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes129(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes129(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes130RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes130(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes130(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes131RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes131(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes131(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes132RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes132(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes132(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes133RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

func TestCloudStackKubernetes134RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(licenseToken2),
			),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
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

// Registry Mirror
func TestCloudStackKubernetes128RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes129RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes130RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes131RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes132RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes128RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes129RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes130RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes131RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes132RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes132RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes133RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes133RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes133RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes134RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes134RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes134RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simple Flow
func TestCloudStackKubernetes128RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes128RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes132RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes133RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes133ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes133MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes133DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes128ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes132ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes128MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes132MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes128DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes132DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

// Cilium Policy
func TestCloudStackKubernetes128CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes129CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes130CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes131CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes132CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes133CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

// Stacked Etcd
func TestCloudStackKubernetes128StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes129StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes130StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes131StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes132StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes133StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes134StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestCloudStackKubernetes128RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat128ProviderWithTaints(t)

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

func TestCloudStackKubernetes129RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat129ProviderWithTaints(t)

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

func TestCloudStackKubernetes130RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat130ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes131RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat131ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes132RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat132ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
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

func TestCloudStackKubernetes133RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat133ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes134RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat134ProviderWithTaints(t)

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

func redhat128ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes128(),
	)
}

func redhat129ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes129(),
	)
}

func redhat130ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes130(),
	)
}

func redhat131ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes131(),
	)
}

func redhat132ProviderWithTaints(t *testing.T) *framework.CloudStack {
	return framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 2),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithCloudStackRedhat9Kubernetes132(),
	)
}

// Upgrade
func TestCloudStackKubernetes128RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes128(),
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
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes129RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes129(),
	)
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

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes130RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes130(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube130),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes131RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes131(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes132RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes132(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes133RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes133(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes134RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes134(),
	)
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

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup("workers-2"),
			api.WithWorkerNodeGroup("workers-1", api.WithCount(1)),
		),
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

func TestCloudStackKubernetes134RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
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

func TestCloudStackKubernetes134RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
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

func TestCloudStackKubernetes132To133Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()),
	)
}

func TestCloudStackKubernetes133To134Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()),
	)
}

func TestCloudStackKubernetes132To133Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()),
	)
}

func TestCloudStackKubernetes133To134Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()),
	)
}

func TestCloudStackKubernetes128To129Redhat8UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat129Template()),
	)
}

func TestCloudStackKubernetes129To130Redhat8UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat130Template()),
	)
}

func TestCloudStackKubernetes130To131Redhat8UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat131Template()),
	)
}

func TestCloudStackKubernetes128To129Redhat8StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat129Template()),
	)
}

func TestCloudStackKubernetes129To130Redhat8StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat130Template()),
	)
}

func TestCloudStackKubernetes130To131Redhat8StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat131Template()),
	)
}

func TestCloudStackKubernetes128To129Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()),
	)
}

func TestCloudStackKubernetes129To130Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()),
	)
}

func TestCloudStackKubernetes130To131Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

func TestCloudStackKubernetes131To132Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()),
	)
}

func TestCloudStackKubernetes128To129Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()),
	)
}

func TestCloudStackKubernetes129To130Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()),
	)
}

func TestCloudStackKubernetes130To131Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

func TestCloudStackKubernetes131To132Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()),
	)
}

func TestCloudStackKubernetes128Redhat8ToRedhat9Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat128())
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
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
	)
}

func TestCloudStackKubernetes129Redhat8ToRedhat9Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat129())
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
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()),
	)
}

func TestCloudStackKubernetes130Redhat8ToRedhat9Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat130())
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
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()),
	)
}

func TestCloudStackKubernetes131Redhat8ToRedhat9Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat131())
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
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()),
	)
}

// TODO: investigate these tests further as they pass even without the expected behavior(upgrade should fail the first time and continue from the checkpoint on second upgrade)
func TestCloudStackKubernetes128RedhatTo129UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube129,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes129RedhatTo130UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes129Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube130,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes130RedhatTo131UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes130Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube131,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes131RedhatTo132UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes131Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube132,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes132RedhatTo133UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes132Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube133,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes128RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
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

func TestCloudStackKubernetes129RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
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

func TestCloudStackKubernetes130RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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

func TestCloudStackKubernetes131RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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

func TestCloudStackKubernetes132RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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

func TestCloudStackKubernetes128RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
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

func TestCloudStackKubernetes129RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
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

func TestCloudStackKubernetes130RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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

func TestCloudStackKubernetes131RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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

func TestCloudStackKubernetes132RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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

func TestCloudStackKubernetes128To129RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes129Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes129To130RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes130Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes130To131RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes131Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes131To132RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes132Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes130To131StackedEtcdRedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes130())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes131Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes131To132StackedEtcdRedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes132Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes132To133RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes133Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes132To133StackedEtcdRedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes133Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes133To134RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes134Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes133To134StackedEtcdRedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes134Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes133RedhatTo134UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes133Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes134Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube134,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

// This test is skipped as registry mirror was not configured for CloudStack
func TestCloudStackKubernetes134RedhatAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes134(),
			framework.WithCloudStackFillers(
				framework.RemoveAllCloudStackAzs(),
				framework.UpdateAddCloudStackAz3(),
			),
		),
		framework.WithClusterFiller(
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "10.0.0.1/8")
}

// Workload API
func TestCloudStackKubernetes134MulticlusterWorkloadClusterAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()

		tests := cloudStackAPIWorkloadUpgradeTests(wc, cloudstack)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				runCloudStackAPIWorkloadUpgradeTest(t, wc, tt)
			})
		}

		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134MulticlusterWorkloadClusterNewCredentialsSecretsAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	))

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	))

	test.CreateManagementCluster()
	test.ManagementCluster.CreateCloudStackCredentialsSecretFromEnvVar("test-creds", framework.CloudStackCredentialsAz1())

	// Create workload clusters
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

func TestCloudStackKubernetes133MulticlusterWorkloadClusterAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes132(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()

		tests := cloudStackAPIWorkloadUpgradeTests(wc, cloudstack)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				runCloudStackAPIWorkloadUpgradeTest(t, wc, tt)
			})
		}

		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133MulticlusterWorkloadClusterNewCredentialsSecretsAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	))

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	))

	test.CreateManagementCluster()
	test.ManagementCluster.CreateCloudStackCredentialsSecretFromEnvVar("test-creds", framework.CloudStackCredentialsAz1())

	// Create workload clusters
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

// Workload GitOps API
func TestCloudStackKubernetes134MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes134(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		tests := cloudStackAPIWorkloadUpgradeTests(wc, cloudstack)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				runCloudStackAPIWorkloadUpgradeTestWithFlux(t, test, wc, tt)
			})
		}

		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134MulticlusterWorkloadClusterNewCredentialsSecretGitHubFluxAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes134(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	))

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	))

	test.CreateManagementCluster()
	test.ManagementCluster.CreateCloudStackCredentialsSecretFromEnvVar("test-creds", framework.CloudStackCredentialsAz1())

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134WorkloadClusterAWSIamAuthAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),

			framework.WithAwsIamEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateAWSIamAuth()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134WorkloadClusterAWSIamAuthGithubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes134(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithAwsIamEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateAWSIamAuth()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134WorkloadClusterOIDCAuthAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes134(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithOIDCEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateOIDC()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134WorkloadClusterOIDCAuthGithubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes134(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithOIDCEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes134(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateOIDC()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes134EtcdEncryption(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.RedHat
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.ValidateEtcdEncryption()
	test.DeleteCluster()
}

func TestCloudStackKubernetes134ValidateDomainFourLevelsSimpleFlow(t *testing.T) {
	provider := framework.NewCloudStack(
		t,
		framework.WithCloudStackRedhat9Kubernetes134(),
		framework.WithCloudStackFillers(
			framework.RemoveAllCloudStackAzs(),
			framework.UpdateAddCloudStackAz4(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube134),
		),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes134EtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
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

func TestCloudStackKubernetes134EtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
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

func TestCloudStackKubernetes134KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes134()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestCloudStackKubernetes133MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes133(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithExternalEtcdTopology(1),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithControlPlaneCount(1),
			),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		tests := cloudStackAPIWorkloadUpgradeTests(wc, cloudstack)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				runCloudStackAPIWorkloadUpgradeTestWithFlux(t, test, wc, tt)
			})
		}

		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133MulticlusterWorkloadClusterNewCredentialsSecretGitHubFluxAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes133(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	))

	test.WithWorkloadClusters(framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
		),
		api.CloudStackToConfigFiller(
			api.WithCloudStackCredentialsRef("test-creds"),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	))

	test.CreateManagementCluster()
	test.ManagementCluster.CreateCloudStackCredentialsSecretFromEnvVar("test-creds", framework.CloudStackCredentialsAz1())

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133WorkloadClusterAWSIamAuthAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),

			framework.WithAwsIamEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateAWSIamAuth()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133WorkloadClusterAWSIamAuthGithubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes133(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithAwsIamEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateAWSIamAuth()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133WorkloadClusterOIDCAuthAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes133(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithOIDCEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateOIDC()
		wc.DeleteClusterWithKubectl()
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133WorkloadClusterOIDCAuthGithubFluxAPI(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
			api.WithLicenseToken(licenseToken),
		),
		cloudstack.WithRedhat9Kubernetes133(),
		framework.WithFluxGithubConfig(),
	)

	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			cloudstack,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithOIDCEnvVarCheck(),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithStackedEtcdTopology(),
				api.WithLicenseToken(licenseToken2),
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes133(),
		),
	)

	test.CreateManagementCluster()

	// Create workload clusters
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		test.PushWorkloadClusterToGit(wc)
		wc.WaitForKubeconfig()
		wc.ValidateClusterState()
		wc.ValidateOIDC()
		test.DeleteWorkloadClusterFromGit(wc)
		wc.ValidateClusterDelete()
	})

	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}

func TestCloudStackKubernetes133EtcdEncryption(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
		),
		framework.WithPodIamConfig(),
	)
	test.OSFamily = v1alpha1.RedHat
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.PostClusterCreateEtcdEncryptionSetup()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{framework.WithEtcdEncrytion()})
	test.StopIfFailed()
	test.ValidateEtcdEncryption()
	test.DeleteCluster()
}

func TestCloudStackKubernetes133ValidateDomainFourLevelsSimpleFlow(t *testing.T) {
	provider := framework.NewCloudStack(
		t,
		framework.WithCloudStackRedhat9Kubernetes133(),
		framework.WithCloudStackFillers(
			framework.RemoveAllCloudStackAzs(),
			framework.UpdateAddCloudStackAz4(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
		),
	)
	runSimpleFlow(test)
}

// Etcd Scale tests
func TestCloudStackKubernetes133EtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(3),
		),
	)
}

func TestCloudStackKubernetes133EtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.WithExternalEtcdTopology(1),
		),
	)
}

// Kubelet Configuration tests
func TestCloudStackKubernetes129KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat129()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestCloudStackKubernetes130KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat130()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestCloudStackKubernetes131KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestCloudStackKubernetes132KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes132()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}

func TestCloudStackKubernetes133KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes133()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationFlow(test)
}
