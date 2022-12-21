package cluster

import (
	"context"
	"path"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func fluxEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.FluxConfigKind: func() APIObject {
				return &anywherev1.FluxConfig{}
			},
		},
		Processors: []ParsedProcessor{processFlux},
		Defaulters: []Defaulter{
			setFluxDefaults,
			SetDefaultFluxConfigPath,
		},
		Validations: []Validation{
			validateFlux,
			validateFluxNamespace,
		},
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

func validateFlux(c *Config) error {
	if c.FluxConfig != nil {
		return c.FluxConfig.Validate()
	}
	return nil
}

func validateFluxNamespace(c *Config) error {
	if c.FluxConfig != nil {
		if err := validateSameNamespace(c, c.FluxConfig); err != nil {
			return err
		}
	}
	return nil
}

func setFluxDefaults(c *Config) error {
	if c.FluxConfig != nil {
		c.FluxConfig.SetDefaults()
	}
	return nil
}

func SetDefaultFluxConfigPath(c *Config) error {
	if c.FluxConfig == nil {
		return nil
	}

	fluxConfig := c.FluxConfig
	if fluxConfig.Spec.ClusterConfigPath != "" {
		return nil
	}

	if c.Cluster.IsSelfManaged() {
		fluxConfig.Spec.ClusterConfigPath = path.Join("clusters", c.Cluster.Name)
	} else {
		fluxConfig.Spec.ClusterConfigPath = path.Join("clusters", c.Cluster.ManagedBy())
	}
	return nil
}

func getFluxConfig(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.GitOpsRef == nil || c.Cluster.Spec.GitOpsRef.Kind != anywherev1.FluxConfigKind {
		return nil
	}

	fluxConfig := &anywherev1.FluxConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.GitOpsRef.Name, c.Cluster.Namespace, fluxConfig); err != nil {
		return err
	}

	c.FluxConfig = fluxConfig

	return nil
}
