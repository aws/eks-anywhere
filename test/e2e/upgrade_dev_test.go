//go:build e2eDev
// +build e2eDev

package e2e

import (
	"testing"

	vsphere2 "github.com/aws/eks-anywhere/internal/pkg/api/vsphere"

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
			vsphere2.WithControlPlaneVMsNumCPUs(vsphereCpVmNumCpuUpdateVar),
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
			vsphere2.WithControlPlaneVMsMemoryMiB(vsphereCpVmMemoryUpdate),
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
			vsphere2.WithControlPlaneDiskGiB(vsphereCpDiskGiBUpdateVar),
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
			vsphere2.WithWorkloadVMsNumCPUs(vsphereWlVmNumCpuUpdateVar),
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
			vsphere2.WithWorkloadVMsMemoryMiB(vsphereWlVmMemoryUpdate),
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
			vsphere2.WithWorkloadDiskGiB(vsphereWlDiskGiBUpdate),
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
			vsphere2.WithFolder(vsphereFolderUpdateVar),
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
			vsphere2.WithNetwork(vsphereNetwork2UpdateVar),
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
			vsphere2.WithNetwork(vsphereNetwork3UpdateVar),
		),
	)
}

func TestCloudStackKubernetes120To121CpComputeOfferingUpgrade(t *testing.T) {
	//t.Skip("Skipping CloudStack in CI/CD")
	provider := framework.NewCloudStack(t, framework.WithRedhat120())
	test := framework.NewE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube120,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube121)),
		provider.WithProviderUpgrade(
			framework.UpdateRedhatTemplate121Var(),
			api.WithCloudStackComputeOffering(cloudstackComputeOfferingUpdateVar),
		),
	)
}
