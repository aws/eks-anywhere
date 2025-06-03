package certificates

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// NodeConfig holds SSH configuration for a node group
type NodeConfig struct {
	Nodes     []string `yaml:"nodes"`
	OS        string   `yaml:"os"`
	SSHKey    string   `yaml:"sshKey"`
	SSHUser   string   `yaml:"sshUser"`
	SSHPasswd string   `yaml:"sshPasswd,omitempty"` // Optional SSH key passphrase
}

type RenewalConfig struct {
	ClusterName  string     `yaml:"clusterName"`
	ControlPlane NodeConfig `yaml:"controlPlane"`
	Etcd         NodeConfig `yaml:"etcd"`
}

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
		return fmt.Errorf("at least one control plane node is required")
	}

	if err := validateNodeConfig(&config.ControlPlane, "control plane"); err != nil {
		return err
	}

	// Etcd nodes are optional (could be embedded in control plane)
	if len(config.Etcd.Nodes) > 0 {
		if err := validateNodeConfig(&config.Etcd, "etcd"); err != nil {
			return err
		}
	}

	return nil
}

func validateNodeConfig(config *NodeConfig, component string) error {
	if len(config.Nodes) == 0 {
		return fmt.Errorf("%s nodes are required", component)
	}
	if config.OS == "" {
		return fmt.Errorf("%s OS is required", component)
	}
	if config.OS != "ubuntu" && config.OS != "rhel" && config.OS != "bottlerocket" {
		return fmt.Errorf("unsupported OS %q for %s", config.OS, component)
	}
	if config.SSHKey == "" {
		return fmt.Errorf("%s SSH key is required", component)
	}
	if config.SSHUser == "" {
		return fmt.Errorf("%s SSH user is required", component)
	}

	if _, err := os.Stat(config.SSHKey); err != nil {
		return fmt.Errorf("SSH key file for %s: %v", component, err)
	}

	return nil
}
