package cluster

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Config struct {
	Cluster               *anywherev1.Cluster
	VSphereDatacenter     *anywherev1.VSphereDatacenterConfig
	DockerDatacenter      *anywherev1.DockerDatacenterConfig
	SnowDatacenter        *anywherev1.SnowDatacenterConfig
	VSphereMachineConfigs map[string]*anywherev1.VSphereMachineConfig
	SnowMachineConfigs    map[string]*anywherev1.SnowMachineConfig
	OIDCConfigs           map[string]*anywherev1.OIDCConfig
	AWSIAMConfigs         map[string]*anywherev1.AWSIamConfig
	GitOpsConfig          *anywherev1.GitOpsConfig
	FluxConfig            *anywherev1.FluxConfig
}

func (c *Config) VsphereMachineConfig(name string) *anywherev1.VSphereMachineConfig {
	return c.VSphereMachineConfigs[name]
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
		Cluster:           c.Cluster.DeepCopy(),
		VSphereDatacenter: c.VSphereDatacenter.DeepCopy(),
		DockerDatacenter:  c.DockerDatacenter.DeepCopy(),
		GitOpsConfig:      c.GitOpsConfig.DeepCopy(),
	}

	if c.VSphereMachineConfigs != nil {
		c2.VSphereMachineConfigs = make(map[string]*anywherev1.VSphereMachineConfig, len(c.VSphereMachineConfigs))
	}
	for k, v := range c.VSphereMachineConfigs {
		c2.VSphereMachineConfigs[k] = v.DeepCopy()
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
