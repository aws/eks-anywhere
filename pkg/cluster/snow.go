package cluster

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func snowEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.SnowDatacenterKind: func() APIObject {
				return &anywherev1.SnowDatacenterConfig{}
			},
			anywherev1.SnowMachineConfigKind: func() APIObject {
				return &anywherev1.SnowMachineConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processSnowDatacenter,
			machineConfigsProcessor(processSnowMachineConfig),
		},
		Defaulters: []Defaulter{
			func(c *Config) error {
				for _, m := range c.SnowMachineConfigs {
					m.SetDefaults()
				}
				return nil
			},
			SetSnowMachineConfigsAnnotations,
		},
		Validations: []Validation{
			func(c *Config) error {
				for _, m := range c.SnowMachineConfigs {
					if err := m.Validate(); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				if c.SnowDatacenter != nil {
					if err := validateSameNamespace(c, c.SnowDatacenter); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, v := range c.SnowMachineConfigs {
					if err := validateSameNamespace(c, v); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processSnowDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.SnowDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		c.SnowDatacenter = datacenter.(*anywherev1.SnowDatacenterConfig)
	}
}

func processSnowMachineConfig(c *Config, objects ObjectLookup, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	if machineRef.Kind != anywherev1.SnowMachineConfigKind {
		return
	}

	if c.SnowMachineConfigs == nil {
		c.SnowMachineConfigs = map[string]*anywherev1.SnowMachineConfig{}
	}

	m := objects.GetFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	c.SnowMachineConfigs[m.GetName()] = m.(*anywherev1.SnowMachineConfig)
}

func SetSnowMachineConfigsAnnotations(c *Config) error {
	if c.SnowMachineConfigs == nil {
		return nil
	}

	c.SnowMachineConfigs[c.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].SetControlPlaneAnnotation()

	if c.Cluster.Spec.ExternalEtcdConfiguration != nil {
		c.SnowMachineConfigs[c.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].SetEtcdAnnotation()
	}

	if c.Cluster.IsManaged() {
		for _, mc := range c.SnowMachineConfigs {
			mc.SetManagedBy(c.Cluster.ManagedBy())
		}
	}
	return nil
}

func getSnowDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.SnowDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.SnowDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.SnowDatacenter = datacenter
	return nil
}

func getSnowMachineConfigs(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.SnowDatacenterKind {
		return nil
	}

	if c.SnowMachineConfigs == nil {
		c.SnowMachineConfigs = map[string]*anywherev1.SnowMachineConfig{}
	}

	for _, machineRef := range c.Cluster.MachineConfigRefs() {
		if machineRef.Kind != anywherev1.SnowMachineConfigKind {
			continue
		}

		machine := &anywherev1.SnowMachineConfig{}
		if err := client.Get(ctx, machineRef.Name, c.Cluster.Namespace, machine); err != nil {
			return err
		}

		c.SnowMachineConfigs[machine.Name] = machine
	}

	return nil
}
