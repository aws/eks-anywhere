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
func TestCloudStackKubernetes131RedHat8APIServerExtraArgsSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat131()),
		framework.WithEnvVar(features.APIServerExtraArgsEnabledEnvVar, "true"),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneAPIServerExtraArgs(),
		),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}

// TODO: Investigate why this test takes long time to pass with service-account-issuer flag
func TestCloudStackKubernetes131Redhat8APIServerExtraArgsUpgradeFlow(t *testing.T) {
	var addAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	var removeAPIServerExtraArgsclusterOpts []framework.ClusterE2ETestOpt
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat131()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
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
func TestCloudStackKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

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

func TestCloudStackKubernetes127to128AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
	)
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

// Curated Packages
func TestCloudStackKubernetes127RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

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

// Emissary
func TestCloudStackKubernetes127RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

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

// Harbor
func TestCloudStackKubernetes127RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

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

// Workload Cluster Curated Packages
func TestCloudStackKubernetes127RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

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

func TestCloudStackKubernetes127RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
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

// Cert Manager
func TestCloudStackKubernetes127RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

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

// ADOT
func TestCloudStackKubernetes127RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

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

func TestCloudStackKubernetes127RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
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

// Cluster Autoscaler
func TestCloudStackKubernetes127RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

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

// Prometheus
func TestCloudStackKubernetes127RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

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
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes131()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestCloudStackKubernetes127GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

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

func TestCloudStackKubernetes127GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
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

func TestCloudStackKubernetes127To128GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
	)
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

func TestCloudStackKubernetes127InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
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

func TestCloudStackKubernetes128UpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128())
	runUpgradeManagementComponentsFlow(t, release, provider, v1alpha1.Kube128, framework.RedHat9)
}

// Labels
func TestCloudStackKubernetes127LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes127(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
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

func TestCloudStackKubernetes127RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat127ProviderWithLabels(t)

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

func redhat127ProviderWithLabels(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat9Kubernetes127(),
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

// Multicluster
func TestCloudStackKubernetes127MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
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

func TestCloudStackUpgradeKubernetes128MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
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
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube127),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat9Kubernetes128Template(),
		),
	)
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

// OIDC
func TestCloudStackKubernetes127WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, framework.RedHat9, anywherev1.Kube127)
}

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

func TestCloudStackKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
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

// Proxy Config
func TestCloudStackKubernetes127RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

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

// Proxy Config Multicluster
func TestCloudStackKubernetes127RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat9Kubernetes127(),
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
			),
			cloudstack.WithRedhat9Kubernetes127(),
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

func TestCloudStackKubernetes128RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
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
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
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
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
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
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
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

// Registry Mirror
func TestCloudStackKubernetes127RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

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

func TestCloudStackKubernetes127RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
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

func TestCloudStackKubernetes127RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes128RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simple Flow
func TestCloudStackKubernetes127RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

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

func TestCloudStackKubernetes127RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
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

func TestCloudStackKubernetes127ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
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

func TestCloudStackKubernetes127MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
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

func TestCloudStackKubernetes127DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
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

// Cilium Policy
func TestCloudStackKubernetes127CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

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

// Stacked Etcd
func TestCloudStackKubernetes127StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

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

// Taints
func TestCloudStackKubernetes127RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat127ProviderWithTaints(t)

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

func redhat127ProviderWithTaints(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat9Kubernetes127(),
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

// Upgrade
func TestCloudStackKubernetes127RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat9Kubernetes127(),
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
		provider.WithNewCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup(
				"workers-3",
				api.WithCount(1),
			),
		),
	)
}

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

func TestCloudStackKubernetes127To128Redhat8UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat128Template()),
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

func TestCloudStackKubernetes127To128Redhat8StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat128Template()),
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

func TestCloudStackKubernetes127To128Redhat9UnstackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
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

func TestCloudStackKubernetes127To128Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
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

func TestCloudStackKubernetes127Redhat8ToRedhat9Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes127Template()),
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
func TestCloudStackKubernetes127RedhatTo128UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes127Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupResourcesVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube128,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

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

func TestCloudStackKubernetes127RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
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

func TestCloudStackKubernetes127RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
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

func TestCloudStackKubernetes127To128RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat9Kubernetes128Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
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

// This test is skipped as registry mirror was not configured for CloudStack
func TestCloudStackKubernetes128RedhatAirgappedProxy(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat9Kubernetes128(),
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
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)

	runAirgapConfigProxyFlow(test, "10.0.0.1/8")
}

// Workload API
func TestCloudStackKubernetes131MulticlusterWorkloadClusterAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes131(),
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
			cloudstack.WithRedhat9Kubernetes131(),
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
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131MulticlusterWorkloadClusterNewCredentialsSecretsAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes131(),
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
		cloudstack.WithRedhat9Kubernetes131(),
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
		cloudstack.WithRedhat9Kubernetes131(),
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
func TestCloudStackKubernetes131MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat9Kubernetes131(),
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
			cloudstack.WithRedhat9Kubernetes131(),
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
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131MulticlusterWorkloadClusterNewCredentialsSecretGitHubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat9Kubernetes131(),
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
		cloudstack.WithRedhat9Kubernetes131(),
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
		cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131WorkloadClusterAWSIamAuthAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes131(),
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
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131WorkloadClusterAWSIamAuthGithubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat9Kubernetes131(),
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
			),
			framework.WithAwsIamConfig(),
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131WorkloadClusterOIDCAuthAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat9Kubernetes131(),
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
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes131WorkloadClusterOIDCAuthGithubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat9Kubernetes131(),
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
			),
			framework.WithOIDCClusterConfig(t),
			cloudstack.WithRedhat9Kubernetes131(),
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

func TestCloudStackKubernetes129EtcdEncryption(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat129())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube129),
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

func TestCloudStackKubernetes127To128RedHatManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)
	runUpgradeFlowWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(3),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.RedHat9, nil),
	)
}

func TestCloudStackKubernetes128ValidateDomainFourLevelsSimpleFlow(t *testing.T) {
	provider := framework.NewCloudStack(
		t,
		framework.WithCloudStackRedhat9Kubernetes128(),
		framework.WithCloudStackFillers(
			framework.RemoveAllCloudStackAzs(),
			framework.UpdateAddCloudStackAz4(),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
		),
	)
	runSimpleFlow(test)
}

// Etcd Scale tests
func TestCloudStackKubernetes128EtcdScaleUp(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
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

func TestCloudStackKubernetes128EtcdScaleDown(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes128()),
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

func TestCloudStackKubernetes127to128EtcdScaleUp(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(3),
		),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
	)
}

func TestCloudStackKubernetes127to128EtcdScaleDown(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithExternalEtcdTopology(3),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithExternalEtcdTopology(1),
		),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes128Template()),
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
