package cluster

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func dockerEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.DockerDatacenterKind: func() APIObject {
				return &anywherev1.DockerDatacenterConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processDockerDatacenter,
		},
		Validations: []Validation{
			func(c *Config) error {
				if c.DockerDatacenter != nil {
					return c.DockerDatacenter.Validate()
				}
				return nil
			},
			func(c *Config) error {
				if c.DockerDatacenter != nil {
					if err := validateSameNamespace(c, c.DockerDatacenter); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processDockerDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.DockerDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.DockerDatacenter = datacenter.(*anywherev1.DockerDatacenterConfig)
		}
	}
}

func getDockerDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.DockerDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.DockerDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.DockerDatacenter = datacenter
	return nil
}
