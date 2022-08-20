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
}

func (c clusterOptions) mountDirs() []string {
	var dirs []string
	if c.managementKubeconfig != "" {
		dirs = append(dirs, filepath.Dir(c.managementKubeconfig))
	}

	return dirs
}

func readAndValidateClusterSpec(clusterConfigPath string, cliVersion version.Info, opts ...cluster.SpecOpt) (*cluster.Spec, error) {
	clusterSpec, err := cluster.NewSpecFromClusterConfig(clusterConfigPath, cliVersion, opts...)
	if err != nil {
		return nil, err
	}
	if err = cluster.ValidateConfig(clusterSpec.Config); err != nil {
		return nil, err
	}

	return clusterSpec, nil
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

	clusterSpec, err := readAndValidateClusterSpec(options.fileName, version.Get(), specOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	return clusterSpec, nil
}
