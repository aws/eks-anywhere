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

func (ipman *E2EIPManager) reserveIP() string {
	return ipman.getUniqueIP(ipman.networkCidr, ipman.networkIPs)
}

func (ipman *E2EIPManager) reserveIPPool(count int) networkutils.IPPool {
	pool := networkutils.NewIPPool()
	for i := 0; i < count; i++ {
		pool.AddIP(ipman.reserveIP())
	}
	return pool
}

func (ipman *E2EIPManager) getUniqueIP(cidr string, usedIPs map[string]bool) string {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr)
	for ; err != nil || usedIPs[ip]; ip, err = ipgen.GenerateUniqueIP(cidr) {
		if err != nil {
			ipman.logger.V(2).Info("Warning: getting unique IP failed", "error", err)
		}
		if usedIPs[ip] {
			ipman.logger.V(2).Info("Warning: generated IP is already taken", "IP", ip)
		}
	}
	usedIPs[ip] = true
	return ip
}
