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
	}
}
