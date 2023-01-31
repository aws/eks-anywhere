package cluster

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

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
			anywherev1.SnowIPPoolKind: func() APIObject {
				return &anywherev1.SnowIPPool{}
			},
		},
		Processors: []ParsedProcessor{
			processSnowDatacenter,
			machineConfigsProcessor(processSnowMachineConfig),
			snowIPPoolsProcessor,
		},
		Defaulters: []Defaulter{
			func(c *Config) error {
				if c.SnowDatacenter != nil {
					SetSnowDatacenterIndentityRefDefault(c.SnowDatacenter)
				}
				return nil
			},
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
				if c.SnowDatacenter != nil {
					return c.SnowDatacenter.Validate()
				}
				return nil
			},
			func(c *Config) error {
				for _, m := range c.SnowMachineConfigs {
					if err := m.Validate(); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *Config) error {
				for _, p := range c.SnowIPPools {
					if err := p.Validate(); err != nil {
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
			func(c *Config) error {
				return ValidateSnowMachineRefExists(c)
			},
			func(c *Config) error {
				return validateSnowUnstackedEtcd(c)
			},
		},
	}
}

func processSnowDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.SnowDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		if datacenter != nil {
			c.SnowDatacenter = datacenter.(*anywherev1.SnowDatacenterConfig)
		}
	}
}

func snowIPPoolsProcessor(c *Config, o ObjectLookup) {
	for _, m := range c.SnowMachineConfigs {
		for _, pool := range m.IPPoolRefs() {
			processSnowIPPool(c, o, &pool)
		}
	}
}

func processSnowIPPool(c *Config, objects ObjectLookup, ipPoolRef *anywherev1.Ref) {
	if ipPoolRef == nil {
		return
	}

	if ipPoolRef.Kind != anywherev1.SnowIPPoolKind {
		return
	}

	if c.SnowIPPools == nil {
		c.SnowIPPools = map[string]*anywherev1.SnowIPPool{}
	}

	p := objects.GetFromRef(c.Cluster.APIVersion, *ipPoolRef)
	if p == nil {
		return
	}

	c.SnowIPPools[p.GetName()] = p.(*anywherev1.SnowIPPool)
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

func getSnowMachineConfigsAndIPPools(ctx context.Context, client Client, c *Config) error {
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

		if err := getSnowIPPools(ctx, client, c, machine); err != nil {
			return err
		}
	}

	return nil
}

func getSnowIPPools(ctx context.Context, client Client, c *Config, machine *anywherev1.SnowMachineConfig) error {
	if c.SnowIPPools == nil {
		c.SnowIPPools = map[string]*anywherev1.SnowIPPool{}
	}

	for _, dni := range machine.Spec.Network.DirectNetworkInterfaces {
		if dni.IPPoolRef == nil {
			continue
		}

		if _, ok := c.SnowIPPools[dni.IPPoolRef.Name]; ok {
			continue
		}

		pool := &anywherev1.SnowIPPool{}
		if err := client.Get(ctx, dni.IPPoolRef.Name, c.Cluster.Namespace, pool); err != nil {
			return err
		}

		c.SnowIPPools[pool.Name] = pool
	}

	return nil
}

func getSnowIdentitySecret(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.SnowDatacenterKind {
		return nil
	}

	if c.SnowDatacenter == nil {
		return errors.New("snow datacenter has to be processed before snow identityRef credentials secret")
	}

	secret := &corev1.Secret{}
	if err := client.Get(ctx, c.SnowDatacenter.Spec.IdentityRef.Name, c.Cluster.Namespace, secret); err != nil {
		return err
	}

	c.SnowCredentialsSecret = secret

	return nil
}

// SetSnowDatacenterIndentityRefDefault sets a default secret as the identity reference
// The secret will need to be created by the CLI flow as it's not provided by the user
// This only runs in CLI. snowDatacenterConfig.SetDefaults() will run in both CLI and webhook.
func SetSnowDatacenterIndentityRefDefault(s *anywherev1.SnowDatacenterConfig) {
	if len(s.Spec.IdentityRef.Kind) == 0 && len(s.Spec.IdentityRef.Name) == 0 {
		s.Spec.IdentityRef = anywherev1.Ref{
			Kind: anywherev1.SnowIdentityKind,
			Name: fmt.Sprintf("%s-snow-credentials", s.GetName()),
		}
	}
}

// ValidateSnowMachineRefExists checks the cluster spec machine refs and makes sure
// the snowmachineconfig object exists for each ref with kind == snowmachineconfig.
func ValidateSnowMachineRefExists(c *Config) error {
	for _, machineRef := range c.Cluster.MachineConfigRefs() {
		if machineRef.Kind == anywherev1.SnowMachineConfigKind && c.SnowMachineConfig(machineRef.Name) == nil {
			return fmt.Errorf("unable to find SnowMachineConfig %s", machineRef.Name)
		}
	}
	return nil
}

func validateSnowUnstackedEtcd(c *Config) error {
	if c.Cluster.Spec.DatacenterRef.Kind != anywherev1.SnowDatacenterKind {
		return nil
	}

	if c.Cluster.Spec.ExternalEtcdConfiguration == nil || c.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}

	mc := c.SnowMachineConfig(c.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)

	for _, dni := range mc.Spec.Network.DirectNetworkInterfaces {
		if dni.DHCP {
			return errors.New("creating unstacked etcd machine with DHCP is not supported for snow. Please use static IP for DNI configuration")
		}
		if dni.IPPoolRef == nil {
			return errors.New("snow machine config ip pool must be specified when using static IP")
		}
	}
	return nil
}
