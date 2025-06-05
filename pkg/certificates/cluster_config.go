package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

// SSH configuration for the cluster
type clusterSSHConfig struct {
	SSHKeyPath  string
	SSHUsername string
}

// get SSH configuration from the cluster's configuration
func getClusterConfig(clusterName string) (*clusterSSHConfig, error) {
	clusterDir := filepath.Join(".", clusterName)
	clusterConfigPath := filepath.Join(clusterDir, fmt.Sprintf("%s-eks-a-cluster.yaml", clusterName))

	data, err := os.ReadFile(clusterConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster config file: %v", err)
	}

	sshConfig := &clusterSSHConfig{}

	// split the YAML file into multiple documents
	documents := strings.Split(string(data), "---")

	// first pass: look for VSphereMachineConfig with control-plane annotation
	for _, doc := range documents {
		if doc == "" {
			continue
		}

		// check if this document is a VSphereMachineConfig with control-plane annotation
		var machineConfig struct {
			Kind     string `yaml:"kind"`
			Metadata struct {
				Annotations map[string]string `yaml:"annotations"`
				Name        string            `yaml:"name"`
			} `yaml:"metadata"`
			Spec struct {
				OSFamily string `yaml:"osFamily"`
				Users    []struct {
					Name string `yaml:"name"`
				} `yaml:"users"`
			} `yaml:"spec"`
		}

		if err := yaml.Unmarshal([]byte(doc), &machineConfig); err != nil {
			// skip documents that don't match this structure
			continue
		}

		// check if this is a VSphereMachineConfig with control-plane annotation
		if machineConfig.Kind == "VSphereMachineConfig" &&
			machineConfig.Metadata.Annotations != nil &&
			machineConfig.Metadata.Annotations["anywhere.eks.amazonaws.com/control-plane"] == "true" {
			// found the control plane machine config
			if len(machineConfig.Spec.Users) > 0 {
				sshConfig.SSHUsername = machineConfig.Spec.Users[0].Name
				fmt.Printf("Found SSH username '%s' in VSphereMachineConfig for control plane\n", sshConfig.SSHUsername)
				break
			}
		}
	}

	// second pass: if no username found, look for any VSphereMachineConfig
	if sshConfig.SSHUsername == "" {
		for _, doc := range documents {
			if doc == "" {
				continue
			}

			var machineConfig struct {
				Kind string `yaml:"kind"`
				Spec struct {
					Users []struct {
						Name string `yaml:"name"`
					} `yaml:"users"`
				} `yaml:"spec"`
			}

			if err := yaml.Unmarshal([]byte(doc), &machineConfig); err != nil {
				continue
			}

			if machineConfig.Kind == "VSphereMachineConfig" && len(machineConfig.Spec.Users) > 0 {
				sshConfig.SSHUsername = machineConfig.Spec.Users[0].Name
				fmt.Printf("Found SSH username '%s' in VSphereMachineConfig\n", sshConfig.SSHUsername)
				break
			}
		}
	}

	// third pass: look for SSHKeyPath in Cluster resource
	for _, doc := range documents {
		if doc == "" {
			continue
		}

		var clusterConfig struct {
			Kind string `yaml:"kind"`
			Spec struct {
				ControlPlaneConfiguration struct {
					SSHKeyPath string `yaml:"sshKeyPath"`
				} `yaml:"controlPlaneConfiguration"`
			} `yaml:"spec"`
		}

		if err := yaml.Unmarshal([]byte(doc), &clusterConfig); err != nil {
			continue
		}

		if clusterConfig.Kind == "Cluster" {
			sshConfig.SSHKeyPath = clusterConfig.Spec.ControlPlaneConfiguration.SSHKeyPath
			break
		}
	}

	// If no username found, use default
	if sshConfig.SSHUsername == "" {
		sshConfig.SSHUsername = "ec2-user"
		fmt.Printf("No SSH username found in config, using default: %s\n", sshConfig.SSHUsername)
	}

	return sshConfig, nil
}

