package cloudstack

import (
	"fmt"
	"net"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

// TemplateBuilder is responsible for building the CAPI templates.
type TemplateBuilder struct {
	now types.NowFunc
}

// NewTemplateBuilder creates a new TemplateBuilder.
func NewTemplateBuilder(now types.NowFunc) *TemplateBuilder {
	return &TemplateBuilder{
		now: now,
	}
}

// GenerateCAPISpecControlPlane builds the CAPI controlplane template containing the CAPI objects for the control plane configuration defined in the cluster.Spec.
func (cs *TemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	if clusterSpec.CloudStackDatacenter == nil {
		return nil, fmt.Errorf("provided clusterSpec CloudStackDatacenter is nil. Unable to generate CAPI spec control plane")
	}
	var etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = etcdMachineConfig(clusterSpec).Spec
	}

	values, err := buildTemplateMapCP(clusterSpec)
	if err != nil {
		return nil, fmt.Errorf("error building template map from CP %v", err)
	}

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := buildControlPlaneTemplate(&controlPlaneMachineConfig(clusterSpec).Spec, values)
	if err != nil {
		return nil, err
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineTemplateBytes, err := buildEtcdTemplate(&etcdMachineSpec, values)
		if err != nil {
			return nil, fmt.Errorf("marshalling etcd machine template to byte array: %v", err)
		}
		bytes = append(bytes, etcdMachineTemplateBytes...)
	}

	return bytes, nil
}

