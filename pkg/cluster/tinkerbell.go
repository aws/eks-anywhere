package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

// tinkerbellEntry is unimplemented. Its boiler plate to mute warnings that could confuse the customer until we
// get round to implementing it.
func tinkerbellEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.TinkerbellDatacenterKind: func() APIObject {
				return &anywherev1.TinkerbellDatacenterConfig{}
			},
			anywherev1.TinkerbellMachineConfigKind: func() APIObject {
				return &anywherev1.TinkerbellMachineConfig{}
			},
			anywherev1.TinkerbellTemplateConfigKind: func() APIObject {
				return &anywherev1.TinkerbellTemplateConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processTinkerbellDatacenter,
			machineConfigsProcessor(processTinkerbellMachineConfig),
			processTinkerbellTemplateConfigs,
		},
	}
}

func processTinkerbellDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.TinkerbellDatacenter = datacenter.(*anywherev1.TinkerbellDatacenterConfig)
		}
	}
}

func processTinkerbellTemplateConfigs(c *Config, objects ObjectLookup) {
	if c.TinkerbellTemplateConfigs == nil {
		c.TinkerbellTemplateConfigs = map[string]*anywherev1.TinkerbellTemplateConfig{}
	}

	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind {
		for _, o := range objects {
			mt, ok := o.(*anywherev1.TinkerbellTemplateConfig)
			if ok {
				c.TinkerbellTemplateConfigs[mt.Name] = mt
			}
		}
	}
}

func processTinkerbellMachineConfig(c *Config, objects ObjectLookup, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	if machineRef.Kind != anywherev1.TinkerbellMachineConfigKind {
		return
	}

	if c.TinkerbellMachineConfigs == nil {
		c.TinkerbellMachineConfigs = map[string]*anywherev1.TinkerbellMachineConfig{}
	}

	m := objects.GetFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	c.TinkerbellMachineConfigs[m.GetName()] = m.(*anywherev1.TinkerbellMachineConfig)
}
