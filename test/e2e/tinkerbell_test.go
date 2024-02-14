//go:build e2e && (tinkerbell || all_providers)
// +build e2e
// +build tinkerbell all_providers

package e2e

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/test/framework"
)

// AWS IAM Auth

func TestTinkerbellKubernetes128AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

// Upgrade
func TestTinkerbellKubernetes125UbuntuTo126Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(framework.Ubuntu126Image()),
	)
}

func TestTinkerbellKubernetes126UbuntuTo127Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(framework.Ubuntu127Image()),
	)
}

func TestTinkerbellKubernetes127UbuntuTo128Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu128Image()),
	)
}

func TestTinkerbellKubernetes127UbuntuTo128UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube127 := v1alpha1.Kube127
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube127)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu128ImageForCP()),
	)
}

func TestTinkerbellKubernetes127UbuntuTo128UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube127 := v1alpha1.Kube127
	kube128 := v1alpha1.Kube128
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube127)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu128ImageForWorker()),
	)
}

func TestTinkerbellKubernetes125To126Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube125, framework.Ubuntu2204, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes126Image()),
	)
}

func TestTinkerbellKubernetes126To127Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube126, framework.Ubuntu2204, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes127Image()),
	)
}

func TestTinkerbellKubernetes127To128Ubuntu2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2204, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes128Image()),
	)
}

func TestTinkerbellKubernetes126Ubuntu2004To2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube126, framework.Ubuntu2004, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube126,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes126Image()),
	)
}

func TestTinkerbellKubernetes127Ubuntu2004To2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes127Image()),
	)
}

func TestTinkerbellKubernetes128Ubuntu2004To2204Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes128Image()),
	)
}

func TestTinkerbellKubernetes128UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes125UbuntuWorkerNodeScaleUpWithAPI(t *testing.T) {
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
	runUpgradeFlowForBareMetalWithAPI(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeCount(2),
		),
	)
}

func TestTinkerbellKubernetes125UbuntuAddWorkerNodeGroupWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithCustomLabelHardware(1, "worker-0"),
	)
	runUpgradeFlowForBareMetalWithAPI(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeGroup("worker-0",
				api.WithCount(1),
				api.WithMachineGroupRef("worker-0", "TinkerbellMachineConfig"),
			),
		),
		api.TinkerbellToConfigFiller(
			api.WithCustomTinkerbellMachineConfig("worker-0",
				framework.UpdateTinkerbellMachineSSHAuthorizedKey(),
				api.WithOsFamilyForTinkerbellMachineConfig(v1alpha1.Ubuntu),
			),
		),
	)
}

func TestTinkerbellKubernetes125UbuntuTo126InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube125, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube126), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu126Image()),
	)
}

func TestTinkerbellKubernetes126UbuntuTo127InPlaceUpgrade_1CP_2Worker(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube126, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu127Image()),
	)
}

func TestTinkerbellKubernetes127UbuntuTo128InPlaceUpgrade_3CP_1Worker(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu128Image()),
	)
}

func TestTinkerbellKubernetes125UbuntuTo126SingleNodeInPlaceUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube125, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube126),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu126Image()),
	)
}

func TestTinkerbellKubernetes126UbuntuTo127SingleNodeInPlaceUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube126, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube127),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu127Image()),
	)
}

func TestTinkerbellKubernetes127UbuntuTo128SingleNodeInPlaceUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2004, nil),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu128Image()),
	)
}

// Curated packages
func TestTinkerbellKubernetes127UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes127UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes127UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes127UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes127UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube127),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes127UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	minNodes := 1
	maxNodes := 2
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube127),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes126UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube126),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes126UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube126),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes126UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube126),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes126UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube126),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes126UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube126),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube126),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
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
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes125UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu125Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube125),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube125),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
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
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	minNodes := 1
	maxNodes := 2
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube128),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

