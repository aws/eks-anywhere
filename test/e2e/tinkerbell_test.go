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

// AWS IAM Auth
func TestTinkerbellKubernetes133AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

func TestTinkerbellKubernetes132AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellAWSIamAuthFlow(test)
}

// Upgrade
func TestTinkerbellKubernetes128UbuntuTo129Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu129Image()),
	)
}

func TestTinkerbellKubernetes129UbuntuTo130Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu129Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(framework.Ubuntu130Image()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu130Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube131,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu131Image()),
	)
}

func TestTinkerbellKubernetes132UbuntuTo133Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133Image()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131UpgradeCPOnly(t *testing.T) {
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
		provider.WithCPKubeVersionAndOS(kube130, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(kube130, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu131ImageForCP()),
	)
}

func TestTinkerbellKubernetes132UbuntuTo133UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
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
		provider.WithCPKubeVersionAndOS(kube132, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(kube132, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133ImageForCP()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132UpgradeCPOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
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
		provider.WithCPKubeVersionAndOS(kube131, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(kube131, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132ImageForCP()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
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
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube131)),
		provider.WithProviderUpgrade(framework.Ubuntu131ImageForWorker()),
	)
}

func TestTinkerbellKubernetes132UbuntuTo133UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
	kube132 := v1alpha1.Kube132
	kube133 := v1alpha1.Kube133
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(),
		framework.WithClusterFiller(api.WithKubernetesVersion(kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133ImageForWorker()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132UpgradeWorkerOnly(t *testing.T) {
	provider := framework.NewTinkerbell(t)
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
		provider.WithCPKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004),
		provider.WithWorkerKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2004),
	)
	runSimpleUpgradeFlowWorkerNodeVersionForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithWorkerKubernetesVersion(nodeGroupLabel1, &kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132ImageForWorker()),
	)
}

func TestTinkerbellKubernetes128To129Ubuntu2204RTOSUpgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129RTOSImage()),
	)
}

func TestTinkerbellKubernetes128To129Ubuntu2204GenericUpgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129GenericImage()),
	)
}

func TestTinkerbellKubernetes128To129Ubuntu2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129Image()),
	)
}

func TestTinkerbellKubernetes129To130Ubuntu2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes130Image()),
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

func TestTinkerbellKubernetes128Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes128Image()),
	)
}

func TestTinkerbellKubernetes129Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129Image()),
	)
}

func TestTinkerbellKubernetes129Ubuntu2004To2204RTOSUpgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129RTOSImage()),
	)
}

func TestTinkerbellKubernetes129Ubuntu2004To2204GenericUpgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube129,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes129GenericImage()),
	)
}

func TestTinkerbellKubernetes130Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration(
		test,
		v1alpha1.Kube130,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130)),
		provider.WithProviderUpgrade(framework.Ubuntu2204Kubernetes130Image()),
	)
}

func TestTinkerbellKubernetes131Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2004, nil),
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

func TestTinkerbellKubernetes133Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
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

func TestTinkerbellKubernetes132Ubuntu2004To2204Upgrade(t *testing.T) {
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
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

func TestTinkerbellKubernetes132UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes133UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes132UbuntuWorkerNodeScaleUpWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
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

func TestTinkerbellKubernetes133UbuntuWorkerNodeScaleUpWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
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

func TestTinkerbellKubernetes132UbuntuAddWorkerNodeGroupWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
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

func TestTinkerbellKubernetes133UbuntuAddWorkerNodeGroupWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
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

func TestTinkerbellKubernetes128UbuntuTo129InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube129), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu129Image()),
	)
}

func TestTinkerbellKubernetes129UbuntuTo130InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube130), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu130Image()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube131), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu131Image()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
}

func TestTinkerbellKubernetes132UbuntuTo133InPlaceUpgrade_1CP_1Worker(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithInPlaceUpgradeStrategy()),
		provider.WithProviderUpgrade(framework.Ubuntu133Image()),
	)
}

