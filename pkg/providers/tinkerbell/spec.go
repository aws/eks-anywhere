package tinkerbell

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type spec struct {
	*cluster.Spec
	datacenterConfig     *anywherev1.TinkerbellDatacenterConfig
	machineConfigsLookup map[string]*anywherev1.TinkerbellMachineConfig
}

func newSpec(clusterSpec *cluster.Spec, machineConfigs map[string]*anywherev1.TinkerbellMachineConfig, datacenterConfig *anywherev1.TinkerbellDatacenterConfig) *spec {
	machineConfigsInCluster := map[string]*anywherev1.TinkerbellMachineConfig{}
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

func (s *spec) controlPlaneMachineConfig() *anywherev1.TinkerbellMachineConfig {
	return s.machineConfigsLookup[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func (s *spec) firstWorkerMachineConfig() *anywherev1.TinkerbellMachineConfig {
	return s.machineConfigsLookup[s.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
}

func (s *spec) etcdMachineConfig() *anywherev1.TinkerbellMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil || s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}
	return s.machineConfigsLookup[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

func (s *spec) machineConfigs() []*anywherev1.TinkerbellMachineConfig {
	machineConfigs := make([]*anywherev1.TinkerbellMachineConfig, 0, len(s.machineConfigsLookup))
	for _, m := range s.machineConfigsLookup {
		machineConfigs = append(machineConfigs, m)
	}

	return machineConfigs
}
