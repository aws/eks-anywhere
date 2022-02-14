package cloudstack

import (
	"bytes"
	"context"
	_ "embed"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/ini.v1"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/crypto"

	etcdv1beta1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	kubeadmnv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

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
	credentialsObjectName                 = "capc-cloudstack-secret"
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
	resourceSetManager          ClusterResourceSetManager
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
	ValidateAccountPresent(ctx context.Context, account string) error
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmnv1beta1.KubeadmControlPlane, error)
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

func NewProvider(deploymentConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCmkClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *cloudstackProvider {
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
		resourceSetManager,
	)
}

func NewProviderCustomNet(deploymentConfig *v1alpha1.CloudStackDatacenterConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCmkClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, netClient networkutils.NetClient, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *cloudstackProvider {
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
		resourceSetManager: resourceSetManager,
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
	controlPlaneUser := p.machineConfigs[p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
	p.controlPlaneSshAuthKey = controlPlaneUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.controlPlaneSshAuthKey); err != nil {
		return err
	}
	controlPlaneUser.SshAuthorizedKeys[0] = p.controlPlaneSshAuthKey
	workerUser := p.machineConfigs[p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec.Users[0]
	p.workerSshAuthKey = workerUser.SshAuthorizedKeys[0]
	if err := p.parseSSHAuthKey(&p.workerSshAuthKey); err != nil {
		return err
	}
	workerUser.SshAuthorizedKeys[0] = p.workerSshAuthKey
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdUser := p.machineConfigs[p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
		p.etcdSshAuthKey = etcdUser.SshAuthorizedKeys[0]
		if err := p.parseSSHAuthKey(&p.etcdSshAuthKey); err != nil {
			return err
		}
		etcdUser.SshAuthorizedKeys[0] = p.etcdSshAuthKey
	}
	return nil
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

func (p *cloudstackProvider) validateControlPlaneHost(pHost *string) error {
	_, port, err := net.SplitHostPort(*pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			port = controlEndpointDefaultPort
			*pHost = fmt.Sprintf("%s:%s", *pHost, port)
		} else {
			return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s (%s)", *pHost, err.Error())
		}
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host has an invalid port: %s (%s)", *pHost, err.Error())
	}
	return nil
}

func (p *cloudstackProvider) validateControlPlaneIpUniqueness(ip string) error {
	// check if control plane endpoint ip is unique
	// TODO: the ip can be existed in cloudstack when using isolated network.
	//ipgen := networkutils.NewIPGenerator(p.netClient)
	//if !ipgen.IsIPUnique(ip) {
	//	return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	//}
	return nil
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

type execConfig struct {
	CloudStackApiKey        string
	CloudStackSecretKey     string
	CloudStackManagementUrl string
}

func parseCloudStackSecret() (*execConfig, error) {
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
	return &execConfig{
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

func (p *cloudstackProvider) setupAndValidateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	var etcdMachineConfig *v1alpha1.CloudStackMachineConfig
	if p.datacenterConfig.Spec.Insecure {
		logger.Info("Warning: CloudStackDatacenterConfig configured in insecure mode")
	}
	if err := os.Setenv(cloudMonkeyInsecure, strconv.FormatBool(p.datacenterConfig.Spec.Insecure)); err != nil {
		return fmt.Errorf("unable to set %s: %v", cloudMonkeyInsecure, err)
	}
	if len(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if len(p.datacenterConfig.Spec.Domain) <= 0 {
		return errors.New("CloudStackDatacenterConfig domain is not set or is empty")
	}
	if len(p.datacenterConfig.Spec.Zone.Value) <= 0 {
		return errors.New("CloudStackDatacenterConfig zone is not set or is empty")
	}
	if len(p.datacenterConfig.Spec.Network.Value) <= 0 {
		return errors.New("CloudStackDatacenterConfig network is not set or is empty")
	}
	if clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}
	controlPlaneMachineConfig, ok := p.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	if !ok {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for control plane", clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}

	if clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig, ok := p.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	if !ok {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for worker nodes", clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		var ok bool
		if clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for etcd machines")
		}
		etcdMachineConfig, ok = p.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		if !ok {
			return fmt.Errorf("cannot find CloudStackMachineConfig %v for etcd machines", clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if len(etcdMachineConfig.Spec.Users) <= 0 {
			etcdMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
		}
		if len(etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
			etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
	}

	if len(controlPlaneMachineConfig.Spec.Users) <= 0 {
		controlPlaneMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
	}
	if len(workerNodeGroupMachineConfig.Spec.Users) <= 0 {
		workerNodeGroupMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
	}
	if len(controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}
	if len(workerNodeGroupMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		workerNodeGroupMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	err := p.validateControlPlaneHost(&clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return err
	}

	// TODO: Validate CloudStack Setup
	//err = p.providerCmkClient.ValidateCloudStackSetup(ctx, p.datacenterConfig, &p.selfSigned)
	//if err != nil {
	//	return fmt.Errorf("error validating CloudStack setup: %v", err)
	//}
	//for _, config := range p.machineConfigs {
	//	err = p.providerCmkClient.ValidateCloudStackSetupMachineConfig(ctx, p.datacenterConfig, config, &p.selfSigned)
	//	if err != nil {
	//		return fmt.Errorf("error validating CloudStack setup for CloudStackMachineConfig %v: %v", config.Name, err)
	//	}
	//}

	for _, machineConfig := range p.machineConfigs {
		if machineConfig.Namespace != clusterSpec.Namespace {
			return errors.New("CloudStackMachineConfig and Cluster objects must have the same namespace specified")
		}
	}
	if p.datacenterConfig.Namespace != clusterSpec.Namespace {
		return errors.New("CloudStackDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	if controlPlaneMachineConfig.Spec.Template.Value == "" {
		return fmt.Errorf("control plane CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
	}

	if err = p.validateMachineConfig(ctx, clusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane machine config validation failed.")
		return err
	}
	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		if workerNodeGroupMachineConfig.Spec.Template.Value == "" {
			return fmt.Errorf("worker CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
		}
		if err = p.validateMachineConfig(ctx, clusterSpec, workerNodeGroupMachineConfig); err != nil {
			logger.V(1).Info("Workload machine config validation failed.")
			return err
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return errors.New("control plane and worker nodes must have the same template specified")
		}
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if etcdMachineConfig.Spec.Template.Value == "" {
			return fmt.Errorf("etcd CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
		}
		if err = p.validateMachineConfig(ctx, clusterSpec, etcdMachineConfig); err != nil {
			logger.V(1).Info("Etcd machine config validation failed.")
			return err
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	return nil
}

func (p *cloudstackProvider) validateMachineConfig(ctx context.Context, clusterSpec *cluster.Spec, machineConfig *v1alpha1.CloudStackMachineConfig) error {
	if err := p.validateTemplatePresence(ctx, p.datacenterConfig, machineConfig.Spec.Template); err != nil {
		return err
	}

	if err := p.validateComputeOfferingPresence(ctx, p.datacenterConfig, machineConfig.Spec.ComputeOffering); err != nil {
		return err
	}

	return nil
}

func (p *cloudstackProvider) validateTemplatePresence(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDatacenterConfig, template v1alpha1.CloudStackResourceRef) error {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	err := p.providerCmkClient.ValidateTemplatePresent(ctx, domain, zone, account, template)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	return nil
}

func (p *cloudstackProvider) validateComputeOfferingPresence(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDatacenterConfig, computeOffering v1alpha1.CloudStackResourceRef) error {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	err := p.providerCmkClient.ValidateServiceOfferingPresent(ctx, domain, zone, account, computeOffering)
	if err != nil {
		return fmt.Errorf("error validating computeOffering: %v", err)
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	err = p.setupAndValidateCluster(ctx, clusterSpec)
	if err != nil {
		return err
	}
	err = p.setupSSHAuthKeysForCreate()
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	if clusterSpec.IsManaged() {
		for _, mc := range p.MachineConfigs() {
			em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, mc.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, mc.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("CloudStackMachineConfig %s already exists", mc.GetName())
			}
		}
		existingDatacenter, err := p.providerKubectlClient.SearchCloudStackDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("CloudStackDeployment %s already exists", p.datacenterConfig.Name)
		}
	}

	if p.skipIpCheck {
		logger.Info("Skipping check for whether control plane ip is in use")
		return nil
	}
	err = p.validateControlPlaneIpUniqueness(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return err
	}
	return nil
}

func (p *cloudstackProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return fmt.Errorf("upgrade is not yet supported for CloudStack cluster!")
}

func (p *cloudstackProvider) validateMachineConfigsNameUniqueness(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.GetName())
	if err != nil {
		return err
	}

	cpMachineConfigName := clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	if prevSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name != cpMachineConfigName {
		em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, cpMachineConfigName, cluster.KubeconfigFile, clusterSpec.GetNamespace())
		if err != nil {
			return err
		}
		if len(em) > 0 {
			return fmt.Errorf("control plane CloudStackMachineConfig %s already exists", cpMachineConfigName)
		}
	}

	workerMachineConfigName := clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	if prevSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name != workerMachineConfigName {
		em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, workerMachineConfigName, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.GetNamespace())
		if err != nil {
			return err
		}
		if len(em) > 0 {
			return fmt.Errorf("worker nodes CloudStackMachineConfig %s already exists", workerMachineConfigName)
		}
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil && prevSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigName := clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		if prevSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name != etcdMachineConfigName {
			em, err := p.providerKubectlClient.SearchCloudStackMachineConfig(ctx, etcdMachineConfigName, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("external etcd machineconfig %s already exists", etcdMachineConfigName)
			}
		}
	}

	return nil
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func AnyImmutableFieldChanged(oldCsdc, newCsdc *v1alpha1.CloudStackDatacenterConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldCsdc.Spec.Network != newCsdc.Spec.Network {
		return true
	}
	if oldCsmc.Spec.Template != newCsmc.Spec.Template {
		return true
	}
	if oldCsmc.Spec.ComputeOffering != newCsmc.Spec.ComputeOffering {
		return true
	}
	return false
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
			bundle.CloudStack.ClusterTemplate,
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
