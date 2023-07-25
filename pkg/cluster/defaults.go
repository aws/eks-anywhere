package cluster

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	NodeStartupTimeout      *metav1.Duration
	UnhealthyMachineTimeout *metav1.Duration
}

// NewMachineHealthCheckDefaulter allows to create a new MachineHealthCheckDefaulter.
func NewMachineHealthCheckDefaulter(nodeStartupTimeout, unhealthyMachineTimeout *metav1.Duration) MachineHealthCheckDefaulter {
	return MachineHealthCheckDefaulter{
		NodeStartupTimeout:      nodeStartupTimeout,
		UnhealthyMachineTimeout: unhealthyMachineTimeout,
	}
}

// MachineHealthCheckDefault sets the defaults for machine health check timeouts.
func (d MachineHealthCheckDefaulter) MachineHealthCheckDefault(ctx context.Context, cluster *anywherev1.Cluster) (*anywherev1.Cluster, error) {
	if cluster.Spec.MachineHealthCheck != nil && cluster.Spec.MachineHealthCheck.NodeStartupTimeout != nil && cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout != nil {
		return cluster, nil
	}

	if cluster.Spec.MachineHealthCheck == nil {
		cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
	}

	if cluster.Spec.MachineHealthCheck.NodeStartupTimeout == nil {
		if cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind && d.NodeStartupTimeout.Duration == constants.DefaultNodeStartupTimeout {
			d.NodeStartupTimeout.Duration = constants.DefaultTinkerbellNodeStartupTimeout
		}
	}

	setMachineHealthCheckTimeoutDefaults(cluster, *d.NodeStartupTimeout, *d.UnhealthyMachineTimeout)

	return cluster, nil
}

// setMachineHealthCheckTimeoutDefaults sets default tiemout values for cluster's machine health checks.
func setMachineHealthCheckTimeoutDefaults(cluster *anywherev1.Cluster, nodeStartupTimeout, unhealthyMachineTimeout metav1.Duration) {
	if cluster.Spec.MachineHealthCheck.NodeStartupTimeout == nil {
		cluster.Spec.MachineHealthCheck.NodeStartupTimeout = &nodeStartupTimeout
	}
	if cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout == nil {
		cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout = &unhealthyMachineTimeout
	}
}
