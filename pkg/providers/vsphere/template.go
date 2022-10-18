package vsphere

import (
	"fmt"
	"net"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

func NewVsphereTemplateBuilder(
	datacenterSpec *anywherev1.VSphereDatacenterConfigSpec,
	controlPlaneMachineSpec, etcdMachineSpec *anywherev1.VSphereMachineConfigSpec,
	workerNodeGroupMachineSpecs map[string]anywherev1.VSphereMachineConfigSpec,
	now types.NowFunc,
	fromController bool,
) *VsphereTemplateBuilder {
	return &VsphereTemplateBuilder{
		datacenterSpec:              datacenterSpec,
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		now:                         now,
		fromController:              fromController,
	}
}

type VsphereTemplateBuilder struct {
	datacenterSpec              *anywherev1.VSphereDatacenterConfigSpec
	controlPlaneMachineSpec     *anywherev1.VSphereMachineConfigSpec
	WorkerNodeGroupMachineSpecs map[string]anywherev1.VSphereMachineConfigSpec
	etcdMachineSpec             *anywherev1.VSphereMachineConfigSpec
	now                         types.NowFunc
	fromController              bool
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecControlPlane(
	clusterSpec *cluster.Spec,
	buildOptions ...providers.BuildMapOption,
) (content []byte, err error) {
	var etcdMachineSpec anywherev1.VSphereMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *vs.etcdMachineSpec
	}
	values := buildTemplateMapCP(clusterSpec, *vs.datacenterSpec, *vs.controlPlaneMachineSpec, etcdMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (vs *VsphereTemplateBuilder) isCgroupDriverSystemd(clusterSpec *cluster.Spec) (bool, error) {
	bundle := clusterSpec.VersionsBundle
	k8sVersion, err := semver.New(bundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return false, fmt.Errorf("parsing kubernetes version %v: %v", bundle.KubeDistro.Kubernetes.Tag, err)
	}
	if vs.fromController && k8sVersion.Major == 1 && k8sVersion.Minor >= 21 {
		return true, nil
	}
	return false, nil
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecWorkers(
	clusterSpec *cluster.Spec,
	workloadTemplateNames,
	kubeadmconfigTemplateNames map[string]string,
) (content []byte, err error) {
	// pin cgroupDriver to systemd for k8s >= 1.21 when generating template in controller
	// remove this check once the controller supports order upgrade.
	// i.e. control plane, etcd upgrade before worker nodes.
	cgroupDriverSystemd, err := vs.isCgroupDriverSystemd(clusterSpec)
	if err != nil {
		return nil, err
	}

	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, *vs.datacenterSpec, vs.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		values["cgroupDriverSystemd"] = cgroupDriverSystemd

		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

func buildTemplateMapCP(
	clusterSpec *cluster.Spec,
	datacenterSpec anywherev1.VSphereDatacenterConfigSpec,
	controlPlaneMachineSpec, etcdMachineSpec anywherev1.VSphereMachineConfigSpec,
) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
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

	vuc := config.NewVsphereUserConfig()

	values := map[string]interface{}{
		"clusterName":                          clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":               clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":                 clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                 bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                    bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                       bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                         bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                    bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                       bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":             bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                   bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":             bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"thumbprint":                           datacenterSpec.Thumbprint,
		"vsphereDatacenter":                    datacenterSpec.Datacenter,
		"controlPlaneVsphereDatastore":         controlPlaneMachineSpec.Datastore,
		"controlPlaneVsphereFolder":            controlPlaneMachineSpec.Folder,
		"managerImage":                         bundle.VSphere.Manager.VersionedImage(),
		"kubeVipImage":                         bundle.VSphere.KubeVip.VersionedImage(),
		"driverImage":                          bundle.VSphere.Driver.VersionedImage(),
		"syncerImage":                          bundle.VSphere.Syncer.VersionedImage(),
		"insecure":                             datacenterSpec.Insecure,
		"vsphereNetwork":                       datacenterSpec.Network,
		"controlPlaneVsphereResourcePool":      controlPlaneMachineSpec.ResourcePool,
		"vsphereServer":                        datacenterSpec.Server,
		"controlPlaneVsphereStoragePolicyName": controlPlaneMachineSpec.StoragePolicyName,
		"vsphereTemplate":                      controlPlaneMachineSpec.Template,
		"controlPlaneVMsMemoryMiB":             controlPlaneMachineSpec.MemoryMiB,
		"controlPlaneVMsNumCPUs":               controlPlaneMachineSpec.NumCPUs,
		"controlPlaneDiskGiB":                  controlPlaneMachineSpec.DiskGiB,
		"controlPlaneSshUsername":              controlPlaneMachineSpec.Users[0].Name,
		"podCidrs":                             clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                         clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"etcdExtraArgs":                        etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                     crypto.SecureCipherSuitesString(),
		"apiserverExtraArgs":                   apiServerExtraArgs.ToPartialYaml(),
		"controllerManagerExtraArgs":           controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                   sharedExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                     kubeletExtraArgs.ToPartialYaml(),
		"format":                               format,
		"externalEtcdVersion":                  bundle.KubeDistro.EtcdVersion,
		"etcdImage":                            bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                  constants.EksaSystemNamespace,
		"auditPolicy":                          common.GetAuditPolicy(),
		"resourceSetName":                      resourceSetName(clusterSpec),
		"eksaVsphereUsername":                  vuc.EksaVsphereUsername,
		"eksaVspherePassword":                  vuc.EksaVspherePassword,
		"eksaCloudProviderUsername":            vuc.EksaVsphereCPUsername,
		"eksaCloudProviderPassword":            vuc.EksaVsphereCPPassword,
		"eksaCSIUsername":                      vuc.EksaVsphereCSIUsername,
		"eksaCSIPassword":                      vuc.EksaVsphereCSIPassword,
		"disableCSI":                           datacenterSpec.DisableCSI,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = net.JoinHostPort(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port)
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}

		if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate {
			values["registryAuth"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate
			username, password, _ := config.ReadCredentials()
			values["registryUsername"] = username
			values["registryPassword"] = password
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
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
			datacenterSpec.Server,
			clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdVsphereDatastore"] = etcdMachineSpec.Datastore
		values["etcdVsphereFolder"] = etcdMachineSpec.Folder
		values["etcdDiskGiB"] = etcdMachineSpec.DiskGiB
		values["etcdVMsMemoryMiB"] = etcdMachineSpec.MemoryMiB
		values["etcdVMsNumCPUs"] = etcdMachineSpec.NumCPUs
		values["etcdVsphereResourcePool"] = etcdMachineSpec.ResourcePool
		values["etcdVsphereStoragePolicyName"] = etcdMachineSpec.StoragePolicyName
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if controlPlaneMachineSpec.OSFamily == anywherev1.Bottlerocket {
		values["format"] = string(anywherev1.Bottlerocket)
		values["pauseRepository"] = bundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = bundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketBootstrap.Bootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketBootstrap.Bootstrap.Tag()
	}

	if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 {
		values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	return values
}

func buildTemplateMapMD(
	clusterSpec *cluster.Spec,
	datacenterSpec anywherev1.VSphereDatacenterConfigSpec,
	workerNodeGroupMachineSpec anywherev1.VSphereMachineConfigSpec,
	workerNodeGroupConfiguration anywherev1.WorkerNodeGroupConfiguration,
) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":                    clusterSpec.Cluster.Name,
		"kubernetesVersion":              bundle.KubeDistro.Kubernetes.Tag,
		"thumbprint":                     datacenterSpec.Thumbprint,
		"vsphereDatacenter":              datacenterSpec.Datacenter,
		"workerVsphereDatastore":         workerNodeGroupMachineSpec.Datastore,
		"workerVsphereFolder":            workerNodeGroupMachineSpec.Folder,
		"vsphereNetwork":                 datacenterSpec.Network,
		"workerVsphereResourcePool":      workerNodeGroupMachineSpec.ResourcePool,
		"vsphereServer":                  datacenterSpec.Server,
		"workerVsphereStoragePolicyName": workerNodeGroupMachineSpec.StoragePolicyName,
		"vsphereTemplate":                workerNodeGroupMachineSpec.Template,
		"workloadVMsMemoryMiB":           workerNodeGroupMachineSpec.MemoryMiB,
		"workloadVMsNumCPUs":             workerNodeGroupMachineSpec.NumCPUs,
		"workloadDiskGiB":                workerNodeGroupMachineSpec.DiskGiB,
		"workerSshUsername":              workerNodeGroupMachineSpec.Users[0].Name,
		"format":                         format,
		"eksaSystemNamespace":            constants.EksaSystemNamespace,
		"kubeletExtraArgs":               kubeletExtraArgs.ToPartialYaml(),
		"vsphereWorkerSshAuthorizedKey":  workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"workerReplicas":                 *workerNodeGroupConfiguration.Count,
		"workerNodeGroupName":            fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":          workerNodeGroupConfiguration.Taints,
		"autoscalingConfig":              workerNodeGroupConfiguration.AutoScalingConfiguration,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = net.JoinHostPort(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port)
		if len(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent
		}

		if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate {
			values["registryAuth"] = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate
			username, password, _ := config.ReadCredentials()
			values["registryUsername"] = username
			values["registryPassword"] = password
		}
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
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
			datacenterSpec.Server,
			clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if workerNodeGroupMachineSpec.OSFamily == anywherev1.Bottlerocket {
		values["format"] = string(anywherev1.Bottlerocket)
		values["pauseRepository"] = bundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = bundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketBootstrap.Bootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketBootstrap.Bootstrap.Tag()
	}

	return values
}
