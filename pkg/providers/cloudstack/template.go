package cloudstack

import (
	"time"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func generateTemplateBuilder(clusterSpec *cluster.Spec) (providers.TemplateBuilder, error) {
	spec := v1alpha1.ClusterSpec{
		ControlPlaneConfiguration:     clusterSpec.Cluster.Spec.ControlPlaneConfiguration,
		WorkerNodeGroupConfigurations: clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		ExternalEtcdConfiguration:     clusterSpec.Cluster.Spec.ExternalEtcdConfiguration,
	}
	controlPlaneMachineSpec, err := getControlPlaneMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	if err != nil {
		return nil, errors.Wrap(err, "generating control plane machine spec")
	}

	workerNodeGroupMachineSpecs, err := getWorkerNodeGroupMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	if err != nil {
		return nil, errors.Wrap(err, "generating worker node group machine specs")
	}

	etcdMachineSpec, err := getEtcdMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	if err != nil {
		return nil, errors.Wrap(err, "generating etcd machine spec")
	}

	templateBuilder := NewCloudStackTemplateBuilder(
		&clusterSpec.CloudStackDatacenter.Spec,
		controlPlaneMachineSpec,
		etcdMachineSpec,
		workerNodeGroupMachineSpecs,
		time.Now,
	)
	return templateBuilder, nil
}

func getEtcdMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) (*v1alpha1.CloudStackMachineConfigSpec, error) {
	var etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ExternalEtcdConfiguration != nil {
		if clusterSpec.ExternalEtcdConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &machineConfigs[clusterSpec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		} else {
			return etcdMachineSpec, errors.Errorf("getting MachineGroupRef")
		}
	} else {
		return etcdMachineSpec, errors.Errorf("getting external etcd config")
	}

	return etcdMachineSpec, nil
}

func getWorkerNodeGroupMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) (map[string]v1alpha1.CloudStackMachineConfigSpec, error) {
	var workerNodeGroupMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(machineConfigs))
	for _, wnConfig := range clusterSpec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && machineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &machineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}

	if len(workerNodeGroupMachineSpecs) == 0 {
		return workerNodeGroupMachineSpecs, errors.Errorf("getting worker node group configs")
	}

	return workerNodeGroupMachineSpecs, nil
}

func getControlPlaneMachineSpec(clusterSpec v1alpha1.ClusterSpec, machineConfigs map[string]*v1alpha1.CloudStackMachineConfig) (*v1alpha1.CloudStackMachineConfigSpec, error) {
	var controlPlaneMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.ControlPlaneConfiguration.MachineGroupRef != nil && machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[clusterSpec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	} else {
		return controlPlaneMachineSpec, errors.Errorf("getting MachineGroupRef")
	}

	return controlPlaneMachineSpec, nil
}
