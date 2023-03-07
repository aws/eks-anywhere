//go:build e2e && (tinkerbell || all_providers)
// +build e2e
// +build tinkerbell all_providers

package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/test/framework"
)

// AWS IAM Auth

func TestTinkerbellKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

func TestTinkerbellKubernetes125AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

func TestTinkerbellKubernetes123BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

// Curated packages
func TestTinkerbellKubernetes125UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)

	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

// Multicluster
func TestTinkerbellKubernetes125UbuntuWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes125UbuntuWorkloadClusterAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	test := framework.NewMulticlusterE2ETest(
		t,
		managementCluster,
	)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runTinkerbellWorkloadClusterAPIFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketWorkloadClusterSimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketWorkloadClusterAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	test := framework.NewMulticlusterE2ETest(
		t,
		managementCluster,
	)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runTinkerbellWorkloadClusterAPIFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(0),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeWorkloadClusterAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	)
	test := framework.NewMulticlusterE2ETest(
		t,
		managementCluster,
	)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterAPIFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(0),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketSingleNodeWorkloadClusterAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	)
	test := framework.NewMulticlusterE2ETest(
		t,
		managementCluster,
	)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube125),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterAPIFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketWorkloadClusterSkipPowerActions(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithNoPowerActions(),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithNoPowerActions(),
		),
	)
	runTinkerbellWorkloadClusterFlowSkipPowerActions(test)
}

func TestTinkerbellUpgrade125MulticlusterWorkloadClusterWorkerScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithWorkerNodeCount(2),
		),
	)
}

func TestTinkerbellUpgrade125MulticlusterWorkloadClusterCPScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithControlPlaneHardware(4),
			framework.WithWorkerHardware(1),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneCount(3),
		),
	)
}

func TestTinkerbellUpgrade125MulticlusterWorkloadClusterWorkerScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithControlPlaneHardware(1),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
			framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithWorkerNodeCount(1),
		),
	)
}

func TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade124To125(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu124Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
			framework.WithControlPlaneHardware(3),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(framework.UpdateTinkerbellUbuntuTemplate125Var()),
	)
}

// Nodes powered on
func TestTinkerbellKubernetes125WithNodesPoweredOn(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)

	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

// OIDC
func TestTinkerbellKubernetes125OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellOIDCFlow(test)
}

// Registry mirror
func TestTinkerbellKubernetes125UbuntuRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes125UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithAuthenticatedRegistryMirror(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes125BottlerocketAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithAuthenticatedRegistryMirror(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

// Simpleflow
func TestTinkerbellKubernetes122UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes123UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes124UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu124Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes126SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes122RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat122Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes123RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat123Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes124RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat124Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes122BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes123BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes124BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuExternalEtcdSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell(), framework.WithTinkerbellExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithExternalEtcdHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125BottleRocketThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

// Single node
func TestTinkerbellKubernetes125BottleRocketSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

// Skip power actions
func TestTinkerbellKubernetes125SkipPowerActions(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithNoPowerActions(),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)

	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.PXEBootHardware()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithForce(), framework.WithControlPlaneWaitTimeout("20m"))
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}

// Upgrade
func TestTinkerbellKubernetes125UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes125UbuntuControlPlaneScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes125UbuntuWorkerNodeScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes125UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes125UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes122UbuntuTo123Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(framework.UpdateTinkerbellUbuntuTemplate123Var()),
	)
}

func TestTinkerbellKubernetes123UbuntuTo124Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateTinkerbellUbuntuTemplate124Var()),
	)
}

func TestTinkerbellKubernetes124UbuntuTo125Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu124Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(framework.UpdateTinkerbellUbuntuTemplate125Var()),
	)
}

// Worker nodegroup taints and labels
func TestTinkerbellKubernetes125UbuntuWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithUbuntu125Tinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
			api.WithControlPlaneTaints([]corev1.Taint{framework.NoScheduleTaint()}),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithWorkerNodeGroup(worker0, api.WithMachineGroupRef(nodeGroupLabel1, "TinkerbellMachineConfig"), api.WithTaint(framework.PreferNoScheduleTaint()), api.WithLabel(key1, val1), api.WithCount(1)),
			api.WithWorkerNodeGroup(worker1, api.WithMachineGroupRef(nodeGroupLabel2, "TinkerbellMachineConfig"), api.WithLabel(key2, val2), api.WithCount(1)),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithCustomLabelHardware(1, nodeGroupLabel1),
		framework.WithCustomLabelHardware(1, nodeGroupLabel2),
	)

	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints, framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints, framework.ValidateControlPlaneLabels)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestTinkerbellKubernetes125BottlerocketWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithBottleRocketTinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
			api.WithControlPlaneTaints([]corev1.Taint{framework.NoScheduleTaint()}),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithWorkerNodeGroup(worker0, api.WithMachineGroupRef(nodeGroupLabel1, "TinkerbellMachineConfig"), api.WithTaint(framework.PreferNoScheduleTaint()), api.WithLabel(key1, val1), api.WithCount(1)),
			api.WithWorkerNodeGroup(worker1, api.WithMachineGroupRef(nodeGroupLabel2, "TinkerbellMachineConfig"), api.WithLabel(key2, val2), api.WithCount(1)),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithCustomLabelHardware(1, nodeGroupLabel1),
		framework.WithCustomLabelHardware(1, nodeGroupLabel2),
	)

	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints, framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints, framework.ValidateControlPlaneLabels)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

// Airgapped tests
func TestTinkerbellAirgappedKubernetes125BottleRocketRegistryMirror(t *testing.T) {
	localIp, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("Cannot get admin machine local IP: %v", err)
	}
	t.Logf("Admin machine's IP is: %s", localIp)

	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithBottleRocketTinkerbell(),
			framework.WithHookImagesURLPath("http://"+localIp.String()+":8080"),
			framework.WithOSImageURL("http://"+localIp.String()+":8080/"+bottlerocketOSFileName),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube125),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)

	runTinkerbellAirgapConfigFlow(test, "10.80.0.0/16")
}
