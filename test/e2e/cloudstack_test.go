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

// AWS IAM Auth
func TestCloudStackKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes124AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes125AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes126AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes127AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

// TODO: Add TestCloudStackKubernetes123to124AWSIamAuthUpgrade
func TestCloudStackKubernetes123to124AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124to125AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125to126AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126to127AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

// Curated packages test
func TestCloudStackKubernetes123RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesHarborSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallSimpleFlowLocalStorageProvisioner(test)
}

func TestCloudStackKubernetes123RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube124)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube125)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube126)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube124)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube125)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube126)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube124)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube125)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube126)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube127)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesAdotUpdateFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallUpdateFlow(test)
}

func TestCloudStackKubernetes123RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes124RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes125RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes126RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes127RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes124RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes125RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes126RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

func TestCloudStackKubernetes127RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Download artifacts
func TestCloudStackDownloadArtifacts(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runDownloadArtifactsFlow(test)
}

// Flux
func TestCloudStackKubernetes123GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes124GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes125GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes126GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes127GithubFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithFluxGithub(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes123GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes124GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes125GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes126GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes127GitFlux(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

// TODO: Add 1.23 to 1.24 flux upgrade tests
func TestCloudStackKubernetes123To124GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124To125GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125To126GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126To127GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

func TestCloudStackKubernetes123InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube123,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes124InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes125InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube125,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes126InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube126,
		framework.WithFluxGit(),
		framework.WithClusterUpgrade(api.WithGitOpsRef(framework.DefaultFluxConfigName, v1alpha1.FluxConfigKind)),
	)
}

func TestCloudStackKubernetes127InstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
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
func TestCloudStackKubernetes123LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat123(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
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

func TestCloudStackKubernetes124LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat124(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
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

func TestCloudStackKubernetes125LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat125(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
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

func TestCloudStackKubernetes126LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat126(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube126),
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

func TestCloudStackKubernetes127LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat127(),
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

func TestCloudStackKubernetes123RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat123ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes124RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat124ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes125RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat125ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes126RedhatLabelsUpgradeFlow(t *testing.T) {
	provider := redhat126ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube126),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
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

func redhat123ProviderWithLabels(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat123(),
	)
}

func redhat124ProviderWithLabels(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat124(),
	)
}

func redhat125ProviderWithLabels(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat125(),
	)
}

func redhat126ProviderWithLabels(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat126(),
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
		framework.WithCloudStackRedhat127(),
	)
}

// Multicluster
func TestCloudStackKubernetes123MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes124MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube124),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes125MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes126MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
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
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes127MulticlusterWorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
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

// WL cluster with prev release version from management cluster
func TestCloudStackKubernetes123MulticlusterWorkloadClusterPrevVersion(t *testing.T) {
	prevLatestRel := prevLatestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterPrevVersionCreateFlow(test, prevLatestRel)
}

// TODO: Add TestCloudStackUpgradeKubernetes124MulticlusterWorkloadClusterWithGithubFlux
func TestCloudStackUpgradeKubernetes124MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube123),
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
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat124Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes125MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
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
				api.WithStackedEtcdTopology(),
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
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.UpdateRedhatTemplate125Var(),
		),
	)
}

func TestCloudStackUpgradeKubernetes126MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
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
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube126),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.UpdateRedhatTemplate126Var(),
		),
	)
}

func TestCloudStackUpgradeKubernetes127MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
			provider.UpdateRedhatTemplate127Var(),
		),
	)
}

func TestCloudStackKubernetes123WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, anywherev1.RedHat, anywherev1.Kube123)
}

func TestCloudStackKubernetes124WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, anywherev1.RedHat, anywherev1.Kube124)
}

func TestCloudStackKubernetes125WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, anywherev1.RedHat, anywherev1.Kube125)
}

func TestCloudStackKubernetes126WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, anywherev1.RedHat, anywherev1.Kube126)
}

func TestCloudStackKubernetes127WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	runTestManagementClusterUpgradeSideEffects(t, cloudstack, anywherev1.RedHat, anywherev1.Kube127)
}

// OIDC
func TestCloudStackKubernetes123OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes124OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes125OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes126OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

func TestCloudStackKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// TODO: Add TestCloudStackKubernetes123To124OIDCUpgrade
func TestCloudStackKubernetes123To124OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124To125OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125To126OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126To127OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

// Proxy config
func TestCloudStackKubernetes123RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes124RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes125RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes126RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

func TestCloudStackKubernetes127RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Proxy config multicluster
func TestCloudStackKubernetes123RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat123(),
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

func TestCloudStackKubernetes124RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat124(),
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
			cloudstack.WithRedhat124(),
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

func TestCloudStackKubernetes125RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat125(),
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
			cloudstack.WithRedhat125(),
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

func TestCloudStackKubernetes126RedhatProxyConfigAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat126(),
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
			cloudstack.WithRedhat126(),
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
		cloudstack.WithRedhat127(),
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
			cloudstack.WithRedhat127(),
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

