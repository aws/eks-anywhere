package framework

import (
	"github.com/aws/eks-anywhere/internal/pkg/api"
)

var incompatiblePathsForVersion = map[string][]string{
	"v0.6.1": {
		"spec.clusterNetwork.dns",
		"spec.workerNodeGroupConfigurations[].name",
	},
	"v0.7.2": {
		"spec.clusterNetwork.cniConfig",
	},
}

func cleanUpClusterForVersion(clusterYaml []byte, version string) ([]byte, error) {
	return api.CleanupPathsFromYaml(clusterYaml, incompatiblePathsForVersion[version])
}
