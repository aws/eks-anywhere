package cluster

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Config struct {
	Cluster               *anywherev1.Cluster
	VSphereDatacenter     *anywherev1.VSphereDatacenterConfig
	DockerDatacenter      *anywherev1.DockerDatacenterConfig
	VSphereMachineConfigs map[string]*anywherev1.VSphereMachineConfig
	OIDCConfigs           map[string]*anywherev1.OIDCConfig
	AWSIAMConfigs         map[string]*anywherev1.AWSIamConfig
	GitOpsConfig          *anywherev1.GitOpsConfig
}

func (c *Config) VsphereMachineConfig(name string) *anywherev1.VSphereMachineConfig {
	return c.VSphereMachineConfigs[name]
}

func (c *Config) OIDCConfig(name string) *anywherev1.OIDCConfig {
	return c.OIDCConfigs[name]
}

func (c *Config) AWSIamConfig(name string) *anywherev1.AWSIamConfig {
	return c.AWSIAMConfigs[name]
}
