package networkutils

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type IPGenerator struct {
	netClient NetClient
}

func NewIPGenerator(netClient NetClient) IPGenerator {
	return IPGenerator{netClient: netClient}
}

func (ipgen IPGenerator) GenerateUniqueIP(cidrBlock string) (string, error) {
	_, cidr, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return "", err
	}
	uniqueIp, err := ipgen.randIp(cidr)
	if err != nil {
		return "", err
	}
	for !IsIPInUse(ipgen.netClient, uniqueIp.String()) {
		uniqueIp, err = ipgen.randIp(cidr)
		if err != nil {
			return "", err
		}
	}
	return uniqueIp.String(), nil
}

// generates a random ip within the specified cidr block
func (ipgen IPGenerator) randIp(cidr *net.IPNet) (net.IP, error) {
	rand.Seed(time.Now().UnixNano())
	newIp := *new(net.IP)
	for i := 0; i < 4; i++ {
		newIp = append(newIp, byte(rand.Intn(256))&^cidr.Mask[i]|cidr.IP[i])
	}
	if !cidr.Contains(newIp) {
		return nil, fmt.Errorf("random IP generation failed")
	}
	return newIp, nil
}
