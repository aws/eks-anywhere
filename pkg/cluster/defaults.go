package cluster

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

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
func (d ControlPlaneIPCheckAnnotationDefaulter) ControlPlaneIPCheckDefault(ctx context.Context, spec *Spec) (*Spec, error) {
	if d.skipCPIPCheck {
		spec.Cluster.DisableControlPlaneIPCheck()
	}

	return spec, nil
}

// MachineHealthCheckDefaulter is the defaulter created to configure the machine health check timeouts.
type MachineHealthCheckDefaulter struct {
	NodeStartupTimeout      time.Duration
	UnhealthyMachineTimeout time.Duration
	MaxUnhealthy            intstr.IntOrString
	WorkerMaxUnhealthy      intstr.IntOrString
}

// NewMachineHealthCheckDefaulter allows to create a new MachineHealthCheckDefaulter.
func NewMachineHealthCheckDefaulter(nodeStartupTimeout, unhealthyMachineTimeout time.Duration, globalMaxUnhealthy, workerMaxUnhealthy intstr.IntOrString) MachineHealthCheckDefaulter {
	return MachineHealthCheckDefaulter{
		NodeStartupTimeout:      nodeStartupTimeout,
		UnhealthyMachineTimeout: unhealthyMachineTimeout,
		MaxUnhealthy:            globalMaxUnhealthy,
		WorkerMaxUnhealthy:      workerMaxUnhealthy,
	}
}

// MachineHealthCheckDefault sets the defaults for machine health check timeouts and maxUnhealthy.
func (d MachineHealthCheckDefaulter) MachineHealthCheckDefault(ctx context.Context, spec *Spec) (*Spec, error) {
	SetMachineHealthCheckTimeoutDefaults(spec.Cluster, d.NodeStartupTimeout, d.UnhealthyMachineTimeout)
	SetMachineHealthCheckMaxUnhealthyDefaults(spec.Cluster, d.MaxUnhealthy, d.WorkerMaxUnhealthy)

	return spec, nil
}

// SetMachineHealthCheckTimeoutDefaults sets default timeouts for MHCs in the EKSA cluster object based on the input.
func SetMachineHealthCheckTimeoutDefaults(cluster *anywherev1.Cluster, nodeStartupTimeout, unhealthyMachineTimeout time.Duration) {
	if cluster.Spec.MachineHealthCheck != nil && cluster.Spec.MachineHealthCheck.NodeStartupTimeout != nil && cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout != nil {
		return
	}

	if cluster.Spec.MachineHealthCheck == nil {
		cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
	}

	if cluster.Spec.MachineHealthCheck.NodeStartupTimeout == nil {
		if cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind && nodeStartupTimeout == constants.DefaultNodeStartupTimeout {
			nodeStartupTimeout = constants.DefaultTinkerbellNodeStartupTimeout
		}
	}

	setMachineHealthCheckTimeoutDefaults(cluster, nodeStartupTimeout, unhealthyMachineTimeout)
}

// SetMachineHealthCheckMaxUnhealthyDefaults sets defaults maxUnhealthy for MHCs in the EKSA cluster object based on the input.
func SetMachineHealthCheckMaxUnhealthyDefaults(cluster *anywherev1.Cluster, globalMaxUnhealthy, workerMaxUnhealthy intstr.IntOrString) {
	if cluster.Spec.MachineHealthCheck != nil && cluster.Spec.MachineHealthCheck.MaxUnhealthy != nil {
		return
	}

	if cluster.Spec.MachineHealthCheck == nil {
		cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
	}

	setMachineHealthCheckMaxUnhealthyDefaults(cluster, globalMaxUnhealthy, workerMaxUnhealthy)
}

// setMachineHealthCheckTimeoutDefaults sets default timeout values for cluster's machine health checks.
func setMachineHealthCheckTimeoutDefaults(cluster *anywherev1.Cluster, nodeStartupTimeout, unhealthyMachineTimeout time.Duration) {
	if cluster.Spec.MachineHealthCheck.NodeStartupTimeout == nil {
		cluster.Spec.MachineHealthCheck.NodeStartupTimeout = &metav1.Duration{
			Duration: nodeStartupTimeout,
		}
	}
	if cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout == nil {
		cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout = &metav1.Duration{
			Duration: unhealthyMachineTimeout,
		}
	}
}

// setMachineHealthCheckMaxUnhealthyDefaults sets default maxUnhealthy values for cluster's machine health checks.
func setMachineHealthCheckMaxUnhealthyDefaults(cluster *anywherev1.Cluster, globalMaxUnhealthy, workerMaxUnhealthy intstr.IntOrString) {
	topLevelMaxUnhealthyUndefined := true
	if cluster.Spec.MachineHealthCheck.MaxUnhealthy == nil {
		cluster.Spec.MachineHealthCheck.MaxUnhealthy = &globalMaxUnhealthy
	} else {
		topLevelMaxUnhealthyUndefined = false
	}

	if cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck == nil {
		cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
	}

	if cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.MaxUnhealthy == nil {
		cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.MaxUnhealthy = &globalMaxUnhealthy
	}

	for i := range cluster.Spec.WorkerNodeGroupConfigurations {
		if cluster.Spec.WorkerNodeGroupConfigurations[i].MachineHealthCheck == nil {
			cluster.Spec.WorkerNodeGroupConfigurations[i].MachineHealthCheck = &anywherev1.MachineHealthCheck{}
		}
		if cluster.Spec.WorkerNodeGroupConfigurations[i].MachineHealthCheck.MaxUnhealthy == nil {
			if topLevelMaxUnhealthyUndefined {
				cluster.Spec.WorkerNodeGroupConfigurations[i].MachineHealthCheck.MaxUnhealthy = &workerMaxUnhealthy
			} else {
				cluster.Spec.WorkerNodeGroupConfigurations[i].MachineHealthCheck.MaxUnhealthy = cluster.Spec.MachineHealthCheck.MaxUnhealthy
			}
		}
	}
}

// NamespaceDefaulter is the defaulter created to configure the cluster's namespace.
type NamespaceDefaulter struct {
	defaultClusterNamespace string
}

// NewNamespaceDefaulter allows to create a new ClusterNamespaceDefaulter.
func NewNamespaceDefaulter(namespace string) NamespaceDefaulter {
	return NamespaceDefaulter{
		defaultClusterNamespace: namespace,
	}
}

// NamespaceDefault sets the defaults for cluster's namespace.
func (c NamespaceDefaulter) NamespaceDefault(ctx context.Context, spec *Spec) (*Spec, error) {
	for _, obj := range spec.ClusterAndChildren() {
		if obj.GetNamespace() == "" {
			obj.SetNamespace(c.defaultClusterNamespace)
		}
	}

	return spec, nil
}
