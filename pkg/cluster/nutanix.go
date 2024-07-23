package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
				} else if c.Cluster.Spec.DatacenterRef.Kind == v1alpha1.NutanixDatacenterKind { // We need this conditional check as NutanixDatacenter will be nil for other providers
					return fmt.Errorf("NutanixDatacenterConfig %s not found", c.Cluster.Spec.DatacenterRef.Name)
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
			func(c *Config) error {
				if c.NutanixMachineConfigs != nil { // We need this conditional check as NutanixMachineConfigs will be nil for other providers
					for _, mcRef := range c.Cluster.MachineConfigRefs() {
						if _, ok := c.NutanixMachineConfigs[mcRef.Name]; !ok {
							return fmt.Errorf("NutanixMachineConfig %s not found", mcRef.Name)
						}
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

func getNutanixDatacenter(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.NutanixDatacenterKind {
		return nil
	}

	datacenter := &anywherev1.NutanixDatacenterConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.DatacenterRef.Name, c.Cluster.Namespace, datacenter); err != nil {
		return err
	}

	c.NutanixDatacenter = datacenter
	return nil
}

func getNutanixMachineConfigs(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.NutanixDatacenterKind {
		return nil
	}

	if c.NutanixMachineConfigs == nil {
		c.NutanixMachineConfigs = map[string]*anywherev1.NutanixMachineConfig{}
	}

	for _, machineConfigRef := range c.Cluster.MachineConfigRefs() {
		if machineConfigRef.Kind != anywherev1.NutanixMachineConfigKind {
			continue
		}

		machineConfig := &anywherev1.NutanixMachineConfig{}
		if err := client.Get(ctx, machineConfigRef.Name, c.Cluster.Namespace, machineConfig); err != nil {
			return err
		}

		c.NutanixMachineConfigs[machineConfig.GetName()] = machineConfig
	}

	return nil
}
