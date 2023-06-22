package validator

import (
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

// IPValidator defines the struct for control plane IP validations.
type IPValidator struct {
	netClient networkutils.NetClient
}

// IPValidatorOpt is the type for optional IPValidator configurations.
type IPValidatorOpt func(e *IPValidator)

// CustomNetClient passes in a custom net client to the IPValidator.
func CustomNetClient(netClient networkutils.NetClient) IPValidatorOpt {
	return func(d *IPValidator) {
		d.netClient = netClient
	}
}

// NewIPValidator initializes a new IPValidator object.
func NewIPValidator(opts ...IPValidatorOpt) *IPValidator {
	v := &IPValidator{
		netClient: &networkutils.DefaultNetClient{},
	}
	for _, opt := range opts {
		opt(v)
	}

	return v
}

// ValidateControlPlaneIPUniqueness checks whether or not the control plane endpoint defined
// in the cluster spec is available.
func (v *IPValidator) ValidateControlPlaneIPUniqueness(cluster *v1alpha1.Cluster) error {
	if cluster.ControlPlaneIPCheckDisabled() {
		return nil
	}

	ip := cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if networkutils.IsIPInUse(v.netClient, ip) {
		return errors.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, control plane IP must be unique", ip)
	}
	return nil
}
