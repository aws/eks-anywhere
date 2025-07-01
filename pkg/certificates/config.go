// Package certificates provides functionality for managing and renewing certificates in EKS Anywhere clusters.
package certificates

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// VerbosityLevel controls the detail level of logging output.
var VerbosityLevel int

// SSHConfig holds the SSH credential information.
type SSHConfig struct {
	User     string `yaml:"sshUser"`
	KeyPath  string `yaml:"sshKey"`
	Password string `yaml:"-"` // enviroment vairables
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

	// get password from env varibales
	if len(config.Etcd.Nodes) > 0 {
		if pass := os.Getenv("EKSA_SSH_KEY_PASSPHRASE_ETCD"); pass != "" {
			config.Etcd.SSH.Password = pass
		}
	}

	if pass := os.Getenv("EKSA_SSH_KEY_PASSPHRASE_CP"); pass != "" {
		config.ControlPlane.SSH.Password = pass
	}

	return config, nil
}

// ValidateConfig validates the certificate renewal configuration and ensures all required fields are present.
func ValidateConfig(config *RenewalConfig, component string) error {
	if config.ClusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	if config.OS == "" {
		return fmt.Errorf("os is required")
	}

	if config.OS != string(v1alpha1.Ubuntu) && config.OS != string(v1alpha1.RedHat) && config.OS != string(v1alpha1.Bottlerocket) {
		return fmt.Errorf("unsupported os %q", config.OS)
	}

	if err := validateNodeConfig(&config.ControlPlane); err != nil {
		return fmt.Errorf("validating control plane config: %w", err)
	}

	// Etcd nodes are only required if using external etcd.
	if len(config.Etcd.Nodes) > 0 {
		if err := validateNodeConfig(&config.Etcd); err != nil {
			return fmt.Errorf("validating etcd config: %w", err)
		}
	}

	if err := ValidateComponentWithConfig(component, config); err != nil {
		return err
	}

	return nil
}

func validateNodeConfig(config *NodeConfig) error {
	if len(config.Nodes) == 0 {
		return fmt.Errorf("nodes list cannot be empty")
	}

	if config.SSH.User == "" {
		return fmt.Errorf("sshUser is required")
	}
	if config.SSH.KeyPath == "" {
		return fmt.Errorf("sshKey is required")
	}

	if _, err := os.Stat(config.SSH.KeyPath); err != nil {
		return fmt.Errorf("validating sshKey path: %v", err)
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

// shouldProcessComponent checks if the specified component should be processed.
func shouldProcessComponent(requestedComponent, targetComponent string) bool {
	return requestedComponent == "" || requestedComponent == targetComponent
}