func TestTinkerbellKubernetes128UbuntuTo129SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube128),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube129),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu129Image()),
	)
}

func TestTinkerbellKubernetes129UbuntuTo130SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube129),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube130),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu130Image()),
	)
}

func TestTinkerbellKubernetes130UbuntuTo131SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube130),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube130, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube131),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu131Image()),
	)
}

func TestTinkerbellKubernetes131UbuntuTo132SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube131),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube131, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
}

func TestTinkerbellKubernetes132UbuntuTo133SingleNodeInPlaceUpgrade(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithEtcdCountIfExternal(0)),
		framework.WithClusterFiller(api.RemoveAllWorkerNodeGroups()),
		framework.WithClusterFiller(api.WithInPlaceUpgradeStrategy()),
		framework.WithControlPlaneHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runInPlaceUpgradeFlowForBareMetal(
		test,
		framework.WithUpgradeClusterConfig(
			api.ClusterToConfigFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithInPlaceUpgradeStrategy(),
			),
			api.TinkerbellToConfigFiller(
				api.RemoveTinkerbellWorkerMachineConfig(),
			),
		),
		provider.WithProviderUpgrade(framework.Ubuntu133Image()),
	)
}

// Curated Packages
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

func TestTinkerbellKubernetes132UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeCuratedPackagesFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube133),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes132UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeCuratedPackagesEmissaryFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube133),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageEmissaryInstallTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes132UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeCuratedPackagesHarborFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube133),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageHarborInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuCuratedPackagesAdotSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube133),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesAdotInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube132),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube133),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	minNodes := 1
	maxNodes := 2
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	minNodes := 1
	maxNodes := 2
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133), api.WithWorkerNodeAutoScalingConfig(minNodes, maxNodes)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube133),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runAutoscalerWithMetricsServerTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)

	runTinkerbellSingleNodeFlow(test)
}

// ISO booting tests
func TestTinkerbellKubernetes131UbuntuHookIsoSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu131Tinkerbell(),
			framework.WithHookIsoBoot(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes131UbuntuHookIsoOverrideSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell(),
			framework.WithHookIsoBoot(),
			framework.WithHookIsoURLPath(framework.HookIsoURLOverride())),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
			api.WithControlPlaneCount(1),
			api.WithEtcdCountIfExternal(0),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithControlPlaneHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

// Multicluster
func TestTinkerbellKubernetes132UbuntuWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes133UbuntuWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
			framework.WithControlPlaneHardware(2),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes132UbuntuWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes133UbuntuWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes132UbuntuWorkloadClusterGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
			),
		),
	)
	runWorkloadClusterGitOpsAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes133UbuntuWorkloadClusterGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
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
	licenseToken := framework.GetLicenseToken()
	licenseToken2 := framework.GetLicenseToken2()
	provider := framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithLicenseToken(licenseToken),
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
				api.WithLicenseToken(licenseToken2),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes132UbuntuSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeWorkloadCluster(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runTinkerbellWorkloadClusterFlow(test)
}

func TestTinkerbellKubernetes132UbuntuSingleNodeWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellKubernetes133UbuntuSingleNodeWorkloadClusterWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(0),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithEtcdCountIfExternal(0),
				api.RemoveAllWorkerNodeGroups(),
			),
		),
	)
	runWorkloadClusterWithAPIFlowForBareMetal(test)
}

func TestTinkerbellUpgrade132MulticlusterWorkloadClusterWorkerScaleupGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube132),
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
				api.WithKubernetesVersion(v1alpha1.Kube132),
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

func TestTinkerbellUpgrade133MulticlusterWorkloadClusterWorkerScaleupGitFluxWithAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
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
			api.WithKubernetesVersion(v1alpha1.Kube133),
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
				api.WithKubernetesVersion(v1alpha1.Kube133),
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

func TestTinkerbellUpgrade132MulticlusterWorkloadClusterCPScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
			framework.WithControlPlaneHardware(4),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(3),
		),
	)
}

