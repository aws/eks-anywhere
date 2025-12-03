package docker

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"regexp"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	githubTokenEnvVar = "GITHUB_TOKEN"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultCAPIConfigMD string

var eksaDockerResourceType = fmt.Sprintf("dockerdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)

type ProviderClient interface {
	GetDockerLBPort(ctx context.Context, clusterName string) (port string, err error)
}

// Provider implements providers.Provider for the docker cluster-api provider.
type Provider struct {
	docker                ProviderClient
	datacenterConfig      *v1alpha1.DockerDatacenterConfig
	providerKubectlClient ProviderKubectlClient
	templateBuilder       *DockerTemplateBuilder
}

// KubeconfigReader reads the kubeconfig secret from the cluster.
type KubeconfigReader interface {
	GetClusterKubeconfig(ctx context.Context, clusterName, kubeconfigPath string) ([]byte, error)
}

// KubeconfigWriter reads the kubeconfig secret on a docker cluster and copies the contents to a writer.
type KubeconfigWriter struct {
	docker ProviderClient
	reader KubeconfigReader
}

// InstallCustomProviderComponents is a no-op. It implements providers.Provider.
func (p *Provider) InstallCustomProviderComponents(ctx context.Context, kubeconfigFile string) error {
	return nil
}

type ProviderKubectlClient interface {
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetMachineDeployment(ctx context.Context, machineDeploymentName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1.KubeadmControlPlane, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1.EtcdadmCluster, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
}

// NewProvider returns a new Provider.
func NewProvider(providerConfig *v1alpha1.DockerDatacenterConfig, docker ProviderClient, providerKubectlClient ProviderKubectlClient, now types.NowFunc) *Provider {
	return &Provider{
		docker:                docker,
		datacenterConfig:      providerConfig,
		providerKubectlClient: providerKubectlClient,
		templateBuilder: &DockerTemplateBuilder{
			now: now,
		},
	}
}

// BootstrapClusterOpts returns a list of options to be used when creating the bootstrap cluster.
func (p *Provider) BootstrapClusterOpts(_ *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithExtraDockerMounts()}, nil
}

// PreCAPIInstallOnBootstrap is a no-op. It implements providers.Provider.
func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

// PostBootstrapSetup is a no-op. It implements providers.Provider.
func (p *Provider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return nil
}

// PostWorkloadInit is a no-op. It implements providers.Provider.
func (p *Provider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

// Name returns the name of the provider.
func (p *Provider) Name() string {
	return constants.DockerProviderName
}

// DatacenterResourceType returns the resource type for the dockerdatacenterconfigs.
func (p *Provider) DatacenterResourceType() string {
	return eksaDockerResourceType
}

// MachineResourceType returns nothing because docker has no machines. It implements providers.Provider.
func (p *Provider) MachineResourceType() string {
	return ""
}

// PostClusterDeleteValidate is a no-op. It implements providers.Provider.
func (p *Provider) PostClusterDeleteValidate(_ context.Context, _ *types.Cluster) error {
	// No validations
	return nil
}

// PostMoveManagementToBootstrap is a no-op. It implements providers.Provider.
func (p *Provider) PostMoveManagementToBootstrap(_ context.Context, _ *types.Cluster) error {
	// NOOP
	return nil
}

// SetupAndValidateCreateCluster validates the cluster spec and sets up any provider-specific resources.
func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The docker infrastructure provider is meant for local development and testing only")
	if err := ValidateControlPlaneEndpoint(clusterSpec); err != nil {
		return err
	}
	return nil
}

// SetupAndValidateDeleteCluster is a no-op. It implements providers.Provider.
func (p *Provider) SetupAndValidateDeleteCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	return nil
}

// SetupAndValidateUpgradeCluster is a no-op. It implements providers.Provider.
func (p *Provider) SetupAndValidateUpgradeCluster(ctx context.Context, _ *types.Cluster, _ *cluster.Spec, _ *cluster.Spec) error {
	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *Provider) SetupAndValidateUpgradeManagementComponents(_ context.Context, _ *cluster.Spec) error {
	return nil
}

