package cluster

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func cloudstackEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.CloudStackDatacenterKind: func() APIObject {
				return &anywherev1.CloudStackDatacenterConfig{}
			},
			anywherev1.CloudStackMachineConfigKind: func() APIObject {
				return &anywherev1.CloudStackMachineConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processCloudStackDatacenter,
			machineConfigsProcessor(processCloudStackMachineConfig),
		},
		Defaulters: []Defaulter{
			func(c *Config) error {
				if c.CloudStackDatacenter != nil {
					c.CloudStackDatacenter.SetDefaults()
				}
				return nil
			},
			func(c *Config) error {
				for _, mc := range c.CloudStackMachineConfigs {
					mc.SetUserDefaults()
				}
				return nil
			},
		},
		Validations: []Validation{
			func(c *Config) error {
				if c.CloudStackDatacenter != nil {
					return c.CloudStackDatacenter.Validate()
				}
				return nil
			},
			func(c *Config) error {
				for _, m := range c.CloudStackMachineConfigs {
					if err := m.Validate(); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				if c.CloudStackDatacenter != nil {
					if err := validateSameNamespace(c, c.CloudStackDatacenter); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, v := range c.CloudStackMachineConfigs {
					if err := validateSameNamespace(c, v); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processCloudStackDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.CloudStackDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.CloudStackDatacenter = datacenter.(*anywherev1.CloudStackDatacenterConfig)
		}
	}
}

func processCloudStackMachineConfig(c *Config, objects ObjectLookup, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	if machineRef.Kind != anywherev1.CloudStackMachineConfigKind {
		return
	}

	if c.CloudStackMachineConfigs == nil {
		c.CloudStackMachineConfigs = map[string]*anywherev1.CloudStackMachineConfig{}
	}

	m := objects.GetFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	c.CloudStackMachineConfigs[m.GetName()] = m.(*anywherev1.CloudStackMachineConfig)
}

func getCloudStackDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.CloudStackDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.CloudStackDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.CloudStackDatacenter = datacenter
	return nil
}

func getCloudStackMachineConfigs(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.CloudStackDatacenterKind {
		return nil
	}

	if c.CloudStackMachineConfigs == nil {
		c.CloudStackMachineConfigs = map[string]*anywherev1.CloudStackMachineConfig{}
	}

	for _, machineRef := range c.Cluster.MachineConfigRefs() {
		if machineRef.Kind != anywherev1.CloudStackMachineConfigKind {
			continue
		}

		machine := &anywherev1.CloudStackMachineConfig{}
		if err := client.Get(ctx, machineRef.Name, c.Cluster.Namespace, machine); err != nil {
			return err
		}

		c.CloudStackMachineConfigs[machine.Name] = machine
	}

	return nil
}
