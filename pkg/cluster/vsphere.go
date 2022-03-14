package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

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
	}
}

func processVSphereDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.VSphereDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		c.VSphereDatacenter = datacenter.(*anywherev1.VSphereDatacenterConfig)
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
