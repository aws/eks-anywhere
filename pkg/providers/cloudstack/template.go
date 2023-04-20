package cloudstack

import (
	"fmt"
	"net"
	"time"

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

// NewTemplateBuilder creates a new CloudStack yaml TemplateBuilder.
func NewTemplateBuilder(CloudStackDatacenterConfigSpec *v1alpha1.CloudStackDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.CloudStackMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &TemplateBuilder{
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
	}
}

// TemplateBuilder is responsible for building the CAPI yaml templates.
type TemplateBuilder struct {
	controlPlaneMachineSpec     *v1alpha1.CloudStackMachineConfigSpec
	WorkerNodeGroupMachineSpecs map[string]v1alpha1.CloudStackMachineConfigSpec
	etcdMachineSpec             *v1alpha1.CloudStackMachineConfigSpec
	now                         types.NowFunc
}

// GenerateCAPISpecControlPlane builds the CAPI yaml controlplane template containing the CAPI objects for the control plane configuration.
// nolint:gocyclo
func (cs *TemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	if clusterSpec.CloudStackDatacenter == nil {
		return nil, fmt.Errorf("provided clusterSpec CloudStackDatacenter is nil. Unable to generate CAPI spec control plane")
	}
	var etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *cs.etcdMachineSpec
	}
	values, err := buildTemplateMapCP(clusterSpec, *cs.controlPlaneMachineSpec, etcdMachineSpec)
	if err != nil {
		return nil, fmt.Errorf("error building template map from CP %v", err)
	}

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	cpTemplateName, ok := values[cpTemplateNameKey]
	if !ok {
		return nil, fmt.Errorf("unable to determine control plane template name")
	}
	cpMachineTemplate := MachineTemplate(fmt.Sprintf("%s", cpTemplateName), cs.controlPlaneMachineSpec)
	cpMachineTemplateBytes, err := templater.ObjectsToYaml(cpMachineTemplate)
	if err != nil {
		return nil, fmt.Errorf("marshalling control plane machine template to byte array: %v", err)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, cpMachineTemplateBytes...)

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineTemplateName, ok := values[etcdTemplateNameKey]
		if !ok {
			return nil, fmt.Errorf("unable to determine etcd template name")
		}
		etcdMachineTemplate := MachineTemplate(fmt.Sprintf("%s", etcdMachineTemplateName), &etcdMachineSpec)
		etcdMachineTemplateBytes, err := templater.ObjectsToYaml(etcdMachineTemplate)
		if err != nil {
			return nil, fmt.Errorf("marshalling etcd machine template to byte array: %v", err)
		}
		bytes = append(bytes, etcdMachineTemplateBytes...)
	}

	return bytes, nil
}

// GenerateCAPISpecWorkers builds the CAPI worker yaml template containing the CAPI objects for the worker node groups configuration defined in the cluster.Spec.
// nolint:gocyclo
func (cs *TemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, cs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		values["autoscalingConfig"] = workerNodeGroupConfiguration.AutoScalingConfiguration

		// TODO: Extract out worker MachineDeployments from templates to use apibuilder instead
		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)

		workerMachineTemplateName := workloadTemplateNames[workerNodeGroupConfiguration.Name]
		machineConfig := cs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		workerMachineTemplate := MachineTemplate(workerMachineTemplateName, &machineConfig)
		workerMachineTemplateBytes, err := templater.ObjectsToYaml(workerMachineTemplate)
		if err != nil {
			return nil, fmt.Errorf("marshalling worker machine template to byte array: %v", err)
		}
		workerSpecs = append(workerSpecs, workerMachineTemplateBytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

// nolint:gocyclo
func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec) (map[string]interface{}, error) {
	datacenterConfigSpec := clusterSpec.CloudStackDatacenter.Spec
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	host, port, _ := net.SplitHostPort(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig)).
		Append(sharedExtraArgs)
	controllerManagerExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))

	values := map[string]interface{}{
		"clusterName":                                clusterSpec.Cluster.Name,
		"controlPlaneEndpointHost":                   host,
		"controlPlaneEndpointPort":                   port,
		"controlPlaneReplicas":                       clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                       bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                          bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                             bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                               bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                          bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                             bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":                   bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                         bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                      bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":                   bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"managerImage":                               bundle.CloudStack.ClusterAPIController.VersionedImage(),
		"kubeRbacProxyImage":                         bundle.CloudStack.KubeRbacProxy.VersionedImage(),
		"kubeVipImage":                               bundle.CloudStack.KubeVip.VersionedImage(),
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
		"podCidrs":                                   clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                               clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                         apiServerExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                           kubeletExtraArgs.ToPartialYaml(),
		"etcdExtraArgs":                              etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                           crypto.SecureCipherSuitesString(),
		"controllermanagerExtraArgs":                 controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                         sharedExtraArgs.ToPartialYaml(),
		"format":                                     format,
		"externalEtcdVersion":                        bundle.KubeDistro.EtcdVersion,
		"etcdImage":                                  bundle.KubeDistro.EtcdImage.VersionedImage(),
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
		fillProxyConfigurations(values, clusterSpec)
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
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

	return values, nil
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

func fillProxyConfigurations(values map[string]interface{}, clusterSpec *cluster.Spec) {
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
	noProxyList = append(noProxyList,
		clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
	)

	values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
	values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
	values["noProxy"] = noProxyList
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.CloudStackMachineConfigSpec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":                      clusterSpec.Cluster.Name,
		"kubernetesVersion":                bundle.KubeDistro.Kubernetes.Tag,
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
		"cloudstackWorkerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
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
		fillProxyConfigurations(values, clusterSpec)
	}

	if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		values["maxSurge"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
		values["maxUnavailable"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable
	}

	return values
}

func generateTemplateBuilder(clusterSpec *cluster.Spec) providers.TemplateBuilder {
	spec := v1alpha1.ClusterSpec{
		ControlPlaneConfiguration:     clusterSpec.Cluster.Spec.ControlPlaneConfiguration,
		WorkerNodeGroupConfigurations: clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		ExternalEtcdConfiguration:     clusterSpec.Cluster.Spec.ExternalEtcdConfiguration,
	}

	controlPlaneMachineSpec := getControlPlaneMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	workerNodeGroupMachineSpecs := getWorkerNodeGroupMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	etcdMachineSpec := getEtcdMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)

	templateBuilder := NewTemplateBuilder(
		&clusterSpec.CloudStackDatacenter.Spec,
		controlPlaneMachineSpec,
		etcdMachineSpec,
		workerNodeGroupMachineSpecs,
		time.Now,
	)
	return templateBuilder
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

func getWorkerNodeGroupMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) map[string]v1alpha1.CloudStackMachineConfigSpec {
	var workerNodeGroupMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(machineConfigs))
	for _, wnConfig := range clusterSpec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && machineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &machineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}

	return workerNodeGroupMachineSpecs
}

func getControlPlaneMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) *v1alpha1.CloudStackMachineConfigSpec {
	var controlPlaneMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec
}
