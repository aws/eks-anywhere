//go:build e2e && (cloudstack || all_providers)
// +build e2e
// +build cloudstack all_providers

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
	corev1 "k8s.io/api/core/v1"
)

// AWS IAM Auth
func TestCloudStackKubernetes122AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat122()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes122to123AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
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

func TestCloudStackKubernetes123RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes122RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
	runCertManagerRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes123RedhatCuratedPackagesCertManagerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	framework.CheckCertManagerCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
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
func TestCloudStackKubernetes123FluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runFluxFlow(test)
}

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

func TestCloudStackKubernetes123ThreeReplicasThreeWorkersFluxLegacy(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithFluxLegacy(),
	)
	runFluxFlow(test)
}

func TestCloudStackKubernetes123GitopsOptionsFluxLegacy(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
	)

	test.RunClusterFlowWithGitOps(
		framework.WithClusterUpgradeGit(
			api.WithWorkerNodeCount(3),
		),
		// Needed in order to replace the CloudStackDatacenterConfig namespace field with the value specified
		// compared to when it was initially created without it.
		provider.WithProviderUpgradeGit(),
	)
}

func TestCloudStackKubernetes122To123FluxUpgradeLegacy(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackKubernetes122To123GitFluxUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxGit(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackInstallGitFluxDuringUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
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

func TestCloudStackKubernetes122RedhatLabelsUpgradeFlow(t *testing.T) {
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

func TestCloudStackUpgradeMulticlusterWorkloadClusterWithFluxLegacy(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxLegacy(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube122),
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
				api.WithKubernetesVersion(v1alpha1.Kube122),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat123Template(),
		),
	)
}

func TestCloudStackUpgradeKubernetes123MulticlusterWorkloadClusterWithGithubFlux(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithFluxGithub(),
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube122),
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
				api.WithKubernetesVersion(v1alpha1.Kube122),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlowWithGitOps(
		test,
		framework.WithClusterUpgradeGit(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(3),
		),
		provider.WithProviderUpgradeGit(
			provider.Redhat123Template(),
		),
	)
}

func TestCloudStackKubernetes123ManagementClusterUpgradeFromLatest(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat123())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithEtcdCountIfExternal(1),
		),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)

	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterName(test.NewWorkloadClusterName()),
			framework.WithClusterFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithKubernetesVersion(v1alpha1.Kube123),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithEtcdCountIfExternal(1),
			),
		),
	)

	runFlowUpgradeManagementClusterCheckForSideEffects(test,
		framework.NewEKSAReleasePackagedBinary(latestMinorRelease(t)),
		newEKSAPackagedBinaryForLocalBinary(t),
	)
}

// OIDC
func TestCloudStackKubernetes122OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat122()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

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

func TestCloudStackKubernetes122To123OIDCUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithOIDC(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
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

// Simpleflow
func TestCloudStackKubernetes122SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
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

func TestCloudStackKubernetes123MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes123DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123(), framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
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

func TestCloudStackKubernetes122RedhatTo123UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

// Stacked etcd
func TestCloudStackKubernetes122StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()))
	runStackedEtcdFlow(test)
}

func TestCloudStackKubernetes123StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
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

func TestCloudStackKubernetes122To123RedhatUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackKubernetes122To123RedhatUnstackedUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackKubernetes122RedhatTo123StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackKubernetes122RedhatTo123UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate122Var()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Redhat123Template()), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube123,
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

func TestCloudStackKubernetes122To123RedhatMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			provider.Redhat123Template(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes122UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube122,
		provider.WithProviderUpgrade(),
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

func TestCloudStackKubernetes122RedhatTo123WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
	)
}

func TestCloudStackKubernetes122RedhatTo123DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat122())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Redhat123Template()),
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

// This test is skipped as registry mirror was not configured for CloudStack
func TestCloudStackKubernetes123UbuntuAirgappedRegistryMirror(t *testing.T) {
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