// Multicluster
func TestTinkerbellKubernetes128UbuntuWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes128UbuntuWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
		),
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
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes128UbuntuWorkloadClusterGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		framework.WithFluxGithubConfig(),
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
		),
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
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterGitOpsAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes128BottlerocketWorkloadClusterSimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes128BottlerocketWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
		),
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
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
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
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSingleNodeWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
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
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellUpgrade128MulticlusterWorkloadClusterWorkerScaleupGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
		framework.WithFluxGithubEnvVarCheck(),
		framework.WithFluxGithubCleanup(),
	).WithClusterConfig(
		framework.WithFluxGithubConfig(),
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.RemoveAllWorkerNodeGroups(),
		),
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
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube128),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterGitOpsAPIUpgradeFlowForBareMetal(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeCount(2),
		),
	)
}

func TestTinkerbellUpgrade128MulticlusterWorkloadClusterCPScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
			framework.WithControlPlaneHardware(4),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(3),
		),
	)
}

func TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade127To128(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
			framework.WithControlPlaneHardware(3),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu128Image()),
	)
}

// OIDC
func TestTinkerbellKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellOIDCFlow(test)
}

// Registry mirror
func TestTinkerbellKubernetes128UbuntuRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes128UbuntuInsecureSkipVerifyRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes128UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithAuthenticatedRegistryMirror(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

// Simpleflow
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

func TestTinkerbellKubernetes126UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu126Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes127UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes125Ubuntu2204SimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube125, framework.Ubuntu2204, nil),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes126Ubuntu2204SimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube126, framework.Ubuntu2204, nil),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes127Ubuntu2204SimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2204, nil),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes128Ubuntu2204SimpleFlow(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
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

func TestTinkerbellKubernetes126RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat126Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes127RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat127Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128BottleRocketSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128UbuntuControlPlaneScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes128UbuntuWorkerNodeScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes128UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes128UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

// Worker nodegroup taints and labels
func TestTinkerbellKubernetes128UbuntuWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithUbuntu128Tinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
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
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeTaints, framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneTaints, framework.ValidateControlPlaneLabels)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

// Proxy tests

func TestTinkerbellAirgappedKubernetes128UbuntuProxyConfigFlow(t *testing.T) {
	localIp, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("Cannot get admin machine local IP: %v", err)
	}
	t.Logf("Admin machine's IP is: %s", localIp)

	kubeVersion := strings.Replace(string(v1alpha1.Kube128), ".", "-", 1)

	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu128Tinkerbell(),
			framework.WithHookImagesURLPath("http://"+localIp.String()+":8080"),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithProxy(framework.TinkerbellProxyRequiredEnvVars),
	)

	runTinkerbellAirgapConfigProxyFlow(test, "10.80.0.0/16", kubeVersion)
}

// OOB test
func TestTinkerbellKubernetes128UbuntuOOB(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellK8sUpgrade127to128WithUbuntuOOB(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu128Image()),
	)
}

func TestTinkerbellSingleNode127To128UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu127Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
	)
	test := framework.NewMulticlusterE2ETest(
		t,
		managementCluster,
	)
	runWorkloadClusterUpgradeFlowWithAPIForBareMetal(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(3),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)
}

func TestTinkerbellKubernetes128UpgradeManagementComponents(t *testing.T) {
	release := latestMinorRelease(t)
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	// create cluster with old eksa
	test.GenerateClusterConfigForVersion(release.Version, framework.ExecuteWithEksaRelease(release))
	test.UpdateClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
	)

	test.GenerateHardwareConfig(framework.ExecuteWithEksaRelease(release))
	test.CreateCluster(framework.ExecuteWithEksaRelease(release), framework.WithControlPlaneWaitTimeout("20m"))
	// upgrade management-components with new eksa
	test.RunEKSA([]string{"upgrade", "management-components", "-f", test.ClusterConfigLocation, "-v", "99"})
	test.DeleteCluster()
}
