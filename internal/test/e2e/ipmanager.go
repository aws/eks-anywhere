package e2e

import (
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type E2EIPManager struct {
	networkCidr string
	networkIPs  map[string]bool
}

func newE2EIPManager(networkCidr string) *E2EIPManager {
	ipman := &E2EIPManager{
		networkCidr: networkCidr,
		networkIPs:  make(map[string]bool),
	}
	return ipman
}

func (ipman *E2EIPManager) reserveIP() string {
	return getUniqueIP(ipman.networkCidr, ipman.networkIPs)
}

func (ipman *E2EIPManager) reserveIPPool(count int) networkutils.IPPool {
	pool := networkutils.NewIPPool()
	for i := 0; i < count; i++ {
		pool.AddIP(ipman.reserveIP())
	}
	return pool
}

func getUniqueIP(cidr string, usedIPs map[string]bool) string {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr)
	for ; err != nil || usedIPs[ip]; ip, err = ipgen.GenerateUniqueIP(cidr) {
		if err != nil {
			logger.V(2).Info("Warning: getting unique IP for vsphere failed", "error", err)
		}
		if usedIPs[ip] {
			logger.V(2).Info("Warning: generated IP is already taken", "IP", ip)
		}
	}
	usedIPs[ip] = true
	return ip
}