// GenerateCAPISpecWorkers builds the CAPI worker template containing the CAPI objects for the worker node groups configuration defined in the cluster.Spec.
func (cs *TemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values, err := buildTemplateMapMD(clusterSpec, workerNodeGroupConfiguration)
		if err != nil {
			return nil, fmt.Errorf("building template map for MD %v", err)
		}

		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		values["autoscalingConfig"] = workerNodeGroupConfiguration.AutoScalingConfiguration

		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			values["upgradeRolloutStrategy"] = true
			values["maxSurge"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
			values["maxUnavailable"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable
		}

		// TODO: Extract out worker MachineDeployments from templates to use apibuilder instead
		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)

		workerMachineTemplateName := workloadTemplateNames[workerNodeGroupConfiguration.Name]
		workerMachineTemplate := MachineTemplate(workerMachineTemplateName, &workerMachineConfig(clusterSpec, workerNodeGroupConfiguration).Spec)
		workerMachineTemplateBytes, err := templater.ObjectsToYaml(workerMachineTemplate)
		if err != nil {
			return nil, fmt.Errorf("marshalling worker machine template to byte array: %v", err)
		}
		workerSpecs = append(workerSpecs, workerMachineTemplateBytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

// nolint:gocyclo
func buildTemplateMapCP(clusterSpec *cluster.Spec) (map[string]interface{}, error) {
	datacenterConfigSpec := clusterSpec.CloudStackDatacenter.Spec
	versionsBundle := clusterSpec.RootVersionsBundle()

	format := "cloud-config"
	host, port, err := getValidControlPlaneHostPort(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return nil, err
	}

	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig)).
		Append(clusterapi.EtcdEncryptionExtraArgs(clusterSpec.Cluster.Spec.EtcdEncryption)).
		Append(sharedExtraArgs)

	controllerManagerExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))

	controlPlaneMachineSpec := controlPlaneMachineConfig(clusterSpec).Spec
	controlPlaneSSHKey, err := common.StripSshAuthorizedKeyComment(controlPlaneMachineSpec.Users[0].SshAuthorizedKeys[0])
	if err != nil {
		return nil, fmt.Errorf("formatting ssh key for cloudstack control plane template: %v", err)
	}

	var etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec
	var etcdSSHAuthorizedKey string
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = etcdMachineConfig(clusterSpec).Spec
		etcdSSHAuthorizedKey, err = common.StripSshAuthorizedKeyComment(etcdMachineSpec.Users[0].SshAuthorizedKeys[0])
		if err != nil {
			return nil, fmt.Errorf("formatting ssh key for cloudstack etcd template: %v", err)
		}
	}

	values := map[string]interface{}{
		"clusterName":                                clusterSpec.Cluster.Name,
		"controlPlaneEndpointHost":                   host,
		"controlPlaneEndpointPort":                   port,
		"controlPlaneReplicas":                       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"apiServerCertSANs":                          clusterSpec.Cluster.Spec.ControlPlaneConfiguration.CertSANs,
		"kubernetesRepository":                       versionsBundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                          versionsBundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                             versionsBundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                               versionsBundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                          versionsBundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                             versionsBundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":                   versionsBundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                         versionsBundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                      versionsBundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":                   versionsBundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"managerImage":                               versionsBundle.CloudStack.ClusterAPIController.VersionedImage(),
		"kubeRbacProxyImage":                         versionsBundle.CloudStack.KubeRbacProxy.VersionedImage(),
		"kubeVipImage":                               versionsBundle.CloudStack.KubeVip.VersionedImage(),
		"cloudstackKubeVip":                          !features.IsActive(features.CloudStackKubeVipDisabled()),
		"cloudstackAvailabilityZones":                datacenterConfigSpec.AvailabilityZones,
		"cloudstackAnnotationSuffix":                 constants.CloudstackAnnotationSuffix,
		"cloudstackControlPlaneComputeOfferingId":    controlPlaneMachineSpec.ComputeOffering.Id,
		"cloudstackControlPlaneComputeOfferingName":  controlPlaneMachineSpec.ComputeOffering.Name,
		"cloudstackControlPlaneTemplateOfferingId":   controlPlaneMachineSpec.Template.Id,
		"cloudstackControlPlaneTemplateOfferingName": controlPlaneMachineSpec.Template.Name,
		"cloudstackControlPlaneCustomDetails":        controlPlaneMachineSpec.UserCustomDetails,
		"cloudstackControlPlaneSymlinks":             controlPlaneMachineSpec.Symlinks,
		"cloudstackControlPlaneAffinity":             controlPlaneMachineSpec.Affinity,
		"cloudstackControlPlaneAffinityGroupIds":     controlPlaneMachineSpec.AffinityGroupIds,
		"cloudstackEtcdComputeOfferingId":            etcdMachineSpec.ComputeOffering.Id,
		"cloudstackEtcdComputeOfferingName":          etcdMachineSpec.ComputeOffering.Name,
		"cloudstackEtcdTemplateOfferingId":           etcdMachineSpec.Template.Id,
		"cloudstackEtcdTemplateOfferingName":         etcdMachineSpec.Template.Name,
		"cloudstackEtcdCustomDetails":                etcdMachineSpec.UserCustomDetails,
		"cloudstackEtcdSymlinks":                     etcdMachineSpec.Symlinks,
		"cloudstackEtcdAffinity":                     etcdMachineSpec.Affinity,
		"cloudstackEtcdAffinityGroupIds":             etcdMachineSpec.AffinityGroupIds,
		"controlPlaneSshUsername":                    controlPlaneMachineSpec.Users[0].Name,
		"cloudstackControlPlaneSshAuthorizedKey":     controlPlaneSSHKey,
		"cloudstackEtcdSshAuthorizedKey":             etcdSSHAuthorizedKey,
		"podCidrs":                                   clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                               clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                         apiServerExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                           kubeletExtraArgs.ToPartialYaml(),
		"etcdExtraArgs":                              etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                           crypto.SecureCipherSuitesString(),
		"controllermanagerExtraArgs":                 controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                         sharedExtraArgs.ToPartialYaml(),
		"format":                                     format,
		"externalEtcdVersion":                        versionsBundle.KubeDistro.EtcdVersion,
		"externalEtcdReleaseUrl":                     versionsBundle.KubeDistro.EtcdURL,
		"etcdImage":                                  versionsBundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                        constants.EksaSystemNamespace,
	}

	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}
	values["auditPolicy"] = auditPolicy

	fillDiskOffering(values, controlPlaneMachineSpec.DiskOffering, "ControlPlane")
	fillDiskOffering(values, etcdMachineSpec.DiskOffering, "Etcd")

	values["cloudstackControlPlaneAnnotations"] = values["cloudstackControlPlaneDiskOfferingProvided"].(bool) || len(controlPlaneMachineSpec.Symlinks) > 0
	values["cloudstackEtcdAnnotations"] = values["cloudstackEtcdDiskOfferingProvided"].(bool) || len(etcdMachineSpec.Symlinks) > 0

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		values["publicMirror"] = containerd.ToAPIEndpoint(registryMirror.CoreEKSAMirror())
		if len(registryMirror.CACertContent) > 0 {
			values["registryCACert"] = registryMirror.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		fillProxyConfigurations(values, clusterSpec, net.JoinHostPort(host, port))
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
		etcdURL, _ := common.GetExternalEtcdReleaseURL(string(*clusterSpec.Cluster.Spec.EksaVersion), versionsBundle)
		if etcdURL != "" {
			values["externalEtcdReleaseUrl"] = etcdURL
		}
	}

	if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 {
		values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		values["maxSurge"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
	}

	if clusterSpec.Cluster.Spec.EtcdEncryption != nil && len(*clusterSpec.Cluster.Spec.EtcdEncryption) != 0 {
		conf, err := common.GenerateKMSEncryptionConfiguration(clusterSpec.Cluster.Spec.EtcdEncryption)
		if err != nil {
			return nil, err
		}
		values["encryptionProviderConfig"] = conf
	}

	return values, nil
}

