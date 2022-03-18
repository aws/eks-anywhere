package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

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
