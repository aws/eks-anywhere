package cluster

import (
	"context"
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
