package cloudstack

import (
	"bytes"
	"context"
	_ "embed"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
	"net"
	"net/url"
	"os"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/eks-anywhere/pkg/crypto"

	etcdv1beta1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/internal/templates"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	eksaLicense                           = "EKSA_LICENSE"
	cloudStackPasswordKey                 = "CLOUDSTACK_PASSWORD"
	eksacloudStackUsernameKey             = "EKSA_CLOUDSTACK_USERNAME"
	eksacloudStackPasswordKey             = "EKSA_CLOUDSTACK_PASSWORD"
	eksacloudStackCloudConfigB64SecretKey = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"
	cloudStackCloudConfigB64SecretKey     = "CLOUDSTACK_B64ENCODED_SECRET"
	cloudMonkeyInsecure                   = "CLOUDMONKEY_INSECURE"
	expClusterResourceSetKey              = "EXP_CLUSTER_RESOURCE_SET"
	privateKeyFileName                    = "eks-a-id_rsa"
	publicKeyFileName                     = "eks-a-id_rsa.pub"

	controlEndpointDefaultPort = "6443"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

//go:embed config/secret.yaml
var defaultSecretObject string

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var requiredEnvs = []string{cloudStackCloudConfigB64SecretKey}

var (
	eksaCloudStackDeploymentResourceType = fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	noProxyDefaults                      = []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
)

type cloudstackProvider struct {
	datacenterConfig            *v1alpha1.CloudStackDatacenterConfig
	machineConfigs              map[string]*v1alpha1.CloudStackMachineConfig
	clusterConfig               *v1alpha1.Cluster
	providerCmkClient           ProviderCmkClient
	providerKubectlClient       ProviderKubectlClient
	writer                      filewriter.FileWriter
	selfSigned                  bool
	controlPlaneSshAuthKey      string
	workerSshAuthKey            string
	etcdSshAuthKey              string
	netClient                   networkutils.NetClient
	controlPlaneTemplateFactory *templates.Factory
	workerTemplateFactory       *templates.Factory
	etcdTemplateFactory         *templates.Factory
	templateBuilder             *CloudStackTemplateBuilder
	skipIpCheck                 bool
	validator              		*Validator
}

func (p *cloudstackProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return fmt.Errorf("cloudstack provider does not support this functionality currently")
}

func (p *cloudstackProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	panic("implement me")
}

func (p *cloudstackProvider) MachineDeploymentsToDelete(workloadCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) []string {
	panic("implement me")
}

func (p *cloudstackProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// Nothing to do
	return nil
}

func (p *cloudstackProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// Nothing to do
	return nil
}

type ProviderCmkClient interface {
	ValidateCloudStackConnection(ctx context.Context) error
	ValidateServiceOfferingPresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, serviceOffering v1alpha1.CloudStackResourceRef) error
	ValidateTemplatePresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, template v1alpha1.CloudStackResourceRef) error
	ValidateAffinityGroupsPresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, affinityGroupIds []string) error
	ValidateZonePresent(ctx context.Context, zone v1alpha1.CloudStackResourceRef) error
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmv1beta1.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1beta1.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	SearchCloudStackMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackMachineConfig, error)
	SearchCloudStackDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackDatacenterConfig, error)
	DeleteEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDeploymentConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error
}

type ClusterResourceSetManager interface {
	ForceUpdate(ctx context.Context, name, namespace string, managementCluster, workloadCluster *types.Cluster) error
}

func NewProvider(deploymentConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCmkClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool) *cloudstackProvider {
	return NewProviderCustomNet(
		deploymentConfig,
		machineConfigs,
		clusterConfig,
		providerCloudMonkeyClient,
		providerKubectlClient,
		writer,
		&networkutils.DefaultNetClient{},
		now,
		skipIpCheck,
	)
}

