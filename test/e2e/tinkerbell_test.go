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
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/test/framework"
)

// Upgrade
func TestTinkerbellKubernetes130UbuntuTo131UpgradeCPOnly(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	kube130 := v1alpha1.Kube130
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube130)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube130, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube130, framework.Ubuntu2204),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu131ImageForCP()),
	)
}

func TestTinkerbellKubernetes134UbuntuTo135UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube134 := v1alpha1.Kube134
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube134)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube134, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube134, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu135ImageForCP()),
	)
}

func TestTinkerbellKubernetes134To135Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes135Image()),
	)
}

func TestTinkerbellKubernetes134To135Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes135Image()),
	)
}

func TestTinkerbellKubernetes134Ubuntu2204To2404RTOSUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes134RTOSImage()),
	)
}

func TestTinkerbellKubernetes134Ubuntu2204To2404GenericUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes134GenericImage()),
	)
}

func TestTinkerbellKubernetes134Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes134Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes134Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes134RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9134Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube134)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuTo133UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	licenseToken := framework.GetLicenseToken()
	kube132 := v1alpha1.Kube132
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube132)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube132, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube132, framework.Ubuntu2204),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133ImageForCP()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	licenseToken := framework.GetLicenseToken()
	kube131 := v1alpha1.Kube131
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube131)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube131, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube131, framework.Ubuntu2204),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132ImageForCP()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	licenseToken := framework.GetLicenseToken()
	kube130 := v1alpha1.Kube130
	kube131 := v1alpha1.Kube131
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube130)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu131ImageForWorker()),
	)
}

func TestTinkerbellKubernetes133UbuntuTo134UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube133 := v1alpha1.Kube133
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube133)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube133, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube133, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu134ImageForCP()),
	)
}

func TestTinkerbellKubernetes133UbuntuTo134UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube133 := v1alpha1.Kube133
	kube134 := v1alpha1.Kube134
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube134)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube134, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu134ImageForWorker()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	licenseToken := framework.GetLicenseToken()
	kube131 := v1alpha1.Kube131
	kube132 := v1alpha1.Kube132
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube131)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132ImageForWorker()),
	)
}

func TestTinkerbellKubernetes130To131Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes131Image()),
	)
}

func TestTinkerbellKubernetes131To132Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes132Image()),
	)
}

func TestTinkerbellKubernetes133To134Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes134Image()),
	)
}

func TestTinkerbellKubernetes132To133Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes133Image()),
	)
}

func TestTinkerbellKubernetes133To134Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube134,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube134)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes134Image()),
	)
}

func TestTinkerbellKubernetes135Ubuntu2204To2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes135Image()),
	)
}

func TestTinkerbellKubernetes135Ubuntu2204To2404RTOSUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes135RTOSImage()),
	)
}

func TestTinkerbellKubernetes135Ubuntu2204To2404GenericUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube135,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube135)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes135GenericImage()),
	)
}

func TestTinkerbellKubernetes133Ubuntu2204To2404RTOSUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes133RTOSImage()),
	)
}

func TestTinkerbellKubernetes133Ubuntu2204To2404GenericUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes133GenericImage()),
	)
}

func TestTinkerbellKubernetes130Ubuntu2204To2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes130Image()),
	)
}

func TestTinkerbellKubernetes131Ubuntu2204To2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes131Image()),
	)
}

func TestTinkerbellKubernetes137UbuntuHookIsoSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu137Tinkerbell(),
			framework.WithHookIsoBoot(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuHookIsoOverrideSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell(),
			framework.WithHookIsoBoot(),
			framework.WithHookIsoURLPath(framework.HookIsoURLOverride())),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

// Multicluster

func TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade136To137(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu136Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube136)),
			framework.WithControlPlaneHardware(3),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube136)),
		),
	)
	test.ManagementCluster.GenerateClusterConfig()
	test.CreateTinkerbellManagementCluster()
	test.RunInWorkloadClusters(func(w *framework.WorkloadCluster) {
		w.GenerateClusterConfigWithLicenseToken(framework.GetLicenseToken2())
		w.GenerateSupportBundleOnCleanupIfTestFailed()
		w.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
		w.UpgradeCluster([]framework.ClusterE2ETestOpt{
			framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
			provider.WithProviderUpgrade(framework.Ubuntu137Image()),
		})
		w.ValidateClusterState()
		w.StopIfFailed()
		w.DeleteCluster()
		w.ValidateHardwareDecommissioned()
	})
	test.DeleteTinkerbellManagementCluster()
}