// UpdateSecrets is a no-op. It implements providers.Provider.
func (p *Provider) UpdateSecrets(ctx context.Context, cluster *types.Cluster, _ *cluster.Spec) error {
	// Not implemented
	return nil
}

// NewDockerTemplateBuilder returns a docker template builder object.
func NewDockerTemplateBuilder(now types.NowFunc) *DockerTemplateBuilder {
	return &DockerTemplateBuilder{
		now: now,
	}
}

// DockerTemplateBuilder builds the docker templates.
type DockerTemplateBuilder struct {
	now types.NowFunc
}

// GenerateCAPISpecControlPlane generates a yaml spec with the CAPI objects representing the control plane.
func (d *DockerTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values, err := buildTemplateMapCP(clusterSpec)
	if err != nil {
		return nil, fmt.Errorf("error building template map for CP %v", err)
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

// GenerateCAPISpecWorkers generates a yaml spec with the CAPI objects representing the worker nodes for a particular eks-a cluster.
func (d *DockerTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values, err := buildTemplateMapMD(clusterSpec, workerNodeGroupConfiguration)
		if err != nil {
			return nil, fmt.Errorf("error building template map for MD %v", err)
		}
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			values["upgradeRolloutStrategy"] = true
			values["maxSurge"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
			values["maxUnavailable"] = workerNodeGroupConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable
		}

		bytes, err := templater.Execute(defaultCAPIConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

// CAPIWorkersSpecWithInitialNames generates a yaml spec with the CAPI objects representing the worker
// nodes for a particular eks-a cluster. It uses default initial names (ended in '-1') for the docker
// machine templates and kubeadm config templates.
func (d *DockerTemplateBuilder) CAPIWorkersSpecWithInitialNames(spec *cluster.Spec) (content []byte, err error) {
	machineTemplateNames, kubeadmConfigTemplateNames := initialNamesForWorkers(spec)
	return d.GenerateCAPISpecWorkers(spec, machineTemplateNames, kubeadmConfigTemplateNames)
}

func initialNamesForWorkers(spec *cluster.Spec) (machineTemplateNames, kubeadmConfigTemplateNames map[string]string) {
	workerGroupsLen := len(spec.Cluster.Spec.WorkerNodeGroupConfigurations)
	machineTemplateNames = make(map[string]string, workerGroupsLen)
	kubeadmConfigTemplateNames = make(map[string]string, workerGroupsLen)
	for _, workerNodeGroupConfiguration := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		machineTemplateNames[workerNodeGroupConfiguration.Name] = clusterapi.WorkerMachineTemplateName(spec, workerNodeGroupConfiguration)
		kubeadmConfigTemplateNames[workerNodeGroupConfiguration.Name] = clusterapi.DefaultKubeadmConfigTemplateName(spec, workerNodeGroupConfiguration)
	}

	return machineTemplateNames, kubeadmConfigTemplateNames
}

func kubeletCgroupDriverExtraArgs(kubeVersion v1alpha1.KubernetesVersion) (clusterapi.ExtraArgs, error) {
	clusterKubeVersionSemver, err := v1alpha1.KubeVersionToSemver(kubeVersion)
	if err != nil {
		return nil, fmt.Errorf("converting kubeVersion %v to semver %v", kubeVersion, err)
	}
	kube124Semver, err := v1alpha1.KubeVersionToSemver(v1alpha1.Kube124)
	if err != nil {
		return nil, fmt.Errorf("error converting kubeVersion %v to semver %v", v1alpha1.Kube124, err)
	}
	if clusterKubeVersionSemver.Compare(kube124Semver) != -1 {
		return clusterapi.CgroupDriverSystemdExtraArgs(), nil
	}

	return clusterapi.CgroupDriverCgroupfsExtraArgs(), nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.RootVersionsBundle()
	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()

	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.APIServerExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.APIServerExtraArgs)).
		Append(sharedExtraArgs)
	clusterapi.SetPodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig, apiServerExtraArgs)
	controllerManagerExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))

	values := map[string]interface{}{
		"clusterName":                   clusterSpec.Cluster.Name,
		"control_plane_replicas":        clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":          versionsBundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":             versionsBundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                versionsBundle.KubeDistro.Etcd.Repository,
		"etcdVersion":                   versionsBundle.KubeDistro.Etcd.Tag,
		"corednsRepository":             versionsBundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                versionsBundle.KubeDistro.CoreDNS.Tag,
		"kindNodeImage":                 versionsBundle.EksD.KindNode.VersionedImage(),
		"etcdExtraArgs":                 etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":              crypto.SecureCipherSuitesString(),
		"apiserverExtraArgs":            apiServerExtraArgs.ToPartialYaml(),
		"controllermanagerExtraArgs":    controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":            sharedExtraArgs.ToPartialYaml(),
		"externalEtcdVersion":           versionsBundle.KubeDistro.EtcdVersion,
		"eksaSystemNamespace":           constants.EksaSystemNamespace,
		"podCidrs":                      clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                  clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"haproxyImageRepository":        getHAProxyImageRepo(versionsBundle.Haproxy.Image),
		"haproxyImageTag":               versionsBundle.Haproxy.Image.Tag(),
		"workerNodeGroupConfigurations": clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		"apiServerCertSANs":             clusterSpec.Cluster.Spec.ControlPlaneConfiguration.CertSANs,
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		values["placeholderExternalEtcdEndpoint"] = constants.PlaceholderExternalEtcdEndpoint
		etcdURL, _ := common.GetExternalEtcdReleaseURL(clusterSpec.Cluster.Spec.EksaVersion, versionsBundle)
		if etcdURL != "" {
			values["externalEtcdReleaseUrl"] = etcdURL
		}
	}
	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints

	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}
	values["auditPolicy"] = auditPolicy

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources != nil &&
		*clusterSpec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources {
		admissionExclusionPolicy, err := common.GetAdmissionPluginExclusionPolicy()
		if err != nil {
			return nil, err
		}
		values["admissionExclusionPolicy"] = admissionExclusionPolicy
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values, err := populateRegistryMirrorValues(clusterSpec, values)
		if err != nil {
			return values, err
		}
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		values["upgradeRolloutStrategy"] = true
		values["maxSurge"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration != nil {
		cpKubeletConfig := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration.Object
		if _, ok := cpKubeletConfig["tlsCipherSuites"]; !ok {
			cpKubeletConfig["tlsCipherSuites"] = crypto.SecureCipherSuiteNames()
		}

		if _, ok := cpKubeletConfig["resolvConf"]; !ok {
			if clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf != nil {
				cpKubeletConfig["resolvConf"] = clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf.Path
			}
		}
		kcString, err := yaml.Marshal(cpKubeletConfig)
		if err != nil {
			return nil, fmt.Errorf("marshaling control plane node Kubelet Configuration while building CAPI template %v", err)
		}

		values["kubeletConfiguration"] = string(kcString)

	} else {
		kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
			Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

		cgroupDriverArgs, err := kubeletCgroupDriverExtraArgs(clusterSpec.Cluster.Spec.KubernetesVersion)
		if err != nil {
			return nil, err
		}
		if cgroupDriverArgs != nil {
			kubeletExtraArgs.Append(cgroupDriverArgs)
		}

		values["kubeletExtraArgs"] = kubeletExtraArgs.ToPartialYaml()
	}

	nodeLabelArgs := clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration)
	if len(nodeLabelArgs) != 0 {
		values["nodeLabelArgs"] = nodeLabelArgs.ToPartialYaml()
	}

	return values, nil
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) (map[string]interface{}, error) {
	versionsBundle := clusterSpec.WorkerNodeGroupVersionsBundle(workerNodeGroupConfiguration)

	values := map[string]interface{}{
		"clusterName":           clusterSpec.Cluster.Name,
		"kubernetesVersion":     versionsBundle.KubeDistro.Kubernetes.Tag,
		"kindNodeImage":         versionsBundle.EksD.KindNode.VersionedImage(),
		"eksaSystemNamespace":   constants.EksaSystemNamespace,
		"workerReplicas":        *workerNodeGroupConfiguration.Count,
		"workerNodeGroupName":   fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints": workerNodeGroupConfiguration.Taints,
		"autoscalingConfig":     workerNodeGroupConfiguration.AutoScalingConfiguration,
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		values, err := populateRegistryMirrorValues(clusterSpec, values)
		if err != nil {
			return values, err
		}
	}

	if workerNodeGroupConfiguration.KubeletConfiguration != nil {
		wnKubeletConfig := workerNodeGroupConfiguration.KubeletConfiguration.Object
		if _, ok := wnKubeletConfig["tlsCipherSuites"]; !ok {
			wnKubeletConfig["tlsCipherSuites"] = crypto.SecureCipherSuiteNames()
		}

		if _, ok := wnKubeletConfig["resolvConf"]; !ok {
			if clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf != nil {
				wnKubeletConfig["resolvConf"] = clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf.Path
			}
		}
		kcString, err := yaml.Marshal(wnKubeletConfig)
		if err != nil {
			return nil, fmt.Errorf("marshaling Kubelet Configuration for worker node %s: %v", workerNodeGroupConfiguration.Name, err)
		}

		values["kubeletConfiguration"] = string(kcString)
	} else {
		kubeVersion := clusterSpec.Cluster.Spec.KubernetesVersion
		if workerNodeGroupConfiguration.KubernetesVersion != nil {
			kubeVersion = *workerNodeGroupConfiguration.KubernetesVersion
		}
		kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
			Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

		cgroupDriverArgs, err := kubeletCgroupDriverExtraArgs(kubeVersion)
		if err != nil {
			return nil, err
		}
		if cgroupDriverArgs != nil {
			kubeletExtraArgs.Append(cgroupDriverArgs)
		}

		values["kubeletExtraArgs"] = kubeletExtraArgs.ToPartialYaml()
	}

	nodeLabelArgs := clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)
	if len(nodeLabelArgs) != 0 {
		values["nodeLabelArgs"] = nodeLabelArgs.ToPartialYaml()
	}

	return values, nil
}