func NewProviderCustomNet(deploymentConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCmkClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, netClient networkutils.NetClient, now types.NowFunc, skipIpCheck bool) *cloudstackProvider {
	var controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	var controlPlaneTemplateFactory, workerNodeGroupTemplateFactory, etcdTemplateFactory *templates.Factory
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
		controlPlaneTemplateFactory = templates.NewFactory(
			providerCloudMonkeyClient,
			deploymentConfig.Spec.Network,
			deploymentConfig.Spec.Domain,
			deploymentConfig.Spec.Zone,
			deploymentConfig.Spec.Account,
		)
	}
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) > 0 && clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name] != nil {
		workerNodeGroupMachineSpec = &machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec
		workerNodeGroupTemplateFactory = templates.NewFactory(
			providerCloudMonkeyClient,
			deploymentConfig.Spec.Network,
			deploymentConfig.Spec.Domain,
			deploymentConfig.Spec.Zone,
			deploymentConfig.Spec.Account,
		)
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
			etcdTemplateFactory = templates.NewFactory(
				providerCloudMonkeyClient,
				deploymentConfig.Spec.Network,
				deploymentConfig.Spec.Domain,
				deploymentConfig.Spec.Zone,
				deploymentConfig.Spec.Account,
			)
		}
	}
	return &cloudstackProvider{
		datacenterConfig:            deploymentConfig,
		machineConfigs:              machineConfigs,
		clusterConfig:               clusterConfig,
		providerCmkClient:           providerCloudMonkeyClient,
		providerKubectlClient:       providerKubectlClient,
		writer:                      writer,
		selfSigned:                  false,
		netClient:                   netClient,
		controlPlaneTemplateFactory: controlPlaneTemplateFactory,
		workerTemplateFactory:       workerNodeGroupTemplateFactory,
		etcdTemplateFactory:         etcdTemplateFactory,
		templateBuilder: &CloudStackTemplateBuilder{
			deploymentConfigSpec:       &deploymentConfig.Spec,
			controlPlaneMachineSpec:    controlPlaneMachineSpec,
			workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
			etcdMachineSpec:            etcdMachineSpec,
			now:                        now,
		},
		skipIpCheck:        skipIpCheck,
		validator: NewValidator(providerCloudMonkeyClient, machineConfigs, netClient),
	}
}

func (p *cloudstackProvider) UpdateKubeConfig(_ *[]byte, _ string) error {
	// customize generated kube config
	return nil
}

func (p *cloudstackProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	execConfig, err := parseCloudStackSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, execConfig.CloudStackManagementUrl)
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

func (p *cloudstackProvider) Name() string {
	return constants.CloudStackProviderName
}

func (p *cloudstackProvider) DatacenterResourceType() string {
	return eksaCloudStackDeploymentResourceType
}

func (p *cloudstackProvider) MachineResourceType() string {
	return eksaCloudStackMachineResourceType
}

