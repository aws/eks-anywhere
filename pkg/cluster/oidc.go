package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

func oidcEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.OIDCConfigKind: func() APIObject {
				return &anywherev1.OIDCConfig{}
			},
		},
		Processors: []ParsedProcessor{processOIDC},
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
