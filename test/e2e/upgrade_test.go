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
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeFlowWithCheckpoint(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts []framework.ClusterE2ETestOpt, clusterOpts2 []framework.ClusterE2ETestOpt, commandOpts []framework.CommandOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeCluster(clusterOpts, commandOpts...)
	test.UpgradeCluster(clusterOpts2)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func runSimpleUpgradeFlowForBareMetal(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func TestVSphereKubernetes120UbuntuTo121Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate121Var()),
	)
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
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var()),
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
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate123Var()),
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
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate124Var()),
	)
}

func TestVSphereKubernetes123UbuntuTo124UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate124Var()),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes123UbuntuTo124MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate124Var(),
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
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes123UbuntuTo124WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate124Var()),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes123UbuntuTo124DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu123(),
		framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)),
	)
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFluxLegacy(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate124Var()),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes124UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes124UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes122BottlerocketTo123Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate123()),
	)
}

func TestVSphereKubernetes122BottlerocketTo123MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate123(),
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

func TestVSphereKubernetes122BottlerocketTo123WithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
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
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate123()),
	)
}

func TestVSphereKubernetes122BottlerocketTo123DifferentNamespaceWithFluxLegacyUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
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
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate123()),
	)
}

func TestVSphereKubernetes123BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
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

func TestVSphereKubernetes123BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
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

func TestDockerKubernetes122To123StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
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
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube124,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube124)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate124Var()),
		framework.WithEnvVar(features.K8s124SupportEnvVar, "true"),
	)
}

func TestVSphereKubernetes122BottlerocketTo123StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
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
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate123()),
	)
}

func TestCloudStackKubernetes120RedhatTo121Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate121Var()),
	)
}

func TestCloudStackKubernetesUnstacked120RedhatTo121Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate121Var()),
	)
}

func TestCloudStackKubernetes120RedhatTo121MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		provider.WithProviderUpgrade(
			framework.UpdateRedhatTemplate121Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes120RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube120,
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
	test.UpgradeCluster([]framework.ClusterE2ETestOpt{
		provider.WithProviderUpgrade(
			framework.UpdateAddCloudStackAz2(),
		),
	})
	test.StopIfFailed()
	test.UpgradeCluster([]framework.ClusterE2ETestOpt{
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
		framework.WithEnvVar("SNOW_PROVIDER", "true"),
		framework.WithEnvVar("FULL_LIFECYCLE_API", "true"),
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
		framework.WithEnvVar("SNOW_PROVIDER", "true"),
		framework.WithEnvVar("FULL_LIFECYCLE_API", "true"),
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
		framework.WithEnvVar("SNOW_PROVIDER", "true"),
		framework.WithEnvVar("FULL_LIFECYCLE_API", "true"),
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
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var(), api.WithResourcePoolForAllMachines(vsphereInvalidResourcePoolUpdateVar)), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "false"))

	commandOpts := []framework.CommandOpt{framework.WithExternalEtcdWaitTimeout("10m")}

	clusterOpts2 = append(clusterOpts, framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)), framework.ExpectFailure(false),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var(), api.WithResourcePoolForAllMachines(os.Getenv(vsphereResourcePoolVar))), framework.WithEnvVar(features.CheckpointEnabledEnvVar, "true"), framework.WithEnvVar(framework.CleanupVmsVar, "true"))

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
