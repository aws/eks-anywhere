package framework

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

func PopIPFromEnv(ipPoolEnvVar string) (string, error) {
	ipPool, err := networkutils.NewIPPoolFromEnv(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("error popping IP from environment: %v", err)
	}

	ip, popErr := ipPool.PopIP()
	if popErr != nil {
		return "", fmt.Errorf("failed to get an ip address from the cluster ip pool env var %s: %v", ipPoolEnvVar, popErr)
	}

	// PopIPFromEnv will remove the ip from the pool.
	// Therefore, we rewrite the envvar to the system so the next caller can pick from remaining ips in the pool
	err = ipPool.ToEnvVar(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("error popping IP from environment: %v", err)
	}

	return ip, nil
}

func GenerateUniqueIp(cidr string) (string, error) {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr)
	if err != nil {
		return "", fmt.Errorf("error getting unique IP for cidr %s: %v", cidr, err)
	}
	return ip, nil
}
