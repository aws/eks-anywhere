package vsphere

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	etcdv1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	kubeadmnv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/templates"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	ProviderName             = "vsphere"
	eksaLicense              = "EKSA_LICENSE"
	vSphereUsernameKey       = "VSPHERE_USERNAME"
	vSpherePasswordKey       = "VSPHERE_PASSWORD"
	eksavSphereUsernameKey   = "EKSA_VSPHERE_USERNAME"
	eksavSpherePasswordKey   = "EKSA_VSPHERE_PASSWORD"
	vSphereServerKey         = "VSPHERE_SERVER"
	govcInsecure             = "GOVC_INSECURE"
	expClusterResourceSetKey = "EXP_CLUSTER_RESOURCE_SET"
	secretObjectType         = "addons.cluster.x-k8s.io/resource-set"
	secretObjectName         = "csi-vsphere-config"
	credentialsObjectName    = "vsphere-credentials"
	privateKeyFileName       = "eks-a-id_rsa"
	publicKeyFileName        = "eks-a-id_rsa.pub"
	defaultTemplateLibrary   = "eks-a-templates"
	defaultTemplatesFolder   = "vm/Templates"
	bottlerocketDefaultUser  = "ec2-user"
	ubuntuDefaultUser        = "capv"
)

//go:embed config/template.yaml
var defaultClusterConfig string

//go:embed config/secret.yaml
var defaultSecretObject string

//go:embed config/defaultStorageClass.yaml
var defaultStorageClass []byte

//go:embed config/machine-health-check-template.yaml
var mhcTemplate []byte

var (
	eksaVSphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType    = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
)

var requiredEnvs = []string{vSphereUsernameKey, vSpherePasswordKey, expClusterResourceSetKey}

type vsphereProvider struct {
	datacenterConfig            *v1alpha1.VSphereDatacenterConfig
	machineConfigs              map[string]*v1alpha1.VSphereMachineConfig
	clusterConfig               *v1alpha1.Cluster
	providerGovcClient          ProviderGovcClient
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
	templateBuilder             *VsphereTemplateBuilder
	skipIpCheck                 bool
}

type ProviderGovcClient interface {
	SearchTemplate(ctx context.Context, datacenter string, machineConfig *v1alpha1.VSphereMachineConfig) (string, error)
	LibraryElementExists(ctx context.Context, library string) (bool, error)
	TemplateHasSnapshot(ctx context.Context, template string) (bool, error)
	GetWorkloadAvailableSpace(ctx context.Context, machineConfig *v1alpha1.VSphereMachineConfig) (float64, error)
	ValidateVCenterSetup(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, selfSigned *bool) error
	ValidateVCenterSetupMachineConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfig *v1alpha1.VSphereMachineConfig, selfSigned *bool) error
	CreateLibrary(ctx context.Context, datastore, library string) error
	DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, resourcePool string, resizeDisk2 bool) error
	ImportTemplate(ctx context.Context, library, ovaURL, name string) error
	GetTags(ctx context.Context, path string) (tags []string, err error)
	ListTags(ctx context.Context) ([]string, error)
	CreateTag(ctx context.Context, tag, category string) error
	AddTag(ctx context.Context, path, tag string) error
	ListCategories(ctx context.Context) ([]string, error)
	CreateCategoryForVM(ctx context.Context, name string) error
}

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error
	LoadSecret(ctx context.Context, secretObject string, secretObjType string, secretObjectName string, kubeConfFile string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster) (*v1alpha1.Cluster, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaVSphereMachineConfig(ctx context.Context, vsphereMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, opts ...executables.KubectlOpt) (*kubeadmnv1alpha3.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, cluster *types.Cluster, opts ...executables.KubectlOpt) (*v1alpha3.MachineDeployment, error)
	GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, opts ...executables.KubectlOpt) (*etcdv1alpha3.EtcdadmCluster, error)
	GetSecret(ctx context.Context, secretObjectName string, opts ...executables.KubectlOpt) (*corev1.Secret, error)
	UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...executables.KubectlOpt) error
}

func NewProvider(datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, providerGovcClient ProviderGovcClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, now types.NowFunc, skipIpCheck bool) *vsphereProvider {
	return NewProviderCustomNet(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		providerGovcClient,
		providerKubectlClient,
		writer,
		&networkutils.DefaultNetClient{},
		now,
		skipIpCheck,
	)
}

func NewProviderCustomNet(datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, providerGovcClient ProviderGovcClient, providerKubectlClient ProviderKubectlClient, writer filewriter.FileWriter, netClient networkutils.NetClient, now types.NowFunc, skipIpCheck bool) *vsphereProvider {
	var controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.VSphereMachineConfigSpec
	var controlPlaneTemplateFactory, workerNodeGroupTemplateFactory, etcdTemplateFactory *templates.Factory
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
		controlPlaneTemplateFactory = templates.NewFactory(
			providerGovcClient,
			controlPlaneMachineSpec.Datastore,
			controlPlaneMachineSpec.ResourcePool,
			defaultTemplateLibrary,
		)
	}
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) > 0 && clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name] != nil {
		workerNodeGroupMachineSpec = &machineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Spec
		workerNodeGroupTemplateFactory = templates.NewFactory(
			providerGovcClient,
			workerNodeGroupMachineSpec.Datastore,
			workerNodeGroupMachineSpec.ResourcePool,
			defaultTemplateLibrary,
		)
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
			etcdTemplateFactory = templates.NewFactory(
				providerGovcClient,
				etcdMachineSpec.Datastore,
				etcdMachineSpec.ResourcePool,
				defaultTemplateLibrary,
			)
		}
	}
	return &vsphereProvider{
		datacenterConfig:            datacenterConfig,
		machineConfigs:              machineConfigs,
		clusterConfig:               clusterConfig,
		providerGovcClient:          providerGovcClient,
		providerKubectlClient:       providerKubectlClient,
		writer:                      writer,
		selfSigned:                  false,
		netClient:                   netClient,
		controlPlaneTemplateFactory: controlPlaneTemplateFactory,
		workerTemplateFactory:       workerNodeGroupTemplateFactory,
		etcdTemplateFactory:         etcdTemplateFactory,
		templateBuilder: &VsphereTemplateBuilder{
			datacenterSpec:             &datacenterConfig.Spec,
			controlPlaneMachineSpec:    controlPlaneMachineSpec,
			workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
			etcdMachineSpec:            etcdMachineSpec,
			now:                        now,
		},
		skipIpCheck: skipIpCheck,
	}
}

func (p *vsphereProvider) UpdateKubeConfig(_ *[]byte, _ string) error {
	// customize generated kube config
	return nil
}

func (p *vsphereProvider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.Server)
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

func (p *vsphereProvider) Name() string {
	return ProviderName
}

func (p *vsphereProvider) DatacenterResourceType() string {
	return eksaVSphereDatacenterResourceType
}

