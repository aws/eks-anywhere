package tinkerbell

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
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	tinkerbellCertURLKey  = "TINKERBELL_CERT_URL"
	tinkerbellGRPCAuthKey = "TINKERBELL_GRPC_AUTHORITY"
	tinkerbellIPKey       = "TINKERBELL_IP"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaTinkerbellDatacenterResourceType = fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaTinkerbellMachineResourceType    = fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	requiredEnvs                         = []string{tinkerbellCertURLKey, tinkerbellGRPCAuthKey, tinkerbellIPKey}
)

type tinkerbellProvider struct {
	clusterConfig    *v1alpha1.Cluster
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig
	// TODO: Update hardwareConfig to proper type
	hardwareConfig string
	// providerKubectlClient ProviderKubectlClient
	// templateBuilder TinkerbellTemplateBuilder
}

// type ProviderKubectlClient interface {
// 	// TODO: Add necessary kubectl functions here
// }

func NewProvider(hardwareConfig string) *tinkerbellProvider {
	return &tinkerbellProvider{
		hardwareConfig: hardwareConfig,
	}
}

func (p *tinkerbellProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	// Adding proxy environment vars to the bootstrap cluster
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.TinkerbellIP)
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

func (p *tinkerbellProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	return nil
}

func (p *tinkerbellProvider) Name() string {
	return constants.TinkerbellProviderName
}

func (p *tinkerbellProvider) DatacenterResourceType() string {
	return eksaTinkerbellDatacenterResourceType
}

func (p *tinkerbellProvider) MachineResourceType() string {
	return eksaTinkerbellMachineResourceType
}

func (p *tinkerbellProvider) DeleteResources(_ context.Context, _ *cluster.Spec) error {
	// TODO: Add delete resource logic
	return nil
}

func (p *tinkerbellProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The tinkerbell infrastructure provider is still in development and should not be used in production")
	// TODO: Add more validations
	return nil
}

func (p *tinkerbellProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	// TODO: validations?
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func (p *tinkerbellProvider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Add validations when this is supported
	return errors.New("upgrade for tinkerbell provider isn't currently supported")
}

func (p *tinkerbellProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	// TODO: implement
	return nil
}

type TinkerbellTemplateBuilder struct {
	controlPlaneMachineSpec    *v1alpha1.TinkerbellMachineConfigSpec
	workerNodeGroupMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	etcdMachineSpec            *v1alpha1.TinkerbellMachineConfigSpec
	now                        types.NowFunc
}

func NewTinkerbellTemplateBuilder(controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &TinkerbellTemplateBuilder{
		controlPlaneMachineSpec:    controlPlaneMachineSpec,
		workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
		etcdMachineSpec:            etcdMachineSpec,
		now:                        now,
	}
}

func (vs *TinkerbellTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapCP(clusterSpec, *vs.controlPlaneMachineSpec)
	for _, buildOption := range buildOptions {
		buildOption(values)
	}
	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (vs *TinkerbellTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapMD(clusterSpec, *vs.workerNodeGroupMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultClusterConfigMD, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (p *tinkerbellProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *tinkerbellProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	// TODO: implement
	return nil, nil, nil
}

func (p *tinkerbellProvider) GenerateStorageClass() []byte {
	// TODO: determine if we need something else here
	return nil
}

func (p *tinkerbellProvider) GenerateMHC() ([]byte, error) {
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

func (p *tinkerbellProvider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) Version(clusterSpec *cluster.Spec) string {
	// TODO: Add Tinkerbell to the bundle and add it's versions
	// return clusterSpec.VersionsBundle.Tinkerbell.Version
	return "tinkerbell-dev"
}

func (p *tinkerbellProvider) EnvMap() (map[string]string, error) {
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

func (p *tinkerbellProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
}

func (p *tinkerbellProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	// TODO: uncomment below when tinkerbell is added to bundle
	// bundle := clusterSpec.VersionsBundle
	// folderName := fmt.Sprintf("infrastructure-tinkerbell/%s/", bundle.Tinkerbell.Version)

	// infraBundle := types.InfrastructureBundle{
	// 	FolderName: folderName,
	// 	Manifests: []releasev1alpha1.Manifest{
	// 		bundle.Tinkerbell.Components,
	// 		bundle.Tinkerbell.Metadata,
	// 		bundle.Tinkerbell.ClusterTemplate,
	// 	},
	// }
	// return &infraBundle
	// TODO - remove below code when tinkerbell is added to bundle
	folderName := fmt.Sprintf("infrastructure-tinkerbell/%s/", "v0.1.0")
	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			{
				URI: "https://github.com/tinkerbell/cluster-api-provider-tinkerbell/releases/download/v0.1.0/infrastructure-components.yaml",
			},
			{
				URI: "https://github.com/tinkerbell/cluster-api-provider-tinkerbell/releases/download/v0.1.0/metadata.yaml",
			},
			{
				URI: "https://github.com/tinkerbell/cluster-api-provider-tinkerbell/releases/download/v0.1.0/cluster-template.yaml",
			},
		},
	}
	return &infraBundle
}

func (p *tinkerbellProvider) DatacenterConfig() providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *tinkerbellProvider) MachineConfigs() []providers.MachineConfig {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	// TODO: implement
	return nil
}

func (p *tinkerbellProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func (p *tinkerbellProvider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func (p *tinkerbellProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// TODO: Figure out if something is needed here
	return nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec v1alpha1.TinkerbellMachineConfigSpec) map[string]interface{} {
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

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.TinkerbellMachineConfigSpec) map[string]interface{} {
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

func (p *tinkerbellProvider) validateEnv(ctx context.Context) error {
	if tinkerbellCertURL, ok := os.LookupEnv(tinkerbellCertURLKey); ok && len(tinkerbellCertURL) > 0 {
		if err := os.Setenv(tinkerbellCertURLKey, tinkerbellCertURL); err != nil {
			return fmt.Errorf("unable to set %s: %v", tinkerbellCertURLKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", tinkerbellCertURLKey)
	}
	if tinkerbellGRPCAuth, ok := os.LookupEnv(tinkerbellGRPCAuthKey); ok && len(tinkerbellGRPCAuth) > 0 {
		if err := os.Setenv(tinkerbellGRPCAuthKey, tinkerbellGRPCAuth); err != nil {
			return fmt.Errorf("unable to set %s: %v", tinkerbellGRPCAuthKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", tinkerbellGRPCAuthKey)
	}
	if tinkerbellIP, ok := os.LookupEnv(tinkerbellIPKey); ok && len(tinkerbellIP) > 0 {
		if err := os.Setenv(tinkerbellIPKey, tinkerbellIP); err != nil {
			return fmt.Errorf("unable to set %s: %v", tinkerbellIPKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", tinkerbellIPKey)
	}
	return nil
}
