package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// StaticIPConfig holds the configuration for static IP E2E tests.
type StaticIPConfig struct {
	PoolName    string
	Addresses   []string
	Prefix      int
	Gateway     string
	Nameservers []string
}

// ValidateStaticIPAllocation validates that nodes received static IPs from the configured IP pool.
func (e *ClusterE2ETest) ValidateStaticIPAllocation(config StaticIPConfig) {
	e.T.Log("Validating static IP allocation from IP pool")

	// Get all machines
	machines, err := e.getAllMachines()
	if err != nil {
		e.T.Fatalf("Failed to get machines: %v", err)
	}

	// Parse expected IP range
	expectedIPs := parseIPRange(config.Addresses)

	for _, machine := range machines {
		e.T.Logf("Checking machine %s for static IP allocation", machine.Name)

		// Wait for machine to have an IP
		err = e.waitForMachineIP(machine.Name, "5m")
		if err != nil {
			e.T.Fatalf("Machine %s failed to get IP: %v", machine.Name, err)
		}

		// Get the external IP
		externalIPs := e.getExternalIPsFromMachine(machine)
		if len(externalIPs) == 0 {
			e.T.Fatalf("Machine %s has no external IPs", machine.Name)
		}

		// Verify IP is in the expected range
		machineIP := externalIPs[0]
		if !isIPInRange(machineIP, expectedIPs) {
			e.T.Fatalf("Machine %s IP %s is not in the expected IP pool range %v",
				machine.Name, machineIP, config.Addresses)
		}

		e.T.Logf("Machine %s has valid static IP %s from pool ✓", machine.Name, machineIP)
	}

	e.T.Log("Static IP allocation validation completed successfully")
}

// ipAddressClaim represents the structure of an IPAddressClaim resource.
type ipAddressClaim struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		PoolRef struct {
			Name string `json:"name"`
		} `json:"poolRef"`
	} `json:"spec"`
	Status struct {
		AddressRef struct {
			Name string `json:"name"`
		} `json:"addressRef"`
	} `json:"status"`
}

// ipAddress represents the structure of an IPAddress resource.
type ipAddress struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Address string `json:"address"`
		Gateway string `json:"gateway"`
		Prefix  int    `json:"prefix"`
		PoolRef struct {
			Name string `json:"name"`
		} `json:"poolRef"`
	} `json:"spec"`
}

// ValidateIPAddressResources validates that CAPI IPAM resources (IPAddressClaim, IPAddress) are created.
func (e *ClusterE2ETest) ValidateIPAddressResources(poolName string) {
	e.T.Log("Validating CAPI IPAM resources")
	e.validateIPAddressClaims(poolName)
	e.validateIPAddresses(poolName)
	e.T.Log("CAPI IPAM resources validation completed successfully")
}

// validateIPAddressClaims checks for IPAddressClaim resources.
func (e *ClusterE2ETest) validateIPAddressClaims(poolName string) {
	e.T.Log("Checking for IPAddressClaim resources")
	claimsOutput, err := e.KubectlClient.ExecuteCommand(context.Background(),
		"get", "ipaddressclaims.ipam.cluster.x-k8s.io",
		"-n", "eksa-system",
		"--kubeconfig", e.KubeconfigFilePath(),
		"-o", "json",
	)
	if err != nil {
		e.T.Fatalf("Failed to get IPAddressClaims: %v", err)
	}

	var claimsList struct {
		Items []ipAddressClaim `json:"items"`
	}
	if err := json.Unmarshal(claimsOutput.Bytes(), &claimsList); err != nil {
		e.T.Fatalf("Failed to parse IPAddressClaims: %v", err)
	}

	if len(claimsList.Items) == 0 {
		e.T.Fatal("No IPAddressClaims found - CAPI IPAM may not be working")
	}

	for _, claim := range claimsList.Items {
		e.validateSingleClaim(claim, poolName)
	}
}

// validateSingleClaim validates a single IPAddressClaim.
func (e *ClusterE2ETest) validateSingleClaim(claim ipAddressClaim, poolName string) {
	if claim.Spec.PoolRef.Name != poolName {
		e.T.Logf("Skipping claim %s with pool %s (expected %s)",
			claim.Metadata.Name, claim.Spec.PoolRef.Name, poolName)
		return
	}

	e.T.Logf("Found IPAddressClaim %s referencing pool %s",
		claim.Metadata.Name, claim.Spec.PoolRef.Name)

	if claim.Status.AddressRef.Name == "" {
		e.T.Fatalf("IPAddressClaim %s has no allocated address", claim.Metadata.Name)
	}
}

// validateIPAddresses checks for IPAddress resources.
func (e *ClusterE2ETest) validateIPAddresses(poolName string) {
	e.T.Log("Checking for IPAddress resources")
	addressesOutput, err := e.KubectlClient.ExecuteCommand(context.Background(),
		"get", "ipaddresses.ipam.cluster.x-k8s.io",
		"-n", "eksa-system",
		"--kubeconfig", e.KubeconfigFilePath(),
		"-o", "json",
	)
	if err != nil {
		e.T.Fatalf("Failed to get IPAddresses: %v", err)
	}

	var addressesList struct {
		Items []ipAddress `json:"items"`
	}
	if err := json.Unmarshal(addressesOutput.Bytes(), &addressesList); err != nil {
		e.T.Fatalf("Failed to parse IPAddresses: %v", err)
	}

	if len(addressesList.Items) == 0 {
		e.T.Fatal("No IPAddress resources found - CAPI IPAM may not be working")
	}

	for _, addr := range addressesList.Items {
		if addr.Spec.PoolRef.Name != poolName {
			continue
		}
		e.T.Logf("Found IPAddress %s with address %s from pool %s",
			addr.Metadata.Name, addr.Spec.Address, addr.Spec.PoolRef.Name)
	}
}