func buildControlPlaneTemplate(machineSpec *v1alpha1.CloudStackMachineConfigSpec, values map[string]interface{}) (content []byte, err error) {
	templateName, ok := values[cpTemplateNameKey]
	if !ok {
		return nil, fmt.Errorf("unable to determine control plane template name")
	}
	machineTemplate := MachineTemplate(fmt.Sprintf("%s", templateName), machineSpec)
	templateBytes, err := templater.ObjectsToYaml(machineTemplate)
	if err != nil {
		return nil, fmt.Errorf("marshalling control plane machine template to byte array: %v", err)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, templateBytes...)

	return bytes, nil
}

func buildEtcdTemplate(machineSpec *v1alpha1.CloudStackMachineConfigSpec, values map[string]interface{}) (content []byte, err error) {
	machineTemplateName, ok := values[etcdTemplateNameKey]
	if !ok {
		return nil, fmt.Errorf("unable to determine etcd template name")
	}
	machineTemplate := MachineTemplate(fmt.Sprintf("%s", machineTemplateName), machineSpec)
	machineTemplateBytes, err := templater.ObjectsToYaml(machineTemplate)
	if err != nil {
		return nil, fmt.Errorf("marshalling etcd machine template to byte array: %v", err)
	}
	return machineTemplateBytes, nil
}

func fillDiskOffering(values map[string]interface{}, diskOffering *v1alpha1.CloudStackResourceDiskOffering, machineType string) {
	if diskOffering != nil {
		values[fmt.Sprintf("cloudstack%sDiskOfferingProvided", machineType)] = len(diskOffering.Id) > 0 || len(diskOffering.Name) > 0
		values[fmt.Sprintf("cloudstack%sDiskOfferingId", machineType)] = diskOffering.Id
		values[fmt.Sprintf("cloudstack%sDiskOfferingName", machineType)] = diskOffering.Name
		values[fmt.Sprintf("cloudstack%sDiskOfferingCustomSize", machineType)] = diskOffering.CustomSize
		values[fmt.Sprintf("cloudstack%sDiskOfferingPath", machineType)] = diskOffering.MountPath
		values[fmt.Sprintf("cloudstack%sDiskOfferingDevice", machineType)] = diskOffering.Device
		values[fmt.Sprintf("cloudstack%sDiskOfferingFilesystem", machineType)] = diskOffering.Filesystem
		values[fmt.Sprintf("cloudstack%sDiskOfferingLabel", machineType)] = diskOffering.Label
	} else {
		values[fmt.Sprintf("cloudstack%sDiskOfferingProvided", machineType)] = false
	}
}