func (p *cloudstackProvider) setupSSHAuthKeysForCreate() error {
	var useKeyGeneratedForControlplane, useKeyGeneratedForWorker bool
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.controlPlaneSshAuthKey); err != nil {
		return err
	}
	if len(p.controlPlaneSshAuthKey) <= 0 {
		generatedKey, err := p.generateSSHAuthKey(controlPlaneUser.Name)
		if err != nil {
			return err
		}
		p.controlPlaneSshAuthKey = generatedKey
		useKeyGeneratedForControlplane = true
	}
	workerUser := p.machineConfigs[p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec.Users[0]
	p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.workerSshAuthKey); err != nil {
		return err
	}
	if len(p.workerSshAuthKey) <= 0 {
		if useKeyGeneratedForControlplane { // use the same key
			p.workerSshAuthKey = p.controlPlaneSshAuthKey
		} else {
			generatedKey, err := p.generateSSHAuthKey(workerUser.Name)
			if err != nil {
				return err
			}
			p.workerSshAuthKey = generatedKey
			useKeyGeneratedForWorker = true
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.etcdSshAuthKey); err != nil {
			return err
		}
		if len(p.etcdSshAuthKey) <= 0 {
			if useKeyGeneratedForControlplane { // use the same key as for controlplane
				p.etcdSshAuthKey = p.controlPlaneSshAuthKey
			} else if useKeyGeneratedForWorker {
				p.etcdSshAuthKey = p.workerSshAuthKey // if cp key was provided by user, check if worker key was generated by cli and use that
			} else {
				generatedKey, err := p.generateSSHAuthKey(etcdUser.Name)
				if err != nil {
					return err
				}
				p.etcdSshAuthKey = generatedKey
			}
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	return nil
}

func (p *cloudstackProvider) setupSSHAuthKeysForUpgrade() error {
	return fmt.Errorf("upgrade is not yet supported by cloudstack provider")
}

func (p *cloudstackProvider) parseSSHAuthKey(key *string) error {
	if len(*key) > 0 {
		// When public key is entered by user in provider config, it may contain email address (or any other comment) at the end. ssh-keygen allows users to add comments as suffixes to public key in
		// public key file. When CLI generates the key pair, no comments will be present. So we get rid of the comment from the public key to ensure unit tests that do string compare on the sshAuthorizedKey
		// will pass
		parts := strings.Fields(strings.TrimSpace(*key))
		if len(parts) >= 3 {
			*key = parts[0] + " " + parts[1]
		}
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*key))
		if err != nil {
			return fmt.Errorf("provided CloudStackMachineConfig sshAuthorizedKey is invalid: %v", err)
		}
	}
	return nil
}

func (p *cloudstackProvider) generateSSHAuthKey(username string) (string, error) {
	logger.Info("Provided CloudStackMachineConfig sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
	keygenerator, _ := crypto.NewKeyGenerator(p.writer)
	sshAuthorizedKeyBytes, err := keygenerator.GenerateSSHKeyPair("", "", privateKeyFileName, publicKeyFileName, username)
	if err != nil || sshAuthorizedKeyBytes == nil {
		return "", fmt.Errorf("VSphereMachineConfig error generating sshAuthorizedKey: %v", err)
	}
	key := string(sshAuthorizedKeyBytes)
	key = strings.TrimRight(key, "\n")
	return key, nil
}

func (p *cloudstackProvider) validateManagementApiEndpoint(rawurl string) error {
	_, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return fmt.Errorf("CloudStack managementApiEndpoint is invalid: #{err}")
	}
	return nil
}

func getHostnameFromUrl(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("#{rawurl} is not a valid url")
	}

	return url.Hostname(), nil
}

