package cloudstack

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type spec struct {
	*cluster.Spec
	datacenterConfig     *anywherev1.CloudStackDatacenterConfig
	machineConfigsLookup map[string]*anywherev1.CloudStackMachineConfig
}

func NewSpec(clusterSpec *cluster.Spec, machineConfigs map[string]*anywherev1.CloudStackMachineConfig, datacenterConfig *anywherev1.CloudStackDatacenterConfig) *spec {
	machineConfigsInCluster := map[string]*anywherev1.CloudStackMachineConfig{}
	for _, m := range clusterSpec.MachineConfigRefs() {
		machineConfig, ok := machineConfigs[m.Name]
		if !ok {
			continue
		}
		machineConfigsInCluster[m.Name] = machineConfig
	}

	return &spec{
		Spec:                 clusterSpec,
		datacenterConfig:     datacenterConfig,
		machineConfigsLookup: machineConfigsInCluster,
	}
}

func (s *spec) controlPlaneMachineConfig() *anywherev1.CloudStackMachineConfig {
	return s.machineConfigsLookup[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func (s *spec) etcdMachineConfig() *anywherev1.CloudStackMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil || s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}
	return s.machineConfigsLookup[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}
