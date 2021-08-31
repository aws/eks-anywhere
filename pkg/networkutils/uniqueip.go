package networkutils

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
)

var ports = []string{"22", "23", "80", "443", "6443"}

type ipgenerator struct {
	netClient NetClient
}

func NewIPGenerator(netClient NetClient) IPGenerator {
	return &ipgenerator{
		netClient: netClient,
	}
}

func (ipgen *ipgenerator) GenerateUniqueIP(cidrBlock string) (string, error) {
	_, cidr, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return "", err
	}
	uniqueIp, err := ipgen.randIp(cidr)
	if err != nil {
		return "", err
	}
	for !ipgen.IsIPUnique(uniqueIp.String()) {
		uniqueIp, err = ipgen.randIp(cidr)
		if err != nil {
			return "", err
		}
	}
	return uniqueIp.String(), nil
}

func (ipgen *ipgenerator) IsIPUnique(ip string) bool {
	// check if the ip is unique
	for _, port := range ports {
		address := net.JoinHostPort(ip, port)
		conn, err := ipgen.netClient.DialTimeout("tcp", address, time.Second)
		if err == nil && conn != nil {
			logger.Info(fmt.Sprintf("%s is already in use", ip))
			return false
		}
	}
	return true
}

// generates a random ip within the specified cidr block
func (ipgen *ipgenerator) randIp(cidr *net.IPNet) (net.IP, error) {
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
