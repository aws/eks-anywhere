package e2e

import (
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

var (
	usedIPs = make(map[string]bool)
	ipgen   = networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
)

func getUniqueIP(cidr string) string {
	ip, err := ipgen.GenerateUniqueIP(cidr)
	for ; err != nil || usedIPs[ip]; ip, err = ipgen.GenerateUniqueIP(cidr) {
		if err != nil {
			logger.V(2).Info("Warning: getting unique IP for vsphere failed", err)
		}
		if usedIPs[ip] {
			logger.V(2).Info("Warning: generated IP is already taken", ip)
		}
	}
	usedIPs[ip] = true
	return ip
}
