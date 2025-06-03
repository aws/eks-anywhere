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

	var config struct {
		Spec struct {
			ControlPlaneConfiguration struct {
				SSHKeyPath string `yaml:"sshKeyPath"`
				Users      []struct {
					Name string `yaml:"name"`
				} `yaml:"users"`
			} `yaml:"controlPlaneConfiguration"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse cluster config: %v", err)
	}

	sshConfig := &clusterSSHConfig{
		SSHKeyPath: config.Spec.ControlPlaneConfiguration.SSHKeyPath,
	}

	// get SSH username
	if len(config.Spec.ControlPlaneConfiguration.Users) > 0 {
		sshConfig.SSHUsername = config.Spec.ControlPlaneConfiguration.Users[0].Name
	} else {
		// if no usernamr, set it to default ec2-user
		sshConfig.SSHUsername = "ec2-user"
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
			renewalConfig.ControlPlane.Nodes = append(renewalConfig.ControlPlane.Nodes, node.Status.Addresses[0].Address)

			osImage := strings.ToLower(node.Status.NodeInfo.OSImage)
			if strings.Contains(osImage, "bottlerocket") {
				renewalConfig.ControlPlane.OS = "bottlerocket"
			} else if strings.Contains(osImage, "ubuntu") {
				renewalConfig.ControlPlane.OS = "ubuntu"
			} else if strings.Contains(osImage, "rhel") {
				renewalConfig.ControlPlane.OS = "rhel"
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
	renewalConfig.ControlPlane.SSHUser = sshConfig.SSHUsername
	renewalConfig.Etcd.SSHKey = sshKeyPath
	renewalConfig.Etcd.SSHUser = sshConfig.SSHUsername

	return renewalConfig, nil
}
