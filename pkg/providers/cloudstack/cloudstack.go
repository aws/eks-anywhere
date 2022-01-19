package cloudstack

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/crypto"

	etcdv1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	kubeadmnv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"

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
	cloudStackUsernameKey                 = "CLOUDSTACK_USERNAME"
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

	ubuntuDefaultUser = "capc"
	redhatDefaultUser = "capc"
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
	eksaCloudStackDeploymentResourceType = fmt.Sprintf("cloudstackdeploymentconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	noProxyDefaults                      = []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
)

type cloudstackProvider struct {
	deploymentConfig            *v1alpha1.CloudStackDeploymentConfig
	machineConfigs              map[string]*v1alpha1.CloudStackMachineConfig
	clusterConfig               *v1alpha1.Cluster
	providerCloudMonkeyClient   ProviderCloudMonkeyClient
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

func (p *cloudstackProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// panic("implement me")
	return nil
}

func (p *cloudstackProvider) RunPostControlPlaneCreation(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	// panic("implement me")
	return nil
}

type ProviderCloudMonkeyClient interface {
	SearchTemplate(ctx context.Context, domain string, zone string, account string, template string) (string, error)
	SearchComputeOffering(ctx context.Context, domain string, zone string, account string, computeOffering string) (string, error)
	SearchDiskOffering(ctx context.Context, domain string, zone string, account string, diskOffering string) (string, error)
	ValidateCloudStackSetup(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, selfSigned *bool) error
	ValidateCloudStackSetupMachineConfig(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfig *v1alpha1.CloudStackMachineConfig, selfSigned *bool) error
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaCloudStackDeploymentConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDeploymentConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmnv1alpha3.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*v1alpha3.MachineDeployment, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*etcdv1alpha3.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
	SearchCloudStackMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackMachineConfig, error)
	SearchCloudStackDeploymentConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackDeploymentConfig, error)
	DeleteEksaCloudStackDeploymentConfig(ctx context.Context, cloudstackDeploymentConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error
}

type ClusterResourceSetManager interface {
	ForceUpdate(ctx context.Context, name, namespace string, managementCluster, workloadCluster *types.Cluster) error
}

