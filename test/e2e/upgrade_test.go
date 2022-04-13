//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	vsphereCpVmNumCpuUpdateVar = 4
	vsphereCpVmMemoryUpdate    = 16384
	vsphereCpDiskGiBUpdateVar  = 40
	vsphereWlVmNumCpuUpdateVar = 4
	vsphereWlVmMemoryUpdate    = 16384
	vsphereWlDiskGiBUpdate     = 40
	vsphereFolderUpdateVar     = "/SDDC-Datacenter/vm/capv/e2eUpdate"
	vsphereNetwork2UpdateVar   = "/SDDC-Datacenter/network/sddc-cgw-network-2"
	vsphereNetwork3UpdateVar   = "/SDDC-Datacenter/network/sddc-cgw-network-3"
	clusterNamespace           = "test-namespace"
)

func runSimpleUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
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

func TestVSphereKubernetes121UbuntuTo122UpgradeCiliumPolicyEnforcementMode(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(v1alpha1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var()),
	)
}

func TestVSphereKubernetes121UbuntuTo122MultipleFieldsUpgrade(t *testing.T) {
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
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate122Var(),
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

func TestVSphereKubernetes121UbuntuTo122WithFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var()),
	)
}

func TestVSphereKubernetes121UbuntuTo122DifferentNamespaceWithFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFlux(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var()),
	)
}

func TestVSphereKubernetes122UbuntuControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes122UbuntuWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestVSphereKubernetes121BottlerocketTo122Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate122()),
	)
}

func TestVSphereKubernetes121BottlerocketTo122MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate122(),
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

func TestVSphereKubernetes121BottlerocketTo122WithFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate122()),
	)
}

func TestVSphereKubernetes121BottlerocketTo122DifferentNamespaceWithFluxUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121(), framework.WithVSphereFillers(api.WithVSphereConfigNamespaceForAllMachinesAndDatacenter(clusterNamespace)))
	test := framework.NewClusterE2ETest(t,
		provider,
		framework.WithFlux(api.WithGitOpsNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithClusterNamespace(clusterNamespace)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate122()),
	)
}

func TestVSphereKubernetes122BottlerocketControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
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
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes122BottlerocketWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
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
		framework.WithClusterUpgrade(api.WithWorkerNodeCount(5)),
	)
}

func TestDockerKubernetes121To122StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
}

func TestVSphereKubernetes121UbuntuTo122StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate122Var()),
	)
}

func TestVSphereKubernetes121BottlerocketTo122StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube122)),
		provider.WithProviderUpgrade(framework.UpdateBottlerocketTemplate122()),
	)
}

func TestCloudStackKubernetes120RedhatTo121Upgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(framework.UpdateRedhatTemplate121Var()),
	)
}

func TestCloudStackKubernetes120RedhatTo121MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithRedhat120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(
			framework.UpdateRedhatTemplate121Var(),
			framework.UpdateLargerCloudStackComputeOffering(),
		),
	)
}

func TestCloudStackKubernetes120RedhatWorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithRedhat120())
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