// The logic here is redundant with the implementation in factory.go
func parseCloudStackSecret() (*v1alpha1.CloudStackExecConfig, error) {
	cloudStackB64EncodedSecret, ok := os.LookupEnv(eksacloudStackCloudConfigB64SecretKey)
	if !ok {
		return nil, fmt.Errorf("%s is not set or is empty", eksacloudStackCloudConfigB64SecretKey)
	}
	decodedString, err := b64.StdEncoding.DecodeString(cloudStackB64EncodedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value for %s with base64: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	cfg, err := ini.Load(decodedString)
	if err != nil {
		return nil, fmt.Errorf("failed to extract values from %s with ini: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	section, err := cfg.GetSection("Global")
	if err != nil {
		return nil, fmt.Errorf("failed to extract section 'Global' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	apiKey, err := section.GetKey("api-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-key' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	secretKey, err := section.GetKey("secret-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'secret-key' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	apiUrl, err := section.GetKey("api-url")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-url' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	return &v1alpha1.CloudStackExecConfig{
		CloudStackApiKey:        apiKey.Value(),
		CloudStackSecretKey:     secretKey.Value(),
		CloudStackManagementUrl: apiUrl.Value(),
	}, nil
}

func (p *cloudstackProvider) validateEnv(ctx context.Context) error {
	var cloudStackB64EncodedSecret string
	var ok bool

	if cloudStackB64EncodedSecret, ok = os.LookupEnv(eksacloudStackCloudConfigB64SecretKey); ok && len(cloudStackB64EncodedSecret) > 0 {
		if err := os.Setenv(cloudStackCloudConfigB64SecretKey, cloudStackB64EncodedSecret); err != nil {
			return fmt.Errorf("unable to set %s: %v", cloudStackCloudConfigB64SecretKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", eksacloudStackCloudConfigB64SecretKey)
	}
	execConfig, err := parseCloudStackSecret()
	if err != nil {
		return fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	if len(execConfig.CloudStackManagementUrl) <= 0 {
		return errors.New("CloudStackDatacenterConfig managementApiEndpoint is not set or is empty")
	}
	if err := p.validateManagementApiEndpoint(execConfig.CloudStackManagementUrl); err != nil {
		return errors.New("CloudStackDatacenterConfig managementApiEndpoint is invalid")
	}
	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}
	return nil
}

func (p *cloudstackProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	cloudStackClusterSpec := NewSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)
	if p.datacenterConfig.Spec.Insecure {
		logger.Info("Warning: CloudStackDatacenterConfig configured in insecure mode")
	}
	if err := os.Setenv(cloudMonkeyInsecure, strconv.FormatBool(p.datacenterConfig.Spec.Insecure)); err != nil {
		return fmt.Errorf("unable to set %s: %v", cloudMonkeyInsecure, err)
	}

	if err := p.validator.validateCloudStackAccess(ctx); err != nil {
		return err
	}

	if err := p.validator.ValidateCloudStackDatacenterConfig(ctx, cloudStackClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, cloudStackClusterSpec); err != nil {
		return err
	}

	if err := p.setupSSHAuthKeysForCreate(); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	if p.skipIpCheck {
		logger.Info("Skipping check for whether control plane ip is in use")
		return nil
	}

	if err := p.validator.validateControlPlaneIpUniqueness(cloudStackClusterSpec); err != nil {
		return err
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return fmt.Errorf("upgrade is not yet supported for CloudStack cluster")
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

type CloudStackTemplateBuilder struct {
	deploymentConfigSpec       *v1alpha1.CloudStackDatacenterConfigSpec
	controlPlaneMachineSpec    *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	etcdMachineSpec            *v1alpha1.CloudStackMachineConfigSpec
	now                        types.NowFunc
}

func (vs *CloudStackTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (vs *CloudStackTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *CloudStackTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *CloudStackTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *vs.etcdMachineSpec
	}
	execConfig, err := parseCloudStackSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	values := buildTemplateMapCP(clusterSpec, *vs.deploymentConfigSpec, *vs.controlPlaneMachineSpec, etcdMachineSpec, execConfig.CloudStackManagementUrl)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (vs *CloudStackTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	execConfig, err := parseCloudStackSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variable exec config: %v", err)
	}
	values := buildTemplateMapMD(clusterSpec, *vs.deploymentConfigSpec, *vs.workerNodeGroupMachineSpec, execConfig.CloudStackManagementUrl)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultClusterConfigMD, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, deploymentConfigSpec v1alpha1.CloudStackDatacenterConfigSpec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec, managementApiEndpoint string) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"
	host, port, _ := net.SplitHostPort(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)

	values := map[string]interface{}{
		"clusterName":                            clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointHost":               host,
		"controlPlaneEndpointPort":               port,
		"controlPlaneReplicas":                   clusterSpec.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":                   bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":                      bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":                         bundle.KubeDistro.Etcd.Repository,
		"etcdImageTag":                           bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":                      bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":                         bundle.KubeDistro.CoreDNS.Tag,
		"nodeDriverRegistrarImage":               bundle.KubeDistro.NodeDriverRegistrar.VersionedImage(),
		"livenessProbeImage":                     bundle.KubeDistro.LivenessProbe.VersionedImage(),
		"externalAttacherImage":                  bundle.KubeDistro.ExternalAttacher.VersionedImage(),
		"externalProvisionerImage":               bundle.KubeDistro.ExternalProvisioner.VersionedImage(),
		"cloudstackManagementApiEndpoint":        managementApiEndpoint,
		"managerImage":                           bundle.CloudStack.Manager.VersionedImage(),
		"insecure":                               deploymentConfigSpec.Insecure,
		"cloudstackNetwork":                      deploymentConfigSpec.Network.Value,
		"cloudstackDomain":                       deploymentConfigSpec.Domain,
		"cloudstackZone":                         deploymentConfigSpec.Zone.Value,
		"cloudstackAccount":                      deploymentConfigSpec.Account,
		"cloudstackControlPlaneComputeOffering":  controlPlaneMachineSpec.ComputeOffering.Value,
		"cloudstackControlPlaneTemplateOffering": controlPlaneMachineSpec.Template.Value,
		"affinityGroupIds":                       controlPlaneMachineSpec.AffinityGroupIds,
		"cloudstackEtcdComputeOffering":          etcdMachineSpec.ComputeOffering.Value,
		"cloudstackEtcdTemplateOffering":         etcdMachineSpec.Template.Value,
		"cloudstackEtcdAffinityGroupIds":         etcdMachineSpec.AffinityGroupIds,
		"controlPlaneSshUsername":                controlPlaneMachineSpec.Users[0].Name,
		"podCidrs":                               clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                           clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                     clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).ToPartialYaml(),
		"format":                                 format,
		"externalEtcdVersion":                    bundle.KubeDistro.EtcdVersion,
		"etcdImage":                              bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                    constants.EksaSystemNamespace,
		"auditPolicy":                            common.GetAuditPolicy(),
	}

	if clusterSpec.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, noProxyDefaults...)
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(managementApiEndpoint)
		if err == nil {
			noProxyList = append(noProxyList, cloudStackManagementApiEndpointHostname)
		}
		noProxyList = append(noProxyList,
			clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Spec.ExternalEtcdConfiguration.Count
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, deploymentConfigSpec v1alpha1.CloudStackDatacenterConfigSpec, workerNodeGroupMachineSpec v1alpha1.CloudStackMachineConfigSpec, managementApiEndpoint string) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                clusterSpec.ObjectMeta.Name,
		"kubernetesVersion":          bundle.KubeDistro.Kubernetes.Tag,
		"cloudstackNetwork":          deploymentConfigSpec.Network.Value,
		"cloudstackTemplate":         workerNodeGroupMachineSpec.Template.Value,
		"cloudstackOffering":         workerNodeGroupMachineSpec.ComputeOffering.Value,
		"cloudstackAffinityGroupIds": workerNodeGroupMachineSpec.AffinityGroupIds,
		"workerReplicas":             clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count,
		"workerSshUsername":          workerNodeGroupMachineSpec.Users[0].Name,
		"format":                     format,
		"eksaSystemNamespace":        constants.EksaSystemNamespace,
	}

	if clusterSpec.Spec.RegistryMirrorConfiguration != nil {
		values["registryMirrorConfiguration"] = clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint
		if len(clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent) > 0 {
			values["registryCACert"] = clusterSpec.Spec.RegistryMirrorConfiguration.CACertContent
		}
	}

	if clusterSpec.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		capacity := len(clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks) +
			len(clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks) +
			len(clusterSpec.Spec.ProxyConfiguration.NoProxy) + 4
		noProxyList := make([]string, 0, capacity)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList, noProxyDefaults...)
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(managementApiEndpoint)
		if err == nil {
			noProxyList = append(noProxyList, cloudStackManagementApiEndpointHostname)
		}
		noProxyList = append(noProxyList,
			clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	return values
}

func (p *cloudstackProvider) generateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, newClusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return nil, nil, fmt.Errorf("cloudstack provider does not support upgrade yet")
}

func (p *cloudstackProvider) generateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	clusterName := clusterSpec.ObjectMeta.Name

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = p.templateBuilder.CPMachineTemplateName(clusterName)
		values["cloudstackControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["cloudstackEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["etcdTemplateName"] = p.templateBuilder.EtcdMachineTemplateName(clusterName)
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}
	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = p.templateBuilder.WorkerMachineTemplateName(clusterName)
		values["cloudstackWorkerSshAuthorizedKey"] = p.workerSshAuthKey
	}
	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workersOpt)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) GenerateCAPISpecForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, currentSpec, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	return nil, nil, fmt.Errorf("cloudstack provider does not support upgrade yet")
}