// the RenewalConfig from a running cluster
func BuildConfigFromCluster(clusterName, sshKeyPath string) (*RenewalConfig, error) {
	if _, err := os.Stat(sshKeyPath); err != nil {
		return nil, fmt.Errorf("SSH key file not found: %v", err)
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	// get kubeadm-config ConfigMap
	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeadm-config: %v", err)
	}

	var clusterConfig struct {
		Etcd struct {
			External struct {
				Endpoints []string `yaml:"endpoints"`
			} `yaml:"external"`
		} `yaml:"etcd"`
	}

	if err := yaml.Unmarshal([]byte(cm.Data["ClusterConfiguration"]), &clusterConfig); err != nil {
		return nil, fmt.Errorf("failed to parse cluster configuration: %v", err)
	}

	// get nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	renewalConfig := &RenewalConfig{
		ClusterName: clusterName,
		ControlPlane: NodeConfig{
			Nodes: []string{},
		},
		Etcd: NodeConfig{
			Nodes: []string{},
		},
	}

	// for control plane nodes
	for _, node := range nodes.Items {
		if _, isControlPlane := node.Labels["node-role.kubernetes.io/control-plane"]; isControlPlane {
			// find InternalIP address
			var nodeIP string
			for _, addr := range node.Status.Addresses {
				if addr.Type == "InternalIP" {
					nodeIP = addr.Address
					break
				}
			}
			// If InternalIP not found, fall back to the first address
			if nodeIP == "" && len(node.Status.Addresses) > 0 {
				nodeIP = node.Status.Addresses[0].Address
				fmt.Printf("Warning: InternalIP not found for node %s, using %s instead\n",
					node.Name, nodeIP)
			}

			renewalConfig.ControlPlane.Nodes = append(renewalConfig.ControlPlane.Nodes, nodeIP)

			osImage := strings.ToLower(node.Status.NodeInfo.OSImage)
			if strings.Contains(osImage, "bottlerocket") {
				renewalConfig.ControlPlane.OS = "bottlerocket"
			} else if strings.Contains(osImage, "ubuntu") {
				renewalConfig.ControlPlane.OS = "ubuntu"
			} else if strings.Contains(osImage, "rhel") || strings.Contains(osImage, "red hat") {
				renewalConfig.ControlPlane.OS = "redhat"
			} else {
				fmt.Printf("DEBUG: Could not detect OS from OSImage: %s\n", osImage)
			}
		}
	}

	// process etcd nodes if external etcd is configured
	if len(clusterConfig.Etcd.External.Endpoints) > 0 {
		for _, endpoint := range clusterConfig.Etcd.External.Endpoints {
			parts := strings.Split(endpoint, "://")
			if len(parts) != 2 {
				continue
			}
			ip := strings.Split(parts[1], ":")[0]
			renewalConfig.Etcd.Nodes = append(renewalConfig.Etcd.Nodes, ip)
		}
		// for external etcd, we assume the same OS type as control plane
		// user can override this using the config file if needed
		renewalConfig.Etcd.OS = renewalConfig.ControlPlane.OS
	}

	// find SSH user
	sshConfig, err := getClusterConfig(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster configuration: %v", err)
	}

	// SSH configuration
	renewalConfig.ControlPlane.SSHKey = sshKeyPath

	// override SSH user based on OS type if needed
	if renewalConfig.ControlPlane.OS == "ubuntu" && sshConfig.SSHUsername != "ubuntu" {
		fmt.Printf("Warning: Overriding SSH user from '%s' to 'ubuntu' for Ubuntu nodes\n", sshConfig.SSHUsername)
		renewalConfig.ControlPlane.SSHUser = "ubuntu"
	} else if renewalConfig.ControlPlane.OS == "bottlerocket" && sshConfig.SSHUsername != "ec2-user" {
		fmt.Printf("Warning: Overriding SSH user from '%s' to 'ec2-user' for Bottlerocket nodes\n", sshConfig.SSHUsername)
		renewalConfig.ControlPlane.SSHUser = "ec2-user"
	} else if renewalConfig.ControlPlane.OS == "rhel" || renewalConfig.ControlPlane.OS == "redhat" {
		renewalConfig.ControlPlane.SSHUser = sshConfig.SSHUsername
		fmt.Printf("Using SSH user '%s' for RHEL/RedHat nodes as specified in cluster config\n", sshConfig.SSHUsername)
	} else {
		renewalConfig.ControlPlane.SSHUser = sshConfig.SSHUsername
	}

	renewalConfig.Etcd.SSHKey = sshKeyPath
	renewalConfig.Etcd.SSHUser = renewalConfig.ControlPlane.SSHUser

	return renewalConfig, nil
}
