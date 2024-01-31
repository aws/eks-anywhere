package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

const defaultTinkerbellNodeStartupTimeout = 20 * time.Minute

const timeoutErrorTemplate = "failed to parse timeout %s: %v"

type timeoutOptions struct {
	cpWaitTimeout           string
	externalEtcdWaitTimeout string
	perMachineWaitTimeout   string
	unhealthyMachineTimeout string
	nodeStartupTimeout      string
	noTimeouts              bool
}

func applyTimeoutFlags(flagSet *pflag.FlagSet, t *timeoutOptions) {
	flagSet.StringVar(&t.cpWaitTimeout, cpWaitTimeoutFlag, clustermanager.DefaultControlPlaneWait.String(), "Override the default control plane wait timeout")
	flagSet.StringVar(&t.externalEtcdWaitTimeout, externalEtcdWaitTimeoutFlag, clustermanager.DefaultEtcdWait.String(), "Override the default external etcd wait timeout")
	flagSet.StringVar(&t.perMachineWaitTimeout, perMachineWaitTimeoutFlag, clustermanager.DefaultMaxWaitPerMachine.String(), "Override the default machine wait timeout per machine")
	flagSet.StringVar(&t.unhealthyMachineTimeout, unhealthyMachineTimeoutFlag, constants.DefaultUnhealthyMachineTimeout.String(), "(DEPRECATED) Override the default unhealthy machine timeout")
	flagSet.StringVar(&t.nodeStartupTimeout, nodeStartupTimeoutFlag, constants.DefaultNodeStartupTimeout.String(), "(DEPRECATED) Override the default node startup timeout (Defaults to 20m for Tinkerbell clusters)")
	flagSet.BoolVar(&t.noTimeouts, noTimeoutsFlag, false, "Disable timeout for all wait operations")
}