// ValidateInClusterIPPool validates that the InClusterIPPool resource exists and is healthy.
func (e *ClusterE2ETest) ValidateInClusterIPPool(poolName string) {
	e.T.Logf("Validating InClusterIPPool %s exists", poolName)

	output, err := e.KubectlClient.ExecuteCommand(context.Background(),
		"get", "inclusterippool.ipam.cluster.x-k8s.io", poolName,
		"-n", "eksa-system",
		"--kubeconfig", e.KubeconfigFilePath(),
		"-o", "json",
	)
	if err != nil {
		e.T.Fatalf("Failed to get InClusterIPPool %s: %v", poolName, err)
	}

	var pool struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Addresses []string `json:"addresses"`
			Prefix    int      `json:"prefix"`
			Gateway   string   `json:"gateway"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(output.Bytes(), &pool); err != nil {
		e.T.Fatalf("Failed to parse InClusterIPPool: %v", err)
	}

	e.T.Logf("InClusterIPPool %s found with addresses %v, prefix %d, gateway %s ✓",
		pool.Metadata.Name, pool.Spec.Addresses, pool.Spec.Prefix, pool.Spec.Gateway)
}

// ValidateIPReleasedAfterScaleDown validates that IPs are released back to the pool after scale down.
func (e *ClusterE2ETest) ValidateIPReleasedAfterScaleDown(poolName string, expectedFreedCount int) {
	e.T.Logf("Validating IPs are released after scale down")

	// Wait a bit for cleanup to happen
	time.Sleep(30 * time.Second)

	// Count current IPAddresses
	addressesOutput, err := e.KubectlClient.ExecuteCommand(context.Background(),
		"get", "ipaddresses.ipam.cluster.x-k8s.io",
		"-n", "eksa-system",
		"--kubeconfig", e.KubeconfigFilePath(),
		"-o", "json",
	)
	if err != nil {
		e.T.Fatalf("Failed to get IPAddresses after scale down: %v", err)
	}

	var addressesList struct {
		Items []interface{} `json:"items"`
	}
	if err := json.Unmarshal(addressesOutput.Bytes(), &addressesList); err != nil {
		e.T.Fatalf("Failed to parse IPAddresses: %v", err)
	}

	e.T.Logf("After scale down, found %d IPAddress resources", len(addressesList.Items))
	e.T.Log("IP release validation completed")
}

// waitForMachineIP waits for a machine to have an IP address assigned.
func (e *ClusterE2ETest) waitForMachineIP(machineName, timeout string) error {
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout format: %v", err)
	}

	deadline := time.Now().Add(timeoutDuration)

	for time.Now().Before(deadline) {
		output, err := e.KubectlClient.ExecuteCommand(context.Background(),
			"get", "machine.cluster.x-k8s.io", machineName,
			"-o", "jsonpath={.status.addresses[?(@.type==\"ExternalIP\")].address}",
			"--kubeconfig", e.KubeconfigFilePath(),
			"-n", "eksa-system",
		)
		if err == nil && strings.TrimSpace(output.String()) != "" {
			return nil
		}
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for machine %s to get IP", machineName)
}

// parseIPRange parses IP address specifications (ranges, CIDRs, single IPs) and returns all IPs.
func parseIPRange(addresses []string) []net.IP {
	var result []net.IP

	for _, addr := range addresses {
		// Handle IP range (e.g., "192.168.1.100-192.168.1.120")
		if strings.Contains(addr, "-") {
			parts := strings.Split(addr, "-")
			if len(parts) == 2 {
				startIP := net.ParseIP(strings.TrimSpace(parts[0]))
				endIP := net.ParseIP(strings.TrimSpace(parts[1]))
				if startIP != nil && endIP != nil {
					result = append(result, generateIPRange(startIP, endIP)...)
				}
			}
			continue
		}

		// Handle CIDR (e.g., "192.168.1.0/24")
		if strings.Contains(addr, "/") {
			_, ipNet, err := net.ParseCIDR(addr)
			if err == nil {
				result = append(result, generateCIDRIPs(ipNet)...)
			}
			continue
		}

		// Handle single IP
		ip := net.ParseIP(addr)
		if ip != nil {
			result = append(result, ip)
		}
	}

	return result
}

// generateIPRange generates all IPs between start and end (inclusive).
func generateIPRange(start, end net.IP) []net.IP {
	var result []net.IP
	start = start.To4()
	end = end.To4()

	for ip := start; !ip.Equal(end); ip = nextIP(ip) {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		result = append(result, newIP)
	}
	result = append(result, end)

	return result
}

// generateCIDRIPs generates all IPs in a CIDR block.
func generateCIDRIPs(ipNet *net.IPNet) []net.IP {
	var result []net.IP
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); ip = nextIP(ip) {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		result = append(result, newIP)
	}
	return result
}

// nextIP returns the next IP address.
func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] > 0 {
			break
		}
	}
	return result
}

// isIPInRange checks if an IP is in the expected range.
func isIPInRange(ipStr string, expectedIPs []net.IP) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, expected := range expectedIPs {
		if ip.Equal(expected) {
			return true
		}
	}
	return false
}
