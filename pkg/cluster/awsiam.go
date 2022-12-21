package cluster

import (
	"context"

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

func processAWSIam(c *Config, objects ObjectLookup) {
	if c.AWSIAMConfigs == nil {
		c.AWSIAMConfigs = map[string]*anywherev1.AWSIamConfig{}
	}

	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		idp := objects.GetFromRef(c.Cluster.APIVersion, idr)
		if idp == nil {
			return
		}
		if idr.Kind == anywherev1.AWSIamConfigKind {
			c.AWSIAMConfigs[idp.GetName()] = idp.(*anywherev1.AWSIamConfig)
		}
	}
}

func getAWSIam(ctx context.Context, client Client, c *Config) error {
	if c.AWSIAMConfigs == nil {
		c.AWSIAMConfigs = map[string]*anywherev1.AWSIamConfig{}
	}

	for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
		if idr.Kind == anywherev1.AWSIamConfigKind {
			iamConfig := &anywherev1.AWSIamConfig{}
			if err := client.Get(ctx, idr.Name, c.Cluster.Namespace, iamConfig); err != nil {
				return err
			}
			c.AWSIAMConfigs[iamConfig.Name] = iamConfig
		}
	}

	return nil
}
