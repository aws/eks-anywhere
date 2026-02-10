package e2e

import (
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type E2EIPManager struct {
	networkCidr string
	networkIPs  map[string]bool
	logger      logr.Logger
}

func newE2EIPManager(logger logr.Logger, networkCidr string) *E2EIPManager {
	return &E2EIPManager{
		networkCidr: networkCidr,
		networkIPs:  make(map[string]bool),
		logger:      logger,
	}
}

func (ipman *E2EIPManager) reserveIP() (string, error) {
	return ipman.getUniqueIP(ipman.networkCidr, ipman.networkIPs)
}

func (ipman *E2EIPManager) reserveIPPool(count int) (networkutils.IPPool, error) {
	pool := networkutils.NewIPPool()
	for i := 0; i < count; i++ {
		ip, err := ipman.reserveIP()
		if err != nil {
			return networkutils.IPPool{}, err
		}
		pool.AddIP(ip)
	}
	return pool, nil
}

func (ipman *E2EIPManager) getUniqueIP(cidr string, usedIPs map[string]bool) (string, error) {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr, usedIPs)
	if err != nil {
		return "", err
	}
	usedIPs[ip] = true
	return ip, nil
}
