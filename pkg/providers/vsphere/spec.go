package vsphere

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type spec struct {
	*cluster.Spec
	datacenterConfig     *anywherev1.VSphereDatacenterConfig
	machineConfigsLookup map[string]*anywherev1.VSphereMachineConfig
}

func newSpec(clusterSpec *cluster.Spec, machineConfigs map[string]*anywherev1.VSphereMachineConfig, datacenterConfig *anywherev1.VSphereDatacenterConfig) *spec {
	machineConfigsInCluster := map[string]*anywherev1.VSphereMachineConfig{}
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

func (s *spec) controlPlaneMachineConfig() *anywherev1.VSphereMachineConfig {
	return s.machineConfigsLookup[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func (s *spec) firstWorkerMachineConfig() *anywherev1.VSphereMachineConfig {
	return s.machineConfigsLookup[s.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
}

func (s *spec) etcdMachineConfig() *anywherev1.VSphereMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil {
		return nil
	}
	return s.machineConfigsLookup[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

func (s *spec) machineConfigs() []*anywherev1.VSphereMachineConfig {
	machineConfigs := make([]*anywherev1.VSphereMachineConfig, 0, len(s.machineConfigsLookup))
	for _, m := range s.machineConfigsLookup {
		machineConfigs = append(machineConfigs, m)
	}

	return machineConfigs
}