// UpdateKubeConfig updates the kubeconfig secret on a docker cluster.
func (p *Provider) UpdateKubeConfig(content *[]byte, clusterName string) error {
	// The Docker provider is for testing only. We don't want to change the interface just for the test
	ctx := context.Background()
	if port, err := p.docker.GetDockerLBPort(ctx, clusterName); err != nil {
		return err
	} else {
		updateKubeconfig(content, port)
		return nil
	}
}

// NewKubeconfigWriter creates a KubeconfigWriter.
func NewKubeconfigWriter(docker ProviderClient, reader KubeconfigReader) KubeconfigWriter {
	return KubeconfigWriter{
		reader: reader,
		docker: docker,
	}
}

// WriteKubeconfig retrieves the contents of the specified cluster's kubeconfig from a secret and copies it to an io.Writer.
func (kr KubeconfigWriter) WriteKubeconfig(ctx context.Context, clusterName, kubeconfigPath string, w io.Writer) error {
	rawkubeconfig, err := kr.reader.GetClusterKubeconfig(ctx, clusterName, kubeconfigPath)
	if err != nil {
		return err
	}

	if err := kr.WriteKubeconfigContent(ctx, clusterName, rawkubeconfig, w); err != nil {
		return err
	}

	return nil
}

// WriteKubeconfigContent retrieves the contents of the specified cluster's kubeconfig from a secret and copies it to an io.Writer.
func (kr KubeconfigWriter) WriteKubeconfigContent(ctx context.Context, clusterName string, content []byte, w io.Writer) error {
	port, err := kr.docker.GetDockerLBPort(ctx, clusterName)
	if err != nil {
		return err
	}

	updateKubeconfig(&content, port)

	if _, err := io.Copy(w, bytes.NewReader(content)); err != nil {
		return err
	}

	return nil
}

