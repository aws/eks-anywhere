package cluster

import (
	"reflect"

	v1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

type Config struct {
	Cluster                   *anywherev1.Cluster
	CloudStackDatacenter      *anywherev1.CloudStackDatacenterConfig
	VSphereDatacenter         *anywherev1.VSphereDatacenterConfig
	DockerDatacenter          *anywherev1.DockerDatacenterConfig
	SnowDatacenter            *anywherev1.SnowDatacenterConfig
	NutanixDatacenter         *anywherev1.NutanixDatacenterConfig
	TinkerbellDatacenter      *anywherev1.TinkerbellDatacenterConfig
	VSphereMachineConfigs     map[string]*anywherev1.VSphereMachineConfig
	CloudStackMachineConfigs  map[string]*anywherev1.CloudStackMachineConfig
	SnowMachineConfigs        map[string]*anywherev1.SnowMachineConfig
	NutanixMachineConfigs     map[string]*anywherev1.NutanixMachineConfig
	TinkerbellMachineConfigs  map[string]*anywherev1.TinkerbellMachineConfig
	TinkerbellTemplateConfigs map[string]*anywherev1.TinkerbellTemplateConfig
	OIDCConfigs               map[string]*anywherev1.OIDCConfig
	AWSIAMConfigs             map[string]*anywherev1.AWSIamConfig
	GitOpsConfig              *anywherev1.GitOpsConfig
	FluxConfig                *anywherev1.FluxConfig
	SnowCredentialsSecret     *v1.Secret
	SnowIPPools               map[string]*anywherev1.SnowIPPool
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

// SnowIPPool returns a SnowIPPool based on a name.
func (c *Config) SnowIPPool(name string) *anywherev1.SnowIPPool {
	return c.SnowIPPools[name]
}

func (c *Config) OIDCConfig(name string) *anywherev1.OIDCConfig {
	return c.OIDCConfigs[name]
}

func (c *Config) AWSIamConfig(name string) *anywherev1.AWSIamConfig {
	return c.AWSIAMConfigs[name]
}

func (c *Config) NutanixMachineConfig(name string) *anywherev1.NutanixMachineConfig {
	return c.NutanixMachineConfigs[name]
}

func (c *Config) DeepCopy() *Config {
	c2 := &Config{
		Cluster:              c.Cluster.DeepCopy(),
		CloudStackDatacenter: c.CloudStackDatacenter.DeepCopy(),
		VSphereDatacenter:    c.VSphereDatacenter.DeepCopy(),
		NutanixDatacenter:    c.NutanixDatacenter.DeepCopy(),
		DockerDatacenter:     c.DockerDatacenter.DeepCopy(),
		SnowDatacenter:       c.SnowDatacenter.DeepCopy(),
		TinkerbellDatacenter: c.TinkerbellDatacenter.DeepCopy(),
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

	if c.NutanixMachineConfigs != nil {
		c2.NutanixMachineConfigs = make(map[string]*anywherev1.NutanixMachineConfig, len(c.NutanixMachineConfigs))
	}
	for k, v := range c.NutanixMachineConfigs {
		c2.NutanixMachineConfigs[k] = v.DeepCopy()
	}

	if c.SnowMachineConfigs != nil {
		c2.SnowMachineConfigs = make(map[string]*anywherev1.SnowMachineConfig, len(c.SnowMachineConfigs))
	}
	for k, v := range c.SnowMachineConfigs {
		c2.SnowMachineConfigs[k] = v.DeepCopy()
	}

	if c.SnowIPPools != nil {
		c2.SnowIPPools = make(map[string]*anywherev1.SnowIPPool, len(c.SnowIPPools))
	}
	for k, v := range c.SnowIPPools {
		c2.SnowIPPools[k] = v.DeepCopy()
	}

	if c.TinkerbellMachineConfigs != nil {
		c2.TinkerbellMachineConfigs = make(map[string]*anywherev1.TinkerbellMachineConfig, len(c.TinkerbellMachineConfigs))
	}
	for k, v := range c.TinkerbellMachineConfigs {
		c2.TinkerbellMachineConfigs[k] = v.DeepCopy()
	}

	if c.TinkerbellTemplateConfigs != nil {
		c2.TinkerbellTemplateConfigs = make(map[string]*anywherev1.TinkerbellTemplateConfig, len(c.TinkerbellTemplateConfigs))
	}
	for k, v := range c.TinkerbellTemplateConfigs {
		c2.TinkerbellTemplateConfigs[k] = v.DeepCopy()
	}

	return c2
}

// ChildObjects returns all API objects in Config except the Cluster.
func (c *Config) ChildObjects() []kubernetes.Object {
	objs := make(
		[]kubernetes.Object,
		0,
		len(c.VSphereMachineConfigs)+len(c.SnowMachineConfigs)+len(c.CloudStackMachineConfigs)+4,
		// machine configs length + datacenter + OIDC + IAM + gitops
	)

	objs = appendIfNotNil(objs,
		c.CloudStackDatacenter,
		c.VSphereDatacenter,
		c.NutanixDatacenter,
		c.DockerDatacenter,
		c.SnowDatacenter,
		c.TinkerbellDatacenter,
		c.GitOpsConfig,
		c.FluxConfig,
	)

	for _, e := range c.VSphereMachineConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.CloudStackMachineConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.SnowMachineConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.SnowIPPools {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.NutanixMachineConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.TinkerbellMachineConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.TinkerbellTemplateConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.OIDCConfigs {
		objs = appendIfNotNil(objs, e)
	}

	for _, e := range c.AWSIAMConfigs {
		objs = appendIfNotNil(objs, e)
	}

	return objs
}

// ClusterAndChildren returns all kubernetes objects in the cluster Config.
// It's equivalent to appending the Cluster to the result of ChildObjects.
func (c *Config) ClusterAndChildren() []kubernetes.Object {
	objs := []kubernetes.Object{c.Cluster}
	return append(objs, c.ChildObjects()...)
}

func appendIfNotNil(objs []kubernetes.Object, elems ...kubernetes.Object) []kubernetes.Object {
	for _, e := range elems {
		// Since we receive interfaces, these will never be nil since they contain
		// the type of the original implementing struct
		// I can't find another clean option of doing this
		if !reflect.ValueOf(e).IsNil() {
			objs = append(objs, e)
		}
	}

	return objs
}
