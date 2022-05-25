package cluster

import (
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func oidcEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.OIDCConfigKind: func() APIObject {
				return &anywherev1.OIDCConfig{}
			},
		},
		Processors: []ParsedProcessor{processOIDC},
		Validations: []Validation{
			func(c *Config) error {
				for _, o := range c.OIDCConfigs {
					if err := o.Validate(); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, o := range c.OIDCConfigs {
					if err := validateSameNamespace(c, o); err != nil {
						return err
					}
				}
				return nil
			},
			validateOIDCConfigName,
		},
	}
}

func processOIDC(c *Config, objects ObjectLookup) {
	if c.OIDCConfigs == nil {
		c.OIDCConfigs = map[string]*anywherev1.OIDCConfig{}
	}

	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		idp := objects.GetFromRef(c.Cluster.APIVersion, idr)
		if idp == nil {
			return
		}
		if idr.Kind == anywherev1.OIDCConfigKind {
			c.OIDCConfigs[idp.GetName()] = idp.(*anywherev1.OIDCConfig)
		}
	}
}

func validateOIDCConfigName(c *Config) error {
	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		if idr.Kind == anywherev1.OIDCConfigKind && c.OIDCConfigs == nil {
			return fmt.Errorf("%s/%s referenced in Cluster but not present in the cluster config", anywherev1.OIDCConfigKind, idr.Name)
		}
	}
	return nil
}
