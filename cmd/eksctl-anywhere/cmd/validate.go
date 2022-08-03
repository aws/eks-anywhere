package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/cmdvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

type validateOptions struct {
	clusterOptions
	skipIpCheck           bool
	hardwareCSVPath       string
	tinkerbellBootstrapIP string
}

var valOpt = &validateOptions{}

var validateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Validate configuration",
	Long:         "This command is used to validate eksctl anywhere configurations",
	PreRunE:      preRunValidate,
	SilenceUsage: true,
	RunE:         valOpt.validateCluster,
}

func init() {
	expCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&valOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	validateCmd.Flags().StringVarP(
		&valOpt.hardwareCSVPath,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)
	validateCmd.Flags().StringVar(&valOpt.tinkerbellBootstrapIP, "tinkerbell-bootstrap-ip", "", "Override the local tinkerbell IP in the bootstrap cluster")
	validateCmd.Flags().BoolVar(&valOpt.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	validateCmd.Flags().StringVar(&valOpt.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	validateCmd.Flags().StringVar(&valOpt.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")

	if err := validateCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunValidate(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (valOpt *validateOptions) validateCluster(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	runner := validations.NewRunner()
	runner.Register(cmdvalidations.PackageDockerValidations(ctx)...)
	runner.StoreValidationResults()

	// Config parse
	clusterConfig, err := v1alpha1.GetClusterConfig(valOpt.fileName)
	if err != nil {
		return runner.ExitError(err)
	}
	runner.Register(cmdvalidations.PackageClusterValidation(clusterConfig)...)
	err = runner.StoreValidationResults()
	if err != nil {
		return runner.ExitError(err)
	}

	runner.Register(cmdvalidations.PackageKubeConfigPath(clusterConfig.Name)...)

	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		flag := cmd.Flags().Lookup(TinkerbellHardwareCSVFlagName)

		// If no flag was returned there is a developer error as the flag has been removed
		// from the program rendering it invalid.
		if flag == nil {
			runner.ReportResults()
			panic("'hardwarefile' flag not configured")
		}

		if !viper.IsSet(TinkerbellHardwareCSVFlagName) || viper.GetString(TinkerbellHardwareCSVFlagName) == "" {
			return runner.ExitError(fmt.Errorf("required flag \"%v\" not set", TinkerbellHardwareCSVFlagName))
		}

		if !validations.FileExists(cc.hardwareCSVPath) {
			return runner.ExitError(fmt.Errorf("hardware config file %s does not exist", cc.hardwareCSVPath))
		}
	}

	runner.StoreValidationResults()
	if err != nil {
		return runner.ExitError(err)
	}

	clusterSpec, err := newClusterSpec(valOpt.clusterOptions)
	if err != nil {
		return runner.ExitError(err)
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := valOpt.directoriesToMount(clusterSpec, cliConfig)
	if err != nil {
		return runner.ExitError(err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithKubectl().
		WithProvider(valOpt.fileName, clusterSpec.Cluster, valOpt.skipIpCheck, valOpt.hardwareCSVPath, true, valOpt.tinkerbellBootstrapIP).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		Build(ctx)
	if err != nil {
		return runner.ExitError(err)
	}
	defer close(ctx, deps)

	runner.Register(cmdvalidations.PackageActiveProvider(deps.Provider)...)
	err = runner.StoreValidationResults()
	if err != nil {
		return runner.ExitError(err)
	}

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

	runner.Register(cmdvalidations.PackageCreatePreflight(ctx, createValidations)...)
	runner.Register(cmdvalidations.PackageProviderValidations(ctx, clusterSpec, deps.Provider)...)
	runner.Register(deps.GitOpsFlux.Validations(ctx, clusterSpec)...)

	runner.StoreValidationResults()
	runner.ReportResults()

	cleanup(deps, &err)
	return err
}

func (valOpt *validateOptions) directoriesToMount(clusterSpec *cluster.Spec, cliConfig *config.CliConfig) ([]string, error) {
	dirs := valOpt.mountDirs()
	fluxConfig := clusterSpec.FluxConfig
	if fluxConfig != nil && fluxConfig.Spec.Git != nil {
		dirs = append(dirs, filepath.Dir(cliConfig.GitPrivateKeyFile))
		dirs = append(dirs, filepath.Dir(cliConfig.GitKnownHostsFile))
		dirs = append(dirs, filepath.Dir(cc.installPackages))
	}

	if clusterSpec.Config.Cluster.Spec.DatacenterRef.Kind == v1alpha1.CloudStackDatacenterKind {
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
	}

	return dirs, nil
}
