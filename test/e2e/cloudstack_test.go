//go:build e2e && (cloudstack || all_providers)
// +build e2e
// +build cloudstack all_providers

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/test/framework"
)

// Curated packages test
func TestCloudStackKubernetes121RedhatCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCloudStackKubernetes121RedhatCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()), // Only 1.20/1.21 is supported at the moment for Redhat.
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageEmissaryInstallSimpleFlow(test)
}

func TestCloudStackKubernetes121RedhatWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes121RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121()) // Only 1.20/1.21 is supported at the moment for Redhat.
	test := SetupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageEmissaryRemoteClusterInstallSimpleFlow(test)
}

func TestCloudStackKubernetes121RedhatCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesAdotInstallSimpleFlow(test)
}

func TestCloudStackKubernetes121RedHatCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	minNodes := 1
	maxNodes := 2
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runAutoscalerWitMetricsServerSimpleFlow(test)
}

func TestCloudStackKubernetes121RedhatCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackagesPrometheusInstallSimpleFlow(test)
}

// Labels
func TestCloudStackKubernetes121LabelsAndNodeNameRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat121(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube121),
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

// Multicluster
func TestCloudStackKubernetes121WorkloadCluster(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube121),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube121),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	runWorkloadClusterFlow(test)
}

func TestCloudStackKubernetes121ManagementClusterUpgradeFromLatest(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube121),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithEtcdCountIfExternal(1),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube121),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithEtcdCountIfExternal(1),
			),
		),
	)
	runWorkloadClusterUpgradeFlowCheckWorkloadRollingUpgrade(test)
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
			framework.UpdateRedhatTemplate123Var(),
		),
	)
}

// OIDC
func TestCloudStackKubernetes121OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// Proxy config
func TestCloudStackKubernetes121RedhatProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithProxy(framework.CloudstackProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Registry mirror
func TestCloudStackKubernetes121RedhatRegistryMirrorAndCert(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)
	runRegistryMirrorConfigFlow(test)
}

// Simpleflow
func TestCloudStackKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
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

func TestCloudStackKubernetes121MultiEndpointSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121(),
			framework.WithCloudStackFillers(framework.UpdateAddCloudStackAz2())),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleFlow(test)
}

func TestCloudStackKubernetes121DifferentNamespaceSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121(), framework.WithCloudStackFillers(api.WithCloudStackConfigNamespace(clusterNamespace))),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
	)
	runSimpleFlow(test)
}

// Stacked etcd
func TestCloudStackKubernetes121StackedEtcdRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
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

// Upgrade
func TestCloudStackKubernetes121RedhatAndRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewCloudStack(t,
		framework.WithCloudStackWorkerNodeGroup(
			"worker-1",
			framework.WithWorkerNodeGroup("workers-1", api.WithCount(2)),
		),
		framework.WithCloudStackWorkerNodeGroup(
			"worker-2",
			framework.WithWorkerNodeGroup("workers-2", api.WithCount(1)),
		),
		framework.WithCloudStackRedhat121(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube121),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
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
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate123Var()),
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
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate123Var()),
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
			framework.UpdateRedhatTemplate123Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes121UpgradeFromLatestMinorRelease(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromReleaseFlow(
		test,
		release,
		v1alpha1.Kube121,
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

func TestCloudStackKubernetes121AddRemoveAz(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
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

// Flux
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

// This test is skipped as registry mirror was not configured for CloudStack
func TestCloudStackKubernetes121UbuntuAirgappedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat121(),
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
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithRegistryMirrorEndpointAndCert(constants.CloudStackProviderName),
	)

	runAirgapConfigFlow(test, "10.0.0.1/8")
}
