package cluster

import eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

func BuildMapForWorkerNodeGroupsByName(workerNodeGroups []eksav1alpha1.WorkerNodeGroupConfiguration) map[string]eksav1alpha1.WorkerNodeGroupConfiguration {
	workerNodeGroupConfigs := make(map[string]eksav1alpha1.WorkerNodeGroupConfiguration, len(workerNodeGroups))
	for _, config := range workerNodeGroups {
		workerNodeGroupConfigs[config.Name] = config
	}
	return workerNodeGroupConfigs
}

func NodeGroupsToDelete(currentSpec, newSpec *Spec) []eksav1alpha1.WorkerNodeGroupConfiguration {
	workerConfigs := BuildMapForWorkerNodeGroupsByName(newSpec.Cluster.Spec.WorkerNodeGroupConfigurations)
	nodeGroupsToDelete := make([]eksav1alpha1.WorkerNodeGroupConfiguration, 0, len(currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, prevWorkerNodeGroupConfig := range currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		// Current spec doesn't have the default name since we never set the defaults at the api server level
		if prevWorkerNodeGroupConfig.Name == "" {
			prevWorkerNodeGroupConfig.Name = "md-0"
		}
		if _, ok := workerConfigs[prevWorkerNodeGroupConfig.Name]; !ok {
			nodeGroupsToDelete = append(nodeGroupsToDelete, prevWorkerNodeGroupConfig)
		}
	}
	return nodeGroupsToDelete
}
