package cluster

import (
	"fmt"
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

func processFlux(c *Config, objects ObjectLookup) error {
	if c.Cluster.Spec.GitOpsRef == nil {
		return nil
	}

	if c.Cluster.Spec.GitOpsRef.Kind == anywherev1.FluxConfigKind {
		flux := objects.GetFromRef(c.Cluster.APIVersion, *c.Cluster.Spec.GitOpsRef)
		if flux == nil {
			return fmt.Errorf("no %s named %s", anywherev1.FluxConfigKind, c.Cluster.Spec.GitOpsRef.Name)
		}

		c.FluxConfig = flux.(*anywherev1.FluxConfig)
	}
	return nil
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
