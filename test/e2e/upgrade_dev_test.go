//go:build e2eDev
// +build e2eDev

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func TestVSphereKubernetes118To119CpVmNumCpuUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithControlPlaneVMsNumCPUs(vsphereCpVmNumCpuUpdateVar),
		),
	)
}

func TestVSphereKubernetes118To119CpVmMemoryUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithControlPlaneVMsMemoryMiB(vsphereCpVmMemoryUpdate),
		),
	)
}

func TestVSphereKubernetes118To119CpDiskGiBUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithControlPlaneDiskGiB(vsphereCpDiskGiBUpdateVar),
		),
	)
}

func TestVSphereKubernetes118To119WlVmNumCpuUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithWorkloadVMsNumCPUs(vsphereWlVmNumCpuUpdateVar),
		),
	)
}

func TestVSphereKubernetes118To119WlVmMemoryUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithWorkloadVMsMemoryMiB(vsphereWlVmMemoryUpdate),
		),
	)
}

func TestVSphereKubernetes118To119WlDiskGiBUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithWorkloadDiskGiB(vsphereWlDiskGiBUpdate),
		),
	)
}

func TestVSphereKubernetes118To119FolderUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithFolder(vsphereFolderUpdateVar),
		),
	)
}

func TestVSphereKubernetes118To119Network1to2Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithNetwork(vsphereNetwork2UpdateVar),
		),
	)
}

func TestVSphereKubernetes118To119Network1to3Upgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu118())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube118)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube119,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube119)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate119Var(),
			api.WithNetwork(vsphereNetwork3UpdateVar),
		),
	)
}

func TestCloudStackKubernetes120To121CpComputeOfferingUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat120())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube120,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(
			framework.UpdateRedhatTemplate121Var(),
			api.WithCloudStackComputeOfferingForAllMachines(cloudstackComputeOfferingUpdateVar),
		),
	)
}
