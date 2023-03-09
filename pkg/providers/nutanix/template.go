package nutanix

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

var jsonMarshal = json.Marshal

// TemplateBuilder builds templates for nutanix.
type TemplateBuilder struct {
	datacenterSpec              *v1alpha1.NutanixDatacenterConfigSpec
	controlPlaneMachineSpec     *v1alpha1.NutanixMachineConfigSpec
	etcdMachineSpec             *v1alpha1.NutanixMachineConfigSpec
	workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec
	creds                       credentials.BasicAuthCredential
	now                         types.NowFunc
}

var _ providers.TemplateBuilder = &TemplateBuilder{}

func NewNutanixTemplateBuilder(
	datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec,
	controlPlaneMachineSpec,
	etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec,
	workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec,
	creds credentials.BasicAuthCredential,
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

	values := buildTemplateMapCP(ntb.datacenterSpec, clusterSpec, *ntb.controlPlaneMachineSpec, etcdMachineSpec)
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
		values["autoscalingConfig"] = workerNodeGroupConfiguration.AutoScalingConfiguration

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

// GenerateCAPISpecSecret generates the secret containing the credentials for the nutanix prism central and is used by the
// CAPX controller. The secret is named after the cluster name.
func (ntb *TemplateBuilder) GenerateCAPISpecSecret(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	return ntb.generateSpecSecret(capxSecretName(clusterSpec), ntb.creds, buildOptions...)
}

func capxSecretName(spec *cluster.Spec) string {
	return fmt.Sprintf("capx-%s", spec.Cluster.Name)
}

// GenerateEKSASpecSecret generates the secret containing the credentials for the nutanix prism central and is used by the
// EKS-A controller. The secret is named nutanix-credentials.
func (ntb *TemplateBuilder) GenerateEKSASpecSecret(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	return ntb.generateSpecSecret(eksaSecretName(clusterSpec), ntb.creds, buildOptions...)
}

func eksaSecretName(spec *cluster.Spec) string {
	return spec.NutanixDatacenter.Spec.CredentialRef.Name
}

func (ntb *TemplateBuilder) generateSpecSecret(secretName string, creds credentials.BasicAuthCredential, buildOptions ...providers.BuildMapOption) ([]byte, error) {
	values, err := buildTemplateMapSecret(secretName, creds)
	if err != nil {
		return nil, err
	}

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
) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig)

	values := map[string]interface{}{
		"apiServerExtraArgs":           apiServerExtraArgs.ToPartialYaml(),
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
		"kubeVipImage":                 bundle.Nutanix.KubeVip.VersionedImage(),
		"kubeVipSvcEnable":             false,
		"kubeVipLBEnable":              false,
		"externalEtcdVersion":          bundle.KubeDistro.EtcdVersion,
		"etcdCipherSuites":             crypto.SecureCipherSuitesString(),
		"nutanixEndpoint":              datacenterSpec.Endpoint,
		"nutanixPort":                  datacenterSpec.Port,
		"nutanixAdditionalTrustBundle": datacenterSpec.AdditionalTrustBundle,
		"nutanixInsecure":              datacenterSpec.Insecure,
		"vcpusPerSocket":               controlPlaneMachineSpec.VCPUsPerSocket,
		"vcpuSockets":                  controlPlaneMachineSpec.VCPUSockets,
		"memorySize":                   controlPlaneMachineSpec.MemorySize.String(),
		"systemDiskSize":               controlPlaneMachineSpec.SystemDiskSize.String(),
		"imageIDType":                  controlPlaneMachineSpec.Image.Type,
		"imageName":                    controlPlaneMachineSpec.Image.Name,
		"imageUUID":                    controlPlaneMachineSpec.Image.UUID,
		"nutanixPEClusterIDType":       controlPlaneMachineSpec.Cluster.Type,
		"nutanixPEClusterName":         controlPlaneMachineSpec.Cluster.Name,
		"nutanixPEClusterUUID":         controlPlaneMachineSpec.Cluster.UUID,
		"secretName":                   capxSecretName(clusterSpec),
		"subnetIDType":                 controlPlaneMachineSpec.Subnet.Type,
		"subnetName":                   controlPlaneMachineSpec.Subnet.Name,
		"subnetUUID":                   controlPlaneMachineSpec.Subnet.UUID,
	}

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
		"imageIDType":            workerNodeGroupMachineSpec.Image.Type,
		"imageName":              workerNodeGroupMachineSpec.Image.Name,
		"imageUUID":              workerNodeGroupMachineSpec.Image.UUID,
		"nutanixPEClusterIDType": workerNodeGroupMachineSpec.Cluster.Type,
		"nutanixPEClusterName":   workerNodeGroupMachineSpec.Cluster.Name,
		"nutanixPEClusterUUID":   workerNodeGroupMachineSpec.Cluster.UUID,
		"subnetIDType":           workerNodeGroupMachineSpec.Subnet.Type,
		"subnetName":             workerNodeGroupMachineSpec.Subnet.Name,
		"subnetUUID":             workerNodeGroupMachineSpec.Subnet.UUID,
		"workerNodeGroupName":    fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
	}
	return values
}

func buildTemplateMapSecret(secretName string, creds credentials.BasicAuthCredential) (map[string]interface{}, error) {
	encodedCreds, err := jsonMarshal(creds)
	if err != nil {
		return nil, err
	}

	nutanixCreds := []credentials.Credential{{
		Type: credentials.BasicAuthCredentialType,
		Data: encodedCreds,
	}}
	credsJSON, err := jsonMarshal(nutanixCreds)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		"secretName":               secretName,
		"eksaSystemNamespace":      constants.EksaSystemNamespace,
		"base64EncodedCredentials": base64.StdEncoding.EncodeToString(credsJSON),
	}

	return values, nil
}