func fillProxyConfigurations(values map[string]interface{}, clusterSpec *cluster.Spec, controlPlaneEndpoint string) {
	datacenterConfigSpec := clusterSpec.CloudStackDatacenter.Spec
	values["proxyConfig"] = true
	capacity := len(clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
		len(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy) + 4
	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy...)

	noProxyList = append(noProxyList, clusterapi.NoProxyDefaults()...)
	for _, az := range datacenterConfigSpec.AvailabilityZones {
		if cloudStackManagementAPIEndpointHostname, err := v1alpha1.GetCloudStackManagementAPIEndpointHostname(az); err == nil {
			noProxyList = append(noProxyList, cloudStackManagementAPIEndpointHostname)
		}
	}

	noProxyList = append(noProxyList, controlPlaneEndpoint)

	values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
	values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
	values["noProxy"] = noProxyList
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.WorkerNodeGroupVersionsBundle(workerNodeGroupConfiguration)
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	workerNodeGroupMachineSpec := workerMachineConfig(clusterSpec, workerNodeGroupConfiguration).Spec
	workerUser := workerNodeGroupMachineSpec.Users[0]
	workerSSHKey, err := common.StripSshAuthorizedKeyComment(workerUser.SshAuthorizedKeys[0])
	if err != nil {
		return nil, fmt.Errorf("formatting ssh key for cloudstack worker template: %v", err)
	}

	values := map[string]interface{}{
		"clusterName":                      clusterSpec.Cluster.Name,
		"kubernetesVersion":                versionsBundle.KubeDistro.Kubernetes.Tag,
		"cloudstackAnnotationSuffix":       constants.CloudstackAnnotationSuffix,
		"cloudstackTemplateId":             workerNodeGroupMachineSpec.Template.Id,
		"cloudstackTemplateName":           workerNodeGroupMachineSpec.Template.Name,
		"cloudstackOfferingId":             workerNodeGroupMachineSpec.ComputeOffering.Id,
		"cloudstackOfferingName":           workerNodeGroupMachineSpec.ComputeOffering.Name,
		"cloudstackCustomDetails":          workerNodeGroupMachineSpec.UserCustomDetails,
		"cloudstackSymlinks":               workerNodeGroupMachineSpec.Symlinks,
		"cloudstackAffinity":               workerNodeGroupMachineSpec.Affinity,
		"cloudstackAffinityGroupIds":       workerNodeGroupMachineSpec.AffinityGroupIds,
		"workerReplicas":                   *workerNodeGroupConfiguration.Count,
		"workerSshUsername":                workerNodeGroupMachineSpec.Users[0].Name,
		"cloudstackWorkerSshAuthorizedKey": workerSSHKey,
		"format":                           format,
		"kubeletExtraArgs":                 kubeletExtraArgs.ToPartialYaml(),
		"eksaSystemNamespace":              constants.EksaSystemNamespace,
		"workerNodeGroupName":              fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":            workerNodeGroupConfiguration.Taints,
	}
	fillDiskOffering(values, workerNodeGroupMachineSpec.DiskOffering, "")
	values["cloudstackAnnotations"] = values["cloudstackDiskOfferingProvided"].(bool) || len(workerNodeGroupMachineSpec.Symlinks) > 0

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		if len(registryMirror.CACertContent) > 0 {
			values["registryCACert"] = registryMirror.CACertContent
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		endpoint, err := controlPlaneEndpointHost(clusterSpec)
		if err != nil {
			return nil, err
		}
		fillProxyConfigurations(values, clusterSpec, endpoint)
	}

	return values, nil
}

func getEtcdMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) *v1alpha1.CloudStackMachineConfigSpec {
	var etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ExternalEtcdConfiguration != nil {
		if clusterSpec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	return etcdMachineSpec
}

func getControlPlaneMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) *v1alpha1.CloudStackMachineConfigSpec {
	var controlPlaneMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec
}