// Registry mirror
func TestCloudStackKubernetes123RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes124RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes125RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes126RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes127RedhatRegistryMirrorInsecureSkipVerify(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes123RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes124RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes125RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes126RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes127RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes123RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes124RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes125RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes126RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

func TestCloudStackKubernetes127RedhatAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithAuthenticatedRegistryMirror(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simpleflow
func TestCloudStackKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes124SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes125SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes126SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes127SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes123ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes124ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes125ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes126ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes127ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes123MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes124MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes125MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes126MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes127MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes123DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes124DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes125DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes126DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes127DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127(),
			framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace),
				api.WithCloudStackConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

// Cilium Policy
func TestCloudStackKubernetes123CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes124CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes125CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes126CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes127CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
	)
	runSimpleFlow(test)
}

// TODO: Add TestCloudStackKubernetes123RedhatTo124UpgradeCiliumPolicyEnforcementMode
func TestCloudStackKubernetes123RedhatTo124UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124RedhatTo125UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125RedhatTo126UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126RedhatTo127UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

// Stacked etcd
func TestCloudStackKubernetes123StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes124StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes125StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes126StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes127StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

// Taints
func TestCloudStackKubernetes123RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat123ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes124RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat124ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes125RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat125ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

func TestCloudStackKubernetes126RedhatTaintsUpgradeFlow(t *testing.T) {
	provider := redhat126ProviderWithTaints(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube126),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

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

func redhat123ProviderWithTaints(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat123(),
	)
}

func redhat124ProviderWithTaints(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat124(),
	)
}

func redhat125ProviderWithTaints(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat125(),
	)
}

func redhat126ProviderWithTaints(t *testing.T) *framework.CloudStack {
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
		framework.WithCloudStackRedhat126(),
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
		framework.WithCloudStackRedhat127(),
	)
}

// Upgrade
func TestCloudStackKubernetes123RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat123(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
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

func TestCloudStackKubernetes124RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat124(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube124),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
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

func TestCloudStackKubernetes125RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat125(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
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

func TestCloudStackKubernetes126RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat126(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube126),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
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
		framework.WithCloudStackRedhat127(),
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

// TODO: 1.23 to 1.24 upgrade tests
func TestCloudStackKubernetes123To124RedhatUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124To125RedhatUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125To126RedhatUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126To127RedhatUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

func TestCloudStackKubernetes123To124RedhatUnstackedUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124To125RedhatUnstackedUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125To126RedhatUnstackedUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126To127RedhatUnstackedUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

func TestCloudStackKubernetes123RedhatTo124StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Redhat124Template()),
	)
}

func TestCloudStackKubernetes124RedhatTo125StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()),
	)
}

func TestCloudStackKubernetes125RedhatTo126StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()),
	)
}

func TestCloudStackKubernetes126RedhatTo127StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()),
	)
}

// TODO: investigate these tests further as they pass even without the expected behavior(upgrade should fail the first time and continue from the checkpoint on second upgrade)wq
func TestCloudStackKubernetes123RedhatTo124UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat123Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat124Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube124,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes124RedhatTo125UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Redhat124Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube125,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes125RedhatTo126UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate125Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube126,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes126RedhatTo127UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate126Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.UpdateRedhatTemplate127Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube127,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestCloudStackKubernetes123RedhatUpgradeFromLatestMinorReleaseAlwaysNetworkPolicy(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
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
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(
			provider.Redhat123Template(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestCloudStackKubernetes124RedhatUpgradeFromLatestMinorReleaseAlwaysNetworkPolicy(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
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
			provider.Redhat124Template(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestCloudStackKubernetes123RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestCloudStackKubernetes124RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestCloudStackKubernetes125RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestCloudStackKubernetes126RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestCloudStackKubernetes127RedhatControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
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

func TestCloudStackKubernetes123RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestCloudStackKubernetes124RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestCloudStackKubernetes125RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestCloudStackKubernetes126RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestCloudStackKubernetes127RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat127())
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

func TestCloudStackKubernetes123UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
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
		v1alpha1.Kube123,
		provider.WithProviderUpgrade(),
	)
}

func TestCloudStackKubernetes124UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
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
		provider.WithProviderUpgrade(),
	)
}

// TODO: More 1.23 to 1.24 Upgrade tests
func TestCloudStackKubernetes123To124RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat124Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes124To125RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.UpdateRedhatTemplate125Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes125To126RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.UpdateRedhatTemplate126Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes126To127RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.UpdateRedhatTemplate127Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes123AddRemoveAz(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{
		provider.WithProviderUpgrade(
			framework.UpdateAddCloudStackAz2(),
		),
	})
	test.StopIfFailed()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{
		provider.WithProviderUpgrade(
			framework.RemoveAllCloudStackAzs(),
			framework.UpdateAddCloudStackAz1(),
		),
	})
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestCloudStackKubernetes124AddRemoveAz(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{
		provider.WithProviderUpgrade(
			framework.UpdateAddCloudStackAz2(),
		),
	})
	test.StopIfFailed()
	test.UpgradeClusterWithNewConfig([]framework.ClusterE2ETestOpt{
		provider.WithProviderUpgrade(
			framework.RemoveAllCloudStackAzs(),
			framework.UpdateAddCloudStackAz1(),
		),
	})
	test.StopIfFailed()
	test.DeleteCluster()
}