func (p *cloudstackProvider) GenerateCAPISpecForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) (controlPlaneSpec, workersSpec []byte, err error) {
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForCreate(ctx, cluster, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
}

func (p *cloudstackProvider) GenerateMHC() ([]byte, error) {
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

func (p *cloudstackProvider) CleanupProviderInfrastructure(_ context.Context) error {
	return nil
}

func (p *cloudstackProvider) createSecret(cluster *types.Cluster, contents *bytes.Buffer) error {
	t, err := template.New("tmpl").Parse(defaultSecretObject)
	if err != nil {
		return fmt.Errorf("error creating secret object template: %v", err)
	}

	values := map[string]string{
		"clusterName":       cluster.Name,
		"insecure":          strconv.FormatBool(p.datacenterConfig.Spec.Insecure),
		"cloudstackNetwork": p.datacenterConfig.Spec.Network.Value,
		"eksaLicense":       os.Getenv(eksaLicense),
	}
	err = t.Execute(contents, values)
	if err != nil {
		return fmt.Errorf("error substituting values for secret object template: %v", err)
	}
	return nil
}

func (p *cloudstackProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// Nothing to do
	return nil
}

func (p *cloudstackProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.CloudStack.Version
}

func (p *cloudstackProvider) EnvMap() (map[string]string, error) {
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

func (p *cloudstackProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capc-system": {"capc-controller-manager"},
	}
}

func (p *cloudstackProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-cloudstack/%s/", bundle.CloudStack.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.CloudStack.Components,
			bundle.CloudStack.Metadata,
		},
	}
	return &infraBundle
}

