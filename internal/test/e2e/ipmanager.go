package e2e

import (
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type E2EIPManager struct {
	vspherenetworkCidr string
	privateNetworkCidr string
	vsphereNetworkIPs  map[string]bool
	privateNetworkIPs  map[string]bool
}

func newE2EIPManager(networkCidr, privateNetworkCidr string) *E2EIPManager {
	ipman := &E2EIPManager{
		vspherenetworkCidr: networkCidr,
		privateNetworkCidr: privateNetworkCidr,
		vsphereNetworkIPs:  make(map[string]bool),
		privateNetworkIPs:  make(map[string]bool),
	}
	return ipman
}

func (ipman *E2EIPManager) getIP() string {
	return getUniqueIP(ipman.vspherenetworkCidr, ipman.vsphereNetworkIPs)
}

func (ipMan *E2EIPManager) getPrivateIP() string {
	return getUniqueIP(ipMan.privateNetworkCidr, ipMan.privateNetworkIPs)
}

func getUniqueIP(cidr string, usedIPs map[string]bool) string {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
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
