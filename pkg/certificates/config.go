// Package certificates provides functionality for managing and renewing certificates in EKS Anywhere clusters.
package certificates

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// NodeConfig holds SSH configuration for a node group.
type NodeConfig struct {
	Nodes     []string `yaml:"nodes"`
	OS        string   `yaml:"os"`
	SSHKey    string   `yaml:"sshKey"`
	SSHUser   string   `yaml:"sshUser"`
	SSHPasswd string   `yaml:"sshPasswd,omitempty"` // Optional SSH key passphrase.
}

// RenewalConfig defines the configuration for certificate renewal operations.
type RenewalConfig struct {
	ClusterName  string     `yaml:"clusterName"`
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
		return fmt.Errorf("cluster name is required")
	}

	if len(config.ControlPlane.Nodes) == 0 {
		return fmt.Errorf("at least one node is required in ControlPlane configuration")
	}

	if err := validateNodeConfig(&config.ControlPlane); err != nil {
		return fmt.Errorf("validating control plane: %v", err)
	}

	// Etcd nodes are optional (could be embedded in control plane).
	if len(config.Etcd.Nodes) > 0 {
		if err := validateNodeConfig(&config.Etcd); err != nil {
			return fmt.Errorf("validating etcd: %v", err)
		}
	}

	return nil
}

func validateNodeConfig(config *NodeConfig) error {
	if len(config.Nodes) == 0 {
		return fmt.Errorf("nodes are required")
	}
	if config.OS == "" {
		return fmt.Errorf("OS is required")
	}
	if config.OS != string(v1alpha1.Ubuntu) && config.OS != string(v1alpha1.RedHat) && config.OS != string(v1alpha1.Bottlerocket) {
		return fmt.Errorf("unsupported OS %q", config.OS)
	}
	if config.SSHKey == "" {
		return fmt.Errorf("SSH key is required")
	}
	if config.SSHUser == "" {
		return fmt.Errorf("SSH user is required")
	}

	if _, err := os.Stat(config.SSHKey); err != nil {
		return fmt.Errorf("retrieving SSH key file information: %v", err)
	}

	return nil
}
