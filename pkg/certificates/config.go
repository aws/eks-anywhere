// Package certificates provides functionality for managing and renewing certificates in EKS Anywhere clusters.
package certificates

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
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

	if pass := os.Getenv("EKSA_SSH_KEY_PASSPHRASE_ETCD"); pass != "" {
		config.Etcd.SSH.Password = pass
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

	if err := ValidateNodeConfig(&config.ControlPlane); err != nil {
		return fmt.Errorf("validating control plane config: %w", err)
	}

	// Etcd nodes are only required if using external etcd.
	if len(config.Etcd.Nodes) > 0 {
		if err := ValidateNodeConfig(&config.Etcd); err != nil {
			return fmt.Errorf("validating etcd config: %w", err)
		}
	}

	if err := ValidateComponentWithConfig(component, config); err != nil {
		return err
	}

	return nil
}

// ValidateNodeConfig validates a node configuration.
func ValidateNodeConfig(config *NodeConfig) error {
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

// PopulateConfig fills in the configuration with control plane and etcd node IPs from the Kubernetes cluster.
func PopulateConfig(ctx context.Context, cfg *RenewalConfig, kubeClient kubernetes.Client, cluster *types.Cluster) error {
	if len(cfg.ControlPlane.Nodes) > 0 {
		return nil
	}

	controlPlaneIPs, err := GetControlPlaneIPs(ctx, kubeClient, cluster)
	if err != nil {
		return fmt.Errorf("cluster is not reachable, please provide control plane and/or external etcd IP addresses: %w", err)
	}

	cfg.ControlPlane.Nodes = controlPlaneIPs

	etcdIPs, err := GetEtcdIPs(ctx, kubeClient, cluster)
	if err != nil {
		return fmt.Errorf("retrieving external etcd IPs for the cluster: %w", err)
	}
	cfg.Etcd.Nodes = etcdIPs

	return nil
}

// GetControlPlaneIPs retrieves the external IP addresses of all control plane nodes.
func GetControlPlaneIPs(ctx context.Context, kubeClient kubernetes.Client, cluster *types.Cluster) ([]string, error) {
	var controlPlaneIPs []string

	machineList := &clusterv1.MachineList{}

	namespaceOpt := kubernetes.ListOptions{
		Namespace: constants.EksaSystemNamespace,
	}

	if err := kubeClient.List(ctx, machineList, namespaceOpt); err != nil {
		return nil, fmt.Errorf("listing machines: %w", err)
	}

	for _, machine := range machineList.Items {
		if machine.Labels[clusterNameLabel] == cluster.Name {
			_, hasControlPlaneLabel := machine.Labels[controlPlaneLabel]
			if hasControlPlaneLabel {
				for _, address := range machine.Status.Addresses {
					if address.Type == clusterv1.MachineExternalIP && address.Address != "" {
						controlPlaneIPs = append(controlPlaneIPs, address.Address)
						break
					}
				}
			}
		}
	}

	return controlPlaneIPs, nil
}

// GetEtcdIPs retrieves the external IP addresses of all etcd nodes.
func GetEtcdIPs(ctx context.Context, kubeClient kubernetes.Client, cluster *types.Cluster) ([]string, error) {
	var etcdIPs []string

	machineList := &clusterv1.MachineList{}

	namespaceOpt := kubernetes.ListOptions{
		Namespace: constants.EksaSystemNamespace,
	}

	if err := kubeClient.List(ctx, machineList, namespaceOpt); err != nil {
		return nil, fmt.Errorf("listing machines: %w", err)
	}

	for _, machine := range machineList.Items {
		if machine.Labels[clusterNameLabel] == cluster.Name && machine.Labels[externalEtcdLabel] == cluster.Name+"-etcd" {
			for _, address := range machine.Status.Addresses {
				if address.Type == clusterv1.MachineExternalIP && address.Address != "" {
					etcdIPs = append(etcdIPs, address.Address)
					break
				}
			}
		}
	}

	return etcdIPs, nil
}
