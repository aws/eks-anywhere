// Package certificates provides functionality for managing and renewing certificates in EKS Anywhere clusters.
package certificates

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// Constants for certificate paths and components.
const (
	tempLocalEtcdCertsDir = "etcd-client-certs"

	// Ubuntu/RHEL paths.
	ubuntuEtcdCertDir           = "/etc/etcd"
	ubuntuControlPlaneCertDir   = "/etc/kubernetes/pki"
	ubuntuControlPlaneManifests = "/etc/kubernetes/manifests"

	// Bottlerocket paths.
	bottlerocketEtcdCertDir         = "/var/lib/etcd"
	bottlerocketControlPlaneCertDir = "/var/lib/kubeadm/pki"
	bottlerocketTmpDir              = "/run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/tmp"
)

// VerbosityLevel controls the detail level of logging output during certificate operations.
var VerbosityLevel int

// NodeConfig holds SSH configuration for a node group.
type NodeConfig struct {
	Nodes []string `yaml:"nodes"`
	// OS        string   `yaml:"os"`
	SSHKey    string `yaml:"sshKey"`
	SSHUser   string `yaml:"sshUser"`
	SSHPasswd string `yaml:"sshPasswd,omitempty"` // Optional SSH key passphrase.
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
		return fmt.Errorf("cluster name is required")
	}

	if config.OS == "" {
		return fmt.Errorf("OS is required")
	}

	if config.OS != string(v1alpha1.Ubuntu) && config.OS != string(v1alpha1.RedHat) && config.OS != string(v1alpha1.Bottlerocket) {
		return fmt.Errorf("unsupported OS %q", config.OS)
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

// ValidateComponent checks if the specified component is valid.
func ValidateComponent(component string) error {
	if component != "" && component != constants.EtcdComponent && component != constants.ControlPlaneComponent {
		return fmt.Errorf("invalid component %q, must be either %q or %q", component, constants.EtcdComponent, constants.ControlPlaneComponent)
	}
	return nil
}

// ValidateComponentWithConfig validates that the specified component is compatible with the configuration.
func ValidateComponentWithConfig(component string, config *RenewalConfig) error {
	if err := ValidateComponent(component); err != nil {
		return err
	}

	if component == constants.EtcdComponent && len(config.Etcd.Nodes) == 0 {
		return fmt.Errorf("no external etcd nodes defined; cannot use --component %s", constants.EtcdComponent)
	}

	return nil
}

// DetermineOSType determines the OS type to use based on the component.
func DetermineOSType(_ string, config *RenewalConfig) string {
	return config.OS
}

// ShouldProcessComponent checks if the specified component should be processed.
func ShouldProcessComponent(requestedComponent, targetComponent string) bool {
	return requestedComponent == "" || requestedComponent == targetComponent
}
