package cloudstack

import (
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func generateTemplateBuilder(clusterSpec *cluster.Spec) providers.TemplateBuilder {
	spec := v1alpha1.ClusterSpec{
		ControlPlaneConfiguration:     clusterSpec.Cluster.Spec.ControlPlaneConfiguration,
		WorkerNodeGroupConfigurations: clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		ExternalEtcdConfiguration:     clusterSpec.Cluster.Spec.ExternalEtcdConfiguration,
	}

	controlPlaneMachineSpec := getControlPlaneMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	workerNodeGroupMachineSpecs := getWorkerNodeGroupMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	etcdMachineSpec := getEtcdMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)

	templateBuilder := NewCloudStackTemplateBuilder(
		&clusterSpec.CloudStackDatacenter.Spec,
		controlPlaneMachineSpec,
		etcdMachineSpec,
		workerNodeGroupMachineSpecs,
		time.Now,
	)
	return templateBuilder
}

func getEtcdMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) *v1alpha1.CloudStackMachineConfigSpec {
	var etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ExternalEtcdConfiguration != nil {
		if clusterSpec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	return etcdMachineSpec
}

func getWorkerNodeGroupMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) map[string]v1alpha1.CloudStackMachineConfigSpec {
	var workerNodeGroupMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(machineConfigs))
	for _, wnConfig := range clusterSpec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && machineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &machineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}

	return workerNodeGroupMachineSpecs
}

func getControlPlaneMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) *v1alpha1.CloudStackMachineConfigSpec {
	var controlPlaneMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec
}
