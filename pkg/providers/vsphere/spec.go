package vsphere

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type Spec struct {
	*cluster.Spec
}

// NewSpec constructs a new vSphere cluster Spec.
func NewSpec(clusterSpec *cluster.Spec) *Spec {
	return &Spec{
		Spec: clusterSpec,
	}
}

func (s *Spec) controlPlaneMachineConfig() *anywherev1.VSphereMachineConfig {
	return controlPlaneMachineConfig(s.Spec)
}

func (s *Spec) workerMachineConfig(c anywherev1.WorkerNodeGroupConfiguration) *anywherev1.VSphereMachineConfig {
	return workerMachineConfig(s.Spec, c)
}

func (s *Spec) etcdMachineConfig() *anywherev1.VSphereMachineConfig {
	return etcdMachineConfig(s.Spec)
}

func (s *Spec) machineConfigs() []*anywherev1.VSphereMachineConfig {
	machineConfigs := make([]*anywherev1.VSphereMachineConfig, 0, len(s.VSphereMachineConfigs))
	for _, m := range s.VSphereMachineConfigs {
		machineConfigs = append(machineConfigs, m)
	}

	return machineConfigs
}

// MachineConfigCount represents a machineConfig with it's associated count.
type MachineConfigCount struct {
	*anywherev1.VSphereMachineConfig
	Count int
}

func (s *Spec) machineConfigsWithCount() []MachineConfigCount {
	machineConfigs := make([]MachineConfigCount, 0, len(s.VSphereMachineConfigs))
	cpMachineConfig := MachineConfigCount{
		VSphereMachineConfig: s.controlPlaneMachineConfig(),
		Count:                s.Cluster.Spec.ControlPlaneConfiguration.Count,
	}
	machineConfigs = append(machineConfigs, cpMachineConfig)
	if s.etcdMachineConfig() != nil {
		etcdMachineConfig := MachineConfigCount{
			VSphereMachineConfig: s.etcdMachineConfig(),
			Count:                s.Cluster.Spec.ExternalEtcdConfiguration.Count,
		}
		machineConfigs = append(machineConfigs, etcdMachineConfig)
	}
	for _, wc := range s.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupConfig := MachineConfigCount{
			VSphereMachineConfig: s.workerMachineConfig(wc),
			Count:                *wc.Count,
		}
		machineConfigs = append(machineConfigs, workerNodeGroupConfig)
	}
	return machineConfigs
}

func etcdMachineConfig(s *cluster.Spec) *anywherev1.VSphereMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil || s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}
	return s.VSphereMachineConfigs[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

func controlPlaneMachineConfig(s *cluster.Spec) *anywherev1.VSphereMachineConfig {
	return s.VSphereMachineConfigs[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func workerMachineConfig(s *cluster.Spec, workers anywherev1.WorkerNodeGroupConfiguration) *anywherev1.VSphereMachineConfig {
	return s.VSphereMachineConfigs[workers.MachineGroupRef.Name]
}
