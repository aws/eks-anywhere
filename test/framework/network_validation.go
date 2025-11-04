package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clusterv2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ValidateNetworkUpUsingMachines validates that worker machines have 2 different external IPs indicating both NICs are up
func (e *ClusterE2ETest) ValidateNetworkUp(workerNodeWithSecondNetwork string) {
	e.T.Log("Validating worker machines have 2 different external IPs")

	// Get all machines.c
	machines, err := e.getAllMachines()
	if err != nil {
		e.T.Fatalf("Failed to get machines: %v", err)
	}

	for _, machine := range machines {
		// Skip non-worker machines (control plane and etcd machines)
		if !e.isWorkerMachine(machine) {
			e.T.Logf("Skipping non-worker machine: %s", machine.Name)
			continue
		}

		// Only validate machines that contain the specified worker node name (the workernode group that has second NIC)
		// if !strings.Contains(machine.Name, workerNodeWithSecondNetwork) {
		// 	e.T.Logf("Skipping worker machine without '%s' in name: %s", workerNodeWithSecondNetwork, machine.Name)
		// 	continue
		// }

		e.T.Logf("Waiting for worker machine %s to have multiple external IPs", machine.Name)

		// Use a custom validation function that checks if we have multiple IPs
		err = e.waitForMultipleExternalIPsOnMachine(machine.Name, "5m")
		if err != nil {
			e.T.Fatalf("Worker machine %s failed to get multiple external IPs within timeout: %v", machine.Name, err)
		}

		e.T.Logf("Worker machine %s successfully has multiple external IPs ✓", machine.Name)
	}

	e.T.Log("Machine network validation completed successfully")
}

// Get all machines in the cluster using kubectl
func (e *ClusterE2ETest) getAllMachines() ([]clusterv2.Machine, error) {
	params := []string{"get", "machines.cluster.x-k8s.io", "-o", "json", "--kubeconfig", e.KubeconfigFilePath(), "-n", "eksa-system"}
	stdOut, err := e.KubectlClient.Execute(context.Background(), params...)
	if err != nil {
		return nil, fmt.Errorf("getting machines: %v", err)
	}

	response := &clusterv2.MachineList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling machines: %v", err)
	}

	return response.Items, nil
}

// Get external IPs from a machine.c
func (e *ClusterE2ETest) getExternalIPsFromMachine(machine clusterv2.Machine) []string {
	var externalIPs []string
	for _, addr := range machine.Status.Addresses {
		if addr.Type == clusterv2.MachineExternalIP {
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

// Check if a machine is a worker machine
func (e *ClusterE2ETest) isWorkerMachine(machine clusterv2.Machine) bool {
	// Check machine labels for control plane or etcd roles
	if _, hasControlPlaneLabel := machine.Labels["cluster.x-k8s.io/control-plane"]; hasControlPlaneLabel {
		return false
	}

	if _, hasEtcdLabel := machine.Labels["cluster.x-k8s.io/etcd"]; hasEtcdLabel {
		return false
	}

	// Check if it's part of a MachineDeployment (worker machines are typically in MachineDeployments)
	if _, hasMachineDeployment := machine.Labels["cluster.x-k8s.io/deployment-name"]; hasMachineDeployment {
		return true
	}

	// If no specific role labels and not in a deployment, assume it's a worker
	return true
}

// Wait for multiple external IPs on a machine
func (e *ClusterE2ETest) waitForMultipleExternalIPsOnMachine(machineName, timeout string) error {
	// Parse timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout format: %v", err)
	}

	deadline := time.Now().Add(timeoutDuration)

	for time.Now().Before(deadline) {
		// Get the specific machine
		output, err := e.KubectlClient.ExecuteCommand(context.Background(),
			"get", "machine.cluster.x-k8s.io", machineName,
			"-o", "json",
			"--kubeconfig", e.KubeconfigFilePath(),
			"-n", "eksa-system")

		if err != nil {
			e.T.Logf("Failed to get machine %s, retrying: %v", machineName, err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Parse the machine JSON
		var machine clusterv2.Machine
		if err := json.Unmarshal(output.Bytes(), &machine); err != nil {
			e.T.Logf("Failed to parse machine JSON, retrying: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Check external IPs
		externalIPs := e.getExternalIPsFromMachine(machine)
		if len(externalIPs) >= 2 && e.areIPsDifferent(externalIPs) {
			e.T.Logf("Machine %s now has %d different external IPs: %v",
				machineName, len(externalIPs), externalIPs)
			return nil
		}

		e.T.Logf("Machine %s has %d external IPs, waiting for 2+ different IPs: %v",
			machineName, len(externalIPs), externalIPs)
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for machine %s to have multiple external IPs", machineName)
}
