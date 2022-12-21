package cluster

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func nutanixEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.NutanixDatacenterKind: func() APIObject {
				return &anywherev1.NutanixDatacenterConfig{}
			},
			anywherev1.NutanixMachineConfigKind: func() APIObject {
				return &anywherev1.NutanixMachineConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processNutanixDatacenter,
			machineConfigsProcessor(processNutanixMachineConfig),
		},
		Validations: []Validation{
			func(c *Config) error {
				if c.NutanixDatacenter != nil {
					return c.NutanixDatacenter.Validate()
				}
				return nil
			},
			func(c *Config) error {
				if c.NutanixDatacenter != nil {
					if err := validateSameNamespace(c, c.NutanixDatacenter); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processNutanixDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.NutanixDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.NutanixDatacenter = datacenter.(*anywherev1.NutanixDatacenterConfig)
		}
	}
}

func processNutanixMachineConfig(c *Config, objects ObjectLookup, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	if machineRef.Kind != anywherev1.NutanixMachineConfigKind {
		return
	}

	if c.NutanixMachineConfigs == nil {
		c.NutanixMachineConfigs = map[string]*anywherev1.NutanixMachineConfig{}
	}

	m := objects.GetFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	c.NutanixMachineConfigs[m.GetName()] = m.(*anywherev1.NutanixMachineConfig)
}
