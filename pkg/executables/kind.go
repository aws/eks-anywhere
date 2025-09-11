package executables

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const kindPath = "kind"

//go:embed config/kind.yaml
var kindConfigTemplate string

//go:embed config/hosts.toml
var hostsTomlTemplate string

const configFileName = "kind_tmp.yaml"

type Kind struct {
	writer filewriter.FileWriter
	Executable
	execConfig *kindExecConfig
}

// kindExecConfig contains transient information for the execution of kind commands
// It's used by BootstrapClusterClientOption's to store/change information prior to a command execution
// It must be cleaned after each execution to prevent side effects from past executions options.
type kindExecConfig struct {
	env                  map[string]string
	ConfigFile           string
	KindImage            string
	KubernetesRepository string
	EtcdRepository       string
	EtcdVersion          string
	CorednsRepository    string
	CorednsVersion       string
	KubernetesVersion    string
	RegistryConfigDir    string
	ExtraPortMappings    []int
	DockerExtraMounts    bool
	DisableDefaultCNI    bool
	PodSubnet            string
	ServiceSubnet        string
	AuditPolicyPath      string
}

func NewKind(executable Executable, writer filewriter.FileWriter) *Kind {
	return &Kind{
		writer:     writer,
		Executable: executable,
	}
}

// CreateAuditPolicy creates an audit policy file to be used by the bootstrap cluster's api server.
func (k *Kind) CreateAuditPolicy(clusterSpec *cluster.Spec) error {
	auditPolicy, err := common.GetAuditPolicy(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return err
	}
	auditPath := filepath.Join(clusterSpec.Cluster.Name, "generated", "kubernetes")
	if err := os.MkdirAll(auditPath, os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(auditPath, "audit-policy.yaml"), []byte(auditPolicy), 0o644); err != nil {
		return fmt.Errorf("error writing the audit policy file: %w", err)
	}
	auditPath = filepath.Join(auditPath, "audit-policy.yaml")
	k.execConfig.AuditPolicyPath = auditPath
	return nil
}

func (k *Kind) CreateBootstrapCluster(ctx context.Context, clusterSpec *cluster.Spec, opts ...bootstrapper.BootstrapClusterClientOption) (kubeconfig string, err error) {
	err = k.setupExecConfig(clusterSpec)
	if err != nil {
		return "", err
	}
	defer k.cleanExecConfig()

	err = processOpts(opts)
	if err != nil {
		return "", err
	}

	serviceCidrs := clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks
	podCidrs := clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks

	if len(serviceCidrs) != 0 {
		k.execConfig.ServiceSubnet = serviceCidrs[0]
	}

	if len(podCidrs) != 0 {
		k.execConfig.PodSubnet = podCidrs[0]
	}

	err = k.buildConfigFile()
	if err != nil {
		return "", err
	}

	kubeconfigName, err := k.createKubeConfig(clusterSpec.Cluster.Name, []byte(""))
	if err != nil {
		return "", err
	}
	executionArgs := k.execArguments(clusterSpec.Cluster.Name, kubeconfigName)

	logger.V(4).Info("Creating kind cluster", "name", getInternalName(clusterSpec.Cluster.Name), "kubeconfig", kubeconfigName)
	_, err = k.ExecuteWithEnv(ctx, k.execConfig.env, executionArgs...)
	if err != nil {
		return "", fmt.Errorf("executing create cluster: %v", err)
	}

	return kubeconfigName, nil
}

