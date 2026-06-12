package networkutils_test

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type DummyNetClient struct{}

func (n *DummyNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "255.255.255.255:22" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}

// MockNetClientAllInUse simulates all IPs being in use.
type MockNetClientAllInUse struct{}

func (n *MockNetClientAllInUse) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// Simulate all IPs are in use by returning ECONNREFUSED
	// IsIPInUse checks for: err == nil OR errors.Is(err, syscall.ECONNREFUSED) OR errors.Is(err, syscall.ECONNRESET)
	return nil, &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: syscall.ECONNREFUSED,
	}
}

// MockNetClientSomeInUse simulates specific IPs being in use.
type MockNetClientSomeInUse struct {
	inUseIPs map[string]bool
}

func (n *MockNetClientSomeInUse) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// Extract IP from address (format is "IP:port")
	ip := strings.Split(address, ":")[0]
	if n.inUseIPs[ip] {
		// Simulate IP in use - return ECONNREFUSED
		return nil, &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: syscall.ECONNREFUSED,
		}
	}
	// Simulate IP not in use - return generic error
	return nil, errors.New("connection timeout")
}

func TestGenerateUniqueIP(t *testing.T) {
	cidrBlock := "1.2.3.4/16"

	ipgen := networkutils.NewIPGenerator(&DummyNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidrBlock, nil)
	if err != nil {
		t.Fatalf("GenerateUniqueIP() ip = %v error: %v", ip, err)
	}
}

func TestGenerateUniqueIPWithUsedIPsMap(t *testing.T) {
	cidrBlock := "192.168.1.0/29" // Small range: .0 to .7 (6 usable IPs)

	// Mark first 3 IPs as used
	usedIPs := map[string]bool{
		"192.168.1.1": true,
		"192.168.1.2": true,
		"192.168.1.3": true,
	}

	ipgen := networkutils.NewIPGenerator(&DummyNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidrBlock, usedIPs)
	if err != nil {
		t.Fatalf("GenerateUniqueIP() error = %v", err)
	}

	// Should skip the used IPs and return .4 or later
	if ip == "192.168.1.1" || ip == "192.168.1.2" || ip == "192.168.1.3" {
		t.Errorf("GenerateUniqueIP() returned used IP: %v", ip)
	}
}

func TestGenerateUniqueIPExhaustion(t *testing.T) {
	cidrBlock := "10.0.0.0/30" // Very small range: only .0 to .3 (2 usable IPs)

	// Use a client that marks all IPs as in use
	ipgen := networkutils.NewIPGenerator(&MockNetClientAllInUse{})
	_, err := ipgen.GenerateUniqueIP(cidrBlock, nil)
	if err == nil {
		t.Fatal("GenerateUniqueIP() expected error for exhausted IP pool, got nil")
	}

	// Verify error message mentions the CIDR
	if !strings.Contains(err.Error(), cidrBlock) {
		t.Errorf("Error message should mention CIDR %s, got: %v", cidrBlock, err)
	}
}

func TestGenerateUniqueIPWithNetworkInUse(t *testing.T) {
	cidrBlock := "10.1.1.0/29" // Small range for testing

	// Mark first 2 IPs as in use on network
	mockClient := &MockNetClientSomeInUse{
		inUseIPs: map[string]bool{
			"10.1.1.1": true,
			"10.1.1.2": true,
		},
	}

	ipgen := networkutils.NewIPGenerator(mockClient)
	ip, err := ipgen.GenerateUniqueIP(cidrBlock, nil)
	if err != nil {
		t.Fatalf("GenerateUniqueIP() error = %v", err)
	}

	// Should skip the in-use IPs
	if ip == "10.1.1.1" || ip == "10.1.1.2" {
		t.Errorf("GenerateUniqueIP() returned in-use IP: %v", ip)
	}
}

func TestGenerateUniqueIPCombinedUsedAndNetwork(t *testing.T) {
	cidrBlock := "172.16.0.0/29"

	// Mark some IPs as used in map
	usedIPs := map[string]bool{
		"172.16.0.1": true,
		"172.16.0.2": true,
	}

	// Mark some IPs as in use on network
	mockClient := &MockNetClientSomeInUse{
		inUseIPs: map[string]bool{
			"172.16.0.3": true,
			"172.16.0.4": true,
		},
	}

	ipgen := networkutils.NewIPGenerator(mockClient)
	ip, err := ipgen.GenerateUniqueIP(cidrBlock, usedIPs)
	if err != nil {
		t.Fatalf("GenerateUniqueIP() error = %v", err)
	}

	// Should skip all used and in-use IPs, return .5 or later
	usedOrInUse := []string{"172.16.0.1", "172.16.0.2", "172.16.0.3", "172.16.0.4"}
	for _, badIP := range usedOrInUse {
		if ip == badIP {
			t.Errorf("GenerateUniqueIP() returned used/in-use IP: %v", ip)
		}
	}
}

func TestGenerateUniqueIPInvalidCIDR(t *testing.T) {
	cidrBlock := "invalid-cidr"

	ipgen := networkutils.NewIPGenerator(&DummyNetClient{})
	_, err := ipgen.GenerateUniqueIP(cidrBlock, nil)
	if err == nil {
		t.Fatal("GenerateUniqueIP() expected error for invalid CIDR, got nil")
	}
}