// this is required for docker provider.
func updateKubeconfig(content *[]byte, dockerLbPort string) {
	mc := regexp.MustCompile("server:.*")
	updatedConfig := mc.ReplaceAllString(string(*content), fmt.Sprintf("server: https://127.0.0.1:%s", dockerLbPort))
	mc = regexp.MustCompile("certificate-authority-data:.*")
	updatedConfig = mc.ReplaceAllString(updatedConfig, "insecure-skip-tls-verify: true")
	updatedContentByte := []byte(updatedConfig)
	*content = updatedContentByte
}

// Version returns the version of the provider.
func (p *Provider) Version(components *cluster.ManagementComponents) string {
	return components.Docker.Version
}

// EnvMap returns a map of environment variables to be set when running the docker clusterctl command.
func (p *Provider) EnvMap(_ *cluster.ManagementComponents, _ *cluster.Spec) (map[string]string, error) {
	envMap := make(map[string]string)
	if env, ok := os.LookupEnv(githubTokenEnvVar); ok && len(env) > 0 {
		envMap[githubTokenEnvVar] = env
	}
	return envMap, nil
}

// GetDeployments returns a map of namespaces to deployments that should be running for the provider.
func (p *Provider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capd-system": {"capd-controller-manager"},
	}
}

// GetInfrastructureBundle returns the infrastructure bundle for the provider.
func (p *Provider) GetInfrastructureBundle(components *cluster.ManagementComponents) *types.InfrastructureBundle {
	folderName := fmt.Sprintf("infrastructure-docker/%s/", components.Docker.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			components.Docker.Components,
			components.Docker.Metadata,
			components.Docker.ClusterTemplate,
		},
	}

	return &infraBundle
}