func TestTinkerbellKubernetes137UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithAuthenticatedRegistryMirror(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

// Simple Flow

func TestTinkerbellKubernetes130Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes131Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes132Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes135Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes135Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes130Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes131Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes132Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes135Ubuntu2404RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes130Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes131Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes132Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes135Ubuntu2404GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes130RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat130Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes131RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat131Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes130RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9130Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes131RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9131Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes135RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9135Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube135)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes133Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil, "rtos"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes133Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes133RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellSingleNode136To137UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu136Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube136),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(4),
	)
	runUpgradeFlowForBareMetalWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
			api.WithControlPlaneCount(3),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil),
	)
}

func TestTinkerbellCustomTemplateRefSimpleFlow(t *testing.T) {

	// Get license token for Ubuntu 2204
	licenseToken := framework.GetLicenseToken()

	// Create a new test with 1 control plane and 1 worker node using Ubuntu 2204
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Get the custom template config
	customTemplateConfig, err := framework.GetCustomTinkerbellConfig(
		test.ClusterConfig.TinkerbellDatacenter.Spec.TinkerbellIP,
		test.ClusterConfig.TinkerbellDatacenter.Spec.OSImageURL)
	if err != nil {
		t.Fatalf("Failed to get custom template config: %v", err)
	}

	// Add the custom template config to the cluster
	test.UpdateClusterConfig(
		provider.WithTinkerbellTemplateConfig(customTemplateConfig),
	)

	// Get the control plane and worker node names
	clusterName := test.ClusterName
	cpName := providers.GetControlPlaneNodeName(clusterName)
	workerName := clusterName

	// Override the machine config for both control plane and worker nodes
	test.UpdateClusterConfig(
		api.TinkerbellToConfigFiller(
			api.WithCustomTinkerbellMachineConfig(cpName,
				framework.WithTemplateRef(customTemplateConfig.Name, anywherev1.TinkerbellTemplateConfigKind),
				framework.UpdateTinkerbellMachineSSHAuthorizedKey(),
				api.WithOsFamilyForTinkerbellMachineConfig(anywherev1.Ubuntu),
			),
			api.WithCustomTinkerbellMachineConfig(workerName,
				framework.WithTemplateRef(customTemplateConfig.Name, anywherev1.TinkerbellTemplateConfigKind),
				framework.UpdateTinkerbellMachineSSHAuthorizedKey(),
				api.WithOsFamilyForTinkerbellMachineConfig(anywherev1.Ubuntu),
			),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes137AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

func TestTinkerbellKubernetes135UbuntuTo136UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube135 := v1alpha1.Kube135
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube135)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube135)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube135, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube135, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube136)),
		provider.WithProviderUpgrade(framework.Ubuntu136ImageForCP()),
	)
}

func TestTinkerbellKubernetes135To136Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube136,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube136)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes136Image()),
	)
}

func TestTinkerbellKubernetes135To136Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube135, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube136,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube136)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes136Image()),
	)
}

func TestTinkerbellKubernetes136UbuntuTo137InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube136)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu137Image()),
	)
}

func TestTinkerbellKubernetes136UbuntuTo137SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube136),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu137Image()),
	)
}

func TestTinkerbellKubernetes136RedHat9To137SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube136),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.RedHat9, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.RedHat9Kubernetes137Image()),
	)
}

func TestTinkerbellKubernetes136Ubuntu2204To2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube136,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube136)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes136Image()),
	)
}

func TestTinkerbellKubernetes136Ubuntu2204To2404GenericUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube136,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube136)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes136GenericImage()),
	)
}

func TestTinkerbellKubernetes137UbuntuWorkerNodeScaleUpWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
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

func TestTinkerbellKubernetes137UbuntuAddWorkerNodeGroupWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
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

func TestTinkerbellKubernetes137UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube137),
		framework.WithControlPlaneHardware(1),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes137UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube137),
		framework.WithControlPlaneHardware(1),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes137UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube137),
		framework.WithControlPlaneHardware(1),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube137),
		framework.WithControlPlaneHardware(1),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube137),
		framework.WithControlPlaneHardware(1),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	minNodes := 1
	maxNodes := 2
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137), api.WithWorkerNodeCount(minNodes), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes137UbuntuWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes137UbuntuWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
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
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(framework.GetLicenseToken2()),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes137UbuntuWorkloadClusterGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube137),
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
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(framework.GetLicenseToken2()),
			),
		),
	)
	runWorkloadClusterGitOpsAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes137UbuntuSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube137),
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
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes137UbuntuSingleNodeWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
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
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(framework.GetLicenseToken2()),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellUpgrade137MulticlusterWorkloadClusterWorkerScaleupGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube137),
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
				api.WithKubernetesVersion(v1alpha1.Kube137),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithLicenseToken(framework.GetLicenseToken2()),
			),
		),
	)
	runWorkloadClusterGitOpsAPIUpgradeFlowForBareMetal(test,
		api.ClusterToConfigFiller(
			api.WithWorkerNodeCount(2),
		),
	)
}

