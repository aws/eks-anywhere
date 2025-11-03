// nolint
package e2e

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	EksaPackagesRegistryMirrorAlias = "curated-packages"
	EksaPackagesSourceRegistry      = "public.ecr.aws/x3k6m8v0"
	EksaPackageBundleURI            = "oci://" + EksaPackagesSourceRegistry + "/eks-anywhere-packages-bundles"
	EksaPackagesNamespace           = "eksa-packages"
	EksaPackagesRegistry            = "067575901363.dkr.ecr.us-west-2.amazonaws.com"

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
	TinkerbellHookOSImagesURLPath       = "http://10.80.18.43:8080/hook"
	TinkerbellNoProxyCIDR               = "10.80.0.0/16"
)

var (
	KubeVersions = []v1alpha1.KubernetesVersion{v1alpha1.Kube128, v1alpha1.Kube129, v1alpha1.Kube130, v1alpha1.Kube131, v1alpha1.Kube132, v1alpha1.Kube133, v1alpha1.Kube134}
)