func NewProvider(deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCloudMonkeyClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *cloudstackProvider {
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

func NewProviderCustomNet(deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig, clusterConfig *v1alpha1.Cluster, providerCloudMonkeyClient ProviderCloudMonkeyClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, netClient networkutils.NetClient, now types.NowFunc, skipIpCheck bool, resourceSetManager ClusterResourceSetManager) *cloudstackProvider {
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
		deploymentConfig:            deploymentConfig,
		machineConfigs:              machineConfigs,
		clusterConfig:               clusterConfig,
		providerCloudMonkeyClient:   providerCloudMonkeyClient,
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
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.deploymentConfig.Spec.ManagementApiEndpoint)
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

// since control plane host can be FQDN or ip, there is no need to parse and confirm it is a valid IP address
func (p *cloudstackProvider) validateControlPlaneIp(ip string) error {
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

func (p *cloudstackProvider) validateEnv(ctx context.Context) error {
	if cloudStackB64EncodedSecret, ok := os.LookupEnv(eksacloudStackCloudConfigB64SecretKey); ok && len(cloudStackB64EncodedSecret) > 0 {
		if err := os.Setenv(cloudStackCloudConfigB64SecretKey, cloudStackB64EncodedSecret); err != nil {
			return fmt.Errorf("unable to set %s: %v", cloudStackCloudConfigB64SecretKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", cloudStackCloudConfigB64SecretKey)
	}
	if len(p.deploymentConfig.Spec.ManagementApiEndpoint) <= 0 {
		return errors.New("CloudStackDeploymentConfig managementApiEndpoint is not set or is empty")
	}
	if err := p.validateManagementApiEndpoint(p.deploymentConfig.Spec.ManagementApiEndpoint); err != nil {
		return errors.New("CloudStackDeploymentConfig managementApiEndpoint is invalid")
	}
	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}
	return nil
}

func (p *cloudstackProvider) validateSSHUsername(machineConfig *v1alpha1.CloudStackMachineConfig) error {
	if len(machineConfig.Spec.Users[0].Name) <= 0 {
		if machineConfig.Spec.OSFamily == v1alpha1.Ubuntu {
			machineConfig.Spec.Users[0].Name = ubuntuDefaultUser
		} else {
			machineConfig.Spec.Users[0].Name = redhatDefaultUser
		}
		logger.V(1).Info(fmt.Sprintf("SSHUsername is not set or is empty for CloudStackMachineConfig %v. Defaulting to %s", machineConfig.Name, machineConfig.Spec.Users[0].Name))
	}
	return nil
}

func (p *cloudstackProvider) setupAndValidateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	var etcdMachineConfig *v1alpha1.CloudStackMachineConfig
	if p.deploymentConfig.Spec.Insecure {
		logger.Info("Warning: CloudStackDeploymentConfig configured in insecure mode")
		p.deploymentConfig.Spec.Thumbprint = ""
	}
	if err := os.Setenv(cloudMonkeyInsecure, strconv.FormatBool(p.deploymentConfig.Spec.Insecure)); err != nil {
		return fmt.Errorf("unable to set %s: %v", cloudMonkeyInsecure, err)
	}
	if len(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if len(p.deploymentConfig.Spec.Domain) <= 0 {
		return errors.New("CloudStackDeploymentConfig domain is not set or is empty")
	}
	if len(p.deploymentConfig.Spec.Zone) <= 0 {
		return errors.New("CloudStackDeploymentConfig zone is not set or is empty")
	}
	if len(p.deploymentConfig.Spec.ManagementApiEndpoint) <= 0 {
		return errors.New("CloudStackDeploymentConfig managementApiEndpoint is not set or is empty")
	}
	if err := p.validateManagementApiEndpoint(p.deploymentConfig.Spec.ManagementApiEndpoint); err != nil {
		return errors.New("CloudStackDeploymentConfig managementApiEndpoint is invalid")
	}
	if len(p.deploymentConfig.Spec.Network) <= 0 {
		return errors.New("CloudStackDeploymentConfig network is not set or is empty")
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

	err := p.validateControlPlaneIp(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return err
	}

	err = p.providerCloudMonkeyClient.ValidateCloudStackSetup(ctx, p.deploymentConfig, &p.selfSigned)
	if err != nil {
		return fmt.Errorf("error validating CloudStack setup: %v", err)
	}
	for _, config := range p.machineConfigs {
		err = p.providerCloudMonkeyClient.ValidateCloudStackSetupMachineConfig(ctx, p.deploymentConfig, config, &p.selfSigned)
		if err != nil {
			return fmt.Errorf("error validating CloudStack setup for CloudStackMachineConfig %v: %v", config.Name, err)
		}
	}

	if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
		return errors.New("control plane and worker nodes must have the same osFamily specified")
	}

	if etcdMachineConfig != nil && controlPlaneMachineConfig.Spec.OSFamily != etcdMachineConfig.Spec.OSFamily {
		return errors.New("control plane and etcd machines must have the same osFamily specified")
	}
	if len(string(controlPlaneMachineConfig.Spec.OSFamily)) <= 0 {
		logger.Info("Warning: OS family not specified in cluster specification. Defaulting to Ubuntu.")
		controlPlaneMachineConfig.Spec.OSFamily = v1alpha1.Ubuntu
		workerNodeGroupMachineConfig.Spec.OSFamily = v1alpha1.Ubuntu
		if etcdMachineConfig != nil {
			etcdMachineConfig.Spec.OSFamily = v1alpha1.Ubuntu
		}
	}

	if err := p.validateSSHUsername(controlPlaneMachineConfig); err == nil {
		if err = p.validateSSHUsername(workerNodeGroupMachineConfig); err != nil {
			return fmt.Errorf("error validating SSHUsername for worker node CloudStackMachineConfig %v: %v", workerNodeGroupMachineConfig.Name, err)
		}
		if etcdMachineConfig != nil {
			if err = p.validateSSHUsername(etcdMachineConfig); err != nil {
				return fmt.Errorf("error validating SSHUsername for etcd CloudStackMachineConfig %v: %v", etcdMachineConfig.Name, err)
			}
		}
	} else {
		return fmt.Errorf("error validating SSHUsername for control plane CloudStackMachineConfig %v: %v", controlPlaneMachineConfig.Name, err)
	}

	for _, machineConfig := range p.machineConfigs {
		if machineConfig.Namespace != clusterSpec.Namespace {
			return errors.New("CloudStackMachineConfig and Cluster objects must have the same namespace specified")
		}
	}
	if p.deploymentConfig.Namespace != clusterSpec.Namespace {
		return errors.New("CloudStackDeploymentConfig and Cluster objects must have the same namespace specified")
	}

	if controlPlaneMachineConfig.Spec.Template == "" {
		logger.V(1).Info("Control plane CloudStackMachineConfig template is not set. Using default template.")
		err := p.setupDefaultTemplate(ctx, clusterSpec, controlPlaneMachineConfig, p.controlPlaneTemplateFactory)
		if err != nil {
			return err
		}
	}

	if err = p.validateMachineConfig(ctx, clusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane machine config validation failed.")
		return err
	}
	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		if workerNodeGroupMachineConfig.Spec.Template == "" {
			logger.V(1).Info("Worker CloudStackMachineConfig template is not set. Using default template.")
			err := p.setupDefaultTemplate(ctx, clusterSpec, workerNodeGroupMachineConfig, p.workerTemplateFactory)
			if err != nil {
				return err
			}
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
		if etcdMachineConfig.Spec.Template == "" {
			logger.V(1).Info("Etcd CloudStackMachineConfig template is not set. Using default template.")
			err := p.setupDefaultTemplate(ctx, clusterSpec, etcdMachineConfig, p.etcdTemplateFactory)
			if err != nil {
				return err
			}
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
	if template, err := p.validateTemplatePresence(ctx, p.deploymentConfig, machineConfig.Spec.Template); err != nil {
		return err
	} else {
		machineConfig.Spec.Template = template
	}

	if computeOffering, err := p.validateComputeOfferingPresence(ctx, p.deploymentConfig, machineConfig.Spec.ComputeOffering); err != nil {
		return err
	} else {
		machineConfig.Spec.ComputeOffering = computeOffering
	}

	if diskOffering, err := p.validateDiskOfferingPresence(ctx, p.deploymentConfig, machineConfig.Spec.DiskOffering); err != nil {
		return err
	} else {
		machineConfig.Spec.DiskOffering = diskOffering
	}

	return nil
}

func (p *cloudstackProvider) validateTemplatePresence(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, template string) (string, error) {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	templateFound, err := p.providerCloudMonkeyClient.SearchTemplate(ctx, domain, zone, account, template)
	if err != nil {
		return "", fmt.Errorf("error validating template: %v", err)
	}

	return templateFound, nil
}

func (p *cloudstackProvider) validateDiskOfferingPresence(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, diskOffering string) (string, error) {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	diskOfferingFound, err := p.providerCloudMonkeyClient.SearchDiskOffering(ctx, domain, zone, account, diskOffering)
	if err != nil {
		return "", fmt.Errorf("error validating diskOffering: %v", err)
	}

	return diskOfferingFound, nil
}

func (p *cloudstackProvider) validateComputeOfferingPresence(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, computeOffering string) (string, error) {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	computeOfferingFound, err := p.providerCloudMonkeyClient.SearchComputeOffering(ctx, domain, zone, account, computeOffering)
	if err != nil {
		return "", fmt.Errorf("error validating computeOffering: %v", err)
	}

	return computeOfferingFound, nil
}

func (p *cloudstackProvider) setupDefaultTemplate(ctx context.Context, clusterSpec *cluster.Spec, machineConfig *v1alpha1.CloudStackMachineConfig, templateFactory *templates.Factory) error {
	// TODO: set up default template for cloudstack, currently template must be specified.
	return fmt.Errorf("default template is not supported in CloudStack, please provide a template name")
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
		existingDatacenter, err := p.providerKubectlClient.SearchCloudStackDeploymentConfig(ctx, p.deploymentConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("CloudStackDeployment %s already exists", p.deploymentConfig.Name)
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
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	err = p.setupAndValidateCluster(ctx, clusterSpec)
	if err != nil {
		return err
	}
	err = p.setupSSHAuthKeysForUpgrade()
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	err = p.validateMachineConfigsNameUniqueness(ctx, cluster, clusterSpec)
	if err != nil {
		return fmt.Errorf("failed validate machineconfig uniqueness: %v", err)
	}
	return nil
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

func (p *cloudstackProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	return nil
}

func (p *cloudstackProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func NeedsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDeploymentConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host != newSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldCsdc, newCsdc, oldCsmc, newCsmc)
}

func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDeploymentConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldCsdc, newCsdc, oldCsmc, newCsmc)
}

func NeedsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldCsdc, newCsdc *v1alpha1.CloudStackDeploymentConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldCsdc, newCsdc, oldCsmc, newCsmc)
}

func AnyImmutableFieldChanged(oldCsdc, newCsdc *v1alpha1.CloudStackDeploymentConfig, oldCsmc, newCsmc *v1alpha1.CloudStackMachineConfig) bool {
	if oldCsdc.Spec.Network != newCsdc.Spec.Network {
		return true
	}
	if oldCsdc.Spec.Thumbprint != newCsdc.Spec.Thumbprint {
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

func NewCloudStackTemplateBuilder(cloudStackDeploymentConfigSpecSpec *v1alpha1.CloudStackDeploymentConfigSpec, controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &CloudStackTemplateBuilder{
		deploymentConfigSpec:       cloudStackDeploymentConfigSpecSpec,
		controlPlaneMachineSpec:    controlPlaneMachineSpec,
		workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
		etcdMachineSpec:            etcdMachineSpec,
		now:                        now,
	}
}

type CloudStackTemplateBuilder struct {
	deploymentConfigSpec       *v1alpha1.CloudStackDeploymentConfigSpec
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
	values := buildTemplateMapCP(clusterSpec, *vs.deploymentConfigSpec, *vs.controlPlaneMachineSpec, etcdMachineSpec)

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
	values := buildTemplateMapMD(clusterSpec, *vs.deploymentConfigSpec, *vs.workerNodeGroupMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultClusterConfigMD, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, deploymentConfigSpec v1alpha1.CloudStackDeploymentConfigSpec, controlPlaneMachineSpec, etcdMachineSpec v1alpha1.CloudStackMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                            clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointIp":                 clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
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
		"thumbprint":                             deploymentConfigSpec.Thumbprint,
		"cloudstackManagementApiEndpoint":        deploymentConfigSpec.ManagementApiEndpoint,
		"managerImage":                           bundle.CloudStack.Manager.VersionedImage(),
		"insecure":                               deploymentConfigSpec.Insecure,
		"cloudstackNetwork":                      deploymentConfigSpec.Network,
		"cloudstackDomain":                       deploymentConfigSpec.Domain,
		"cloudstackZone":                         deploymentConfigSpec.Zone,
		"cloudstackAccount":                      deploymentConfigSpec.Account,
		"cloudstackControlPlaneComputeOffering":  controlPlaneMachineSpec.ComputeOffering,
		"cloudstackControlPlaneTemplateOffering": controlPlaneMachineSpec.Template,
		"cloudstackControlPlaneDiskOffering":     controlPlaneMachineSpec.DiskOffering,
		"cloudstackControlPlaneDetails":          controlPlaneMachineSpec.Details,
		"cloudstackEtcdComputeOffering":          etcdMachineSpec.ComputeOffering,
		"cloudstackEtcdTemplateOffering":         etcdMachineSpec.Template,
		"cloudstackEtcdDiskOffering":             etcdMachineSpec.DiskOffering,
		"cloudstackEtcdDetails":                  etcdMachineSpec.Details,
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
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(deploymentConfigSpec.ManagementApiEndpoint)
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

func buildTemplateMapMD(clusterSpec *cluster.Spec, deploymentConfigSpec v1alpha1.CloudStackDeploymentConfigSpec, workerNodeGroupMachineSpec v1alpha1.CloudStackMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.ObjectMeta.Name,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"thumbprint":             deploymentConfigSpec.Thumbprint,
		"cloudstackNetwork":      deploymentConfigSpec.Network,
		"cloudstackTemplate":     workerNodeGroupMachineSpec.Template,
		"cloudstackOffering":     workerNodeGroupMachineSpec.ComputeOffering,
		"cloudstackDiskOffering": workerNodeGroupMachineSpec.DiskOffering,
		"cloudstackDetails":      workerNodeGroupMachineSpec.Details,
		"workerReplicas":         clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count,
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
		"format":                 format,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
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
		cloudStackManagementApiEndpointHostname, err := getHostnameFromUrl(deploymentConfigSpec.ManagementApiEndpoint)
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
	clusterName := newClusterSpec.ObjectMeta.Name
	var controlPlaneTemplateName, workloadTemplateName, etcdTemplateName string
	var needsNewEtcdTemplate bool

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster, newClusterSpec.Name)
	if err != nil {
		return nil, nil, err
	}
	vdc, err := p.providerKubectlClient.GetEksaCloudStackDeploymentConfig(ctx, p.deploymentConfig.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
	if err != nil {
		return nil, nil, err
	}
	controlPlaneMachineConfig := p.machineConfigs[newClusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneCsmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, c.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
	if err != nil {
		return nil, nil, err
	}
	workerMachineConfig := p.machineConfigs[newClusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	workerCsmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, c.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
	if err != nil {
		return nil, nil, err
	}

	needsNewControlPlaneTemplate := NeedsNewControlPlaneTemplate(currentSpec, newClusterSpec, vdc, p.deploymentConfig, controlPlaneCsmc, controlPlaneMachineConfig)
	if !needsNewControlPlaneTemplate {
		cp, err := p.providerKubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, c.Name, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, nil, err
		}
		controlPlaneTemplateName = cp.Spec.InfrastructureTemplate.Name
	} else {
		controlPlaneTemplateName = p.templateBuilder.CPMachineTemplateName(clusterName)
	}

	needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(currentSpec, newClusterSpec, vdc, p.deploymentConfig, workerCsmc, workerMachineConfig)
	if !needsNewWorkloadTemplate {
		md, err := p.providerKubectlClient.GetMachineDeployment(ctx, workloadCluster, clusterName, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, nil, err
		}
		workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
	} else {
		workloadTemplateName = p.templateBuilder.WorkerMachineTemplateName(clusterName)
	}

	if newClusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := p.machineConfigs[newClusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdMachineCsmc, err := p.providerKubectlClient.GetEksaCloudStackMachineConfig(ctx, c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, newClusterSpec.Namespace)
		if err != nil {
			return nil, nil, err
		}
		needsNewEtcdTemplate = NeedsNewEtcdTemplate(currentSpec, newClusterSpec, vdc, p.deploymentConfig, etcdMachineCsmc, etcdMachineConfig)
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
				map[string]string{etcdv1alpha3.UpgradeInProgressAnnotation: "true"},
				executables.WithCluster(bootstrapCluster),
				executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, nil, err
			}
			etcdTemplateName = p.templateBuilder.EtcdMachineTemplateName(clusterName)
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["cloudstackControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["cloudstackEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["etcdTemplateName"] = etcdTemplateName
	}
	controlPlaneSpec, err = p.templateBuilder.GenerateCAPISpecControlPlane(newClusterSpec, cpOpt)
	if err != nil {
		return nil, nil, err
	}

	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = workloadTemplateName
		values["cloudstackWorkerSshAuthorizedKey"] = p.workerSshAuthKey
	}
	workersSpec, err = p.templateBuilder.GenerateCAPISpecWorkers(newClusterSpec, workersOpt)
	if err != nil {
		return nil, nil, err
	}
	return controlPlaneSpec, workersSpec, nil
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
	controlPlaneSpec, workersSpec, err = p.generateCAPISpecForUpgrade(ctx, bootstrapCluster, workloadCluster, currentSpec, clusterSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating cluster api spec contents: %v", err)
	}
	return controlPlaneSpec, workersSpec, nil
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
		"clusterName": cluster.Name,
		"insecure":    strconv.FormatBool(p.deploymentConfig.Spec.Insecure),
		"cloudstackNetwork":  p.deploymentConfig.Spec.Network,
		"eksaLicense":        os.Getenv(eksaLicense),
	}
	err = t.Execute(contents, values)
	if err != nil {
		return fmt.Errorf("error substituting values for secret object template: %v", err)
	}
	return nil
}

func (p *cloudstackProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	//var contents bytes.Buffer
	//err := p.createSecret(cluster, &contents)
	//if err != nil {
	//	return err
	//}
	//
	//var loadContents bytes.Buffer
	//loadContents.WriteString("data=")
	//loadContents.WriteString(contents.String())
	//err = p.providerKubectlClient.LoadSecret(ctx, loadContents.String(), secretObjectType, secretObjectName, cluster.KubeconfigFile)
	//if err != nil {
	//	return fmt.Errorf("error loading csi-cloudstack-secret object: %v", err)
	//}
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
		"capc-system":         {"capc-controller-manager"},
		"capi-webhook-system": {"capi-controller-manager"},
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
	return p.deploymentConfig
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

func (p *cloudstackProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Name)
	if err != nil {
		return err
	}

	prevDatacenter, err := p.providerKubectlClient.GetEksaCloudStackDeploymentConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	datacenter := p.deploymentConfig

	oSpec := prevDatacenter.Spec
	nSpec := datacenter.Spec

	for _, machineConfigRef := range clusterSpec.MachineConfigRefs() {
		machineConfig, ok := p.machineConfigs[machineConfigRef.Name]
		if !ok {
			return fmt.Errorf("cannot find machine config %s in cloudstack provider machine configs", machineConfigRef.Name)
		}

		err = p.validateMachineConfigImmutability(ctx, cluster, machineConfig, clusterSpec)
		if err != nil {
			return err
		}
	}

	if nSpec.ManagementApiEndpoint != oSpec.ManagementApiEndpoint {
		return fmt.Errorf("spec.managementApiEndpoint is immutable. Previous value %s, new value %s", oSpec.ManagementApiEndpoint, nSpec.ManagementApiEndpoint)
	}
	if nSpec.Domain != oSpec.Domain {
		return fmt.Errorf("spec.domain is immutable. Previous value %s, new value %s", oSpec.Domain, nSpec.Domain)
	}
	if nSpec.Zone != oSpec.Zone {
		return fmt.Errorf("spec.zone is immutable. Previous value %s, new value %s", oSpec.Zone, nSpec.Zone)
	}
	if nSpec.Account != oSpec.Account {
		return fmt.Errorf("spec.account is immutable. Previous value %s, new value %s", oSpec.Account, nSpec.Account)
	}

	if nSpec.Network != oSpec.Network {
		return fmt.Errorf("spec.network is immutable. Previous value %s, new value %s", oSpec.Network, nSpec.Network)
	}

	if nSpec.Insecure != oSpec.Insecure {
		return fmt.Errorf("spec.insecure is immutable. Previous value %t, new value %t", oSpec.Insecure, nSpec.Insecure)
	}

	if nSpec.Thumbprint != oSpec.Thumbprint {
		return fmt.Errorf("spec.thumbprint is immutable. Previous value %s, new value %s", oSpec.Thumbprint, nSpec.Thumbprint)
	}
	// TODO: Add support for changing cloudstack credentials

	return nil
}

func (p *cloudstackProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.CloudStackMachineConfig, clusterSpec *cluster.Spec) error {
	// allow template, compute offering, details, users to mutate
	return nil
}

func (p *cloudstackProvider) secretContentsChanged(ctx context.Context, workloadCluster *types.Cluster) (bool, error) {
	cloudConfig := os.Getenv(eksacloudStackCloudConfigB64SecretKey)
	oSecret, err := p.providerKubectlClient.GetSecret(ctx, credentialsObjectName, executables.WithCluster(workloadCluster), executables.WithNamespace(constants.CapcSystemNamespace))
	if err != nil {
		return false, fmt.Errorf("error when obtaining CloudStack secret %s from workload cluster: %v", credentialsObjectName, err)
	}
	if string(oSecret.Data["cloud-config"]) != cloudConfig {
		return true, nil
	}
	return false, nil
}

func (p *cloudstackProvider) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.CloudStack.Version == newSpec.VersionsBundle.CloudStack.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: constants.CloudStackProviderName,
		NewVersion:    newSpec.VersionsBundle.CloudStack.Version,
		OldVersion:    currentSpec.VersionsBundle.CloudStack.Version,
	}
}

func (p *cloudstackProvider) RunPostUpgrade(ctx context.Context, clusterSpec *cluster.Spec, managementCluster, workloadCluster *types.Cluster) error {
	if err := p.resourceSetManager.ForceUpdate(ctx, resourceSetName(clusterSpec), constants.EksaSystemNamespace, managementCluster, workloadCluster); err != nil {
		return fmt.Errorf("failed updating the cloudstack provider resource set post upgrade: %v", err)
	}
	return nil
}

func resourceSetName(clusterSpec *cluster.Spec) string {
	return fmt.Sprintf("%s-crs-0", clusterSpec.Name)
}

func (p *cloudstackProvider) UpgradeNeeded(ctx context.Context, newSpec, currentSpec *cluster.Spec) (bool, error) {
	newV, oldV := newSpec.VersionsBundle.CloudStack, currentSpec.VersionsBundle.CloudStack

	return newV.Manager.ImageDigest != oldV.Manager.ImageDigest, nil
	// || newV.KubeVip.ImageDigest != oldV.KubeVip.ImageDigest, nil
}

func (p *cloudstackProvider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaCloudStackMachineConfig(ctx, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaCloudStackDeploymentConfig(ctx, p.deploymentConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.deploymentConfig.Namespace)
}

func (p *cloudstackProvider) GenerateStorageClass() []byte {
	panic("implement me")
}
