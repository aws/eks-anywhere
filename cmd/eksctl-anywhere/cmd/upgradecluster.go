package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeClusterOptions struct {
	clusterOptions
	wConfig         string
	forceClean      bool
	hardwareCSVPath string
}

var uc = &upgradeClusterOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Upgrade workload cluster",
	Long:         "This command is used to upgrade workload clusters",
	PreRunE:      preRunUpgradeCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradeCluster(cmd); err != nil {
			return fmt.Errorf("failed to upgrade cluster: %v", err)
		}
		return nil
	},
}

func preRunUpgradeCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func init() {
	upgradeCmd.AddCommand(upgradeClusterCmd)
	upgradeClusterCmd.Flags().StringVarP(&uc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	upgradeClusterCmd.Flags().StringVarP(&uc.wConfig, "w-config", "w", "", "Kubeconfig file to use when upgrading a workload cluster")
	upgradeClusterCmd.Flags().BoolVar(&uc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	upgradeClusterCmd.Flags().StringVar(&uc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	upgradeClusterCmd.Flags().StringVar(&uc.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
	upgradeClusterCmd.Flags().StringVarP(
		&cc.hardwareCSVPath,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)

	if err := upgradeClusterCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradeCluster(cmd *cobra.Command) error {
	ctx := cmd.Context()

	clusterConfigFileExist := validations.FileExists(uc.fileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", uc.fileName)
	}

	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(uc.fileName)
	if err != nil {
		return fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		flag := cmd.Flags().Lookup(TinkerbellHardwareCSVFlagName)

		// If no flag was returned there is a developer error as the flag has been removed
		// from the program rendering it invalid.
		if flag == nil {
			panic("'hardwarefile' flag not configured")
		}

		if len(uc.hardwareCSVPath) != 0 && !validations.FileExists(uc.hardwareCSVPath) {
			return fmt.Errorf("hardware config file %s does not exist", uc.hardwareCSVPath)
		}
	}

	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}
	clusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := cc.directoriesToMount(clusterSpec, cliConfig)
	if err != nil {
		return err
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster).
		WithProvider(uc.fileName, clusterSpec.Cluster, cc.skipIpCheck, uc.hardwareCSVPath, uc.forceClean).
		WithFluxAddonClient(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithCAPIManager().
		WithEksdUpgrader().
		WithEksdInstaller().
		WithKubectl().
		Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	if deps.Provider.Name() == "tinkerbell" {
		return fmt.Errorf("Error: upgrade operation is not supported for provider tinkerbell")
	}

	upgradeCluster := workflows.NewUpgrade(
		deps.Bootstrapper,
		deps.Provider,
		deps.CAPIManager,
		deps.ClusterManager,
		deps.FluxAddonClient,
		deps.Writer,
		deps.EksdUpgrader,
		deps.EksdInstaller,
	)

	workloadCluster := &types.Cluster{
		Name:           clusterSpec.Cluster.Name,
		KubeconfigFile: getKubeconfigPath(clusterSpec.Cluster.Name, uc.wConfig),
	}

	var managementCluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		managementCluster = workloadCluster
	} else {
		managementCluster = clusterSpec.ManagementCluster
	}

	validationOpts := &validations.Opts{
		Kubectl:           deps.Kubectl,
		Spec:              clusterSpec,
		WorkloadCluster:   workloadCluster,
		ManagementCluster: managementCluster,
		Provider:          deps.Provider,
	}
	upgradeValidations := upgradevalidations.New(validationOpts)

	err = upgradeCluster.Run(ctx, clusterSpec, managementCluster, workloadCluster, upgradeValidations, uc.forceClean)
	cleanup(deps, &err)
	return err
}

func (uc *upgradeClusterOptions) commonValidations(ctx context.Context) (cluster *v1alpha1.Cluster, err error) {
	clusterConfig, err := commonValidation(ctx, uc.fileName)
	if err != nil {
		return nil, err
	}

	kubeconfigPath := getKubeconfigPath(clusterConfig.Name, uc.wConfig)
	if !validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return nil, kubeconfig.NewMissingFileError(kubeconfigPath)
	}

	return clusterConfig, nil
}
