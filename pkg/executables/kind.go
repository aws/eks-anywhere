package executables

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const kindPath = "kind"

//go:embed config/kind.yaml
var kindConfigTemplate string

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
	RegistryMirrorMap    map[string]string
	MirrorBase           string
	RegistryCACertPath   string
	RegistryAuth         bool
	RegistryUsername     string
	RegistryPassword     string
	ExtraPortMappings    []int
	DockerExtraMounts    bool
	DisableDefaultCNI    bool
}

func NewKind(executable Executable, writer filewriter.FileWriter) *Kind {
	return &Kind{
		writer:     writer,
		Executable: executable,
	}
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
		k.execConfig.MirrorBase = registryMirror.BaseRegistry
		k.execConfig.RegistryMirrorMap = containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap)
		if registryMirror.CACertContent != "" {
			path := filepath.Join(clusterSpec.Cluster.Name, "generated", "certs.d", registryMirror.BaseRegistry)
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(path, "ca.crt"), []byte(registryMirror.CACertContent), 0o644); err != nil {
				return errors.New("error writing the registry certification file")
			}
			k.execConfig.RegistryCACertPath = filepath.Join(clusterSpec.Cluster.Name, "generated", "certs.d")
		}
		if registryMirror.Auth {
			k.execConfig.RegistryAuth = registryMirror.Auth
			username, password, err := config.ReadCredentials()
			if err != nil {
				return err
			}
			k.execConfig.RegistryUsername = username
			k.execConfig.RegistryPassword = password
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