// DatacenterConfig returns the datacenter config for the provider.
func (p *Provider) DatacenterConfig(_ *cluster.Spec) providers.DatacenterConfig {
	return p.datacenterConfig
}

// MachineConfigs is a no-op. It implements providers.Provider.
func (p *Provider) MachineConfigs(_ *cluster.Spec) []providers.MachineConfig {
	return nil
}

// ValidateNewSpec is a no-op. It implements providers.Provider.
func (p *Provider) ValidateNewSpec(_ context.Context, _ *types.Cluster, _ *cluster.Spec) error {
	return nil
}

// ChangeDiff returns the component change diff for the provider.
func (p *Provider) ChangeDiff(currentComponents, newComponents *cluster.ManagementComponents) *types.ComponentChangeDiff {
	if currentComponents.Docker.Version == newComponents.Docker.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.DockerProviderName,
		NewVersion:    newComponents.Docker.Version,
		OldVersion:    currentComponents.Docker.Version,
	}
}

// RunPostControlPlaneUpgrade is a no-op. It implements providers.Provider.
func (p *Provider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	return nil
}

// RunPostControlPlaneCreation is a no-op. It implements providers.Provider.
func (p *Provider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	return nil
}

func getHAProxyImageRepo(haProxyImage releasev1alpha1.Image) string {
	var haproxyImageRepo string

	regexStr := `(?P<HAProxyImageRepoPrefix>public.ecr.aws/[a-z0-9._-]+/kubernetes-sigs/kind)/haproxy`
	regex := regexp.MustCompile(regexStr)
	matches := regex.FindStringSubmatch(haProxyImage.Image())
	if len(matches) > 0 {
		haproxyImageRepo = matches[regex.SubexpIndex("HAProxyImageRepoPrefix")]
	}

	return haproxyImageRepo
}

// PreCoreComponentsUpgrade staisfies the Provider interface.
func (p *Provider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	return nil
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
