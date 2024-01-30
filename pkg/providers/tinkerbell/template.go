package tinkerbell

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	unstructuredutil "github.com/aws/eks-anywhere/pkg/utils/unstructured"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

const (
	TinkerbellMachineTemplateKind = "TinkerbellMachineTemplate"
	defaultRegistry               = "public.ecr.aws"
)

type TemplateBuilder struct {
	controlPlaneMachineSpec     *v1alpha1.TinkerbellMachineConfigSpec
	datacenterSpec              *v1alpha1.TinkerbellDatacenterConfigSpec
	WorkerNodeGroupMachineSpecs map[string]v1alpha1.TinkerbellMachineConfigSpec
	etcdMachineSpec             *v1alpha1.TinkerbellMachineConfigSpec
	tinkerbellIP                string
	now                         types.NowFunc
}

// NewTemplateBuilder creates a new TemplateBuilder instance.
func NewTemplateBuilder(datacenterSpec *v1alpha1.TinkerbellDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.TinkerbellMachineConfigSpec, tinkerbellIP string, now types.NowFunc) providers.TemplateBuilder {
	return &TemplateBuilder{
		controlPlaneMachineSpec:     controlPlaneMachineSpec,
		datacenterSpec:              datacenterSpec,
		WorkerNodeGroupMachineSpecs: workerNodeGroupMachineSpecs,
		etcdMachineSpec:             etcdMachineSpec,
		tinkerbellIP:                tinkerbellIP,
		now:                         now,
	}
}

func (tb *TemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	cpTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[tb.controlPlaneMachineSpec.TemplateRef.Name]
	bundle := clusterSpec.RootVersionsBundle()
	if err != nil {
		return nil, err
	}
	var OSImageURL string

	if cpTemplateConfig == nil {
		OSImageURL = clusterSpec.TinkerbellDatacenter.Spec.OSImageURL
		if tb.controlPlaneMachineSpec.OSImageURL != "" {
			OSImageURL = tb.controlPlaneMachineSpec.OSImageURL
		}
		versionBundle := bundle.VersionsBundle
		cpTemplateConfig = v1alpha1.NewDefaultTinkerbellTemplateConfigCreate(clusterSpec.Cluster, *versionBundle, OSImageURL, tb.tinkerbellIP, tb.datacenterSpec.TinkerbellIP, tb.controlPlaneMachineSpec.OSFamily)
	}

	cpTemplateString, err := cpTemplateConfig.ToTemplateString()
	if err != nil {
		return nil, fmt.Errorf("failed to get Control Plane TinkerbellTemplateConfig: %v", err)
	}

	var etcdMachineSpec v1alpha1.TinkerbellMachineConfigSpec
	var etcdTemplateString string
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *tb.etcdMachineSpec
		OSImageURL = clusterSpec.TinkerbellDatacenter.Spec.OSImageURL
		if etcdMachineSpec.OSImageURL != "" {
			OSImageURL = etcdMachineSpec.OSImageURL
		}
		etcdTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[tb.etcdMachineSpec.TemplateRef.Name]
		if etcdTemplateConfig == nil {
			versionBundle := bundle.VersionsBundle
			etcdTemplateConfig = v1alpha1.NewDefaultTinkerbellTemplateConfigCreate(clusterSpec.Cluster, *versionBundle, OSImageURL, tb.tinkerbellIP, tb.datacenterSpec.TinkerbellIP, tb.etcdMachineSpec.OSFamily)
		}
		etcdTemplateString, err = etcdTemplateConfig.ToTemplateString()
		if err != nil {
			return nil, fmt.Errorf("failed to get ETCD TinkerbellTemplateConfig: %v", err)
		}
	}
	values, err := buildTemplateMapCP(clusterSpec, *tb.controlPlaneMachineSpec, etcdMachineSpec, cpTemplateString, etcdTemplateString, *tb.datacenterSpec)
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

