package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func vsphereEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.VSphereDatacenterKind: func() APIObject {
				return &anywherev1.VSphereDatacenterConfig{}
			},
			anywherev1.VSphereMachineConfigKind: func() APIObject {
				return &anywherev1.VSphereMachineConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processVSphereDatacenter,
			machineConfigsProcessor(processVSphereMachineConfig),
		},
		Defaulters: []Defaulter{
			func(c *Config) error {
				if c.VSphereDatacenter != nil {
					c.VSphereDatacenter.SetDefaults()
				}
				return nil
			},
			func(c *Config) error {
				for _, m := range c.VSphereMachineConfigs {
					m.SetDefaults()
					m.SetUserDefaults()
				}
				return nil
			},
		},
		Validations: []Validation{
			func(c *Config) error {
				if c.VSphereDatacenter != nil {
					return c.VSphereDatacenter.Validate()
				} else if c.Cluster.Spec.DatacenterRef.Kind == v1alpha1.VSphereDatacenterKind { // We need this conditional check as VSphereDatacenter will be nil for other providers
					return fmt.Errorf("VSphereDatacenterConfig %s not found", c.Cluster.Spec.DatacenterRef.Name)
				}
				return nil
			},
			func(c *Config) error {
				if c.VSphereMachineConfigs != nil { // We need this conditional check as VSphereMachineConfigs will be nil for other providers
					for _, mcRef := range c.Cluster.MachineConfigRefs() {
						m, ok := c.VSphereMachineConfigs[mcRef.Name]
						if !ok {
							return fmt.Errorf("VSphereMachineConfig %s not found", mcRef.Name)
						}
						if err := m.Validate(); err != nil {
							return err
						}
					}
				}
				return nil
			},
			func(c *Config) error {
				if c.VSphereDatacenter != nil {
					if err := validateSameNamespace(c, c.VSphereDatacenter); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, v := range c.VSphereMachineConfigs {
					if err := validateSameNamespace(c, v); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processVSphereDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.VSphereDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.VSphereDatacenter = datacenter.(*anywherev1.VSphereDatacenterConfig)
		}
	}
}

func processVSphereMachineConfig(c *Config, objects ObjectLookup, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	if machineRef.Kind != anywherev1.VSphereMachineConfigKind {
		return
	}

	if c.VSphereMachineConfigs == nil {
		c.VSphereMachineConfigs = map[string]*anywherev1.VSphereMachineConfig{}
	}

	m := objects.GetFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	c.VSphereMachineConfigs[m.GetName()] = m.(*anywherev1.VSphereMachineConfig)
}

func getVSphereDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.VSphereDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.VSphereDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.VSphereDatacenter = datacenter
	return nil
}

func getVSphereMachineConfigs(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.VSphereDatacenterKind {
		return nil
	}

	if c.VSphereMachineConfigs == nil {
		c.VSphereMachineConfigs = map[string]*anywherev1.VSphereMachineConfig{}
	}

	for _, machineRef := range c.Cluster.MachineConfigRefs() {
		if machineRef.Kind != anywherev1.VSphereMachineConfigKind {
			continue
		}

		machine := &anywherev1.VSphereMachineConfig{}
		if err := client.Get(ctx, machineRef.Name, c.Cluster.Namespace, machine); err != nil {
			return err
		}

		c.VSphereMachineConfigs[machine.Name] = machine
	}

	return nil
}
