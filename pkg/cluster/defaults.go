package cluster

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func SetConfigDefaults(c *Config) error {
	return manager().SetDefaults(c)
}

// ControlPlaneIPCheckAnnotationDefaulter is the defaulter created to set the skip ip value.
type ControlPlaneIPCheckAnnotationDefaulter struct {
	skipCPIPCheck bool
}

// NewControlPlaneIPCheckAnnotationDefaulter allows to create a new ControlPlaneIPCheckAnnotationDefaulter.
func NewControlPlaneIPCheckAnnotationDefaulter(skipIPCheck bool) ControlPlaneIPCheckAnnotationDefaulter {
	return ControlPlaneIPCheckAnnotationDefaulter{
		skipCPIPCheck: skipIPCheck,
	}
}

// ControlPlaneIPCheckDefault sets the annotation for control plane skip ip check if the flag is set to true.
func (d ControlPlaneIPCheckAnnotationDefaulter) ControlPlaneIPCheckDefault(ctx context.Context, cluster *anywherev1.Cluster) (*anywherev1.Cluster, error) {
	if d.skipCPIPCheck {
		cluster.DisableControlPlaneIPCheck()
	}

	return cluster, nil
}

// MachineHealthCheckDefaulter is the defaulter created to configure the machine health check timeouts.
type MachineHealthCheckDefaulter struct {
	nodeStartupTimeout      string
	unhealthyMachineTimeout string
}

// NewMachineHealthCheckDefaulter allows to create a new MachineHealthCheckDefaulter.
func NewMachineHealthCheckDefaulter(nodeStartupTimeout, unhealthyMachineTimeout string) MachineHealthCheckDefaulter {
	return MachineHealthCheckDefaulter{
		nodeStartupTimeout:      nodeStartupTimeout,
		unhealthyMachineTimeout: unhealthyMachineTimeout,
	}
}

// MachineHealthCheckDefault sets the defaults for machine health check timeouts.
func (d MachineHealthCheckDefaulter) MachineHealthCheckDefault(ctx context.Context, cluster *anywherev1.Cluster) (*anywherev1.Cluster, error) {
	if cluster.Spec.MachineHealthCheck != nil && len(cluster.Spec.MachineHealthCheck.NodeStartupTimeout) != 0 && len(cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout) != 0 {
		return cluster, nil
	}

	if cluster.Spec.MachineHealthCheck == nil {
		cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
	}

	if len(cluster.Spec.MachineHealthCheck.NodeStartupTimeout) == 0 {
		if cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind && d.nodeStartupTimeout == constants.DefaultNodeStartupTimeout.String() {
			d.nodeStartupTimeout = constants.DefaultTinkerbellNodeStartupTimeout.String()
		}
	}

	setMachineHealthCheckTimeoutDefaults(cluster, d.nodeStartupTimeout, d.unhealthyMachineTimeout)

	return cluster, nil
}

// setMachineHealthCheckTimeoutDefaults sets default tiemout values for cluster's machine health checks.
func setMachineHealthCheckTimeoutDefaults(cluster *anywherev1.Cluster, nodeStartupTimeout, unhealthyMachineTimeout string) {
	if len(cluster.Spec.MachineHealthCheck.NodeStartupTimeout) == 0 {
		cluster.Spec.MachineHealthCheck.NodeStartupTimeout = nodeStartupTimeout
	}

	if len(cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout) == 0 {
		cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout = unhealthyMachineTimeout
	}
}