func TestTinkerbellUpgrade137MulticlusterWorkloadClusterCPScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
			framework.WithControlPlaneHardware(4),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube137,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube137),
			api.WithControlPlaneCount(3),
		),
	)
}

func TestTinkerbellKubernetes137OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellOIDCFlow(test)
}

func TestTinkerbellKubernetes137UbuntuRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes137UbuntuInsecureSkipVerifyRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes136Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes136Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes136Ubuntu2404GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2404, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes136RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9136Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube136)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137UbuntuControlPlaneScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube137,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes137UbuntuWorkerNodeScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube137,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes137UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube137,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes137UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube137,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes137UbuntuWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithUbuntu137Tinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
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

func TestTinkerbellAirgappedKubernetes137UbuntuProxyConfigFlow(t *testing.T) {
	localIp, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("Cannot get admin machine local IP: %v", err)
	}
	t.Logf("Admin machine's IP is: %s", localIp)

	kubeVersion := strings.Replace(string(v1alpha1.Kube137), ".", "-", 1)

	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu137Tinkerbell(),
			framework.WithHookImagesURLPath(TinkerbellHookOSImagesURLPath),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube137),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithProxy(framework.TinkerbellProxyRequiredEnvVars),
	)

	runTinkerbellAirgapConfigProxyFlow(test, TinkerbellNoProxyCIDR, kubeVersion)
}

func TestTinkerbellKubernetes137UbuntuOOB(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellK8sUpgrade136to137WithUbuntuOOB(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu136Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube136)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	test.GenerateClusterConfig()
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu137Image()),
	)
}

func TestTinkerbellKubernetes137KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationTinkerbellFlow(test)
}

// TestTinkerbellKubernetes137Ubuntu2404MultiNICBootstrapIPSimpleFlow validates cluster
// creation on a multi-NIC admin machine: it creates a secondary network interface on the
// admin machine with an IP from the Tinkerbell provisioning network and passes that IP as
// --tinkerbell-bootstrap-ip. The default route stays on the primary interface, so the
// Tinkerbell stack must bind its TFTP/HTTP/syslog listeners to the bootstrap IP explicitly
// rather than auto-detecting the default-route NIC. Regression test for the smee
// advertise-vs-bind IP mismatch on multi-NIC admin machines.

func TestTinkerbellKubernetes137Ubuntu2404MultiNICBootstrapIPSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	bootstrapIP := framework.CreateSecondaryNetworkInterface(t, provider.CPNetworkCIDR())
	framework.SetTinkerbellBootstrapIP(t, bootstrapIP)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes136UbuntuTo137UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube136 := v1alpha1.Kube136
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(kube136)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube136)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(kube136, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(kube136, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu137ImageForCP()),
	)
}

func TestTinkerbellKubernetes136To137Ubuntu2204Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes137Image()),
	)
}

func TestTinkerbellKubernetes136To137Ubuntu2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes137Image()),
	)
}

func TestTinkerbellKubernetes136To137RedHat9Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube136, framework.RedHat9, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.RedHat9Kubernetes137Image()),
	)
}

func TestTinkerbellKubernetes136UbuntuTo137UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube136 := v1alpha1.Kube136
	kube137 := v1alpha1.Kube137
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube137)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube136)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube136, framework.Ubuntu2204),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu137ImageForWorker()),
	)
}

func TestTinkerbellKubernetes137Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes137Ubuntu2404SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2404, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes137Ubuntu2404GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2404, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes137RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9137Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube137)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes137Ubuntu2204To2404Upgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes137Image()),
	)
}

func TestTinkerbellKubernetes137Ubuntu2204To2404GenericUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube137)),
		provider.WithProviderUpgrade(framework.Ubuntu2404Kubernetes137GenericImage()),
	)
}

func TestTinkerbellKubernetes137Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube137, framework.Ubuntu2204, nil, "generic"),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}
