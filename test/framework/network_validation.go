package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// ValidateNetworkUp validates that nodes have 2 different external IPs indicating both NICs are up
func (e *ClusterE2ETest) ValidateNetworkUp() {
	e.T.Log("Validating worker nodes have 2 different external IP")

	// First get all node names
	nodes, err := e.getAllNodes()
	if err != nil {
		e.T.Fatalf("Failed to get nodes: %v", err)
	}

	for _, node := range nodes {
		// Skip non-worker nodes (control plane nodes)
		if !e.isWorkerNode(node) {
			e.T.Logf("Skipping non-worker node: %s", node.Name)
			continue
		}

		e.T.Logf("Waiting for worker node %s to have multiple external IPs", node.Name)

		// Use a custom validation function that checks if we have multiple IPs
		err = e.waitForMultipleExternalIPs(node.Name, "5m")
		if err != nil {
			e.T.Fatalf("Worker node %s failed to get multiple external IPs within timeout: %v", node.Name, err)
		}

		e.T.Logf("Worker node %s successfully has multiple external IPs ✓", node.Name)
	}

	e.T.Log("WaitLoop network validation completed successfully")
}

// Get all nodes in the cluster using kubectl
func (e *ClusterE2ETest) getAllNodes() ([]corev1.Node, error) {
	params := []string{"get", "nodes", "-o", "json", "--kubeconfig", e.KubeconfigFilePath()}
	stdOut, err := e.KubectlClient.Execute(context.Background(), params...)
	if err != nil {
		return nil, fmt.Errorf("getting nodes: %v", err)
	}

	response := &corev1.NodeList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling nodes: %v", err)
	}

	return response.Items, nil
}

// Get external IPs from a node
func (e *ClusterE2ETest) getExternalIPsFromNode(node corev1.Node) []string {
	var externalIPs []string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			externalIPs = append(externalIPs, addr.Address)
		}
	}
	return externalIPs
}

// Check if IPs are different
func (e *ClusterE2ETest) areIPsDifferent(ips []string) bool {
	if len(ips) < 2 {
		return false
	}

	seen := make(map[string]bool)
	for _, ip := range ips {
		if seen[ip] {
			return false // Found duplicate
		}
		seen[ip] = true
	}
	return true
}

// Check if a node is a worker node
func (e *ClusterE2ETest) isWorkerNode(node corev1.Node) bool {
	// Check if node has control-plane role label
	if _, hasControlPlaneRole := node.Labels["node-role.kubernetes.io/control-plane"]; hasControlPlaneRole {
		return false
	}

	// Check for etcd role label
	if _, hasEtcdRole := node.Labels["node-role.kubernetes.io/etcd"]; hasEtcdRole {
		return false
	}

	// If it doesn't have control-plane or etcd role, it's a worker node
	return true
}

// Wait for multiple external IPs using a custom approach
func (e *ClusterE2ETest) waitForMultipleExternalIPs(nodeName, timeout string) error {
	// Parse timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout format: %v", err)
	}

	deadline := time.Now().Add(timeoutDuration)

	for time.Now().Before(deadline) {
		// Get the specific node
		output, err := e.KubectlClient.ExecuteCommand(context.Background(),
			"get", "node", nodeName,
			"-o", "json",
			"--kubeconfig", e.KubeconfigFilePath())

		if err != nil {
			e.T.Logf("Failed to get node %s, retrying: %v", nodeName, err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Parse the node JSON
		var node corev1.Node
		if err := json.Unmarshal(output.Bytes(), &node); err != nil {
			e.T.Logf("Failed to parse node JSON, retrying: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Check external IPs
		externalIPs := e.getExternalIPsFromNode(node)
		if len(externalIPs) >= 2 && e.areIPsDifferent(externalIPs) {
			e.T.Logf("Node %s now has %d different external IPs: %v",
				nodeName, len(externalIPs), externalIPs)
			return nil
		}

		e.T.Logf("Node %s has %d external IPs, waiting for 2+ different IPs: %v",
			nodeName, len(externalIPs), externalIPs)
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for node %s to have multiple external IPs", nodeName)
}
