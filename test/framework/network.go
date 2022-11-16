package framework

import (
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

func PopIPFromEnv(ipPoolEnvVar string) (string, error) {
	ipPool, err := networkutils.NewIPPoolFromEnv(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("popping IP from environment: %v", err)
	}

	ip, popErr := ipPool.PopIP()
	if popErr != nil {
		return "", fmt.Errorf("failed to get an ip address from the cluster ip pool env var %s: %v", ipPoolEnvVar, popErr)
	}

	// PopIPFromEnv will remove the ip from the pool.
	// Therefore, we rewrite the envvar to the system so the next caller can pick from remaining ips in the pool
	err = ipPool.ToEnvVar(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("popping IP from environment: %v", err)
	}

	return ip, nil
}

func GenerateUniqueIp(cidr string) (string, error) {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr)
	if err != nil {
		return "", fmt.Errorf("getting unique IP for cidr %s: %v", cidr, err)
	}
	return ip, nil
}

func GetIP(cidr, ipEnvVar string) (string, error) {
	value, ok := os.LookupEnv(ipEnvVar)
	var ip string
	var err error
	if ok && value != "" {
		ip, err = PopIPFromEnv(ipEnvVar)
		if err != nil {
			logger.V(2).Info("WARN: failed to pop ip from environment, attempting to generate unique ip")
			ip, err = GenerateUniqueIp(cidr)
			if err != nil {
				return "", fmt.Errorf("failed to generate ip for cidr %s: %v", cidr, err)
			}
		}
	} else {
		ip, err = GenerateUniqueIp(cidr)
		if err != nil {
			return "", fmt.Errorf("failed to generate ip for cidr %s: %v", cidr, err)
		}
	}
	return ip, nil
}
