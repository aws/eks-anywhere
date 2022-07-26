package cluster

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

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
					if errs := o.Validate(); len(errs) != 0 {
						return apierrors.NewInvalid(anywherev1.GroupVersion.WithKind(anywherev1.OIDCConfigKind).GroupKind(), o.Name, errs)
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

func getOIDC(ctx context.Context, client Client, c *Config) error {
	if c.OIDCConfigs == nil {
		c.OIDCConfigs = map[string]*anywherev1.OIDCConfig{}
	}

	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		if idr.Kind == anywherev1.OIDCConfigKind {
			oidc := &anywherev1.OIDCConfig{}
			if err := client.Get(ctx, idr.Name, c.Cluster.Namespace, oidc); err != nil {
				return err
			}
			c.OIDCConfigs[oidc.Name] = oidc
		}
	}

	return nil
}
