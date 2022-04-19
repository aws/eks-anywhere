package validator

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

// TODO: update vsphere, cloudstack, tinkerbell validators to use this common validation method
func ValidateControlPlaneIpUniqueness(cluster *v1alpha1.Cluster, netClient networkutils.NetClient) error {
	ip := cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if networkutils.IsIPInUse(netClient, ip) {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	}
	return nil
}