func TestTinkerbellUpgrade133MulticlusterWorkloadClusterCPScaleup(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
			framework.WithControlPlaneHardware(4),
			framework.WithWorkerHardware(2),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(3),
		),
	)
}

func TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade131To132(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
			framework.WithControlPlaneHardware(3),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
}

func TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade132To133(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
			framework.WithControlPlaneHardware(3),
			framework.WithWorkerHardware(3),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		),
	)
	runSimpleWorkloadUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133Image()),
	)
}

// OIDC
func TestTinkerbellKubernetes132OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellOIDCFlow(test)
}

func TestTinkerbellKubernetes133OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellOIDCFlow(test)
}

// Registry Mirror
func TestTinkerbellKubernetes132UbuntuRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes133UbuntuRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorEndpointAndCert(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes132UbuntuInsecureSkipVerifyRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes133UbuntuInsecureSkipVerifyRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithRegistryMirrorInsecureSkipVerify(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

func TestTinkerbellKubernetes131UbuntuAuthenticatedRegistryMirror(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithAuthenticatedRegistryMirror(constants.TinkerbellProviderName),
	)
	runTinkerbellRegistryMirrorFlow(test)
}

// Simple Flow
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

func TestTinkerbellKubernetes129UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes130UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu130Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes131UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes128Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

func TestTinkerbellKubernetes129Ubuntu2204SimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)
}

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

func TestTinkerbellKubernetes129Ubuntu2204RTOSSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil, "rtos"),
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

func TestTinkerbellKubernetes129Ubuntu2204GenericSimpleFlow(t *testing.T) {
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube129, framework.Ubuntu2204, nil, "generic"),
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

func TestTinkerbellKubernetes129RedHatSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
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

func TestTinkerbellKubernetes128RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9128Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes129RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithRedHat9129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
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

func TestTinkerbellKubernetes133UbuntuThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuThreeControlPlaneReplicasSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuThreeWorkersSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(3),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes132UbuntuControlPlaneScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes133UbuntuControlPlaneScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes132UbuntuWorkerNodeScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes133UbuntuWorkerNodeScaleUp(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes132UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes133UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes133UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes132UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

// Worker Nodegroup Taints and Labels
func TestTinkerbellKubernetes132UbuntuWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithUbuntu132Tinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
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

func TestTinkerbellKubernetes133UbuntuWorkerNodeGroupsTaintsAndLabels(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(
			t,
			framework.WithUbuntu133Tinkerbell(),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel1),
			framework.WithCustomTinkerbellMachineConfig(nodeGroupLabel2),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
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
func TestTinkerbellAirgappedKubernetes132UbuntuProxyConfigFlow(t *testing.T) {
	localIp, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("Cannot get admin machine local IP: %v", err)
	}
	t.Logf("Admin machine's IP is: %s", localIp)

	kubeVersion := strings.Replace(string(v1alpha1.Kube132), ".", "-", 1)

	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu132Tinkerbell(),
			framework.WithHookImagesURLPath("http://"+localIp.String()+":8080"),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithProxy(framework.TinkerbellProxyRequiredEnvVars),
	)

	runTinkerbellAirgapConfigProxyFlow(test, "10.80.0.0/16", kubeVersion)
}

func TestTinkerbellAirgappedKubernetes133UbuntuProxyConfigFlow(t *testing.T) {
	localIp, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("Cannot get admin machine local IP: %v", err)
	}
	t.Logf("Admin machine's IP is: %s", localIp)

	kubeVersion := strings.Replace(string(v1alpha1.Kube133), ".", "-", 1)

	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t,
			framework.WithUbuntu133Tinkerbell(),
			framework.WithHookImagesURLPath("http://"+localIp.String()+":8080"),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube133),
		),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithProxy(framework.TinkerbellProxyRequiredEnvVars),
	)

	runTinkerbellAirgapConfigProxyFlow(test, "10.80.0.0/16", kubeVersion)
}

