package bootstrap

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

const kindPath = "kind"

//go:embed kind.yaml
var kindConfigTemplate string

const configFileName = "kind_tmp.yaml"

type CreateOption func(cfg *kindExecConfig)

type Kind struct {
	executables.Executable
	execConfig *kindExecConfig
	spec       *cluster.Spec
}

func NewKind(executable executables.Executable, spec *cluster.Spec) *Kind {
	return &Kind{
		Executable: executable,
		spec:       spec,
	}
}

func (k *Kind) Create(ctx context.Context, opts ...bootstrapper.BootstrapClusterClientOption) error {
	err := k.setupExecConfig(k.spec)
	if err != nil {
		return err
	}
	defer k.cleanExecConfig()

	err = processOpts(opts)
	if err != nil {
		return err
	}

	err = k.buildConfigFile()
	if err != nil {
		return err
	}

	kubeconfigName, err := k.createKubeConfig(k.spec.Cluster.Name, []byte(""))
	if err != nil {
		return err
	}
	executionArgs := k.execArguments(k.spec.Cluster.Name, kubeconfigName)

	logger.V(4).Info("Creating kind cluster", "name", toKindClusterName(k.spec.Cluster.Name), "kubeconfig", kubeconfigName)
	_, err = k.ExecuteWithEnv(ctx, k.execConfig.env, executionArgs...)
	if err != nil {
		return fmt.Errorf("executing create cluster: %v", err)
	}

	return nil
}

func (k *Kind) Delete(ctx context.Context) error {
	internalName := toKindClusterName(k.spec.Cluster.Name)
	logger.V(4).Info("Deleting kind cluster", "name", internalName)
	_, err := k.Execute(ctx, "delete", "cluster", "--name", internalName)
	if err != nil {
		return fmt.Errorf("executing delete cluster: %v", err)
	}
	return err
}

func (k *Kind) WriteKubeconfig(ctx context.Context, w io.Writer) error {
	stdOut, err := k.Execute(ctx, "get", "kubeconfig", "--name", toKindClusterName(k.spec.Cluster.Name))
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, &stdOut); err != nil {
		return err
	}

	return nil
}

func (k *Kind) setupExecConfig(clusterSpec *cluster.Spec) error {
	bundle := clusterSpec.VersionsBundle
	k.execConfig = &kindExecConfig{
		KindImage:            urls.ReplaceHost(bundle.EksD.KindNode.VersionedImage(), clusterSpec.Cluster.RegistryMirror()),
		KubernetesRepository: bundle.KubeDistro.Kubernetes.Repository,
		KubernetesVersion:    bundle.KubeDistro.Kubernetes.Tag,
		EtcdRepository:       bundle.KubeDistro.Etcd.Repository,
		EtcdVersion:          bundle.KubeDistro.Etcd.Tag,
		CorednsRepository:    bundle.KubeDistro.CoreDNS.Repository,
		CorednsVersion:       bundle.KubeDistro.CoreDNS.Tag,
		env:                  make(map[string]string),
	}
	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		k.execConfig.RegistryMirrorEndpoint = net.JoinHostPort(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port)
		if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent != "" {
			path := filepath.Join(clusterSpec.Cluster.Name, "generated", "certs.d", k.execConfig.RegistryMirrorEndpoint)
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(path, "ca.crt"), []byte(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent), 0o644); err != nil {
				return errors.New("error writing the registry certification file")
			}
			k.execConfig.RegistryCACertPath = filepath.Join(clusterSpec.Cluster.Name, "generated", "certs.d")
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
		"--name", toKindClusterName(clusterName),
		"--kubeconfig", kubeconfigName,
		"--image", k.execConfig.KindImage,
		"--config", k.execConfig.ConfigFile,
	}
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

func toKindClusterName(clusterName string) string {
	return fmt.Sprintf("%s-eks-a-cluster", clusterName)
}

// kindExecConfig contains transient information for the execution of kind commands
// It's used by BootstrapClusterClientOption's to store/change information prior to a command execution
// It must be cleaned after each execution to prevent side effects from past executions options.
type kindExecConfig struct {
	env                    map[string]string
	ConfigFile             string
	KindImage              string
	KubernetesRepository   string
	EtcdRepository         string
	EtcdVersion            string
	CorednsRepository      string
	CorednsVersion         string
	KubernetesVersion      string
	RegistryMirrorEndpoint string
	RegistryCACertPath     string
	ExtraPortMappings      []int
	DockerExtraMounts      bool
	DisableDefaultCNI      bool
}

func newKindExecConfig(spec *cluster.Spec) kindExecConfig {
	bundle := spec.VersionsBundle

	cfg := kindExecConfig{
		KindImage:            urls.ReplaceHost(bundle.EksD.KindNode.VersionedImage(), spec.Cluster.RegistryMirror()),
		KubernetesRepository: bundle.KubeDistro.Kubernetes.Repository,
		KubernetesVersion:    bundle.KubeDistro.Kubernetes.Tag,
		EtcdRepository:       bundle.KubeDistro.Etcd.Repository,
		EtcdVersion:          bundle.KubeDistro.Etcd.Tag,
		CorednsRepository:    bundle.KubeDistro.CoreDNS.Repository,
		CorednsVersion:       bundle.KubeDistro.CoreDNS.Tag,
		env:                  make(map[string]string),
	}
	if spec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		cfg.RegistryMirrorEndpoint = net.JoinHostPort(spec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint, spec.Cluster.Spec.RegistryMirrorConfiguration.Port)
		if spec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent != "" {
			path := filepath.Join(spec.Cluster.Name, "generated", "certs.d", cfg.RegistryMirrorEndpoint)
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(path, "ca.crt"), []byte(spec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent), 0o644); err != nil {
				return errors.New("error writing the registry certification file")
			}
			cfg.RegistryCACertPath = filepath.Join(spec.Cluster.Name, "generated", "certs.d")
		}
	}
	return nil
}
