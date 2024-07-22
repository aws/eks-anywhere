package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

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
		Validations: []Validation{
			func(c *Config) error {
				if c.TinkerbellDatacenter != nil {
					if err := validateSameNamespace(c, c.TinkerbellDatacenter); err != nil {
						return err
					}
				} else if c.Cluster.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind { // We need this conditional check as TinkerbellDatacenter will be nil for other providers
					return fmt.Errorf("TinkerbellDatacenterConfig %s not found", c.Cluster.Spec.DatacenterRef.Name)
				}
				return nil
			},
			func(c *Config) error {
				if c.TinkerbellMachineConfigs != nil { // We need this conditional check as TinkerbellMachineConfigs will be nil for other providers
					for _, mcRef := range c.Cluster.MachineConfigRefs() {
						m, ok := c.TinkerbellMachineConfigs[mcRef.Name]
						if !ok {
							return fmt.Errorf("TinkerbellMachineConfig %s not found", mcRef.Name)
						}
						if err := validateSameNamespace(c, m); err != nil {
							return err
						}
					}
				}
				return nil
			},
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

func getTinkerbellDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.TinkerbellDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.TinkerbellDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.TinkerbellDatacenter = datacenter
	return nil
}

func getTinkerbellMachineAndTemplateConfigs(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.TinkerbellDatacenterKind {
		return nil
	}

	if c.TinkerbellMachineConfigs == nil {
		c.TinkerbellMachineConfigs = map[string]*anywherev1.TinkerbellMachineConfig{}
	}
	for _, machineRef := range c.Cluster.MachineConfigRefs() {
		if machineRef.Kind != anywherev1.TinkerbellMachineConfigKind {
			continue
		}

		machineConfig := &anywherev1.TinkerbellMachineConfig{}
		if err := client.Get(ctx, machineRef.Name, c.Cluster.Namespace, machineConfig); err != nil {
			return err
		}

		c.TinkerbellMachineConfigs[machineConfig.Name] = machineConfig

		if !machineConfig.Spec.TemplateRef.IsEmpty() {
			if c.TinkerbellTemplateConfigs == nil {
				c.TinkerbellTemplateConfigs = map[string]*anywherev1.TinkerbellTemplateConfig{}
			}

			templateRefName := machineConfig.Spec.TemplateRef.Name
			templateConfig := &anywherev1.TinkerbellTemplateConfig{}
			if err := client.Get(ctx, templateRefName, c.Cluster.Namespace, templateConfig); err != nil {
				return err
			}
			c.TinkerbellTemplateConfigs[templateRefName] = templateConfig
		}
	}
	return nil
}
