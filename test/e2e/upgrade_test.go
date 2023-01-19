//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	vsphereCpVmNumCpuUpdateVar          = 4
	vsphereCpVmMemoryUpdate             = 16384
	vsphereCpDiskGiBUpdateVar           = 40
	vsphereWlVmNumCpuUpdateVar          = 4
	vsphereWlVmMemoryUpdate             = 16384
	vsphereWlDiskGiBUpdate              = 40
	vsphereFolderUpdateVar              = "/SDDC-Datacenter/vm/capv/e2eUpdate"
	vsphereNetwork2UpdateVar            = "/SDDC-Datacenter/network/sddc-cgw-network-2"
	vsphereNetwork3UpdateVar            = "/SDDC-Datacenter/network/sddc-cgw-network-3"
	vsphereInvalidResourcePoolUpdateVar = "*/Resources/INVALID-ResourcePool"
	clusterNamespace                    = "test-namespace"
	vsphereResourcePoolVar              = "T_VSPHERE_RESOURCE_POOL"
)

func runSimpleUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeFlowWithAPI(test *framework.ClusterE2ETest, fillers ...api.ClusterConfigFiller) {
	test.CreateCluster()
	test.UpgradeClusterWithKubectl(fillers...)
	test.ValidateClusterState()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeFlowWithCheckpoint(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts []framework.ClusterE2ETestOpt, clusterOpts2 []framework.ClusterE2ETestOpt, commandOpts []framework.CommandOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeClusterWithNewConfig(clusterOpts, commandOpts...)
	test.UpgradeClusterWithNewConfig(clusterOpts2)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func runSimpleUpgradeFlowForBareMetal(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.UpgradeClusterWithNewConfig(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestVSphereKubernetes121UbuntuTo122Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(provider.Ubuntu122Template()),
	)
}

func TestVSphereKubernetes122UbuntuTo123Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(provider.Ubuntu123Template()),
	)
}

func TestVSphereKubernetes123UbuntuTo124Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Ubuntu124Template()),
	)
}

func TestVSphereKubernetes124UbuntuTo125Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.Ubuntu125Template()),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes124UbuntuTo125UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
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
		provider.WithProviderUpgrade(provider.Ubuntu125Template()),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes124UbuntuTo125MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
		provider.WithProviderUpgrade(
			provider.Ubuntu125Template(),
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

func TestVSphereKubernetes124UbuntuTo125WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.Ubuntu125Template()),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes124UbuntuTo125DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124(),
		framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)),
	)
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		provider.WithProviderUpgrade(provider.Ubuntu125Template()),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes125UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes125UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu125())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes123BottlerocketTo124Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Bottlerocket124Template()),
	)
}

func TestVSphereKubernetes123BottlerocketTo124MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(
			provider.Bottlerocket124Template(),
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

func TestVSphereKubernetes123BottlerocketTo124WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Bottlerocket124Template()),
	)
}

func TestVSphereKubernetes123BottlerocketTo124DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123(),
		framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(provider.Bottlerocket124Template()),
	)
}

func TestVSphereKubernetes124BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket124())
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

func TestVSphereKubernetes124BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket124())
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

func TestDockerKubernetes124To125StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestDockerKubernetes124To125ExternalEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube125)),
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes123UbuntuTo124StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
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
		provider.WithProviderUpgrade(provider.Ubuntu124Template()),
	)
}

func TestVSphereKubernetes123BottlerocketTo124StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
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
		provider.WithProviderUpgrade(provider.Bottlerocket124Template()),
	)
}

func TestCloudStackKubernetes122RedhatTo123Upgrade(t *testing.T) {
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

func TestCloudStackKubernetesUnstacked122RedhatTo123Upgrade(t *testing.T) {
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

func TestCloudStackKubernetes122RedhatTo123MultipleFieldsUpgrade(t *testing.T) {
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

func TestSnowKubernetes121To122UbuntuUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateSnowUbuntuTemplate122Var()),
	)
}

func TestSnowKubernetes122To123UbuntuMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube122),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(1),
		),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(2),
		),
		provider.WithProviderUpgrade(
			framework.UpdateSnowUbuntuTemplate123Var(),
			api.WithSnowInstanceTypeForAllMachines(v1alpha1.SbeCXLarge),
			api.WithSnowPhysicalNetworkConnectorForAllMachines(v1alpha1.QSFP),
		),
	)
}

func TestVSphereKubernetes121UbuntuTo122UpgradeWithCheckpoint(t *testing.T) {
	var clusterOpts []framework.ClusterE2ETestOpt
	var clusterOpts2 []framework.ClusterE2ETestOpt

	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)

	clusterOpts = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)), framework.ExpectFailure(true),
		provider.WithProviderUpgrade(provider.Ubuntu122Template(), api.WithResourcePoolForAllMachines(vsphereInvalidResourcePoolUpdateVar)), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(provider.Ubuntu122Template(), api.WithResourcePoolForAllMachines(os.Getenv(vsphereResourcePoolVar))), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

	runUpgradeFlowWithCheckpoint(
		test,
		v1alpha1.Kube122,
		clusterOpts,
		clusterOpts2,
		commandOpts,
	)
}

