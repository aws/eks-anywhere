package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
)

type clusterOptions struct {
	fileName             string
	bundlesOverride      string
	managementKubeconfig string
	hostPathsToMount     []string
}

func (c clusterOptions) mountDirs() []string {
	var dirs []string
	dirs = append(dirs, c.hostPathsToMount...)
	if c.managementKubeconfig != "" {
		dirs = append(dirs, filepath.Dir(c.managementKubeconfig))
	}

	return dirs
}

func newClusterSpec(options clusterOptions) (*cluster.Spec, error) {
	var specOpts []cluster.SpecOpt
	if options.bundlesOverride != "" {
		specOpts = append(specOpts, cluster.WithOverrideBundlesManifest(options.bundlesOverride))
	}
	if options.managementKubeconfig != "" {
		managementCluster, err := cluster.LoadManagement(options.managementKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("unable to get management cluster from kubeconfig: %v", err)
		}
		specOpts = append(specOpts, cluster.WithManagementCluster(managementCluster))
	}

	clusterSpec, err := cluster.NewSpecFromClusterConfig(options.fileName, version.Get(), specOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	return clusterSpec, nil
}