func (k *Kind) ClusterExists(ctx context.Context, clusterName string) (bool, error) {
	internalName := getInternalName(clusterName)
	stdOut, err := k.Execute(ctx, "get", "clusters")
	if err != nil {
		return false, fmt.Errorf("executing get clusters: %v", err)
	}

	logger.V(5).Info("Executed kind get clusters", "response", stdOut.String())

	scanner := bufio.NewScanner(&stdOut)
	for scanner.Scan() {
		if kindClusterName := scanner.Text(); kindClusterName == internalName {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed checking if cluster exists when reading kind get cluster response: %v", err)
	}

	return false, nil
}

func (k *Kind) GetKubeconfig(ctx context.Context, clusterName string) (string, error) {
	internalName := getInternalName(clusterName)
	stdOut, err := k.Execute(ctx, "get", "kubeconfig", "--name", internalName)
	if err != nil {
		return "", fmt.Errorf("executing get kubeconfig: %v", err)
	}
	return k.createKubeConfig(clusterName, stdOut.Bytes())
}

func (k *Kind) WithExtraDockerMounts() bootstrapper.BootstrapClusterClientOption {
	return func() error {
		if k.execConfig == nil {
			return errors.New("kind exec config is not ready")
		}

		k.execConfig.DockerExtraMounts = true
		return nil
	}
}

func (k *Kind) WithExtraPortMappings(ports []int) bootstrapper.BootstrapClusterClientOption {
	return func() error {
		if k.execConfig == nil {
			return errors.New("kind exec config is not ready")
		}

		if len(ports) == 0 {
			return errors.New("no ports found in the list")
		}

		k.execConfig.ExtraPortMappings = ports

		return nil
	}
}

func (k *Kind) WithEnv(env map[string]string) bootstrapper.BootstrapClusterClientOption {
	return func() error {
		if k.execConfig == nil {
			return errors.New("kind exec config is not ready")
		}

		for name, value := range env {
			k.execConfig.env[name] = value
		}

		return nil
	}
}

func (k *Kind) DeleteBootstrapCluster(ctx context.Context, cluster *types.Cluster) error {
	internalName := getInternalName(cluster.Name)
	logger.V(4).Info("Deleting kind cluster", "name", internalName)
	_, err := k.Execute(ctx, "delete", "cluster", "--name", internalName)
	if err != nil {
		return fmt.Errorf("executing delete cluster: %v", err)
	}
	return err
}

// RegistryConfig contains configuration for setting up a registry with containerd.
type RegistryConfig struct {
	Server     string // The server URL for the registry
	Host       string // The host URL to redirect to
	CACertPath string // CA certificate path
	AuthHeader string
	OutputDir  string // Directory where to write hosts.toml
}

// setupRegistryConfig creates a registry configuration hosts.toml file.
func setupRegistryConfig(config RegistryConfig) error {
	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory %s: %w", config.OutputDir, err)
	}

	// Generate hosts.toml content using template
	content, err := templater.Execute(hostsTomlTemplate, config)
	if err != nil {
		return fmt.Errorf("executing hosts.toml template: %w", err)
	}

	// Write hosts.toml file
	hostsPath := filepath.Join(config.OutputDir, "hosts.toml")
	if err := os.WriteFile(hostsPath, content, 0o644); err != nil {
		return fmt.Errorf("writing hosts.toml: %w", err)
	}

	return nil
}

func (k *Kind) setupExecConfig(clusterSpec *cluster.Spec) error {
	versionsBundle := clusterSpec.RootVersionsBundle()
	registryMirror := registrymirror.FromCluster(clusterSpec.Cluster)
	k.execConfig = &kindExecConfig{
		KindImage:            registryMirror.ReplaceRegistry(versionsBundle.EksD.KindNode.VersionedImage()),
		KubernetesRepository: registryMirror.ReplaceRegistry(versionsBundle.KubeDistro.Kubernetes.Repository),
		KubernetesVersion:    versionsBundle.KubeDistro.Kubernetes.Tag,
		EtcdRepository:       registryMirror.ReplaceRegistry(versionsBundle.KubeDistro.Etcd.Repository),
		EtcdVersion:          versionsBundle.KubeDistro.Etcd.Tag,
		CorednsRepository:    registryMirror.ReplaceRegistry(versionsBundle.KubeDistro.CoreDNS.Repository),
		CorednsVersion:       versionsBundle.KubeDistro.CoreDNS.Tag,
		env:                  make(map[string]string),
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		if err := k.setupRegistryMirror(clusterSpec, registryMirror); err != nil {
			return err
		}
	}

	if err := k.CreateAuditPolicy(clusterSpec); err != nil {
		return err
	}
	return nil
}