func (tb *TemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	bundle := clusterSpec.RootVersionsBundle()
	OSImageURL := clusterSpec.TinkerbellDatacenter.Spec.OSImageURL

	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeMachineSpec := tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		wTemplateConfig := clusterSpec.TinkerbellTemplateConfigs[workerNodeMachineSpec.TemplateRef.Name]
		if wTemplateConfig == nil {
			versionBundle := bundle.VersionsBundle
			if workerNodeMachineSpec.OSImageURL != "" {
				OSImageURL = workerNodeMachineSpec.OSImageURL
			}
			wTemplateConfig = v1alpha1.NewDefaultTinkerbellTemplateConfigCreate(clusterSpec.Cluster, *versionBundle, OSImageURL, tb.tinkerbellIP, tb.datacenterSpec.TinkerbellIP, workerNodeMachineSpec.OSFamily)
		}

		wTemplateString, err := wTemplateConfig.ToTemplateString()
		if err != nil {
			return nil, fmt.Errorf("failed to get worker TinkerbellTemplateConfig: %v", err)
		}

		values, err := buildTemplateMapMD(clusterSpec, tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name], workerNodeGroupConfiguration, wTemplateString, *tb.datacenterSpec)
		if err != nil {
			return nil, err
		}

		_, ok := workloadTemplateNames[workerNodeGroupConfiguration.Name]
		if workloadTemplateNames == nil || !ok {
			return nil, fmt.Errorf("workloadTemplateNames invalid in GenerateCAPISpecWorkers: %v", err)
		}
		_, ok = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		if kubeadmconfigTemplateNames == nil || !ok {
			return nil, fmt.Errorf("kubeadmconfigTemplateNames invalid in GenerateCAPISpecWorkers: %v", err)
		}
		values["workerSshAuthorizedKey"] = tb.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name].Users[0].SshAuthorizedKeys[0]
		values["workerReplicas"] = *workerNodeGroupConfiguration.Count
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workerNodeGroupName"] = workerNodeGroupConfiguration.Name
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]
		values["autoscalingConfig"] = workerNodeGroupConfiguration.AutoScalingConfiguration

		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			values["upgradeRolloutStrategy"] = true
			if workerNodeGroupConfiguration.UpgradeRolloutStrategy.Type == v1alpha1.InPlaceStrategyType {
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

func (p *Provider) generateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := newClusterSpec.Cluster.Name
	var controlPlaneTemplateName, workloadTemplateName, kubeadmconfigTemplateName, etcdTemplateName string
	var needsNewEtcdTemplate bool

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return nil, nil, err
	}
	vdc, err := p.providerKubectlClient.GetEksaTinkerbellDatacenterConfig(ctx, p.datacenterConfig.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
	if err != nil {
		return nil, nil, err
	}
	needsNewControlPlaneTemplate := needsNewControlPlaneTemplate(currentSpec, newClusterSpec)
	if !needsNewControlPlaneTemplate {
		cp, err := p.providerKubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, c.Name, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, nil, err
		}
		controlPlaneTemplateName = cp.Spec.MachineTemplate.InfrastructureRef.Name
	} else {
		controlPlaneTemplateName = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
	}

	previousWorkerNodeGroupConfigs := cluster.BuildMapForWorkerNodeGroupsByName(currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations)

	workloadTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		needsNewWorkloadTemplate, err := p.needsNewMachineTemplate(ctx, workloadCluster, currentSpec, newClusterSpec, workerNodeGroupConfiguration, vdc, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}

		needsNewKubeadmConfigTemplate, err := p.needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration, previousWorkerNodeGroupConfigs)
		if err != nil {
			return nil, nil, err
		}
		if !needsNewKubeadmConfigTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			kubeadmconfigTemplateName = md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		} else {
			kubeadmconfigTemplateName = common.KubeadmConfigTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = kubeadmconfigTemplateName
		}

		if !needsNewWorkloadTemplate {
			mdName := machineDeploymentName(newClusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name)
			md, err := p.providerKubectlClient.GetMachineDeployment(ctx, mdName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			workloadTemplateName = common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	// @TODO: upgrade of external etcd
	if newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		// etcdMachineConfig := p.machineConfigs[newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		// etcdMachineTmc, err := p.providerKubectlClient.GetEksaTinkerbellMachineConfig(ctx, c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Cluster.Namespace)
		// if err != nil {
		//	return nil, nil, err
		// }
		// needsNewEtcdTemplate = NeedsNewEtcdTemplate(currentSpec, newClusterSpec, vdc, p.datacenterConfig, etcdMachineTmc, etcdMachineConfig)
		/*** @TODO: hardcoding this to false, remove later *****/
		needsNewEtcdTemplate = false
		if !needsNewEtcdTemplate {
			etcdadmCluster, err := p.providerKubectlClient.GetEtcdadmCluster(ctx, workloadCluster, clusterName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = etcdadmCluster.Spec.InfrastructureTemplate.Name
		} else {
			/* During a cluster upgrade, etcd machines need to be upgraded first, so that the etcd machines with new spec get created and can be used by controlplane machines
			as etcd endpoints. KCP rollout should not start until then. As a temporary solution in the absence of static etcd endpoints, we annotate the etcd cluster as "upgrading",
			so that KCP checks this annotation and does not proceed if etcd cluster is upgrading. The etcdadm controller removes this annotation once the etcd upgrade is complete.
			*/
			err = p.providerKubectlClient.UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", clusterName),
				map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"},
				executables.WithCluster(bootstrapCluster),
				executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["controlPlaneSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		if newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
			values["etcdSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		}
		values["etcdTemplateName"] = etcdTemplateName
	}

	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(newClusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(newClusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}

	if p.isScaleUpDown(currentSpec.Cluster, newClusterSpec.Cluster) {
		cpSpec, err := omitTinkerbellMachineTemplate(controlPlaneSpec)
		if err == nil {
			if wSpec, err := omitTinkerbellMachineTemplate(workersSpec); err == nil {
				return cpSpec, wSpec, nil
			}
		}
	}

	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForUpgrade(ctx, bootstrapCluster, workloadCluster, currentSpec, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) GenerateCAPISpecForCreate(ctx context.Context, _ *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, clusterSpec)

	if err != nil {
		return nil, nil, fmt.Errorf("generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) generateCAPISpecForCreate(ctx context.Context, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.Cluster.Name
	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, p.templateBuilder.now)
		values["controlPlaneSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
			values["etcdSshAuthorizedKey"] = p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0].SshAuthorizedKeys[0]
		}
		values["etcdTemplateName"] = common.EtcdMachineTemplateName(clusterName, p.templateBuilder.now)
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = common.WorkerMachineTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name, p.templateBuilder.now)
		p.templateBuilder.WorkerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = p.machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}
	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *Provider) needsNewMachineTemplate(ctx context.Context, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, vdc *v1alpha1.TinkerbellDatacenterConfig, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if prevWorkerNode, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		return needsNewWorkloadTemplate(currentSpec, newClusterSpec, prevWorkerNode, workerNodeGroupConfiguration), nil
	}
	return true, nil
}

func (p *Provider) needsNewKubeadmConfigTemplate(workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration, prevWorkerNodeGroupConfigs map[string]v1alpha1.WorkerNodeGroupConfiguration) (bool, error) {
	if _, ok := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]; ok {
		existingWorkerNodeGroupConfig := prevWorkerNodeGroupConfigs[workerNodeGroupConfiguration.Name]
		return needsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, &existingWorkerNodeGroupConfig), nil
	}
	return true, nil
}

