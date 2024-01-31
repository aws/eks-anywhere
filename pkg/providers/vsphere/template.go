package vsphere

import (
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

func NewVsphereTemplateBuilder(
	now types.NowFunc,
) *VsphereTemplateBuilder {
	return &VsphereTemplateBuilder{
		now: now,
	}
}

type VsphereTemplateBuilder struct {
	now types.NowFunc
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecControlPlane(
	clusterSpec *cluster.Spec,
	buildOptions ...providers.BuildMapOption,
) (content []byte, err error) {
	var etcdMachineSpec anywherev1.VSphereMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = etcdMachineConfig(clusterSpec).Spec
	}
	values, err := buildTemplateMapCP(
		clusterSpec,
		clusterSpec.VSphereDatacenter.Spec,
		controlPlaneMachineConfig(clusterSpec).Spec,
		etcdMachineSpec,
	)
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

func (vs *VsphereTemplateBuilder) isCgroupDriverSystemd(clusterSpec *cluster.Spec, worker anywherev1.WorkerNodeGroupConfiguration) (bool, error) {
	bundle := clusterSpec.WorkerNodeGroupVersionsBundle(worker)
	k8sVersion, err := semver.New(bundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return false, fmt.Errorf("parsing kubernetes version %v: %v", bundle.KubeDistro.Kubernetes.Tag, err)
	}
	if k8sVersion.Major == 1 && k8sVersion.Minor == 21 {
		return true, nil
	}
	return false, nil
}

// CAPIWorkersSpecWithInitialNames generates a yaml spec with the CAPI objects representing the worker
// nodes for a particular eks-a cluster. It uses default initial names (ended in '-1') for the vsphere
// machine templates and kubeadm config templates.
func (vs *VsphereTemplateBuilder) CAPIWorkersSpecWithInitialNames(spec *cluster.Spec) (content []byte, err error) {
	machineTemplateNames, kubeadmConfigTemplateNames := clusterapi.InitialTemplateNamesForWorkers(spec)
	return vs.GenerateCAPISpecWorkers(spec, machineTemplateNames, kubeadmConfigTemplateNames)
}

func (vs *VsphereTemplateBuilder) GenerateCAPISpecWorkers(
	clusterSpec *cluster.Spec,
	workloadTemplateNames,
	kubeadmconfigTemplateNames map[string]string,
) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		// pin cgroupDriver to systemd for k8s >= 1.21 when generating template in controller
		// remove this check once the controller supports order upgrade.
		// i.e. control plane, etcd upgrade before worker nodes.
		cgroupDriverSystemd, err := vs.isCgroupDriverSystemd(clusterSpec, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		values, err := buildTemplateMapMD(
			clusterSpec,
			clusterSpec.VSphereDatacenter.Spec,
			workerMachineConfig(clusterSpec, workerNodeGroupConfiguration).Spec,
			workerNodeGroupConfiguration,
		)
		if err != nil {
			return nil, err
		}

		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		values["cgroupDriverSystemd"] = cgroupDriverSystemd

		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			values["upgradeRolloutStrategy"] = true
			if workerNodeGroupConfiguration.UpgradeRolloutStrategy.Type == anywherev1.InPlaceStrategyType {
				values["upgradeRolloutStrategyType"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.Type
			} else {
				values["maxSurge"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
				values["maxUnavailable"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable
			}
		}

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
) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.RootVersionsBundle()
	format := "cloud-config"
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

	vuc := config.NewVsphereUserConfig()

	firstControlPlaneMachinesUser := controlPlaneMachineSpec.Users[0]
	controlPlaneSSHKey, err := common.StripSshAuthorizedKeyComment(firstControlPlaneMachinesUser.SshAuthorizedKeys[0])
	if err != nil {
		return nil, fmt.Errorf("formatting ssh key for vsphere control plane template: %v", err)
	}

	values := map[string]interface{}{
		"clusterName":                          clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":               clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":                 clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"apiServerCertSANs":                    clusterSpec.Cluster.Spec.ControlPlaneConfiguration.CertSANs,
		"kubernetesRepository":                 versionsBundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                    versionsBundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                       versionsBundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                         versionsBundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                    versionsBundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                       versionsBundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":             versionsBundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                   versionsBundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                versionsBundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":             versionsBundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"thumbprint":                           datacenterSpec.Thumbprint,
		"vsphereDatacenter":                    datacenterSpec.Datacenter,
		"controlPlaneVsphereDatastore":         controlPlaneMachineSpec.Datastore,
		"controlPlaneVsphereFolder":            controlPlaneMachineSpec.Folder,
		"managerImage":                         versionsBundle.VSphere.Manager.VersionedImage(),
		"kubeVipImage":                         versionsBundle.VSphere.KubeVip.VersionedImage(),
		"insecure":                             datacenterSpec.Insecure,
		"vsphereNetwork":                       datacenterSpec.Network,
		"controlPlaneVsphereResourcePool":      controlPlaneMachineSpec.ResourcePool,
		"vsphereServer":                        datacenterSpec.Server,
		"controlPlaneVsphereStoragePolicyName": controlPlaneMachineSpec.StoragePolicyName,
		"controlPlaneTemplate":                 controlPlaneMachineSpec.Template,
		"etcdTemplate":                         etcdMachineSpec.Template,
		"controlPlaneVMsMemoryMiB":             controlPlaneMachineSpec.MemoryMiB,
		"controlPlaneVMsNumCPUs":               controlPlaneMachineSpec.NumCPUs,
		"controlPlaneDiskGiB":                  controlPlaneMachineSpec.DiskGiB,
		"controlPlaneTagIDs":                   controlPlaneMachineSpec.TagIDs,
		"etcdTagIDs":                           etcdMachineSpec.TagIDs,
		"controlPlaneSshUsername":              firstControlPlaneMachinesUser.Name,
		"vsphereControlPlaneSshAuthorizedKey":  controlPlaneSSHKey,
		"podCidrs":                             clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                         clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"etcdExtraArgs":                        etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":                     crypto.SecureCipherSuitesString(),
		"apiserverExtraArgs":                   apiServerExtraArgs.ToPartialYaml(),
		"controllerManagerExtraArgs":           controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":                   sharedExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":                     kubeletExtraArgs.ToPartialYaml(),
		"format":                               format,
		"externalEtcdVersion":                  versionsBundle.KubeDistro.EtcdVersion,
		"etcdImage":                            versionsBundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                  constants.EksaSystemNamespace,
		"cpiResourceSetName":                   cpiResourceSetName(clusterSpec),
		"eksaVsphereUsername":                  vuc.EksaVsphereUsername,
		"eksaVspherePassword":                  vuc.EksaVspherePassword,
		"eksaCloudProviderUsername":            vuc.EksaVsphereCPUsername,
		"eksaCloudProviderPassword":            vuc.EksaVsphereCPPassword,
		"controlPlaneCloneMode":                controlPlaneMachineSpec.CloneMode,
		"etcdCloneMode":                        etcdMachineSpec.CloneMode,
	}

	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}
	values["auditPolicy"] = auditPolicy

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		values["publicMirror"] = containerd.ToAPIEndpoint(registryMirror.CoreEKSAMirror())
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
		firstEtcdMachinesUser := etcdMachineSpec.Users[0]
		etcdSSHKey, err := common.StripSshAuthorizedKeyComment(firstEtcdMachinesUser.SshAuthorizedKeys[0])
		if err != nil {
			return nil, fmt.Errorf("formatting ssh key for vsphere etcd template: %v", err)
		}

		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdVsphereDatastore"] = etcdMachineSpec.Datastore
		values["etcdVsphereFolder"] = etcdMachineSpec.Folder
		values["etcdDiskGiB"] = etcdMachineSpec.DiskGiB
		values["etcdVMsMemoryMiB"] = etcdMachineSpec.MemoryMiB
		values["etcdVMsNumCPUs"] = etcdMachineSpec.NumCPUs
		values["etcdVsphereResourcePool"] = etcdMachineSpec.ResourcePool
		values["etcdVsphereStoragePolicyName"] = etcdMachineSpec.StoragePolicyName
		values["etcdSshUsername"] = firstEtcdMachinesUser.Name
		values["vsphereEtcdSshAuthorizedKey"] = etcdSSHKey

		if etcdMachineSpec.HostOSConfiguration != nil {
			if etcdMachineSpec.HostOSConfiguration.NTPConfiguration != nil {
				values["etcdNtpServers"] = etcdMachineSpec.HostOSConfiguration.NTPConfiguration.Servers
			}

			if etcdMachineSpec.HostOSConfiguration.CertBundles != nil {
				values["etcdCertBundles"] = etcdMachineSpec.HostOSConfiguration.CertBundles
			}

			if etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration != nil {
				if etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Kernel != nil &&
					etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Kernel.SysctlSettings != nil {
					values["etcdKernelSettings"] = etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Kernel.SysctlSettings
				}
				if etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Boot != nil &&
					etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Boot.BootKernelParameters != nil {
					values["etcdBootParameters"] = etcdMachineSpec.HostOSConfiguration.BottlerocketConfiguration.Boot.BootKernelParameters
				}
			}
		}
		etcdURL, _ := common.GetExternalEtcdReleaseURL(string(*clusterSpec.Cluster.Spec.EksaVersion), versionsBundle)
		if etcdURL != "" {
			values["externalEtcdReleaseUrl"] = etcdURL
		}
	}

	if controlPlaneMachineSpec.OSFamily == anywherev1.Bottlerocket {
		values["format"] = string(anywherev1.Bottlerocket)
		values["pauseRepository"] = versionsBundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = versionsBundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Image()
		values["bottlerocketBootstrapVersion"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Tag()
	}

	if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 {
		values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	if controlPlaneMachineSpec.HostOSConfiguration != nil {
		if controlPlaneMachineSpec.HostOSConfiguration.NTPConfiguration != nil {
			values["cpNtpServers"] = controlPlaneMachineSpec.HostOSConfiguration.NTPConfiguration.Servers
		}

		if controlPlaneMachineSpec.HostOSConfiguration.CertBundles != nil {
			values["certBundles"] = controlPlaneMachineSpec.HostOSConfiguration.CertBundles
		}

		brSettings, err := common.GetCAPIBottlerocketSettingsConfig(controlPlaneMachineSpec.HostOSConfiguration.BottlerocketConfiguration)
		if err != nil {
			return nil, err
		}
		values["bottlerocketSettings"] = brSettings
	}

	if clusterSpec.Cluster.Spec.EtcdEncryption != nil && len(*clusterSpec.Cluster.Spec.EtcdEncryption) != 0 {
		conf, err := common.GenerateKMSEncryptionConfiguration(clusterSpec.Cluster.Spec.EtcdEncryption)
		if err != nil {
			return nil, err
		}
		values["encryptionProviderConfig"] = conf
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type == anywherev1.InPlaceStrategyType {
			values["upgradeRolloutStrategyType"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type
		} else {
			values["maxSurge"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
		}
	}

	return values, nil
}

func buildTemplateMapMD(
	clusterSpec *cluster.Spec,
	datacenterSpec anywherev1.VSphereDatacenterConfigSpec,
	workerNodeGroupMachineSpec anywherev1.VSphereMachineConfigSpec,
	workerNodeGroupConfiguration anywherev1.WorkerNodeGroupConfiguration,
) (map[string]interface{}, error) {
	bundle := clusterSpec.WorkerNodeGroupVersionsBundle(workerNodeGroupConfiguration)
	if bundle == nil {
		return nil, fmt.Errorf("could not find VersionsBundle")
	}
	format := "cloud-config"
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	firstUser := workerNodeGroupMachineSpec.Users[0]
	sshKey, err := common.StripSshAuthorizedKeyComment(firstUser.SshAuthorizedKeys[0])
	if err != nil {
		return nil, fmt.Errorf("formatting ssh key for vsphere workers template: %v", err)
	}

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
		"workerTemplate":                 workerNodeGroupMachineSpec.Template,
		"workloadVMsMemoryMiB":           workerNodeGroupMachineSpec.MemoryMiB,
		"workloadVMsNumCPUs":             workerNodeGroupMachineSpec.NumCPUs,
		"workloadDiskGiB":                workerNodeGroupMachineSpec.DiskGiB,
		"workerTagIDs":                   workerNodeGroupMachineSpec.TagIDs,
		"workerSshUsername":              firstUser.Name,
		"vsphereWorkerSshAuthorizedKey":  sshKey,
		"format":                         format,
		"eksaSystemNamespace":            constants.EksaSystemNamespace,
		"kubeletExtraArgs":               kubeletExtraArgs.ToPartialYaml(),
		"workerReplicas":                 *workerNodeGroupConfiguration.Count,
		"workerNodeGroupName":            fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints":          workerNodeGroupConfiguration.Taints,
		"autoscalingConfig":              workerNodeGroupConfiguration.AutoScalingConfiguration,
		"workerCloneMode":                workerNodeGroupMachineSpec.CloneMode,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
		values["registryMirrorMap"] = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		values["mirrorBase"] = registryMirror.BaseRegistry
		values["insecureSkip"] = registryMirror.InsecureSkipVerify
		values["publicMirror"] = containerd.ToAPIEndpoint(registryMirror.CoreEKSAMirror())
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
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketHostContainers.KubeadmBootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketHostContainers.KubeadmBootstrap.Tag()
	}

	if workerNodeGroupMachineSpec.HostOSConfiguration != nil {
		if workerNodeGroupMachineSpec.HostOSConfiguration.NTPConfiguration != nil {
			values["ntpServers"] = workerNodeGroupMachineSpec.HostOSConfiguration.NTPConfiguration.Servers
		}

		if workerNodeGroupMachineSpec.HostOSConfiguration.CertBundles != nil {
			values["certBundles"] = workerNodeGroupMachineSpec.HostOSConfiguration.CertBundles
		}

		brSettings, err := common.GetCAPIBottlerocketSettingsConfig(workerNodeGroupMachineSpec.HostOSConfiguration.BottlerocketConfiguration)
		if err != nil {
			return nil, err
		}
		values["bottlerocketSettings"] = brSettings
	}

	return values, nil
}