func TestTinkerbellKubernetes121UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes121UbuntuControlPlaneUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes122UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes122UbuntuControlPlaneUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes123UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes124UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu124Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

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
		framework.WithEnvVar(features.K8s125SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube125,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

func TestTinkerbellKubernetes123UbuntuControlPlaneUpgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestTinkerbellKubernetes121UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes121UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes122UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes122UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes123UbuntuWorkerNodeScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
		framework.WithControlPlaneHardware(1),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

func TestTinkerbellKubernetes123UbuntuControlPlaneScaleDown(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(3),
		framework.WithWorkerHardware(1),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(1)),
	)
}

func TestTinkerbellKubernetes121UbuntuTo122Upgrade(t *testing.T) {
	provider := framework.NewTinkerbell(t, framework.WithUbuntu121Tinkerbell())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithControlPlaneHardware(2),
		framework.WithWorkerHardware(2),
	)
	runSimpleUpgradeFlowForBareMetal(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateTinkerbellUbuntuTemplate122Var()),
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

// Nutanix Upgrade tests START
func TestNutanixKubernetes121To122UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate122Var()),
	)
}

func TestNutanixKubernetes122To123UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate123Var()),
	)
}

func TestNutanixKubernetes123To124UbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateNutanixUbuntuTemplate124Var()),
	)
}

//
// Nutanix Upgrade tests END
//

//
// Nutanix Worker Scale Up tests START
//

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes121UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 2 worker nodes clusters scaled up to 5
func TestNutanixKubernetes121UbuntuWorkerNodeScaleUp2To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes122UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 2 worker nodes clusters scaled up to 5
func TestNutanixKubernetes122UbuntuWorkerNodeScaleUp2To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes123UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 2 worker nodes clusters scaled up to 5
func TestNutanixKubernetes123UbuntuWorkerNodeScaleUp2To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(2)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

// 1 worker node cluster scaled up to 3
func TestNutanixKubernetes124UbuntuWorkerNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(3)),
	)
}

// 2 worker nodes clusters scaled up to 5
func TestNutanixKubernetes124UbuntuWorkerNodeScaleUp2To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

//
// Nutanix Worker Scale Up tests END
//

//
// Nutanix Control Plane Scale Up tests START
//

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes121UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled up to 5
func TestNutanixKubernetes121UbuntuControlPlaneNodeScaleUp3To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(5)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled up to 5
func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleUp3To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(5)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled up to 5
func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleUp3To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(5)),
	)
}

// 1 node control plane cluster scaled up to 3
func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleUp1To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
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
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled up to 5
func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleUp3To5(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(5)),
	)
}

//
// Nutanix Control Plane Scale Up tests END
//

//
// Nutanix Worker Scale Down tests START
//

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes121UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 5 worker nodes clusters scaled down to 2
func TestNutanixKubernetes121UbuntuWorkerNodeScaleDown5To2(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes122UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 5 worker nodes clusters scaled down to 2
func TestNutanixKubernetes122UbuntuWorkerNodeScaleDown5To2(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes123UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 5 worker nodes clusters scaled down to 2
func TestNutanixKubernetes123UbuntuWorkerNodeScaleDown5To2(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(5)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

// 3 worker node cluster scaled down to 1
func TestNutanixKubernetes124UbuntuWorkerNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(1)),
	)
}

// 5 worker nodes clusters scaled down to 2
func TestNutanixKubernetes124UbuntuWorkerNodeScaleDown5To2(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(5)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(2)),
	)
}

//
// Nutanix Worker Scale Down tests END
//

//
// Nutanix Control Plane Scale Down tests START
//

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes121UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 5 node control plane cluster scaled down to 3
func TestNutanixKubernetes121UbuntuControlPlaneNodeScaleDown5To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu121Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(5)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 5 node control plane cluster scaled down to 3
func TestNutanixKubernetes122UbuntuControlPlaneNodeScaleDown5To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu122Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithControlPlaneCount(5)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 5 node control plane cluster scaled down to 3
func TestNutanixKubernetes123UbuntuControlPlaneNodeScaleDown5To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu123Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithControlPlaneCount(5)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

// 3 node control plane cluster scaled down to 1
func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleDown3To1(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
}

// 5 node control plane cluster scaled down to 3
func TestNutanixKubernetes124UbuntuControlPlaneNodeScaleDown5To3(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu124Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(5)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

//
// Nutanix Control Plane Scale Down tests END
//