// This test is skipped as registry mirror was not configured for CloudStack
func TestCloudStackKubernetes123RedhatAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat123(),
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
		// framework.WithClusterFiller(api.WithExternalEtcdTopology(1)), there is a bug that the etcd node download etcd from internet
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)

	runAirgapConfigFlow(test, "10.0.0.1/8")
}

func TestCloudStackKubernetes124RedhatAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat124(),
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
		// framework.WithClusterFiller(api.WithExternalEtcdTopology(1)), there is a bug that the etcd node download etcd from internet
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runAirgapConfigFlow(test, "10.0.0.1/8")
}

func TestCloudStackKubernetes125RedhatAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat125(),
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
		// framework.WithClusterFiller(api.WithExternalEtcdTopology(1)), there is a bug that the etcd node download etcd from internet
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runAirgapConfigFlow(test, "10.0.0.1/8")
}

func TestCloudStackKubernetes126RedhatAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat126(),
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
		// framework.WithClusterFiller(api.WithExternalEtcdTopology(1)), there is a bug that the etcd node download etcd from internet
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runAirgapConfigFlow(test, "10.0.0.1/8")
}

// Workload API
func TestCloudStackKubernetes123RedHatAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	test := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat123(),
	)

	test.CreateCluster()
	test.LoadClusterConfigGeneratedByCLI()
	// Run mgmt cluster API tests
	tests := cloudstackAPIManagementClusterUpgradeTests(test, cloudstack)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCloudStackAPIUpgradeTest(t, test, tt)
		})
	}
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestCloudStackKubernetes124RedHatAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	test := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithStackedEtcdTopology(),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		cloudstack.WithRedhat124(),
	)

	test.CreateCluster()
	test.LoadClusterConfigGeneratedByCLI()
	// Run mgmt cluster API tests
	tests := cloudstackAPIManagementClusterUpgradeTests(test, cloudstack)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCloudStackAPIUpgradeTest(t, test, tt)
		})
	}
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestCloudStackMulticlusterWorkloadClusterAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat124(),
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

func TestCloudStackMulticlusterWorkloadClusterNewCredentialsSecretsAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat124(),
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
		cloudstack.WithRedhat123(),
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
		cloudstack.WithRedhat124(),
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

func TestCloudStackKubernetesRedHat123To124UpgradeFromLatestMinorReleaseAPI(t *testing.T) {
	release := latestMinorRelease(t)
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
		),
		cloudstack.WithRedhat123(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),
	)
	wc.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
		),
		cloudstack.WithRedhat123(),
	)
	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPI(
		test,
		release,
		cloudstack.WithRedhat124(),
	)
}

// Workload GitOps API
func TestCloudStackKubernetesRedHat123to124UpgradeFromLatestMinorReleaseGitHubFluxAPI(t *testing.T) {
	release := latestMinorRelease(t)
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t,
		cloudstack,

		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	)
	managementCluster.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	managementCluster.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
		),

		cloudstack.WithRedhat123(),
		framework.WithFluxGithubConfig(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	wc := framework.NewClusterE2ETest(
		t,
		cloudstack,
		framework.WithClusterName(test.NewWorkloadClusterName()),

		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	)
	wc.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	wc.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithManagementCluster(managementCluster.ClusterName),
		),
		cloudstack.WithRedhat123(),
		framework.WithFluxGithubConfig(),
	)
	test.WithWorkloadClusters(wc)

	runMulticlusterUpgradeFromReleaseFlowAPIWithFlux(
		test,
		release,
		cloudstack.WithRedhat124(),
	)
}

func TestCloudStackMulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat124(),
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

func TestCloudStackMulticlusterWorkloadClusterNewCredentialsSecretGitHubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat124(),
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
		cloudstack.WithRedhat123(),
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
		cloudstack.WithRedhat124(),
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

func TestCloudStackWorkloadClusterAWSIamAuthAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat124(),
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
			cloudstack.WithRedhat123(),
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

func TestCloudStackWorkloadClusterAWSIamAuthGithubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat123(),
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

func TestCloudStackWorkloadClusterOIDCAuthAPI(t *testing.T) {
	cloudstack := framework.NewCloudStack(t)
	managementCluster := framework.NewClusterE2ETest(
		t, cloudstack,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		cloudstack.WithRedhat124(),
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
			cloudstack.WithRedhat123(),
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

func TestCloudStackWorkloadClusterOIDCAuthGithubFluxAPI(t *testing.T) {
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
		cloudstack.WithRedhat123(),
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
			cloudstack.WithRedhat123(),
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
