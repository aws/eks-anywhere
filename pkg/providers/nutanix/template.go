package nutanix

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

// TemplateBuilder builds templates for nutanix
type TemplateBuilder struct {
	datacenterSpec              *v1alpha1.NutanixDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.NutanixMachineConfigSpec
	etcdMachineSpec             *v1alpha1.NutanixMachineConfigSpec
	workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec
	creds                       basicAuthCreds
	now                         types.NowFunc
}

var _ providers.TemplateBuilder = &TemplateBuilder{}

func NewNutanixTemplateBuilder(
	datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec,
	controlPlaneMachineSpec,
	etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec,
	workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec,
	creds basicAuthCreds,
	now types.NowFunc,
) *TemplateBuilder {
	return &TemplateBuilder{
		datacenterSpec:              datacenterSpec,
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		etcdMachineSpec:             etcdMachineSpec,
		workerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		creds:                       creds,
		now:                         now,
	}
}

func (ntb *TemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.NutanixMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *ntb.etcdMachineSpec
	}

	values := buildTemplateMapCP(ntb.datacenterSpec, clusterSpec, *ntb.controlPlaneMachineSpec, etcdMachineSpec, ntb.creds)
	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (ntb *TemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, ntb.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

func (ntb *TemplateBuilder) GenerateCAPISpecSecret(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapSecret(clusterSpec, ntb.creds)
	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(secretTemplate, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

func buildTemplateMapCP(
	datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec,
	clusterSpec *cluster.Spec,
	controlPlaneMachineSpec v1alpha1.NutanixMachineConfigSpec,
	etcdMachineSpec v1alpha1.NutanixMachineConfigSpec,
	creds basicAuthCreds,
) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                  clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys[0],
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"podCidrs":                     clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"kubernetesVersion":            bundle.KubeDistro.Kubernetes.Tag,
		"kubernetesRepository":         bundle.KubeDistro.Kubernetes.Repository,
		"corednsRepository":            bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":               bundle.KubeDistro.CoreDNS.Tag,
		"etcdRepository":               bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                 bundle.KubeDistro.Etcd.Tag,
		"kubeVipImage":                 "ghcr.io/kube-vip/kube-vip:latest",
		"externalEtcdVersion":          bundle.KubeDistro.EtcdVersion,
		"etcdCipherSuites":             crypto.SecureCipherSuitesString(),
		"nutanixEndpoint":              datacenterSpec.Endpoint,
		"nutanixPort":                  datacenterSpec.Port,
		"nutanixAdditionalTrustBundle": datacenterSpec.AdditionalTrustBundle,
		"nutanixUser":                  creds.username,
		"nutanixPassword":              creds.password,
		"vcpusPerSocket":               controlPlaneMachineSpec.VCPUsPerSocket,
		"vcpuSockets":                  controlPlaneMachineSpec.VCPUSockets,
		"memorySize":                   controlPlaneMachineSpec.MemorySize.String(),
		"systemDiskSize":               controlPlaneMachineSpec.SystemDiskSize.String(),
		"imageName":                    controlPlaneMachineSpec.Image.Name,   // TODO(nutanix): pass name or uuid based on type of identifier
		"nutanixPEClusterName":         controlPlaneMachineSpec.Cluster.Name, // TODO(nutanix): pass name or uuid based on type of identifier
		"subnetName":                   controlPlaneMachineSpec.Subnet.Name,  // TODO(nutanix): pass name or uuid based on type of identifier
	}
	values["nutanixInsecure"] = datacenterSpec.AdditionalTrustBundle != ""

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.NutanixMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.Cluster.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"workerReplicas":         *workerNodeGroupConfiguration.Count,
		"workerPoolName":         "md-0",
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
		"vcpusPerSocket":         workerNodeGroupMachineSpec.VCPUsPerSocket,
		"vcpuSockets":            workerNodeGroupMachineSpec.VCPUSockets,
		"memorySize":             workerNodeGroupMachineSpec.MemorySize.String(),
		"systemDiskSize":         workerNodeGroupMachineSpec.SystemDiskSize.String(),
		"imageName":              workerNodeGroupMachineSpec.Image.Name,   // TODO(nutanix): pass name or uuid based on type of identifier
		"nutanixPEClusterName":   workerNodeGroupMachineSpec.Cluster.Name, // TODO(nutanix): pass name or uuid based on type of identifier
		"subnetName":             workerNodeGroupMachineSpec.Subnet.Name,  // TODO(nutanix): pass name or uuid based on type of identifier
		"workerNodeGroupName":    fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
	}
	return values
}

func buildTemplateMapSecret(
	clusterSpec *cluster.Spec,
	creds basicAuthCreds,
) map[string]interface{} {
	values := map[string]interface{}{
		"clusterName":         clusterSpec.Cluster.Name,
		"eksaSystemNamespace": constants.EksaSystemNamespace,
		"nutanixUser":         creds.username,
		"nutanixPassword":     creds.password,
	}
	return values
}
