// Package certificates provides functionality for managing and renewing certificates in EKS Anywhere clusters.
package certificates

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

// VerbosityLevel controls the detail level of logging output.
var VerbosityLevel int

// SSHConfig holds the SSH credential information.
type SSHConfig struct {
	User      string `yaml:"sshUser"`
	KeyPath   string `yaml:"sshKey"`
	Password  string `yaml:"sshPasswd,omitempty"` // Optional SSH key passphrase.
	component string
}

// NodeConfig holds configuration for a group of nodes.
type NodeConfig struct {
	Nodes []string  `yaml:"nodes"`
	SSH   SSHConfig `yaml:"ssh"`
}

// RenewalConfig defines the configuration for certificate renewal operations.
type RenewalConfig struct {
	ClusterName  string     `yaml:"clusterName"`
	OS           string     `yaml:"os"`
	ControlPlane NodeConfig `yaml:"controlPlane"`
	Etcd         NodeConfig `yaml:"etcd"`
}

// ParseConfig reads and parses a certificate renewal configuration file.
func ParseConfig(path string) (*RenewalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %v", err)
	}

	config := &RenewalConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing config file: %v", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("validating config: %v", err)
	}

	return config, nil
}

func validateConfig(config *RenewalConfig) error {
	if config.ClusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	if config.OS == "" {
		return fmt.Errorf("os is required")
	}

	if config.OS != string(v1alpha1.Ubuntu) && config.OS != string(v1alpha1.RedHat) && config.OS != string(v1alpha1.Bottlerocket) {
		return fmt.Errorf("unsupported os %q", config.OS)
	}

	if err := validateNodeConfig(&config.ControlPlane, "controlPlane"); err != nil {
		return err
	}

	// Etcd nodes are optional (for stacked etcd).
	if len(config.Etcd.Nodes) > 0 {
		if err := validateNodeConfig(&config.Etcd, "etcd"); err != nil {
			return err
		}
	}

	return nil
}

func validateNodeConfig(config *NodeConfig, componentName string) error {
	if len(config.Nodes) == 0 {
		return fmt.Errorf("%s: nodes list cannot be empty", componentName)
	}

	if config.SSH.User == "" {
		return fmt.Errorf("%s: sshUser is required", componentName)
	}
	if config.SSH.KeyPath == "" {
		return fmt.Errorf("%s: sshKey is required", componentName)
	}

	if _, err := os.Stat(config.SSH.KeyPath); err != nil {
		return fmt.Errorf("%s: validating sshKey path: %v", componentName, err)
	}

	return nil
}

// ValidateComponentWithConfig validates that the specified component is compatible with the configuration.
func ValidateComponentWithConfig(component string, config *RenewalConfig) error {
	if component == "" {
		return nil
	}

	if component != constants.EtcdComponent && component != constants.ControlPlaneComponent {
		return fmt.Errorf("invalid component %q, must be either %q or %q", component, constants.EtcdComponent, constants.ControlPlaneComponent)
	}

	if component == constants.EtcdComponent && len(config.Etcd.Nodes) == 0 {
		return fmt.Errorf("no external etcd nodes defined; cannot use --component %s", constants.EtcdComponent)
	}

	return nil
}

// ShouldProcessComponent checks if the specified component should be processed.
func ShouldProcessComponent(requestedComponent, targetComponent string) bool {
	return requestedComponent == "" || requestedComponent == targetComponent
}

// ValidateNodesPresence ensures that the slice of node ip is not empty.
func ValidateNodesPresence(nodes []string, componentName string) error {
	if len(nodes) == 0 {
		return fmt.Errorf("%s: nodes list cannot be empty", componentName)
	}
	return nil
}

// GetSSHKeyDirs returns a list of directories containing SSH keys used in container.
func GetSSHKeyDirs(cfg *RenewalConfig) []string {
	set := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		dir := filepath.Dir(p)
		set[dir] = struct{}{}
	}
	add(cfg.ControlPlane.SSH.KeyPath)
	add(cfg.Etcd.SSH.KeyPath)
	dirs := make([]string, 0, len(set))
	for d := range set {
		dirs = append(dirs, d)
	}
	return dirs
}

// PreloadAllSSHKeys initializes SSH configurations for all keys in the renewal config.
func PreloadAllSSHKeys(r SSHRunner, cfg *RenewalConfig) error {
	seen := map[string]struct{}{}
	load := func(sc SSHConfig, component string) error {
		if _, ok := seen[sc.KeyPath]; ok {
			return nil
		}
		sc.component = component
		if err := r.InitSSHConfig(sc); err != nil {
			return err
		}
		seen[sc.KeyPath] = struct{}{}
		return nil
	}
	if err := load(cfg.ControlPlane.SSH, "CP"); err != nil {
		return fmt.Errorf("loading control-plane key: %w", err)
	}
	if len(cfg.Etcd.Nodes) > 0 {
		if err := load(cfg.Etcd.SSH, "ETCD"); err != nil {
			return fmt.Errorf("loading etcd key: %w", err)
		}
	}
	return nil
}

// ResolveKubeconfigPath attempts to find a valid kubeconfig file for the given cluster name.
func ResolveKubeconfigPath(clusterName string) (string, error) {
	var kubeconfigPath string

	if clusterName != "" {
		pwd, err := os.Getwd()
		if err == nil {
			possiblePath := filepath.Join(pwd, clusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName))
			logger.Info("Trying kubeconfig path based on clusterName", "possiblePath", possiblePath)
			if _, err := os.Stat(possiblePath); err == nil {
				kubeconfigPath = possiblePath
				logger.Info("Using kubeconfig from cluster directory", "path", kubeconfigPath)
				return kubeconfigPath, nil
			}
		}
	}

	envPath := os.Getenv("KUBECONFIG")
	if envPath != "" {
		kubeconfigPath = envPath
		logger.Info("Using kubeconfig from environment variable", "path", kubeconfigPath)
	}

	if kubeconfigPath == "" {
		return "", fmt.Errorf("could not find kubeconfig for cluster %s and KUBECONFIG environment variable is not set. "+
			"Try setting KUBECONFIG environment variable: export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig",
			clusterName)
	}

	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return "", fmt.Errorf("kubeconfig file does not exist: %s", kubeconfigPath)
	}

	logger.Info("Using kubeconfig from cluster directory", "path", kubeconfigPath)
	return kubeconfigPath, nil
}
