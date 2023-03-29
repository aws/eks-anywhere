package cloudstack

import (
	"time"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func generateTemplateBuilder(clusterSpec *cluster.Spec) (providers.TemplateBuilder, error) {
	controlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating control plane machine spec")
	}

	workerNodeGroupMachineSpecs, err := getWorkerNodeGroupMachineSpec(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating worker node group machine specs")
	}

	etcdMachineSpec, err := getEtcdMachineSpec(clusterSpec)
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

func getEtcdMachineSpec(clusterSpec *cluster.Spec) (*v1alpha1.CloudStackMachineConfigSpec, error) {
	var etcdMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		}
	}

	return etcdMachineSpec, nil
}

func getWorkerNodeGroupMachineSpec(clusterSpec *cluster.Spec) (map[string]v1alpha1.CloudStackMachineConfigSpec, error) {
	var workerNodeGroupMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(clusterSpec.CloudStackMachineConfigs))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && clusterSpec.CloudStackMachineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &clusterSpec.CloudStackMachineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
		}
	}

	return workerNodeGroupMachineSpecs, nil
}

func getControlPlaneMachineSpec(clusterSpec *cluster.Spec) (*v1alpha1.CloudStackMachineConfigSpec, error) {
	var controlPlaneMachineSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec, nil
}