// setupRegistryMirror handles the complete setup of registry mirror configuration.
func (k *Kind) setupRegistryMirror(clusterSpec *cluster.Spec, registryMirror *registrymirror.RegistryMirror) error {
	mirrorBase := registryMirror.BaseRegistry
	registryMirrorMap := containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)

	// Create the base certs.d directory
	certsBasePath := filepath.Join(clusterSpec.Cluster.Name, "generated", "certs.d")
	k.execConfig.RegistryConfigDir = certsBasePath

	// Generate authorization header if authentication is required
	var authHeader string
	if registryMirror.Auth {
		username, password, err := config.ReadCredentials()
		if err != nil {
			return err
		}
		if username != "" {
			credentials := fmt.Sprintf("%s:%s", username, password)
			encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
			authHeader = fmt.Sprintf("Basic %s", encoded)
		}
	}

	// Create CA certificate only once for the base registry and determine shared CA path
	var mountedCACertPath string
	if registryMirror.CACertContent != "" {
		// Create the base registry directory and CA certificate file
		mirrorBaseDir := filepath.Join(certsBasePath, mirrorBase)
		if err := os.MkdirAll(mirrorBaseDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating directory %s: %w", mirrorBaseDir, err)
		}

		// Write the CA certificate file
		caPath := filepath.Join(mirrorBaseDir, "ca.crt")
		if err := os.WriteFile(caPath, []byte(registryMirror.CACertContent), 0o644); err != nil {
			return fmt.Errorf("writing CA certificate: %w", err)
		}

		mountedCACertPath = fmt.Sprintf("/etc/containerd/certs.d/%s/ca.crt", mirrorBase)
	}

	// Setup configuration for the mirror registry
	err := setupRegistryConfig(RegistryConfig{
		Server:     mirrorBase,
		Host:       mirrorBase,
		CACertPath: mountedCACertPath,
		AuthHeader: authHeader,
		OutputDir:  filepath.Join(certsBasePath, mirrorBase),
	})
	if err != nil {
		return fmt.Errorf("setting up configuration for local registry %s: %w", mirrorBase, err)
	}

	// Setup configuration for each original registry that should be mirrored
	for originalRegistry, mirrorEndpoint := range registryMirrorMap {
		err := setupRegistryConfig(RegistryConfig{
			Server:     originalRegistry,
			Host:       mirrorEndpoint,
			CACertPath: mountedCACertPath,
			AuthHeader: authHeader,
			OutputDir:  filepath.Join(certsBasePath, originalRegistry),
		})
		if err != nil {
			return fmt.Errorf("setting up mirror configuration for registry %s: %w", originalRegistry, err)
		}
	}

	return nil
}

func (k *Kind) cleanExecConfig() {
	k.execConfig = nil
}

func (k *Kind) buildConfigFile() error {
	t := templater.New(k.writer)
	writtenFileName, err := t.WriteToFile(kindConfigTemplate, k.execConfig, configFileName)
	if err != nil {
		return fmt.Errorf("creating file for kind config: %v", err)
	}

	k.execConfig.ConfigFile = writtenFileName

	return nil
}

func (k *Kind) execArguments(clusterName string, kubeconfigName string) []string {
	return []string{
		"create", "cluster",
		"--name", getInternalName(clusterName),
		"--kubeconfig", kubeconfigName,
		"--image", k.execConfig.KindImage,
		"--config", k.execConfig.ConfigFile,
	}
}

func (k *Kind) createKubeConfig(clusterName string, content []byte) (string, error) {
	fileName, err := k.writer.Write(fmt.Sprintf("%s.kind.kubeconfig", clusterName), content)
	if err != nil {
		return "", fmt.Errorf("generating temp file for storing kind kubeconfig: %v", err)
	}
	return fileName, nil
}

func processOpts(opts []bootstrapper.BootstrapClusterClientOption) error {
	for _, opt := range opts {
		err := opt()
		if err != nil {
			return err
		}
	}

	return nil
}

func getInternalName(clusterName string) string {
	return fmt.Sprintf("%s-eks-a-cluster", clusterName)
}
