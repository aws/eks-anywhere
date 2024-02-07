// nolint
package e2e

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	EksaPackageControllerHelmChartName = "eks-anywhere-packages"
	EksaPackagesSourceRegistry         = "public.ecr.aws/l0g8r8j6"
	EksaPackageControllerHelmURI       = "oci://" + EksaPackagesSourceRegistry + "/eks-anywhere-packages"
	EksaPackageControllerHelmVersion   = "0.2.20-eks-a-v0.0.0-dev-build.4894"
	EksaPackageBundleURI               = "oci://" + EksaPackagesSourceRegistry + "/eks-anywhere-packages-bundles"
	EksaPackagesNamespace              = "eksa-packages"

	clusterNamespace = "test-namespace"

	key1            = framework.LabelPrefix + "/" + "key1"
	key2            = framework.LabelPrefix + "/" + "key2"
	cpKey1          = framework.LabelPrefix + "/" + "cp-key1"
	val1            = "val1"
	val2            = "val2"
	cpVal1          = "cp-val1"
	nodeGroupLabel1 = "md-0"
	nodeGroupLabel2 = "md-1"
	worker0         = "worker-0"
	worker1         = "worker-1"
	worker2         = "worker-2"

	fluxUserProvidedBranch    = "testbranch"
	fluxUserProvidedNamespace = "testns"
	fluxUserProvidedPath      = "test/testerson"

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
	vsphereResourcePoolVar              = "T_VSPHERE_RESOURCE_POOL"
	TinkerbellHardwareCountFile         = "TINKERBELL_HARDWARE_COUNT.yaml"
)

var (
	EksaPackageControllerHelmValues = []string{"sourceRegistry=public.ecr.aws/l0g8r8j6"}
	KubeVersions                    = []v1alpha1.KubernetesVersion{v1alpha1.Kube125, v1alpha1.Kube126, v1alpha1.Kube127, v1alpha1.Kube128, v1alpha1.Kube129}
)
