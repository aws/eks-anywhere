package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/version"
)

const timeoutErrorTemplate = "failed to parse timeout %s: %v"

type timeoutOptions struct {
	cpWaitTimeout           string
	externalEtcdWaitTimeout string
	perMachineWaitTimeout   string
}

func setupTimeoutFlags(cmd *cobra.Command, t *timeoutOptions) {
	cmd.Flags().StringVar(&t.cpWaitTimeout, cpWaitTimeoutFlag, clustermanager.DefaultControlPlaneWait.String(), "Override the default control plane wait timeout (60m).")
	markFlagHidden(cmd, cpWaitTimeoutFlag)

	cmd.Flags().StringVar(&t.externalEtcdWaitTimeout, externalEtcdWaitTimeoutFlag, clustermanager.DefaultEtcdWait.String(), "Override the default external etcd wait timeout (60m)")
	markFlagHidden(cmd, externalEtcdWaitTimeoutFlag)

	cmd.Flags().StringVar(&t.perMachineWaitTimeout, perMachineWaitTimeoutFlag, clustermanager.DefaultMaxWaitPerMachine.String(), "Override the default machine wait timeout (10m)/per machine ")
	markFlagHidden(cmd, perMachineWaitTimeoutFlag)
}

func buildClusterManagerOpts(t timeoutOptions) ([]clustermanager.ClusterManagerOpt, error) {
	cpWaitTimeout, err := time.ParseDuration(t.cpWaitTimeout)
	if err != nil {
		return nil, fmt.Errorf(timeoutErrorTemplate, cpWaitTimeoutFlag, err)
	}

	externalEtcdWaitTimeout, err := time.ParseDuration(t.externalEtcdWaitTimeout)
	if err != nil {
		return nil, fmt.Errorf(timeoutErrorTemplate, externalEtcdWaitTimeoutFlag, err)
	}

	perMachineWaitTimeout, err := time.ParseDuration(t.perMachineWaitTimeout)
	if err != nil {
		return nil, fmt.Errorf(timeoutErrorTemplate, perMachineWaitTimeoutFlag, err)
	}

	return []clustermanager.ClusterManagerOpt{
		clustermanager.WithControlPlaneWaitTimeout(cpWaitTimeout),
		clustermanager.WithExternalEtcdWaitTimeout(externalEtcdWaitTimeout),
		clustermanager.WithMachineMaxWait(perMachineWaitTimeout),
	}, nil
}

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

func markFlagHidden(cmd *cobra.Command, flagName string) {
	if err := cmd.Flags().MarkHidden(flagName); err != nil {
		logger.V(5).Info("Warning: Failed to mark flag as hidden: " + flagName)
	}
}
