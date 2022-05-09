package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createClusterOptions struct {
	clusterOptions
	forceClean       bool
	skipIpCheck      bool
	hardwareFileName string
	skipPowerActions bool
	setupTinkerbell  bool
	installPackages  string
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:          "cluster -f <cluster-config-file> [flags]",
	Short:        "Create workload cluster",
	Long:         "This command is used to create workload clusters",
	PreRunE:      preRunCreateCluster,
	SilenceUsage: true,
	RunE:         cc.createCluster,
}

func init() {
	createCmd.AddCommand(createClusterCmd)
	createClusterCmd.Flags().StringVarP(&cc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	if features.IsActive(features.TinkerbellProvider()) {
		createClusterCmd.Flags().StringVarP(&cc.hardwareFileName, "hardwarefile", "w", "", "Filename that contains datacenter hardware information")
		createClusterCmd.Flags().BoolVar(&cc.skipPowerActions, "skip-power-actions", false, "Skip IPMI power actions on the hardware for Tinkerbell provider")
		if features.IsActive(features.TinkerbellStackSetup()) {
			createClusterCmd.Flags().BoolVar(&cc.setupTinkerbell, "setup-tinkerbell", false, "Setup Tinkerbell stack during baremetal cluster creation")
		}
	}
	createClusterCmd.Flags().BoolVar(&cc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	createClusterCmd.Flags().BoolVar(&cc.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	createClusterCmd.Flags().StringVar(&cc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	createClusterCmd.Flags().StringVar(&cc.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
	createClusterCmd.Flags().StringVar(&cc.installPackages, "install-packages", "", "Location of curated packages configuration files to install to the cluster")

	if err := createClusterCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunCreateCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (cc *createClusterOptions) createCluster(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	clusterConfigFileExist := validations.FileExists(cc.fileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", cc.fileName)
	}

	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(cc.fileName)
	if err != nil {
		return fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		flag := cmd.Flags().Lookup("hardwarefile")

		// If no flag was returned there is a developer error as the flag has been removed
		// from the program rendering it invalid.
		if flag == nil {
			panic("'hardwarefile' flag not configured")
		}

		if !viper.IsSet("hardwarefile") || viper.GetString("hardwarefile") == "" {
			return fmt.Errorf("required flag \"hardwarefile\" not set")
		}

		if !validations.FileExists(cc.hardwareFileName) {
			return fmt.Errorf("hardware config file %s does not exist", cc.hardwareFileName)
		}
	}

	docker := executables.BuildDockerExecutable()

	if err := validations.CheckMinimumDockerVersion(ctx, docker); err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
	}

	if runtime.GOOS == "darwin" {
		if err = validations.CheckDockerDesktopVersion(ctx, docker); err != nil {
			return fmt.Errorf("failed to validate docker desktop: %v", err)
		}
	}

	validations.CheckDockerAllocatedMemory(ctx, docker)

	kubeconfigPath := kubeconfig.FromClusterName(clusterConfig.Name)
	if validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return fmt.Errorf(
			"old cluster config file exists under %s, please use a different clusterName to proceed",
			clusterConfig.Name,
		)
	}

	clusterSpec, err := newClusterSpec(cc.clusterOptions)
	if err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs := cc.directoriesToMount(clusterSpec, cliConfig)

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithClusterManager(clusterSpec.Cluster).
		WithProvider(cc.fileName, clusterSpec.Cluster, cc.skipIpCheck, cc.hardwareFileName, cc.skipPowerActions, cc.setupTinkerbell, cc.forceClean).
		WithFluxAddonClient(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithEksdInstaller().
		Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	if !features.IsActive(features.TinkerbellProvider()) && deps.Provider.Name() == constants.TinkerbellProviderName {
		return fmt.Errorf("provider tinkerbell is not supported in this release")
	}

	if !features.IsActive(features.CloudStackProvider()) && deps.Provider.Name() == constants.CloudStackProviderName {
		return fmt.Errorf("provider cloudstack is not supported in this release")
	}

	if !features.IsActive(features.SnowProvider()) && deps.Provider.Name() == constants.SnowProviderName {
		return fmt.Errorf("provider snow is not supported in this release")
	}

	if !features.IsActive(features.CuratedPackagesSupport()) && cc.installPackages != "" {
		return fmt.Errorf("curated packages installation is not supported in this release")
	}

	createCluster := workflows.NewCreate(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.FluxAddonClient,
		deps.Writer,
		deps.EksdInstaller,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.ManagementCluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	validationOpts := &validations.Opts{
		Kubectl: deps.Kubectl,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster: cluster,
		Provider:          deps.Provider,
		CliConfig:         cliConfig,
	}
	createValidations := createvalidations.New(validationOpts)

	err = createCluster.Run(ctx, clusterSpec, createValidations, cc.forceClean, cc.installPackages)
	cleanup(deps, &err)
	return err
}

func (cc *createClusterOptions) directoriesToMount(clusterSpec *cluster.Spec, cliConfig *config.CliConfig) []string {
	dirs := cc.mountDirs()
	fluxConfig := clusterSpec.FluxConfig
	if fluxConfig == nil || fluxConfig.Spec.Git == nil {
		return dirs
	}
	dirs = append(dirs, filepath.Dir(cliConfig.GitPrivateKeyFile))
	dirs = append(dirs, filepath.Dir(cliConfig.GitKnownHostsFile))

	return dirs
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