func machineDeploymentName(clusterName, nodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, nodeGroupName)
}

// nolint:gocyclo
func buildTemplateMapCP(
	clusterSpec *cluster.Spec,
	controlPlaneMachineSpec,
	etcdMachineSpec v1alpha1.TinkerbellMachineConfigSpec,
	cpTemplateOverride,
	etcdTemplateOverride string,
	datacenterSpec v1alpha1.TinkerbellDatacenterConfigSpec,
) (map[string]interface{}, error) {
	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle := clusterSpec.RootVersionsBundle()
	format := "cloud-config"

	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig))

	// LoadBalancerClass is feature gated in K8S v1.21 and needs to be enabled manually
	if clusterSpec.Cluster.Spec.KubernetesVersion == v1alpha1.Kube121 {
		apiServerExtraArgs.Append(clusterapi.FeatureGatesExtraArgs("ServiceLoadBalancerClass=true"))
	}

	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))

	values := map[string]interface{}{
		"auditPolicy":                   auditPolicy,
		"clusterName":                   clusterSpec.Cluster.Name,
		"controlPlaneEndpointIp":        clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":          clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"apiServerCertSANs":             clusterSpec.Cluster.Spec.ControlPlaneConfiguration.CertSANs,
		"controlPlaneSshAuthorizedKey":  controlPlaneMachineSpec.Users[0].SshAuthorizedKeys[0],
		"controlPlaneSshUsername":       controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":           constants.EksaSystemNamespace,
		"format":                        format,
		"kubernetesVersion":             versionsBundle.KubeDistro.Kubernetes.Tag,
		"kubeVipImage":                  versionsBundle.Tinkerbell.KubeVip.VersionedImage(),
		"podCidrs":                      clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                  clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":            apiServerExtraArgs.ToPartialYaml(),
		"baseRegistry":                  "", // TODO: need to get this values for creating template IMAGE_URL
		"osDistro":                      "", // TODO: need to get this values for creating template IMAGE_URL
		"osVersion":                     "", // TODO: need to get this values for creating template IMAGE_URL
		"kubernetesRepository":          versionsBundle.KubeDistro.Kubernetes.Repository,
		"corednsRepository":             versionsBundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                versionsBundle.KubeDistro.CoreDNS.Tag,
		"etcdRepository":                versionsBundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                  versionsBundle.KubeDistro.Etcd.Tag,
		"externalEtcdVersion":           versionsBundle.KubeDistro.EtcdVersion,
		"etcdCipherSuites":              crypto.SecureCipherSuitesString(),
		"kubeletExtraArgs":              kubeletExtraArgs.ToPartialYaml(),
		"hardwareSelector":              controlPlaneMachineSpec.HardwareSelector,
		"controlPlaneTaints":            clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints,
		"workerNodeGroupConfigurations": clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		"skipLoadBalancerDeployment":    datacenterSpec.SkipLoadBalancerDeployment,
		"cpSkipLoadBalancerDeployment":  clusterSpec.Cluster.Spec.ControlPlaneConfiguration.SkipLoadBalancerDeployment,
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type == v1alpha1.InPlaceStrategyType {
			values["upgradeRolloutStrategyType"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type
		} else {
			values["maxSurge"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
		}
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values, err := populateRegistryMirrorValues(clusterSpec, values)
		if err != nil {
			return values, err
		}

		// Replace public.ecr.aws endpoint with the endpoint given in the cluster config file
		localRegistry := values["publicMirror"].(string)
		cpTemplateOverride = strings.ReplaceAll(cpTemplateOverride, defaultRegistry, localRegistry)
		etcdTemplateOverride = strings.ReplaceAll(etcdTemplateOverride, defaultRegistry, localRegistry)
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = generateNoProxyList(clusterSpec.Cluster, datacenterSpec)
	}

	values["controlPlanetemplateOverride"] = cpTemplateOverride

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
		values["etcdTemplateOverride"] = etcdTemplateOverride
		values["etcdHardwareSelector"] = etcdMachineSpec.HardwareSelector
		etcdURL, _ := common.GetExternalEtcdReleaseURL(string(*clusterSpec.Cluster.Spec.EksaVersion), versionsBundle)
		if etcdURL != "" {
			values["externalEtcdReleaseUrl"] = etcdURL
		}
	}

	if controlPlaneMachineSpec.OSFamily == v1alpha1.Bottlerocket {
		values["format"] = string(v1alpha1.Bottlerocket)
		values["pauseRepository"] = versionsBundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = versionsBundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Image()
		values["bottlerocketBootstrapVersion"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Tag()
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

	return values, nil
}

func buildTemplateMapMD(
	clusterSpec *cluster.Spec,
	workerNodeGroupMachineSpec v1alpha1.TinkerbellMachineConfigSpec,
	workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration,
	workerTemplateOverride string,
	datacenterSpec v1alpha1.TinkerbellDatacenterConfigSpec,
) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.WorkerNodeGroupVersionsBundle(workerNodeGroupConfiguration)
	format := "cloud-config"

	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":            clusterSpec.Cluster.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"kubeletExtraArgs":       kubeletExtraArgs.ToPartialYaml(),
		"format":                 format,
		"kubernetesVersion":      versionsBundle.KubeDistro.Kubernetes.Tag,
		"workerNodeGroupName":    workerNodeGroupConfiguration.Name,
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys[0],
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
		"hardwareSelector":       workerNodeGroupMachineSpec.HardwareSelector,
		"workerNodeGroupTaints":  workerNodeGroupConfiguration.Taints,
	}

	if workerNodeGroupMachineSpec.OSFamily == v1alpha1.Bottlerocket {
		values["format"] = string(v1alpha1.Bottlerocket)
		values["pauseRepository"] = versionsBundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = versionsBundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Image()
		values["bottlerocketBootstrapVersion"] = versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.Tag()
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values, err := populateRegistryMirrorValues(clusterSpec, values)
		if err != nil {
			return values, err
		}

		// Replace public.ecr.aws endpoint with the endpoint given in the cluster config file
		localRegistry := values["publicMirror"].(string)
		workerTemplateOverride = strings.ReplaceAll(workerTemplateOverride, defaultRegistry, localRegistry)
	}

	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		values["httpProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = generateNoProxyList(clusterSpec.Cluster, datacenterSpec)
	}

	values["workertemplateOverride"] = workerTemplateOverride

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

// omitTinkerbellMachineTemplate removes TinkerbellMachineTemplate API objects from yml. yml is
// typically an EKSA cluster configuration.
func omitTinkerbellMachineTemplate(yml []byte) ([]byte, error) {
	var filtered []unstructured.Unstructured

	r := yamlutil.NewYAMLReader(bufio.NewReader(bytes.NewReader(yml)))
	for {
		d, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := yamlutil.Unmarshal(d, &m); err != nil {
			return nil, err
		}
		var u unstructured.Unstructured
		u.SetUnstructuredContent(m)

		// Omit TinkerbellMachineTemplate kind.
		if u.GetKind() == TinkerbellMachineTemplateKind {
			continue
		}

		filtered = append(filtered, u)
	}

	return unstructuredutil.UnstructuredToYaml(filtered)
}

func populateRegistryMirrorValues(clusterSpec *cluster.Spec, values map[string]interface{}) (map[string]interface{}, error) {
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
	return values, nil
}

func getControlPlaneMachineSpec(clusterSpec *cluster.Spec) (*v1alpha1.TinkerbellMachineConfigSpec, error) {
	var controlPlaneMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec, nil
}

func getWorkerNodeGroupMachineSpec(clusterSpec *cluster.Spec) (map[string]v1alpha1.TinkerbellMachineConfigSpec, error) {
	var workerNodeGroupMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.TinkerbellMachineConfigSpec, len(clusterSpec.TinkerbellMachineConfigs))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &clusterSpec.TinkerbellMachineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}

	return workerNodeGroupMachineSpecs, nil
}

