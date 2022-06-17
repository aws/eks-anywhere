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
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"machine1": {}, "machine2": {},
		},
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
			"machine1": {}, "machine2": {},
		},
		CloudStackMachineConfigs: map[string]*anywherev1.CloudStackMachineConfig{
			"machine1": {}, "machine2": {},
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
	g.Expect(objs).To(HaveLen(12))
	for _, o := range objs {
		g.Expect(reflect.ValueOf(o).IsNil()).To(BeFalse())
	}
}
