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
)

func runSimpleUpgradeFlow(test *framework.E2ETest, updateVersion v1alpha1.KubernetesVersion, opts ...framework.E2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.UpgradeCluster(opts...)
	test.ValidateCluster(updateVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestVSphereKubernetes120To121Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewE2ETest(
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

func TestVSphereKubernetes120To121MultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate121Var(),
			api.WithNumCPUs(vsphereCpVmNumCpuUpdateVar),
			api.WithMemoryMiB(vsphereCpVmMemoryUpdate),
			api.WithDiskGiB(vsphereCpDiskGiBUpdateVar),
			api.WithFolder(vsphereFolderUpdateVar),
			// Uncomment once we support tests with multiple machine configs
			/*api.WithWorkloadVMsNumCPUs(vsphereWlVmNumCpuUpdateVar),
			api.WithWorkloadVMsMemoryMiB(vsphereWlVmMemoryUpdate),
			api.WithWorkloadDiskGiB(vsphereWlDiskGiBUpdate),*/
			// Uncomment the network field once upgrade starts working with it
			// api.WithNetwork(vsphereNetwork2UpdateVar),
		),
	)
}

func TestVSphereKubernetes120To121UpgradeWithFlux(t *testing.T) {
	t.Skip("Needs extra setup for GitOps/Flux")
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewE2ETest(t,
		provider,
		framework.WithFlux(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runUpgradeFlowWithFlux(
		test,
		v1alpha1.Kube121,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate121Var()),
	)
}

func TestVSphereKubernetes120ControlPlaneNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(3)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube120,
		framework.WithClusterUpgrade(api.WithControlPlaneCount(3)),
	)
}

func TestVSphereKubernetes120WorkerNodeUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewE2ETest(
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
