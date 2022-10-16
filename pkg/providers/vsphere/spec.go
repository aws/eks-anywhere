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
	return s.VSphereMachineConfigs[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func (s *Spec) workerMachineConfig(c anywherev1.WorkerNodeGroupConfiguration) *anywherev1.VSphereMachineConfig {
	return s.VSphereMachineConfigs[c.MachineGroupRef.Name]
}

func (s *Spec) etcdMachineConfig() *anywherev1.VSphereMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil || s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}
	return s.VSphereMachineConfigs[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

func (s *Spec) machineConfigs() []*anywherev1.VSphereMachineConfig {
	machineConfigs := make([]*anywherev1.VSphereMachineConfig, 0, len(s.VSphereMachineConfigs))
	for _, m := range s.VSphereMachineConfigs {
		machineConfigs = append(machineConfigs, m)
	}

	return machineConfigs
}
