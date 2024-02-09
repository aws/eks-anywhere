package nutanix

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
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

	values, err := buildTemplateMapCP(ntb.datacenterSpec, clusterSpec, *ntb.controlPlaneMachineSpec, etcdMachineSpec)
	if err != nil {
		return nil, err
	}
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
		values, err := buildTemplateMapMD(clusterSpec, ntb.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		values["autoscalingConfig"] = workerNodeGroupConfiguration.AutoScalingConfiguration

		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			values["upgradeRolloutStrategy"] = true
			values["maxSurge"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
			values["maxUnavailable"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable
		}

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
	return ntb.generateSpecSecret(clusterSpec, CAPXSecretName(clusterSpec), ntb.creds, buildOptions...)
}

// CAPXSecretName returns the name of the secret containing the credentials for the nutanix prism central and is used by the
// CAPX controller.
func CAPXSecretName(spec *cluster.Spec) string {
	return fmt.Sprintf("capx-%s", spec.Cluster.Name)
}

// GenerateEKSASpecSecret generates the secret containing the credentials for the nutanix prism central and is used by the
// EKS-A controller. The secret is named nutanix-credentials.
func (ntb *TemplateBuilder) GenerateEKSASpecSecret(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	return ntb.generateSpecSecret(clusterSpec, EKSASecretName(clusterSpec), ntb.creds, buildOptions...)
}

// EKSASecretName returns the name of the secret containing the credentials for the nutanix prism central and is used by the
// EKS-Anywhere controller.
func EKSASecretName(spec *cluster.Spec) string {
	if spec.NutanixDatacenter.Spec.CredentialRef != nil {
		return spec.NutanixDatacenter.Spec.CredentialRef.Name
	}
	return constants.NutanixCredentialsName
}

func (ntb *TemplateBuilder) generateSpecSecret(clusterSpec *cluster.Spec, secretName string, creds credentials.BasicAuthCredential, buildOptions ...providers.BuildMapOption) ([]byte, error) {
	values, err := buildTemplateMapSecret(clusterSpec, secretName, creds)
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
) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.RootVersionsBundle()
	format := "cloud-config"
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig))
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))

	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		"auditPolicy":                  auditPolicy,
		"apiServerExtraArgs":           apiServerExtraArgs.ToPartialYaml(),
		"cloudProviderImage":           versionsBundle.Nutanix.CloudProvider.VersionedImage(),
		"clusterName":                  clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys[0],
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"controlPlaneTaints":           clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"podCidrs":                     clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"kubernetesVersion":            versionsBundle.KubeDistro.Kubernetes.Tag,
		"kubernetesRepository":         versionsBundle.KubeDistro.Kubernetes.Repository,
		"corednsRepository":            versionsBundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":               versionsBundle.KubeDistro.CoreDNS.Tag,
		"etcdRepository":               versionsBundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                 versionsBundle.KubeDistro.Etcd.Tag,
		"kubeletExtraArgs":             kubeletExtraArgs.ToPartialYaml(),
		"kubeVipImage":                 versionsBundle.Nutanix.KubeVip.VersionedImage(),
		"kubeVipSvcEnable":             false,
		"kubeVipLBEnable":              false,
		"externalEtcdVersion":          versionsBundle.KubeDistro.EtcdVersion,
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
		"secretName":                   CAPXSecretName(clusterSpec),
		"subnetIDType":                 controlPlaneMachineSpec.Subnet.Type,
		"subnetName":                   controlPlaneMachineSpec.Subnet.Name,
		"subnetUUID":                   controlPlaneMachineSpec.Subnet.UUID,
		"apiServerCertSANs":            clusterSpec.Cluster.Spec.ControlPlaneConfiguration.CertSANs,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["publicMirror"] = containerd.ToAPIEndpoint(registryMirror.CoreEKSAMirror())
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		if len(registryMirror.CACertContent) > 0 {
			values["registryCACert"] = registryMirror.CACertContent
		}

		if registryMirror.Auth {
			values["registryAuth"] = registryMirror.Auth
			username, password, err := config.ReadCredentials()
			if err != nil {
				return values, err
			}
			values["registryUsername"] = username
			values["registryPassword"] = password
		}
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if controlPlaneMachineSpec.Project != nil {
		values["projectIDType"] = controlPlaneMachineSpec.Project.Type
		values["projectName"] = controlPlaneMachineSpec.Project.Name
		values["projectUUID"] = controlPlaneMachineSpec.Project.UUID
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = generateNoProxyList(clusterSpec)
	}

	if len(controlPlaneMachineSpec.AdditionalCategories) > 0 {
		values["additionalCategories"] = controlPlaneMachineSpec.AdditionalCategories
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		values["maxSurge"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
	}

	return values, nil
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.NutanixMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.WorkerNodeGroupVersionsBundle(workerNodeGroupConfiguration)
	format := "cloud-config"

	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration))
	values := map[string]interface{}{
		"clusterName":            clusterSpec.Cluster.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      versionsBundle.KubeDistro.Kubernetes.Tag,
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
		"kubeletExtraArgs":       kubeletExtraArgs.ToPartialYaml(),
		"nutanixPEClusterIDType": workerNodeGroupMachineSpec.Cluster.Type,
		"nutanixPEClusterName":   workerNodeGroupMachineSpec.Cluster.Name,
		"nutanixPEClusterUUID":   workerNodeGroupMachineSpec.Cluster.UUID,
		"subnetIDType":           workerNodeGroupMachineSpec.Subnet.Type,
		"subnetName":             workerNodeGroupMachineSpec.Subnet.Name,
		"subnetUUID":             workerNodeGroupMachineSpec.Subnet.UUID,
		"workerNodeGroupName":    fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":  workerNodeGroupConfiguration.Taints,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["publicMirror"] = containerd.ToAPIEndpoint(registryMirror.CoreEKSAMirror())
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		if len(registryMirror.CACertContent) > 0 {
			values["registryCACert"] = registryMirror.CACertContent
		}

		if registryMirror.Auth {
			values["registryAuth"] = registryMirror.Auth
			username, password, err := config.ReadCredentials()
			if err != nil {
				return values, err
			}
			values["registryUsername"] = username
			values["registryPassword"] = password
		}
	}

	if workerNodeGroupMachineSpec.Project != nil {
		values["projectIDType"] = workerNodeGroupMachineSpec.Project.Type
		values["projectName"] = workerNodeGroupMachineSpec.Project.Name
		values["projectUUID"] = workerNodeGroupMachineSpec.Project.UUID
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true

		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = generateNoProxyList(clusterSpec)
	}

	if len(workerNodeGroupMachineSpec.AdditionalCategories) > 0 {
		values["additionalCategories"] = workerNodeGroupMachineSpec.AdditionalCategories
	}

	return values, nil
}

func buildTemplateMapSecret(clusterSpec *cluster.Spec, secretName string, creds credentials.BasicAuthCredential) (map[string]interface{}, error) {
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
		"clusterName":              clusterSpec.Cluster.Name,
		"secretName":               secretName,
		"eksaSystemNamespace":      constants.EksaSystemNamespace,
		"base64EncodedCredentials": base64.StdEncoding.EncodeToString(credsJSON),
		"nutanixPCUsername":        creds.PrismCentral.BasicAuth.Username,
		"nutanixPCPassword":        creds.PrismCentral.BasicAuth.Password,
	}

	return values, nil
}

func generateNoProxyList(clusterSpec *cluster.Spec) []string {
	capacity := len(clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy) + 4

	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy...)

	// Add no-proxy defaults
	noProxyList = append(noProxyList, clusterapi.NoProxyDefaults()...)
	noProxyList = append(noProxyList,
		clusterSpec.Config.NutanixDatacenter.Spec.Endpoint,
		clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
	)

	return noProxyList
}
