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

	// Count total usable IPs in the CIDR range (skip network address)
	totalIPs := cidrHostCount(cidr)
	if totalIPs == 0 {
		return "", fmt.Errorf("no usable IPs in CIDR %s", cidrBlock)
	}

	// Pick a random starting offset to reduce collisions between parallel instances
	startOffset := ipgen.rand.Intn(totalIPs)

	// Walk through all IPs starting from the random offset, wrapping around
	for i := 0; i < totalIPs; i++ {
		offset := (startOffset + i) % totalIPs
		// offset 0 = first host IP (network address + 1)
		ip := addToIP(cidr.IP, offset+1)

		if !cidr.Contains(ip) {
			continue
		}

		ipStr := ip.String()

		if usedIPs != nil && usedIPs[ipStr] {
			continue
		}

		if IsIPInUse(ipgen.netClient, ipStr) {
			continue
		}

		return ipStr, nil
	}

	return "", fmt.Errorf("no available IPs in CIDR %s (checked %d IPs)", cidrBlock, totalIPs)
}

// cidrHostCount returns the number of usable host IPs in a CIDR range (excluding network address).
func cidrHostCount(cidr *net.IPNet) int {
	ones, bits := cidr.Mask.Size()
	if bits-ones <= 0 {
		return 0
	}
	// Total addresses = 2^(bits-ones), subtract 1 for network address
	total := 1 << uint(bits-ones)
	return total - 1
}

// addToIP returns a new IP that is base + offset.
func addToIP(base net.IP, offset int) net.IP {
	ip := copyIP(base)
	carry := offset
	for i := len(ip) - 1; i >= 0 && carry > 0; i-- {
		sum := int(ip[i]) + carry
		ip[i] = byte(sum & 0xff)
		carry = sum >> 8
	}
	return ip
}

// copyIP creates a copy of the IP address.
func copyIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