func (p *cloudstackProvider) DatacenterConfig() providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *cloudstackProvider) MachineConfigs() []providers.MachineConfig {
	var configs []providers.MachineConfig
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerMachineName := p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	if p.clusterConfig.IsManaged() {
		p.machineConfigs[controlPlaneMachineName].SetManagement(p.clusterConfig.ManagedBy())
	}

	configs = append(configs, p.machineConfigs[controlPlaneMachineName])
	if workerMachineName != controlPlaneMachineName {
		configs = append(configs, p.machineConfigs[workerMachineName])
		if p.clusterConfig.IsManaged() {
			p.machineConfigs[workerMachineName].SetManagement(p.clusterConfig.ManagedBy())
		}
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName && etcdMachineName != workerMachineName {
			configs = append(configs, p.machineConfigs[etcdMachineName])
			p.machineConfigs[etcdMachineName].SetManagement(p.clusterConfig.ManagedBy())
		}
	}
	return configs
}

func (p *cloudstackProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.CloudStackMachineConfig, clusterSpec *cluster.Spec) error {
	// allow template, compute offering, details, users to mutate
	return nil
}

func (p *cloudstackProvider) RunPostUpgrade(ctx context.Context, clusterSpec *cluster.Spec, managementCluster, workloadCluster *types.Cluster) error {
	return fmt.Errorf("upgrade is not supported for CloudStack provider yet")
}

func (p *cloudstackProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec) (bool, error) {
	return false, fmt.Errorf("upgrade is not supported for CloudStack provider yet")
}

func (p *cloudstackProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaCloudStackMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaCloudStackDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *cloudstackProvider) GenerateStorageClass() []byte {
	return nil
}