func (p *vsphereProvider) MachineResourceType() string {
	return eksaVSphereMachineResourceType
}

func (p *vsphereProvider) setupSSHAuthKeysForCreate() error {
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

func (p *vsphereProvider) setupSSHAuthKeysForUpgrade() error {
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

func (p *vsphereProvider) parseSSHAuthKey(key *string) error {
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
			return fmt.Errorf("provided VSphereMachineConfig sshAuthorizedKey is invalid: %v", err)
		}
	}
	return nil
}

func (p *vsphereProvider) generateSSHAuthKey(username string) (string, error) {
	logger.Info("Provided VSphereMachineConfig sshAuthorizedKey is not set or is empty, auto-generating new key pair...")
	keygenerator, _ := crypto.NewKeyGenerator(p.writer)
	sshAuthorizedKeyBytes, err := keygenerator.GenerateSSHKeyPair("", "", privateKeyFileName, publicKeyFileName, username)
	if err != nil || sshAuthorizedKeyBytes == nil {
		return "", fmt.Errorf("VSphereMachineConfig error generating sshAuthorizedKey: %v", err)
	}
	key := string(sshAuthorizedKeyBytes)
	key = strings.TrimRight(key, "\n")
	return key, nil
}

func (p *vsphereProvider) validateControlPlaneIp(ip string) error {
	// check if controlPlaneEndpointIp is valid
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s", ip)
	}
	return nil
}

func (p *vsphereProvider) validateControlPlaneIpUniqueness(ip string) error {
	// check if control plane endpoint ip is unique
	ipgen := networkutils.NewIPGenerator(p.netClient)
	if !ipgen.IsIPUnique(ip) {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	}
	return nil
}

func (p *vsphereProvider) validateEnv(ctx context.Context) error {
	if vSphereUsername, ok := os.LookupEnv(eksavSphereUsernameKey); ok && len(vSphereUsername) > 0 {
		if err := os.Setenv(vSphereUsernameKey, vSphereUsername); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksavSphereUsernameKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", eksavSphereUsernameKey)
	}
	if vSpherePassword, ok := os.LookupEnv(eksavSpherePasswordKey); ok && len(vSpherePassword) > 0 {
		if err := os.Setenv(vSpherePasswordKey, vSpherePassword); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksavSpherePasswordKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", eksavSpherePasswordKey)
	}
	if len(p.datacenterConfig.Spec.Server) <= 0 {
		return errors.New("VSphereDatacenterConfig server is not set or is empty")
	}
	if err := os.Setenv(vSphereServerKey, p.datacenterConfig.Spec.Server); err != nil {
		return fmt.Errorf("unable to set %s: %v", vSphereServerKey, err)
	}
	if err := os.Setenv(expClusterResourceSetKey, "true"); err != nil {
		return fmt.Errorf("unable to set %s: %v", expClusterResourceSetKey, err)
	}
	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}
	return nil
}

func (p *vsphereProvider) validateSSHUsername(machineConfig *v1alpha1.VSphereMachineConfig) error {
	if len(machineConfig.Spec.Users[0].Name) <= 0 {
		if machineConfig.Spec.OSFamily == v1alpha1.Bottlerocket {
			machineConfig.Spec.Users[0].Name = bottlerocketDefaultUser
		} else {
			machineConfig.Spec.Users[0].Name = ubuntuDefaultUser
		}
		logger.V(1).Info(fmt.Sprintf("SSHUsername is not set or is empty for VSphereMachineConfig %v. Defaulting to %s", machineConfig.Name, machineConfig.Spec.Users[0].Name))
	} else {
		if machineConfig.Spec.OSFamily == v1alpha1.Bottlerocket {
			if machineConfig.Spec.Users[0].Name != bottlerocketDefaultUser {
				return fmt.Errorf("SSHUsername %s is invalid. Please use 'ec2-user' for Bottlerocket.", machineConfig.Spec.Users[0].Name)
			}
		}
	}
	return nil
}

