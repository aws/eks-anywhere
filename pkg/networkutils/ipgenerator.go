package networkutils

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type IPGenerator struct {
	netClient NetClient
	rand      *rand.Rand
}

func NewIPGenerator(netClient NetClient) IPGenerator {
	return IPGenerator{
		netClient: netClient,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (ipgen IPGenerator) GenerateUniqueIP(cidrBlock string, usedIPs map[string]bool) (string, error) {
	_, cidr, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return "", err
	}

	// Start from the first IP in the CIDR range
	ip := incrementIP(copyIP(cidr.IP))
	checked := 0

	// Iterate sequentially through all IPs in the range
	for cidr.Contains(ip) {
		ipStr := ip.String()
		checked++

		if usedIPs != nil && usedIPs[ipStr] {
			ip = incrementIP(ip)
			continue
		}

		if IsIPInUse(ipgen.netClient, ipStr) {
			ip = incrementIP(ip)
			continue
		}

		return ipStr, nil
	}

	return "", fmt.Errorf("no available IPs in CIDR %s (checked %d IPs)", cidrBlock, checked)
}

// incrementIP returns a new IP address incremented by 1.
func incrementIP(ip net.IP) net.IP {
	// Make a copy to avoid modifying the original
	newIP := copyIP(ip)

	// Increment from the last byte
	for i := len(newIP) - 1; i >= 0; i-- {
		newIP[i]++
		if newIP[i] != 0 {
			// No overflow, we're done
			break
		}
		// Overflow occurred, continue to next byte
	}
	return newIP
}

// copyIP creates a copy of the IP address.
func copyIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
