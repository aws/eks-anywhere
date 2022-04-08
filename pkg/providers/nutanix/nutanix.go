package nutanix

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaNutanixDatacenterResourceType = fmt.Sprintf("nutanixdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaNutanixMachineResourceType    = fmt.Sprintf("nutanixmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	requiredEnvs                      = []string{}
)

type nutanixProvider struct {
	clusterConfig    *v1alpha1.Cluster
	datacenterConfig *v1alpha1.NutanixDatacenterConfig
	// providerKubectlClient ProviderKubectlClient
	// templateBuilder NutanixTemplateBuilder
}

// type ProviderKubectlClient interface {
// 	// TODO: Add necessary kubectl functions here
// }

func NewProvider(hardwareConfig string) *nutanixProvider {
	return &nutanixProvider{}
}

func (p *nutanixProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	// Adding proxy environment vars to the bootstrap cluster
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.NutanixIP)
		for _, s := range p.clusterConfig.Spec.ProxyConfiguration.NoProxy {
			if s != "" {
				noProxy += "," + s
			}
		}
		env["HTTP_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = noProxy
	}
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithEnv(env)}, nil
}

func (p *nutanixProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	return nil
}

func (p *nutanixProvider) Name() string {
	return constants.NutanixProviderName
}

func (p *nutanixProvider) DatacenterResourceType() string {
	return eksaNutanixDatacenterResourceType
}

func (p *nutanixProvider) MachineResourceType() string {
	return eksaNutanixMachineResourceType
}

func (p *nutanixProvider) DeleteResources(_ context.Context, _ *cluster.Spec) error {
	// TODO: Add delete resource logic
	return nil
}

func (p *nutanixProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The nutanix infrastructure provider is still in development and should not be used in production")
	// TODO: Add more validations
	return nil
}

func (p *nutanixProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	// TODO: validations?
	return nil
}

func (p *nutanixProvider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Add validations when this is supported
	return errors.New("upgrade for nutanix provider isn't currently supported")
}

func (p *nutanixProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	// TODO: implement
	return nil
}

type NutanixTemplateBuilder struct {
	now types.NowFunc
}

func NewNutanixTemplateBuilder(datacenterSpec *v1alpha1.NutanixDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec, workerNodeGroupMachineSpecs map[string]v1alpha1.NutanixMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &NutanixTemplateBuilder{
		now: now,
	}
}

func (vs *NutanixTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (vs *NutanixTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *NutanixTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *NutanixTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var values map[string]interface{}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (vs *NutanixTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, _ = range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		var values map[string]interface{}
		bytes, err := templater.Execute(defaultClusterConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}
	return templater.AppendYamlResources(workerSpecs...), nil
}

func (p *nutanixProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *nutanixProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *nutanixProvider) GenerateStorageClass() []byte {
	// TODO: determine if we need something else here
	return nil
}

func (p *nutanixProvider) GenerateMHC() ([]byte, error) {
	data := map[string]string{
		"clusterName":         p.clusterConfig.Name,
		"eksaSystemNamespace": constants.EksaSystemNamespace,
	}
	mhc, err := templater.Execute(string(mhcTemplate), data)
	if err != nil {
		return nil, err
	}
	return mhc, nil
}

func (p *nutanixProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) Version(clusterSpec *cluster.Spec) string {
	// TODO: Add Nutanix to the bundle and add it's versions
	// return clusterSpec.VersionsBundle.Nutanix.Version
	return "nutanix-dev"
}

func (p *nutanixProvider) EnvMap() (map[string]string, error) {
	// TODO: determine if any env vars are needed and add them to requiredEnvs
	envMap := make(map[string]string)
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return envMap, fmt.Errorf("warning required env not set %s", key)
		}
	}
	return envMap, nil
}

func (p *nutanixProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
}

func (p *nutanixProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	// TODO: uncomment below when nutanix is added to bundle
	// bundle := clusterSpec.VersionsBundle
	// folderName := fmt.Sprintf("infrastructure-nutanix/%s/", bundle.Nutanix.Version)

	// infraBundle := types.InfrastructureBundle{
	// 	FolderName: folderName,
	// 	Manifests: []releasev1alpha1.Manifest{
	// 		bundle.Nutanix.Components,
	// 		bundle.Nutanix.Metadata,
	// 		bundle.Nutanix.ClusterTemplate,
	// 	},
	// }
	// return &infraBundle
	return nil
}

func (p *nutanixProvider) DatacenterConfig() providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *nutanixProvider) MachineConfigs() []providers.MachineConfig {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	// TODO: implement
	return nil
}

func (p *nutanixProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *nutanixProvider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func (p *nutanixProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec v1alpha1.NutanixMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                  clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointIp":       clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys,
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"kubernetesVersion":            bundle.KubeDistro.Kubernetes.Tag,
		"kubeVipImage":                 "ghcr.io/kube-vip/kube-vip:latest",
		"podCidrs":                     clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
	}
	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.NutanixMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.ObjectMeta.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"workerReplicas":         clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count,
		"workerPoolName":         "md-0",
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys,
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
	}
	return values
}
