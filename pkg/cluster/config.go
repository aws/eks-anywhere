package cluster

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Config struct {
	Cluster                  *anywherev1.Cluster
	CloudStackDatacenter     *anywherev1.CloudStackDatacenterConfig
	VSphereDatacenter        *anywherev1.VSphereDatacenterConfig
	DockerDatacenter         *anywherev1.DockerDatacenterConfig
	SnowDatacenter           *anywherev1.SnowDatacenterConfig
	VSphereMachineConfigs    map[string]*anywherev1.VSphereMachineConfig
	CloudStackMachineConfigs map[string]*anywherev1.CloudStackMachineConfig
	SnowMachineConfigs       map[string]*anywherev1.SnowMachineConfig
	OIDCConfigs              map[string]*anywherev1.OIDCConfig
	AWSIAMConfigs            map[string]*anywherev1.AWSIamConfig
	GitOpsConfig             *anywherev1.GitOpsConfig
	FluxConfig               *anywherev1.FluxConfig
}

func (c *Config) VsphereMachineConfig(name string) *anywherev1.VSphereMachineConfig {
	return c.VSphereMachineConfigs[name]
}

func (c *Config) CloudStackMachineConfig(name string) *anywherev1.CloudStackMachineConfig {
	return c.CloudStackMachineConfigs[name]
}

func (c *Config) SnowMachineConfig(name string) *anywherev1.SnowMachineConfig {
	return c.SnowMachineConfigs[name]
}

func (c *Config) OIDCConfig(name string) *anywherev1.OIDCConfig {
	return c.OIDCConfigs[name]
}

func (c *Config) AWSIamConfig(name string) *anywherev1.AWSIamConfig {
	return c.AWSIAMConfigs[name]
}

func (c *Config) DeepCopy() *Config {
	c2 := &Config{
		Cluster:              c.Cluster.DeepCopy(),
		CloudStackDatacenter: c.CloudStackDatacenter.DeepCopy(),
		VSphereDatacenter:    c.VSphereDatacenter.DeepCopy(),
		DockerDatacenter:     c.DockerDatacenter.DeepCopy(),
		GitOpsConfig:         c.GitOpsConfig.DeepCopy(),
		FluxConfig:           c.FluxConfig.DeepCopy(),
	}

	if c.VSphereMachineConfigs != nil {
		c2.VSphereMachineConfigs = make(map[string]*anywherev1.VSphereMachineConfig, len(c.VSphereMachineConfigs))
	}

	for k, v := range c.VSphereMachineConfigs {
		c2.VSphereMachineConfigs[k] = v.DeepCopy()
	}

	if c.CloudStackMachineConfigs != nil {
		c2.CloudStackMachineConfigs = make(map[string]*anywherev1.CloudStackMachineConfig, len(c.CloudStackMachineConfigs))
	}
	for k, v := range c.CloudStackMachineConfigs {
		c2.CloudStackMachineConfigs[k] = v.DeepCopy()
	}

	if c.OIDCConfigs != nil {
		c2.OIDCConfigs = make(map[string]*anywherev1.OIDCConfig, len(c.OIDCConfigs))
	}
	for k, v := range c.OIDCConfigs {
		c2.OIDCConfigs[k] = v.DeepCopy()
	}

	if c.AWSIAMConfigs != nil {
		c2.AWSIAMConfigs = make(map[string]*anywherev1.AWSIamConfig, len(c.AWSIAMConfigs))
	}
	for k, v := range c.AWSIAMConfigs {
		c2.AWSIAMConfigs[k] = v.DeepCopy()
	}

	return c2
}
