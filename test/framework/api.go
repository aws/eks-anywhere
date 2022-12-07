package framework

import (
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

var incompatiblePathsForVersion = map[string][]string{
	"v0.6.1": {
		"spec.clusterNetwork.dns",
		"spec.workerNodeGroupConfigurations[].name",
	},
}

func cleanUpClusterForVersion(config *cluster.Config, version string) error {
	return api.CleanupPathsInObject(config.Cluster, incompatiblePathsForVersion[version])
}