// buildClusterManagerOpts builds options for constructing a ClusterManager from CLI flags.
// datacenterKind is an API kind such as v1alpha1.TinkerbellDatacenterKind.
func buildClusterManagerOpts(t timeoutOptions, datacenterKind string) (*dependencies.ClusterManagerTimeoutOptions, error) {
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

	unhealthyMachineTimeout, err := time.ParseDuration(t.unhealthyMachineTimeout)
	if err != nil {
		return nil, fmt.Errorf(timeoutErrorTemplate, unhealthyMachineTimeoutFlag, err)
	}

	if t.nodeStartupTimeout == clustermanager.DefaultNodeStartupTimeout.String() &&
		datacenterKind == v1alpha1.TinkerbellDatacenterKind {
		t.nodeStartupTimeout = defaultTinkerbellNodeStartupTimeout.String()
	}

	nodeStartupTimeout, err := time.ParseDuration(t.nodeStartupTimeout)
	if err != nil {
		return nil, fmt.Errorf(timeoutErrorTemplate, nodeStartupTimeoutFlag, err)
	}

	return &dependencies.ClusterManagerTimeoutOptions{
		ControlPlaneWait:     cpWaitTimeout,
		ExternalEtcdWait:     externalEtcdWaitTimeout,
		MachineWait:          perMachineWaitTimeout,
		UnhealthyMachineWait: unhealthyMachineTimeout,
		NodeStartupWait:      nodeStartupTimeout,
		NoTimeouts:           t.noTimeouts,
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

func readClusterSpec(clusterConfigPath string, cliVersion version.Info, opts ...cluster.FileSpecBuilderOpt) (*cluster.Spec, error) {
	b := cluster.NewFileSpecBuilder(
		files.NewReader(files.WithEKSAUserAgent("cli", cliVersion.GitVersion)),
		cliVersion,
		opts...,
	)
	return b.Build(clusterConfigPath)
}

func readAndValidateClusterSpec(clusterConfigPath string, cliVersion version.Info, opts ...cluster.FileSpecBuilderOpt) (*cluster.Spec, error) {
	clusterSpec, err := readClusterSpec(clusterConfigPath, cliVersion, opts...)
	if err != nil {
		return nil, err
	}
	if err = cluster.ValidateConfig(clusterSpec.Config); err != nil {
		return nil, err
	}

	return clusterSpec, nil
}

func newClusterSpec(options clusterOptions) (*cluster.Spec, error) {
	var opts []cluster.FileSpecBuilderOpt
	if options.bundlesOverride != "" {
		opts = append(opts, cluster.WithOverrideBundlesManifest(options.bundlesOverride))
	}

	clusterSpec, err := readAndValidateClusterSpec(options.fileName, version.Get(), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	if clusterSpec.Cluster.IsManaged() {
		if options.managementKubeconfig == "" {
			managementKubeconfig, err := getManagementClusterKubeconfig(clusterSpec.Cluster.Spec.ManagementCluster.Name)
			if err != nil {
				return nil, err
			}
			options.managementKubeconfig = managementKubeconfig
		}
		managementCluster, err := cluster.LoadManagement(options.managementKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("unable to get management cluster from kubeconfig: %v", err)
		}
		clusterSpec.ManagementCluster = managementCluster
	}

	return clusterSpec, nil
}

func markFlagHidden(flagSet *pflag.FlagSet, flagName string) {
	if err := flagSet.MarkHidden(flagName); err != nil {
		logger.V(5).Info("Warning: Failed to mark flag as hidden: " + flagName)
	}
}

func buildCliConfig(clusterSpec *cluster.Spec) *config.CliConfig {
	cliConfig := &config.CliConfig{}
	if clusterSpec.FluxConfig != nil && clusterSpec.FluxConfig.Spec.Git != nil {
		cliConfig.GitSshKeyPassphrase = os.Getenv(config.EksaGitPassphraseTokenEnv)
		cliConfig.GitPrivateKeyFile = os.Getenv(config.EksaGitPrivateKeyTokenEnv)
		cliConfig.GitKnownHostsFile = os.Getenv(config.EksaGitKnownHostsFileEnv)
	}

	return cliConfig
}

func buildCreateCliConfig(clusterOptions *createClusterOptions) (*config.CreateClusterCLIConfig, error) {
	createCliConfig := &config.CreateClusterCLIConfig{}
	createCliConfig.SkipCPIPCheck = clusterOptions.skipIpCheck
	if clusterOptions.noTimeouts {
		maxTime := time.Duration(math.MaxInt64)
		createCliConfig.NodeStartupTimeout = maxTime
		createCliConfig.UnhealthyMachineTimeout = maxTime

		return createCliConfig, nil
	}

	unhealthyMachineTimeout, err := time.ParseDuration(clusterOptions.unhealthyMachineTimeout)
	if err != nil {
		return nil, err
	}

	nodeStartupTimeout, err := time.ParseDuration(clusterOptions.nodeStartupTimeout)
	if err != nil {
		return nil, err
	}

	createCliConfig.NodeStartupTimeout = nodeStartupTimeout
	createCliConfig.UnhealthyMachineTimeout = unhealthyMachineTimeout

	return createCliConfig, nil
}

func buildUpgradeCliConfig(clusterOptions *upgradeClusterOptions) (*config.UpgradeClusterCLIConfig, error) {
	upgradeCliConfig := config.UpgradeClusterCLIConfig{}
	if clusterOptions.noTimeouts {
		maxTime := time.Duration(math.MaxInt64)
		upgradeCliConfig.NodeStartupTimeout = maxTime
		upgradeCliConfig.UnhealthyMachineTimeout = maxTime

		return &upgradeCliConfig, nil
	}

	unhealthyMachineTimeout, err := time.ParseDuration(clusterOptions.unhealthyMachineTimeout)
	if err != nil {
		return nil, err
	}

	nodeStartupTimeout, err := time.ParseDuration(clusterOptions.nodeStartupTimeout)
	if err != nil {
		return nil, err
	}

	upgradeCliConfig.NodeStartupTimeout = nodeStartupTimeout
	upgradeCliConfig.UnhealthyMachineTimeout = unhealthyMachineTimeout

	return &upgradeCliConfig, nil
}

func buildDeleteCliConfig() *config.DeleteClusterCLIConfig {
	deleteCliConfig := &config.DeleteClusterCLIConfig{
		ClusterNamespace: "default",
	}
	return deleteCliConfig
}

func getManagementClusterKubeconfig(clusterName string) (string, error) {
	envKubeconfig := kubeconfig.FromEnvironment()
	if envKubeconfig != "" {
		return envKubeconfig, nil
	}
	// check if kubeconfig for management cluster exists locally
	managementKubeconfigPath := kubeconfig.FromClusterName(clusterName)
	if validations.FileExistsAndIsNotEmpty(managementKubeconfigPath) {
		return managementKubeconfigPath, nil
	}
	return "", fmt.Errorf("management kubeconfig file not found, must be present for workload cluster operations")
}

func getManagementCluster(clusterSpec *cluster.Spec) *types.Cluster {
	if clusterSpec.ManagementCluster == nil {
		return &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}
	} else {
		return &types.Cluster{
			Name:           clusterSpec.ManagementCluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}
}

func (c *clusterOptions) directoriesToMount(clusterSpec *cluster.Spec, cliConfig *config.CliConfig, addDirs ...string) ([]string, error) {
	dirs := c.mountDirs()
	fluxConfig := clusterSpec.FluxConfig
	if fluxConfig != nil && fluxConfig.Spec.Git != nil {
		dirs = append(dirs, filepath.Dir(cliConfig.GitPrivateKeyFile))
		dirs = append(dirs, filepath.Dir(cliConfig.GitKnownHostsFile))
	}

	if clusterSpec.Config.Cluster.Spec.DatacenterRef.Kind == v1alpha1.CloudStackDatacenterKind {
		if extraDirs, err := c.cloudStackDirectoriesToMount(); err == nil {
			dirs = append(dirs, extraDirs...)
		}
	}

	for _, addDir := range addDirs {
		dirs = append(dirs, filepath.Dir(addDir))
	}

	return dirs, nil
}

func (c *clusterOptions) cloudStackDirectoriesToMount() ([]string, error) {
	dirs := []string{}
	env, found := os.LookupEnv(decoder.EksaCloudStackHostPathToMount)
	if found && len(env) > 0 {
		mountDirs := strings.Split(env, ",")
		for _, dir := range mountDirs {
			if _, err := os.Stat(dir); err != nil {
				return nil, fmt.Errorf("invalid host path to mount: %v", err)
			}
			dirs = append(dirs, dir)
		}
	}
	return dirs, nil
}