func getEtcdMachineSpec(clusterSpec *cluster.Spec) (*v1alpha1.TinkerbellMachineConfigSpec, error) {
	var etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	return etcdMachineSpec, nil
}

func generateTemplateBuilder(clusterSpec *cluster.Spec) (providers.TemplateBuilder, error) {
	controlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating control plane machine spec")
	}

	workerNodeGroupMachineSpecs, err := getWorkerNodeGroupMachineSpec(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating worker node group machine specs")
	}

	etcdMachineSpec, err := getEtcdMachineSpec(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating etcd machine spec")
	}

	templateBuilder := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec,
		controlPlaneMachineSpec,
		etcdMachineSpec,
		workerNodeGroupMachineSpecs,
		clusterSpec.TinkerbellDatacenter.Spec.TinkerbellIP,
		time.Now,
	)
	return templateBuilder, nil
}

// GenerateNoProxyList generates NOPROXY list for tinkerbell provider based on HTTP_PROXY, HTTPS_PROXY, NOPROXY and tinkerbellIP.
func generateNoProxyList(clusterSpec *v1alpha1.Cluster, datacenterSpec v1alpha1.TinkerbellDatacenterConfigSpec) []string {
	capacity := len(clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks) +
		len(clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks) +
		len(clusterSpec.Spec.ProxyConfiguration.NoProxy) + 4

	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

	noProxyList = append(noProxyList, clusterapi.NoProxyDefaults()...)

	noProxyList = append(noProxyList,
		clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		datacenterSpec.TinkerbellIP,
	)

	return noProxyList
}
