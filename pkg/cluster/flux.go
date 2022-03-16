package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

func fluxEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.FluxConfigKind: func() APIObject {
				return &anywherev1.FluxConfig{}
			},
		},
		Processors: []ParsedProcessor{processFlux},
	}
}

func processFlux(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.GitOpsRef == nil {
		return
	}

	if c.Cluster.Spec.GitOpsRef.Kind == anywherev1.FluxConfigKind {
		flux := objects.GetFromRef(c.Cluster.APIVersion, *c.Cluster.Spec.GitOpsRef)
		if flux == nil {
			return
		}

		c.FluxConfig = flux.(*anywherev1.FluxConfig)
	}
}