// OOB tests
func TestTinkerbellKubernetes132UbuntuOOB(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellKubernetes133UbuntuOOB(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
	)
	runTinkerbellSimpleFlow(test)
}

func TestTinkerbellK8sUpgrade131to132WithUbuntuOOB(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
}

func TestTinkerbellK8sUpgrade132to133WithUbuntuOOB(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithOOBConfiguration(),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube133,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube133)),
		provider.WithProviderUpgrade(framework.Ubuntu133Image()),
	)
}

func TestTinkerbellSingleNode131To132UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu131Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube131),
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
			api.WithKubernetesVersion(v1alpha1.Kube132),
			api.WithControlPlaneCount(3),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube132, framework.Ubuntu2004, nil),
	)
}

func TestTinkerbellSingleNode132To133UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell())
	managementCluster := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube132),
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
			api.WithKubernetesVersion(v1alpha1.Kube133),
			api.WithControlPlaneCount(3),
		),
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2004, nil),
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
	test.GenerateClusterConfigForVersion(release.Version, "", framework.ExecuteWithEksaRelease(release))
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

// TestTinkerbellKubernetes128UbuntuTo132MultipleUpgrade creates a single 1.28 cluster and upgrades it
// all the way until 1.32. This tests each K8s version upgrade in a single test and saves up
// hardware which would otherwise be needed for each test as part of both create and upgrade.
func TestTinkerbellKubernetes128UbuntuTo132MultipleUpgrade(t *testing.T) {
	var kube129clusterOpts []framework.ClusterE2ETestOpt
	var kube130clusterOpts []framework.ClusterE2ETestOpt
	var kube131clusterOpts []framework.ClusterE2ETestOpt
	var kube132clusterOpts []framework.ClusterE2ETestOpt
	licenseToken := framework.GetLicenseToken()
	provider := framework.NewTinkerbell(t, framework.WithUbuntu128Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube128, framework.Ubuntu2004, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)
	kube129clusterOpts = append(
		kube129clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube129),
		),
		provider.WithProviderUpgrade(framework.Ubuntu129Image()),
	)
	kube130clusterOpts = append(
		kube130clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube130),
		),
		provider.WithProviderUpgrade(framework.Ubuntu130Image()),
	)
	kube131clusterOpts = append(
		kube131clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube131),
		),
		provider.WithProviderUpgrade(framework.Ubuntu131Image()),
	)
	kube132clusterOpts = append(
		kube132clusterOpts,
		framework.WithClusterUpgrade(
			api.WithKubernetesVersion(v1alpha1.Kube132),
		),
		provider.WithProviderUpgrade(framework.Ubuntu132Image()),
	)
	runMultipleUpgradesFlowForBareMetal(
		test,
		kube129clusterOpts,
		kube130clusterOpts,
		kube131clusterOpts,
		kube132clusterOpts,
	)
}

// Kubelet Configuration tests
func TestTinkerbellKubernetes129KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu129Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube129)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationTinkerbellFlow(test)
}

func TestTinkerbellKubernetes130KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu130Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube130)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationTinkerbellFlow(test)
}

func TestTinkerbellKubernetes132KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu132Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationTinkerbellFlow(test)
}

func TestTinkerbellKubernetes133KubeletConfigurationSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewTinkerbell(t, framework.WithUbuntu133Tinkerbell()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube133)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(1),
		framework.WithKubeletConfig(),
	)
	runKubeletConfigurationTinkerbellFlow(test)
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
		provider.WithKubeVersionAndOS(v1alpha1.Kube133, framework.Ubuntu2204, nil),
		api.ClusterToConfigFiller(
			api.WithLicenseToken(licenseToken),
		),
	)

	// Get the custom template config
	customTemplateConfig, err := framework.GetCustomTinkerbellConfig(provider.GetTinkerbellLBIP())
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
