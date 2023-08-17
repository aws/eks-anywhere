package cluster_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigChildObjects(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster:              &anywherev1.Cluster{},
		SnowDatacenter:       &anywherev1.SnowDatacenterConfig{},
		CloudStackDatacenter: &anywherev1.CloudStackDatacenterConfig{},
		VSphereDatacenter:    &anywherev1.VSphereDatacenterConfig{},
		NutanixDatacenter:    &anywherev1.NutanixDatacenterConfig{},
		TinkerbellDatacenter: &anywherev1.TinkerbellDatacenterConfig{},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"machine1": {}, "machine2": {},
		},
		SnowIPPools: map[string]*anywherev1.SnowIPPool{
			"pool1": {}, "pool2": {},
		},
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
			"machine1": {}, "machine2": {},
		},
		CloudStackMachineConfigs: map[string]*anywherev1.CloudStackMachineConfig{
			"machine1": {}, "machine2": {},
		},
		NutanixMachineConfigs: map[string]*anywherev1.NutanixMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellMachineConfigs: map[string]*anywherev1.TinkerbellMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellTemplateConfigs: map[string]*anywherev1.TinkerbellTemplateConfig{
			"template1": {}, "tenplate2": {},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"machine1": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"config1": {},
		},
		FluxConfig: &anywherev1.FluxConfig{},
	}

	objs := config.ChildObjects()
	g.Expect(objs).To(HaveLen(22))
	for _, o := range objs {
		g.Expect(reflect.ValueOf(o).IsNil()).To(BeFalse())
	}
}

func TestConfigClusterAndChildren(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster:              &anywherev1.Cluster{},
		SnowDatacenter:       &anywherev1.SnowDatacenterConfig{},
		CloudStackDatacenter: &anywherev1.CloudStackDatacenterConfig{},
		VSphereDatacenter:    &anywherev1.VSphereDatacenterConfig{},
		NutanixDatacenter:    &anywherev1.NutanixDatacenterConfig{},
		TinkerbellDatacenter: &anywherev1.TinkerbellDatacenterConfig{},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"machine1": {}, "machine2": {},
		},
		SnowIPPools: map[string]*anywherev1.SnowIPPool{
			"pool1": {}, "pool2": {},
		},
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
			"machine1": {}, "machine2": {},
		},
		CloudStackMachineConfigs: map[string]*anywherev1.CloudStackMachineConfig{
			"machine1": {}, "machine2": {},
		},
		NutanixMachineConfigs: map[string]*anywherev1.NutanixMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellMachineConfigs: map[string]*anywherev1.TinkerbellMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellTemplateConfigs: map[string]*anywherev1.TinkerbellTemplateConfig{
			"template1": {}, "tenplate2": {},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"machine1": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"config1": {},
		},
		FluxConfig: &anywherev1.FluxConfig{},
	}

	objs := config.ClusterAndChildren()
	g.Expect(objs).To(HaveLen(23))
	for _, o := range objs {
		g.Expect(reflect.ValueOf(o).IsNil()).To(BeFalse())
	}
}

func TestConfigDeepCopy(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster:              &anywherev1.Cluster{},
		CloudStackDatacenter: &anywherev1.CloudStackDatacenterConfig{},
		VSphereDatacenter:    &anywherev1.VSphereDatacenterConfig{},
		DockerDatacenter:     &anywherev1.DockerDatacenterConfig{},
		SnowDatacenter:       &anywherev1.SnowDatacenterConfig{},
		NutanixDatacenter:    &anywherev1.NutanixDatacenterConfig{},
		TinkerbellDatacenter: &anywherev1.TinkerbellDatacenterConfig{},
		GitOpsConfig:         &anywherev1.GitOpsConfig{},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"machine1": {}, "machine2": {},
		},
		SnowIPPools: map[string]*anywherev1.SnowIPPool{
			"pool1": {}, "pool2": {},
		},
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
			"machine1": {}, "machine2": {},
		},
		CloudStackMachineConfigs: map[string]*anywherev1.CloudStackMachineConfig{
			"machine1": {}, "machine2": {},
		},
		NutanixMachineConfigs: map[string]*anywherev1.NutanixMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellMachineConfigs: map[string]*anywherev1.TinkerbellMachineConfig{
			"machine1": {}, "machine2": {},
		},
		TinkerbellTemplateConfigs: map[string]*anywherev1.TinkerbellTemplateConfig{
			"template1": {}, "tenplate2": {},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"machine1": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"config1": {},
		},
		FluxConfig: &anywherev1.FluxConfig{},
	}

	copyConf := config.DeepCopy()
	g.Expect(copyConf).To(BeEquivalentTo(config))
}