func (p *vsphereProvider) setupAndValidateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	var etcdMachineConfig *v1alpha1.VSphereMachineConfig
	if p.datacenterConfig.Spec.Insecure {
		logger.Info("Warning: VSphereDatacenterConfig configured in insecure mode")
		p.datacenterConfig.Spec.Thumbprint = ""
	}
	if err := os.Setenv(govcInsecure, strconv.FormatBool(p.datacenterConfig.Spec.Insecure)); err != nil {
		return fmt.Errorf("unable to set %s: %v", govcInsecure, err)
	}
	if clusterSpec.Spec.ControlPlaneConfiguration.Count != 3 && clusterSpec.Spec.ControlPlaneConfiguration.Count != 5 && clusterSpec.Spec.ExternalEtcdConfiguration == nil {
		logger.Info("Warning: The recommended number of control plane nodes is 3 or 5")
	}
	if clusterSpec.Spec.ExternalEtcdConfiguration != nil && clusterSpec.Spec.ExternalEtcdConfiguration.Count != 3 && clusterSpec.Spec.ExternalEtcdConfiguration.Count != 5 {
		logger.Info("Warning: The recommended number of etcd machines is 3 or 5")
	}
	if len(clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if len(p.datacenterConfig.Spec.Datacenter) <= 0 {
		return errors.New("VSphereDatacenterConfig datacenter is not set or is empty")
	}
	if len(p.datacenterConfig.Spec.Network) <= 0 {
		return errors.New("VSphereDatacenterConfig VM network is not set or is empty")
	}
	if clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}
	controlPlaneMachineConfig, ok := p.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	if !ok {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for control plane", clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}
	if controlPlaneMachineConfig.Spec.MemoryMiB <= 0 {
		logger.V(1).Info("VSphereMachineConfig MemoryMiB for control plane is not set or is empty. Defaulting to 8192.")
		controlPlaneMachineConfig.Spec.MemoryMiB = 8192
	}
	if controlPlaneMachineConfig.Spec.MemoryMiB < 2048 {
		logger.Info("Warning: VSphereMachineConfig MemoryMiB for control plane should not be less than 2048. Defaulting to 2048. Recommended memory is 8192.")
		controlPlaneMachineConfig.Spec.MemoryMiB = 2048
	}
	if controlPlaneMachineConfig.Spec.NumCPUs <= 0 {
		logger.V(1).Info("VSphereMachineConfig NumCPUs for control plane is not set or is empty. Defaulting to 2.")
		controlPlaneMachineConfig.Spec.NumCPUs = 2
	}
	if len(controlPlaneMachineConfig.Spec.Datastore) <= 0 {
		return errors.New("VSphereMachineConfig datastore for control plane is not set or is empty")
	}
	if len(controlPlaneMachineConfig.Spec.Folder) <= 0 {
		logger.Info("VSphereMachineConfig folder for control plane is not set or is empty. Will default to root vSphere folder.")
	}
	if len(controlPlaneMachineConfig.Spec.ResourcePool) <= 0 {
		return errors.New("VSphereMachineConfig VM resourcePool for control plane is not set or is empty")
	}
	if clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig, ok := p.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	if !ok {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for worker nodes", clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
	}
	if workerNodeGroupMachineConfig.Spec.MemoryMiB <= 0 {
		logger.V(1).Info("VSphereMachineConfig MemoryMiB for worker nodes is not set or is empty. Defaulting to 8192.")
		workerNodeGroupMachineConfig.Spec.MemoryMiB = 8192
	}
	if workerNodeGroupMachineConfig.Spec.MemoryMiB < 2048 {
		logger.Info("Warning: VSphereMachineConfig MemoryMiB for worker nodes should not be less than 2048. Defaulting to 2048. Recommended memory is 8192.")
		workerNodeGroupMachineConfig.Spec.MemoryMiB = 2048
	}
	if workerNodeGroupMachineConfig.Spec.NumCPUs <= 0 {
		logger.V(1).Info("VSphereMachineConfig NumCPUs for worker nodes is not set or is empty. Defaulting to 2.")
		workerNodeGroupMachineConfig.Spec.NumCPUs = 2
	}
	if len(workerNodeGroupMachineConfig.Spec.Datastore) <= 0 {
		return errors.New("VSphereMachineConfig datastore for worker nodes is not set or is empty")
	}
	if len(workerNodeGroupMachineConfig.Spec.Folder) <= 0 {
		logger.Info("VSphereMachineConfig folder for worker nodes is not set or is empty. Will default to root vSphere folder.")
	}
	if len(workerNodeGroupMachineConfig.Spec.ResourcePool) <= 0 {
		return errors.New("VSphereMachineConfig VM resourcePool for worker nodes is not set or is empty")
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		var ok bool
		if clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for etcd machines")
		}
		etcdMachineConfig, ok = p.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		if !ok {
			return fmt.Errorf("cannot find VSphereMachineConfig %v for etcd machines", clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if etcdMachineConfig.Spec.MemoryMiB <= 0 {
			logger.V(1).Info("VSphereMachineConfig MemoryMiB for etcd machines is not set or is empty. Defaulting to 8192.")
			etcdMachineConfig.Spec.MemoryMiB = 8192
		}
		if etcdMachineConfig.Spec.MemoryMiB < 8192 {
			logger.Info("Warning: VSphereMachineConfig MemoryMiB for etcd machines should not be less than 8192. Defaulting to 8192")
			etcdMachineConfig.Spec.MemoryMiB = 8192
		}
		if etcdMachineConfig.Spec.NumCPUs <= 0 {
			logger.V(1).Info("VSphereMachineConfig NumCPUs for etcd machines is not set or is empty. Defaulting to 2.")
			etcdMachineConfig.Spec.NumCPUs = 2
		}
		if len(etcdMachineConfig.Spec.Datastore) <= 0 {
			return errors.New("VSphereMachineConfig datastore for etcd machines is not set or is empty")
		}
		if len(etcdMachineConfig.Spec.Folder) <= 0 {
			logger.Info("VSphereMachineConfig folder for etcd machines is not set or is empty. Will default to root vSphere folder.")
		}
		if len(etcdMachineConfig.Spec.ResourcePool) <= 0 {
			return errors.New("VSphereMachineConfig VM resourcePool for etcd machines is not set or is empty")
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

	err = p.providerGovcClient.ValidateVCenterSetup(ctx, p.datacenterConfig, &p.selfSigned)
	if err != nil {
		return fmt.Errorf("error validating vCenter setup: %v", err)
	}
	for _, config := range p.machineConfigs {
		err = p.providerGovcClient.ValidateVCenterSetupMachineConfig(ctx, p.datacenterConfig, config, &p.selfSigned)
		if err != nil {
			return fmt.Errorf("error validating vCenter setup for VSphereMachineConfig %v: %v", config.Name, err)
		}
	}

	if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
		return errors.New("control plane and worker nodes must have the same osFamily specified")
	}

	if etcdMachineConfig != nil && controlPlaneMachineConfig.Spec.OSFamily != etcdMachineConfig.Spec.OSFamily {
		return errors.New("control plane and etcd machines must have the same osFamily specified")
	}
	if len(string(controlPlaneMachineConfig.Spec.OSFamily)) <= 0 {
		logger.Info("Warning: OS family not specified in cluster specification. Defaulting to Bottlerocket.")
		controlPlaneMachineConfig.Spec.OSFamily = v1alpha1.Bottlerocket
		workerNodeGroupMachineConfig.Spec.OSFamily = v1alpha1.Bottlerocket
		if etcdMachineConfig != nil {
			etcdMachineConfig.Spec.OSFamily = v1alpha1.Bottlerocket
		}
	}

	if err := p.validateSSHUsername(controlPlaneMachineConfig); err == nil {
		if err = p.validateSSHUsername(workerNodeGroupMachineConfig); err != nil {
			return fmt.Errorf("error validating SSHUsername for worker node VSphereMachineConfig %v: %v", workerNodeGroupMachineConfig.Name, err)
		}
		if etcdMachineConfig != nil {
			if err = p.validateSSHUsername(etcdMachineConfig); err != nil {
				return fmt.Errorf("error validating SSHUsername for etcd VSphereMachineConfig %v: %v", etcdMachineConfig.Name, err)
			}
		}
	} else {
		return fmt.Errorf("error validating SSHUsername for control plane VSphereMachineConfig %v: %v", controlPlaneMachineConfig.Name, err)
	}

	for _, machineConfig := range p.machineConfigs {
		if machineConfig.Namespace != clusterSpec.Namespace {
			return errors.New("VSphereMachineConfig and Cluster objects must have the same namespace specified")
		}
	}
	if p.datacenterConfig.Namespace != clusterSpec.Namespace {
		return errors.New("VSphereDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	if controlPlaneMachineConfig.Spec.Template == "" {
		logger.V(1).Info("Control plane VSphereMachineConfig template is not set. Using default template.")
		err := p.setupDefaultTemplate(ctx, clusterSpec, controlPlaneMachineConfig, p.controlPlaneTemplateFactory)
		if err != nil {
			return err
		}
	}

	if err = p.validateTemplate(ctx, clusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane template validation failed.")
		return err
	}
	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		if workerNodeGroupMachineConfig.Spec.Template == "" {
			logger.V(1).Info("Worker VSphereMachineConfig template is not set. Using default template.")
			err := p.setupDefaultTemplate(ctx, clusterSpec, workerNodeGroupMachineConfig, p.workerTemplateFactory)
			if err != nil {
				return err
			}
		}
		if err = p.validateTemplate(ctx, clusterSpec, workerNodeGroupMachineConfig); err != nil {
			logger.V(1).Info("Workload template validation failed.")
			return err
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return errors.New("control plane and worker nodes must have the same template specified")
		}
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if etcdMachineConfig.Spec.Template == "" {
			logger.V(1).Info("Etcd VSphereMachineConfig template is not set. Using default template.")
			err := p.setupDefaultTemplate(ctx, clusterSpec, etcdMachineConfig, p.etcdTemplateFactory)
			if err != nil {
				return err
			}
		}
		if err = p.validateTemplate(ctx, clusterSpec, etcdMachineConfig); err != nil {
			logger.V(1).Info("Etcd template validation failed.")
			return err
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	templateHasSnapshot, err := p.providerGovcClient.TemplateHasSnapshot(ctx, controlPlaneMachineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("error getting template details: %v", err)
	}

	if !templateHasSnapshot {
		logger.Info("Warning: Your VM template has no snapshots. Defaulting to FullClone mode. VM provisioning might take longer.")

		if workerNodeGroupMachineConfig.Spec.DiskGiB < 20 {
			logger.Info("Warning: VSphereMachineConfig DiskGiB for worker nodes cannot be less than 20. Defaulting to 20.")
			workerNodeGroupMachineConfig.Spec.DiskGiB = 20
		}
		if controlPlaneMachineConfig.Spec.DiskGiB < 20 {
			logger.Info("Warning: VSphereDatacenterConfig DiskGiB for control plane cannot be less than 20. Defaulting to 20.")
			controlPlaneMachineConfig.Spec.DiskGiB = 20
		}
		if etcdMachineConfig != nil && etcdMachineConfig.Spec.DiskGiB < 20 {
			logger.Info("Warning: VSphereDatacenterConfig DiskGiB for etcd machines cannot be less than 20. Defaulting to 20.")
			etcdMachineConfig.Spec.DiskGiB = 20
		}
	} else if workerNodeGroupMachineConfig.Spec.DiskGiB != 25 || controlPlaneMachineConfig.Spec.DiskGiB != 25 || (etcdMachineConfig != nil && etcdMachineConfig.Spec.DiskGiB != 25) {
		logger.Info("Warning: Your VM template includes snapshot(s). LinkedClone mode will be used. DiskGiB cannot be customizable as disks cannot be expanded when using LinkedClone mode. Using default of 25 for DiskGiBs.")
		workerNodeGroupMachineConfig.Spec.DiskGiB = 25
		controlPlaneMachineConfig.Spec.DiskGiB = 25
		if etcdMachineConfig != nil {
			etcdMachineConfig.Spec.DiskGiB = 25
		}
	}

	return p.checkDatastoreUsage(ctx, clusterSpec, controlPlaneMachineConfig, workerNodeGroupMachineConfig, etcdMachineConfig)
}

type datastoreUsage struct {
	availableSpace float64
	needGiBSpace   int
}

func (p *vsphereProvider) checkDatastoreUsage(ctx context.Context, clusterSpec *cluster.Spec, controlPlaneMachineConfig *v1alpha1.VSphereMachineConfig, workerNodeGroupMachineConfig *v1alpha1.VSphereMachineConfig, etcdMachineConfig *v1alpha1.VSphereMachineConfig) error {
	usage := make(map[string]*datastoreUsage)
	var etcdAvailableSpace float64
	controlPlaneAvailableSpace, err := p.providerGovcClient.GetWorkloadAvailableSpace(ctx, controlPlaneMachineConfig)
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	workerAvailableSpace, err := p.providerGovcClient.GetWorkloadAvailableSpace(ctx, workerNodeGroupMachineConfig)
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	if etcdMachineConfig != nil {
		etcdAvailableSpace, err = p.providerGovcClient.GetWorkloadAvailableSpace(ctx, etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("error getting datastore details: %v", err)
		}
		if etcdAvailableSpace == -1 {
			logger.Info("Warning: Unable to get datastore available space for etcd machines. Using default of 25 for DiskGiBs.")
			etcdMachineConfig.Spec.DiskGiB = 25
		}
	}
	if controlPlaneAvailableSpace == -1 {
		logger.Info("Warning: Unable to get control plane datastore available space. Using default of 25 for DiskGiBs.")
		controlPlaneMachineConfig.Spec.DiskGiB = 25
	}
	if workerAvailableSpace == -1 {
		logger.Info("Warning: Unable to get worker node datastore available space. Using default of 25 for DiskGiBs.")
		workerNodeGroupMachineConfig.Spec.DiskGiB = 25
	}
	controlPlaneNeedGiB := controlPlaneMachineConfig.Spec.DiskGiB * clusterSpec.Spec.ControlPlaneConfiguration.Count
	usage[controlPlaneMachineConfig.Spec.Datastore] = &datastoreUsage{
		availableSpace: controlPlaneAvailableSpace,
		needGiBSpace:   controlPlaneNeedGiB,
	}
	workerNeedGiB := workerNodeGroupMachineConfig.Spec.DiskGiB * clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count
	_, ok := usage[workerNodeGroupMachineConfig.Spec.Datastore]
	if ok {
		usage[workerNodeGroupMachineConfig.Spec.Datastore].needGiBSpace += workerNeedGiB
	} else {
		usage[workerNodeGroupMachineConfig.Spec.Datastore] = &datastoreUsage{
			availableSpace: workerAvailableSpace,
			needGiBSpace:   workerNeedGiB,
		}
	}

	if etcdMachineConfig != nil {
		etcdNeedGiB := etcdMachineConfig.Spec.DiskGiB * clusterSpec.Spec.ExternalEtcdConfiguration.Count
		if _, ok := usage[etcdMachineConfig.Spec.Datastore]; ok {
			usage[etcdMachineConfig.Spec.Datastore].needGiBSpace += etcdNeedGiB
		} else {
			usage[etcdMachineConfig.Spec.Datastore] = &datastoreUsage{
				availableSpace: etcdAvailableSpace,
				needGiBSpace:   etcdNeedGiB,
			}
		}
	}

	for datastore, usage := range usage {
		if float64(usage.needGiBSpace) > usage.availableSpace {
			return fmt.Errorf("not enough space in datastore %v for given diskGiB and count for respective machine groups", datastore)
		}
	}
	return nil
}

func (p *vsphereProvider) validateTemplate(ctx context.Context, clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) error {
	if err := p.validateTemplatePresence(ctx, p.datacenterConfig.Spec.Datacenter, machineConfig); err != nil {
		return err
	}

	if err := p.validateTemplateTags(ctx, clusterSpec, machineConfig); err != nil {
		return err
	}

	return nil
}

func (p *vsphereProvider) validateTemplatePresence(ctx context.Context, datacenter string, machineConfig *v1alpha1.VSphereMachineConfig) error {
	templateFullPath, err := p.providerGovcClient.SearchTemplate(ctx, datacenter, machineConfig)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found. Has the template been imported?", machineConfig.Spec.Template)
	}

	machineConfig.Spec.Template = templateFullPath

	return nil
}

func (p *vsphereProvider) validateTemplateTags(ctx context.Context, clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) error {
	tags, err := p.providerGovcClient.GetTags(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("error validating template tags: %v", err)
	}

	tagsLookup := types.SliceToLookup(tags)
	for _, t := range requiredTemplateTags(clusterSpec, machineConfig) {
		if !tagsLookup.IsPresent(t) {
			// TODO: maybe add help text about to how to tag a template?
			return fmt.Errorf("template %s is missing tag %s", machineConfig.Spec.Template, t)
		}
	}

	return nil
}

func requiredTemplateTags(clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) []string {
	tagsByCategory := requiredTemplateTagsByCategory(clusterSpec, machineConfig)
	tags := make([]string, 0, len(tagsByCategory))
	for _, t := range tagsByCategory {
		tags = append(tags, t...)
	}

	return tags
}

func requiredTemplateTagsByCategory(clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) map[string][]string {
	osFamily := machineConfig.Spec.OSFamily
	return map[string][]string{
		"eksdRelease": {fmt.Sprintf("eksdRelease:%s", clusterSpec.VersionsBundle.EksD.Name)},
		"os":          {fmt.Sprintf("os:%s", strings.ToLower(string(osFamily)))},
	}
}

func (p *vsphereProvider) setupDefaultTemplate(ctx context.Context, clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig, templateFactory *templates.Factory) error {
	ovaURL := p.defaultTemplateForClusterSpec(clusterSpec, machineConfig)

	tags := requiredTemplateTagsByCategory(clusterSpec, machineConfig)

	if err := templateFactory.CreateIfMissing(ctx, p.datacenterConfig.Spec.Datacenter, machineConfig, ovaURL, tags); err != nil {
		return err
	}

	return nil
}

func (p *vsphereProvider) defaultTemplateForClusterSpec(clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) string {
	osFamily := machineConfig.Spec.OSFamily
	eksd := clusterSpec.VersionsBundle.EksD

	var ova releasev1alpha1.OvaArchive

	switch osFamily {
	case v1alpha1.Bottlerocket:
		ova = eksd.Ova.Bottlerocket
	case v1alpha1.Ubuntu:
		ova = eksd.Ova.Ubuntu
	}

	templateName := fmt.Sprintf("%s-%s-%s-%s-%s", osFamily, eksd.KubeVersion, eksd.Name, strings.Join(ova.Arch, "-"), ova.SHA256[:7])
	machineConfig.Spec.Template = filepath.Join("/", p.datacenterConfig.Spec.Datacenter, defaultTemplatesFolder, templateName)
	return ova.URI
}

func (p *vsphereProvider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
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

func (p *vsphereProvider) SetupAndValidateUpgradeCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
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
	return nil
}

func (p *vsphereProvider) UpdateSecrets(ctx context.Context, cluster *types.Cluster) error {
	var contents bytes.Buffer
	err := p.createSecret(cluster, &contents)
	if err != nil {
		return err
	}

	err = p.providerKubectlClient.ApplyKubeSpecFromBytes(ctx, cluster, contents.Bytes())
	if err != nil {
		return fmt.Errorf("error loading csi-vsphere-secret object: %v", err)
	}
	return nil
}

func (p *vsphereProvider) SetupAndValidateDeleteCluster(ctx context.Context) error {
	err := p.validateEnv(ctx)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	return nil
}

func NeedsNewControlPlaneTemplate(oldC, newC *v1alpha1.Cluster, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldC.Spec.KubernetesVersion != newC.Spec.KubernetesVersion {
		return true
	}
	if oldC.Spec.ControlPlaneConfiguration.Endpoint.Host != newC.Spec.ControlPlaneConfiguration.Endpoint.Host {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func NeedsNewWorkloadTemplate(oldC, newC *v1alpha1.Cluster, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldC.Spec.KubernetesVersion != newC.Spec.KubernetesVersion {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func NeedsNewEtcdTemplate(oldC, newC *v1alpha1.Cluster, oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldC.Spec.KubernetesVersion != newC.Spec.KubernetesVersion {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldVmc, newVmc)
}

func AnyImmutableFieldChanged(oldVdc, newVdc *v1alpha1.VSphereDatacenterConfig, oldVmc, newVmc *v1alpha1.VSphereMachineConfig) bool {
	if oldVmc.Spec.NumCPUs != newVmc.Spec.NumCPUs {
		return true
	}
	if oldVmc.Spec.MemoryMiB != newVmc.Spec.MemoryMiB {
		return true
	}
	if oldVmc.Spec.DiskGiB != newVmc.Spec.DiskGiB {
		return true
	}
	if oldVmc.Spec.Datastore != newVmc.Spec.Datastore {
		return true
	}
	if oldVmc.Spec.Folder != newVmc.Spec.Folder {
		return true
	}
	if oldVdc.Spec.Network != newVdc.Spec.Network {
		return true
	}
	if oldVmc.Spec.ResourcePool != newVmc.Spec.ResourcePool {
		return true
	}
	if oldVdc.Spec.Thumbprint != newVdc.Spec.Thumbprint {
		return true
	}
	if oldVmc.Spec.Template != newVmc.Spec.Template {
		return true
	}
	return false
}

func NewVsphereTemplateBuilder(datacenterSpec *v1alpha1.VSphereDatacenterConfigSpec, controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.VSphereMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &VsphereTemplateBuilder{
		datacenterSpec:             datacenterSpec,
		controlPlaneMachineSpec:    controlPlaneMachineSpec,
		workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
		etcdMachineSpec:            etcdMachineSpec,
		now:                        now,
	}
}

type VsphereTemplateBuilder struct {
	datacenterSpec             *v1alpha1.VSphereDatacenterConfigSpec
	controlPlaneMachineSpec    *v1alpha1.VSphereMachineConfigSpec
	workerNodeGroupMachineSpec *v1alpha1.VSphereMachineConfigSpec
	etcdMachineSpec            *v1alpha1.VSphereMachineConfigSpec
	now                        types.NowFunc
}

func (vs *VsphereTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (vs *VsphereTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *VsphereTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *VsphereTemplateBuilder) GenerateDeploymentFile(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	var etcdMachineSpec v1alpha1.VSphereMachineConfigSpec
	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineSpec = *vs.etcdMachineSpec
	}
	values, err := BuildTemplateMap(clusterSpec, *vs.datacenterSpec, *vs.controlPlaneMachineSpec, *vs.workerNodeGroupMachineSpec, etcdMachineSpec)
	if err != nil {
		return nil, err
	}

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultClusterConfig, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func BuildTemplateMap(clusterSpec *cluster.Spec, datacenterSpec v1alpha1.VSphereDatacenterConfigSpec, controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec v1alpha1.VSphereMachineConfigSpec) (map[string]interface{}, error) {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	/*
	 * These values are for testing only. This data should probably be pulled from a configuration file
	 * and I'm not sure how we'd handle secrets in a safe way for that.
	 */
	values := map[string]interface{}{
		"clusterName":                          clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointIp":               clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":                 strconv.Itoa(clusterSpec.Spec.ControlPlaneConfiguration.Count),
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
		"expClusterResourceSet":                "true",
		"thumbprint":                           datacenterSpec.Thumbprint,
		"vsphereDatacenter":                    datacenterSpec.Datacenter,
		"controlPlaneVsphereDatastore":         controlPlaneMachineSpec.Datastore,
		"controlPlaneVsphereFolder":            controlPlaneMachineSpec.Folder,
		"workerVsphereDatastore":               workerNodeGroupMachineSpec.Datastore,
		"workerVsphereFolder":                  workerNodeGroupMachineSpec.Folder,
		"managerImage":                         bundle.VSphere.Manager.VersionedImage(),
		"kubeVipImage":                         bundle.VSphere.KubeVip.VersionedImage(),
		"driverImage":                          bundle.VSphere.Driver.VersionedImage(),
		"syncerImage":                          bundle.VSphere.Syncer.VersionedImage(),
		"insecure":                             datacenterSpec.Insecure,
		"vsphereNetwork":                       datacenterSpec.Network,
		"controlPlaneVsphereResourcePool":      controlPlaneMachineSpec.ResourcePool,
		"workerVsphereResourcePool":            workerNodeGroupMachineSpec.ResourcePool,
		"vsphereServer":                        datacenterSpec.Server,
		"controlPlaneVsphereStoragePolicyName": controlPlaneMachineSpec.StoragePolicyName,
		"workerVsphereStoragePolicyName":       workerNodeGroupMachineSpec.StoragePolicyName,
		"vsphereTemplate":                      controlPlaneMachineSpec.Template,
		"workerReplicas":                       strconv.Itoa(clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count),
		"workloadVMsMemoryMiB":                 strconv.Itoa(workerNodeGroupMachineSpec.MemoryMiB),
		"workloadVMsNumCPUs":                   strconv.Itoa(workerNodeGroupMachineSpec.NumCPUs),
		"controlPlaneVMsMemoryMiB":             strconv.Itoa(controlPlaneMachineSpec.MemoryMiB),
		"controlPlaneVMsNumCPUs":               strconv.Itoa(controlPlaneMachineSpec.NumCPUs),
		"controlPlaneDiskGiB":                  strconv.Itoa(controlPlaneMachineSpec.DiskGiB),
		"workloadDiskGiB":                      strconv.Itoa(workerNodeGroupMachineSpec.DiskGiB),
		"controlPlaneSshUsername":              controlPlaneMachineSpec.Users[0].Name,
		"workerSshUsername":                    workerNodeGroupMachineSpec.Users[0].Name,
		"podCidrs":                             clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                         clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
		"apiserverExtraArgs":                   clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).ToPartialYaml(),
		"format":                               format,
		"externalEtcdVersion":                  bundle.KubeDistro.EtcdVersion,
		"etcdImage":                            bundle.KubeDistro.EtcdImage.VersionedImage(),
		"eksaSystemNamespace":                  constants.EksaSystemNamespace,
		"auditPolicy":                          common.GetAuditPolicy(),
	}

	k8sVersion, err := semver.New(bundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return nil, fmt.Errorf("error parsing kubernetes version %v: %v", bundle.KubeDistro.Kubernetes.Tag, err)
	}
	if k8sVersion.Major == 1 && k8sVersion.Minor >= 21 {
		values["cgroupDriverSystemd"] = true
	}

	if clusterSpec.Spec.ECRMirror != nil {
		values["ecrMirrorEndpoint"] = clusterSpec.Spec.ECRMirror.Endpoint
		if clusterSpec.Spec.ECRMirror.CACert != "" {
			cert, err := ioutil.ReadFile(clusterSpec.Spec.ECRMirror.CACert)
			if err != nil {
				return nil, fmt.Errorf("unable to read file %s: %v", clusterSpec.Spec.ECRMirror.CACert, err)
			}
			values["ecrMirrorCert"] = cert
		}
	}

	if clusterSpec.Spec.ProxyConfiguration != nil {
		values["proxyConfig"] = true
		var noProxyList []string
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Spec.ProxyConfiguration.NoProxy...)

		// Add no-proxy defaults
		noProxyList = append(noProxyList,
			"localhost",
			"127.0.0.1",
			datacenterSpec.Server,
			clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)

		values["httpProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpProxy
		values["httpsProxy"] = clusterSpec.Spec.ProxyConfiguration.HttpsProxy
		values["noProxy"] = noProxyList
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Spec.ExternalEtcdConfiguration.Count
		values["etcdVsphereDatastore"] = etcdMachineSpec.Datastore
		values["etcdVsphereFolder"] = etcdMachineSpec.Folder
		values["etcdDiskGiB"] = etcdMachineSpec.DiskGiB
		values["etcdVMsMemoryMiB"] = etcdMachineSpec.MemoryMiB
		values["etcdVMsNumCPUs"] = etcdMachineSpec.NumCPUs
		values["etcdVsphereResourcePool"] = etcdMachineSpec.ResourcePool
		values["etcdVsphereStoragePolicyName"] = etcdMachineSpec.StoragePolicyName
		values["etcdSshUsername"] = etcdMachineSpec.Users[0].Name
	}

	if strings.EqualFold(string(controlPlaneMachineSpec.OSFamily), string(v1alpha1.Bottlerocket)) {
		values["format"] = string(v1alpha1.Bottlerocket)
		values["pauseRepository"] = bundle.KubeDistro.Pause.Image()
		values["pauseVersion"] = bundle.KubeDistro.Pause.Tag()
		values["bottlerocketBootstrapRepository"] = bundle.BottleRocketBootstrap.Bootstrap.Image()
		values["bottlerocketBootstrapVersion"] = bundle.BottleRocketBootstrap.Bootstrap.Tag()
	}

	return values, nil
}

func (p *vsphereProvider) generateTemplateValuesForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) ([]byte, error) {
	clusterName := clusterSpec.ObjectMeta.Name
	var controlPlaneTemplateName, workloadTemplateName, etcdTemplateName string
	var needsNewEtcdTemplate bool

	c, err := p.providerKubectlClient.GetEksaCluster(ctx, workloadCluster)
	if err != nil {
		return nil, err
	}
	vdc, err := p.providerKubectlClient.GetEksaVSphereDatacenterConfig(ctx, p.datacenterConfig.Name, workloadCluster.KubeconfigFile, clusterSpec.Namespace)
	if err != nil {
		return nil, err
	}
	controlPlaneMachineConfig := p.machineConfigs[clusterSpec.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	controlPlaneVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, c.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, clusterSpec.Namespace)
	if err != nil {
		return nil, err
	}
	workerMachineConfig := p.machineConfigs[clusterSpec.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	workerVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, c.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, workloadCluster.KubeconfigFile, clusterSpec.Namespace)
	if err != nil {
		return nil, err
	}

	needsNewControlPlaneTemplate := NeedsNewControlPlaneTemplate(c, clusterSpec.Cluster, vdc, p.datacenterConfig, controlPlaneVmc, controlPlaneMachineConfig)
	if !needsNewControlPlaneTemplate {
		cp, err := p.providerKubectlClient.GetKubeadmControlPlane(ctx, workloadCluster, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, err
		}
		controlPlaneTemplateName = cp.Spec.InfrastructureTemplate.Name
	} else {
		controlPlaneTemplateName = p.templateBuilder.CPMachineTemplateName(clusterName)
	}

	needsNewWorkloadTemplate := NeedsNewWorkloadTemplate(c, clusterSpec.Cluster, vdc, p.datacenterConfig, workerVmc, workerMachineConfig)
	if !needsNewWorkloadTemplate {
		md, err := p.providerKubectlClient.GetMachineDeployment(ctx, workloadCluster, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return nil, err
		}
		workloadTemplateName = md.Spec.Template.Spec.InfrastructureRef.Name
	} else {
		workloadTemplateName = p.templateBuilder.WorkerMachineTemplateName(clusterName)
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := p.machineConfigs[clusterSpec.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdMachineVmc, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, workloadCluster.KubeconfigFile, clusterSpec.Namespace)
		if err != nil {
			return nil, err
		}
		needsNewEtcdTemplate = NeedsNewEtcdTemplate(c, clusterSpec.Cluster, vdc, p.datacenterConfig, etcdMachineVmc, etcdMachineConfig)
		if !needsNewEtcdTemplate {
			etcdadmCluster, err := p.providerKubectlClient.GetEtcdadmCluster(ctx, workloadCluster, executables.WithCluster(bootstrapCluster), executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, err
			}
			etcdTemplateName = etcdadmCluster.Spec.InfrastructureTemplate.Name
		} else {
			/* During a cluster upgrade, etcd machines need to be upgraded first, so that the etcd machines with new spec get created and can be used by controlplane machines
			as etcd endpoints. KCP rollout should not start until then. As a temporary solution in the absence of static etcd endpoints, we annotate the etcd cluster as "upgrading",
			so that KCP checks this annotation and does not proceed if etcd cluster is upgrading. The etcdadm controller removes this annotation once the etcd upgrade is complete.
			*/
			err = p.providerKubectlClient.UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", workloadCluster.Name),
				map[string]string{etcdv1alpha3.UpgradeInProgressAnnotation: "true"},
				executables.WithCluster(bootstrapCluster),
				executables.WithNamespace(constants.EksaSystemNamespace))
			if err != nil {
				return nil, err
			}
			etcdTemplateName = p.templateBuilder.EtcdMachineTemplateName(clusterName)
		}
	}

	valuesOpt := func(values map[string]interface{}) {
		values["needsNewControlPlaneTemplate"] = needsNewControlPlaneTemplate
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["needsNewWorkloadTemplate"] = needsNewWorkloadTemplate
		values["workloadTemplateName"] = workloadTemplateName
		values["vsphereControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["vsphereWorkerSshAuthorizedKey"] = p.workerSshAuthKey
		values["vsphereEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["vspherePassword"] = os.Getenv(vSpherePasswordKey)
		values["vsphereUsername"] = os.Getenv(vSphereUsernameKey)
		values["needsNewEtcdTemplate"] = needsNewEtcdTemplate
		values["etcdTemplateName"] = etcdTemplateName
	}
	return p.templateBuilder.GenerateDeploymentFile(clusterSpec, valuesOpt)
}

func (p *vsphereProvider) generateTemplateValuesForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) ([]byte, error) {
	clusterName := clusterSpec.ObjectMeta.Name

	valuesOpt := func(values map[string]interface{}) {
		values["needsNewControlPlaneTemplate"] = true
		values["needsNewWorkloadTemplate"] = true
		values["controlPlaneTemplateName"] = p.templateBuilder.CPMachineTemplateName(clusterName)
		values["workloadTemplateName"] = p.templateBuilder.WorkerMachineTemplateName(clusterName)
		values["vsphereControlPlaneSshAuthorizedKey"] = p.controlPlaneSshAuthKey
		values["vsphereWorkerSshAuthorizedKey"] = p.workerSshAuthKey
		values["vsphereEtcdSshAuthorizedKey"] = p.etcdSshAuthKey
		values["vspherePassword"] = os.Getenv(vSpherePasswordKey)
		values["vsphereUsername"] = os.Getenv(vSphereUsernameKey)
		values["needsNewEtcdTemplate"] = clusterSpec.Spec.ExternalEtcdConfiguration != nil
		values["etcdTemplateName"] = p.templateBuilder.EtcdMachineTemplateName(clusterName)
	}
	return p.templateBuilder.GenerateDeploymentFile(clusterSpec, valuesOpt)
}

func (p *vsphereProvider) generateDeploymentFile(ctx context.Context, fileName string, content []byte) (string, error) {
	t := templater.New(p.writer)
	writtenFile, err := t.WriteBytesToFile(content, fileName)
	if err != nil {
		return "", fmt.Errorf("error creating cluster config file: %v", err)
	}

	return writtenFile, nil
}

func (p *vsphereProvider) GenerateDeploymentFileForUpgrade(ctx context.Context, bootstrapCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec, fileName string) (string, error) {
	content, err := p.generateTemplateValuesForUpgrade(ctx, bootstrapCluster, workloadCluster, clusterSpec)
	if err != nil {
		return "", fmt.Errorf("error generating template values for cluster config file: %v", err)
	}
	return p.generateDeploymentFile(ctx, fileName, content)
}

func (p *vsphereProvider) GenerateDeploymentFileForCreate(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, fileName string) (string, error) {
	content, err := p.generateTemplateValuesForCreate(ctx, cluster, clusterSpec)
	if err != nil {
		return "", fmt.Errorf("error generating template values for cluster config file: %v", err)
	}
	return p.generateDeploymentFile(ctx, fileName, content)
}

func (p *vsphereProvider) GenerateStorageClass() []byte {
	return defaultStorageClass
}

func (p *vsphereProvider) GenerateMHC() ([]byte, error) {
	data := map[string]string{
		"clusterName": p.clusterConfig.Name,
	}
	mhc, err := templater.Execute(string(mhcTemplate), data)
	if err != nil {
		return nil, err
	}
	return mhc, nil
}

func (p *vsphereProvider) CleanupProviderInfrastructure(_ context.Context) error {
	return nil
}

func (p *vsphereProvider) createSecret(cluster *types.Cluster, contents *bytes.Buffer) error {
	t, err := template.New("tmpl").Parse(defaultSecretObject)
	if err != nil {
		return fmt.Errorf("error creating secret object template: %v", err)
	}

	thumbprint := p.datacenterConfig.Spec.Thumbprint
	if !p.selfSigned {
		thumbprint = ""
	}

	values := map[string]string{
		"clusterName":       cluster.Name,
		"insecure":          strconv.FormatBool(p.datacenterConfig.Spec.Insecure),
		"thumbprint":        thumbprint,
		"vspherePassword":   os.Getenv(vSpherePasswordKey),
		"vsphereUsername":   os.Getenv(vSphereUsernameKey),
		"vsphereServer":     p.datacenterConfig.Spec.Server,
		"vsphereDatacenter": p.datacenterConfig.Spec.Datacenter,
		"vsphereNetwork":    p.datacenterConfig.Spec.Network,
		"eksaLicense":       os.Getenv(eksaLicense),
	}
	err = t.Execute(contents, values)
	if err != nil {
		return fmt.Errorf("error substituting values for secret object template: %v", err)
	}
	return nil
}

func (p *vsphereProvider) BootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	var contents bytes.Buffer
	err := p.createSecret(cluster, &contents)
	if err != nil {
		return err
	}

	var loadContents bytes.Buffer
	loadContents.WriteString("data=")
	loadContents.WriteString(contents.String())
	err = p.providerKubectlClient.LoadSecret(ctx, loadContents.String(), secretObjectType, secretObjectName, cluster.KubeconfigFile)
	if err != nil {
		return fmt.Errorf("error loading csi-vsphere-secret object: %v", err)
	}
	return nil
}

func (p *vsphereProvider) Version(clusterSpec *cluster.Spec) string {
	return clusterSpec.VersionsBundle.VSphere.Version
}

func (p *vsphereProvider) EnvMap() (map[string]string, error) {
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

func (p *vsphereProvider) GetDeployments() map[string][]string {
	return map[string][]string{
		"capv-system":         {"capv-controller-manager"},
		"capi-webhook-system": {"capv-controller-manager"},
	}
}

func (p *vsphereProvider) GetInfrastructureBundle(clusterSpec *cluster.Spec) *types.InfrastructureBundle {
	bundle := clusterSpec.VersionsBundle
	folderName := fmt.Sprintf("infrastructure-vsphere/%s/", bundle.VSphere.Version)

	infraBundle := types.InfrastructureBundle{
		FolderName: folderName,
		Manifests: []releasev1alpha1.Manifest{
			bundle.VSphere.Components,
			bundle.VSphere.Metadata,
			bundle.VSphere.ClusterTemplate,
		},
	}
	return &infraBundle
}

func (p *vsphereProvider) DatacenterConfig() providers.DatacenterConfig {
	return p.datacenterConfig
}

func (p *vsphereProvider) MachineConfigs() []providers.MachineConfig {
	var configs []providers.MachineConfig
	controlPlaneMachineName := p.clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerMachineName := p.clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	p.machineConfigs[controlPlaneMachineName].Annotations = map[string]string{p.clusterConfig.ControlPlaneAnnotation(): "true"}
	configs = append(configs, p.machineConfigs[controlPlaneMachineName])
	if workerMachineName != controlPlaneMachineName {
		configs = append(configs, p.machineConfigs[workerMachineName])
	}
	if p.clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineName := p.clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		p.machineConfigs[etcdMachineName].Annotations = map[string]string{p.clusterConfig.EtcdAnnotation(): "true"}
		if etcdMachineName != controlPlaneMachineName && etcdMachineName != workerMachineName {
			configs = append(configs, p.machineConfigs[etcdMachineName])
		}
	}
	return configs
}

func (p *vsphereProvider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster)
	if err != nil {
		return err
	}

	prevDatacenter, err := p.providerKubectlClient.GetEksaVSphereDatacenterConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	datacenterConfig := p.DatacenterConfig()
	datacenter := datacenterConfig.(*v1alpha1.VSphereDatacenterConfig)

	oSpec := prevDatacenter.Spec
	nSpec := datacenter.Spec

	for _, machineConfig := range p.machineConfigs {
		err := p.validateMachineConfigImmutability(ctx, cluster, machineConfig, clusterSpec)
		if err != nil {
			return err
		}
	}

	if nSpec.Server != oSpec.Server {
		return fmt.Errorf("spec.server is immutable. Previous value %s, new value %s", oSpec.Server, nSpec.Server)
	}
	if nSpec.Datacenter != oSpec.Datacenter {
		return fmt.Errorf("spec.datacenter is immutable. Previous value %s, new value %s", oSpec.Datacenter, nSpec.Datacenter)
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

	secretChanged, err := p.secretContentsChanged(ctx, cluster)
	if err != nil {
		return err
	}

	if secretChanged {
		return fmt.Errorf("the VSphere credentials derived from %s and %s are immutable; please use the same credentials for the upgraded cluster", vSpherePasswordKey, vSphereUsernameKey)
	}
	return nil
}

func (p *vsphereProvider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.VSphereMachineConfig, clusterSpec *cluster.Spec) error {
	prevMachineConfig, err := p.providerKubectlClient.GetEksaVSphereMachineConfig(ctx, newConfig.Name, cluster.KubeconfigFile, clusterSpec.Namespace)
	if err != nil {
		return err
	}

	if newConfig.Spec.StoragePolicyName != prevMachineConfig.Spec.StoragePolicyName {
		return fmt.Errorf("spec.storagePolicyName is immutable. Previous value %s, new value %s", prevMachineConfig.Spec.StoragePolicyName, newConfig.Spec.StoragePolicyName)
	}

	if !reflect.DeepEqual(newConfig.Spec.Users, prevMachineConfig.Spec.Users) {
		return fmt.Errorf("vsphereMachineConfig %s users are immutable; new user: %v; old user: %v", newConfig.Name, newConfig.Spec.Users, prevMachineConfig.Spec.Users)
	}

	return nil
}

func (p *vsphereProvider) secretContentsChanged(ctx context.Context, workloadCluster *types.Cluster) (bool, error) {
	nPassword := os.Getenv(vSpherePasswordKey)
	oSecret, err := p.providerKubectlClient.GetSecret(ctx, credentialsObjectName, executables.WithCluster(workloadCluster), executables.WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return false, fmt.Errorf("error when obtaining VSphere secret %s from workload cluster: %v", credentialsObjectName, err)
	}

	if string(oSecret.Data["password"]) != nPassword {
		return true, nil
	}

	nUser := os.Getenv(vSphereUsernameKey)
	if string(oSecret.Data["username"]) != nUser {
		return true, nil
	}
	return false, nil
}
