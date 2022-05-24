package cluster

import (
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func awsIamEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.AWSIamConfigKind: func() APIObject {
				return &anywherev1.AWSIamConfig{}
			},
		},
		Processors: []ParsedProcessor{processAWSIam},
		Defaulters: []Defaulter{
			func(c *Config) error {
				for _, a := range c.AWSIAMConfigs {
					a.SetDefaults()
				}
				return nil
			},
		},
		Validations: []Validation{
			func(c *Config) error {
				for _, a := range c.AWSIAMConfigs {
					if err := a.Validate(); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, a := range c.AWSIAMConfigs {
					if err := validateSameNamespace(c, a); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processAWSIam(c *Config, objects ObjectLookup) error {
	if c.AWSIAMConfigs == nil {
		c.AWSIAMConfigs = map[string]*anywherev1.AWSIamConfig{}
	}

	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		idp := objects.GetFromRef(c.Cluster.APIVersion, idr)
		if idr.Kind == anywherev1.AWSIamConfigKind {
			if idp == nil {
				return fmt.Errorf("no %s named %s", anywherev1.AWSIamConfigKind, idr.Name)
			}
			c.AWSIAMConfigs[idp.GetName()] = idp.(*anywherev1.AWSIamConfig)
		}
	}
	return nil
}
