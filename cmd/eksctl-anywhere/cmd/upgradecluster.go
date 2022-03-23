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
	wConfig          string
	forceClean       bool
	hardwareFileName string
}

var uc = &upgradeClusterOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Upgrade workload cluster",
	Long:         "This command is used to upgrade workload clusters",
	PreRunE:      preRunUpgradeCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradeCluster(cmd.Context()); err != nil {
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
	err := upgradeClusterCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradeCluster(ctx context.Context) error {
	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}
	clusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(cc.mountDirs()...).
		WithBootstrapper().
		WithClusterManager(clusterSpec.Cluster).
		WithProvider(uc.fileName, clusterSpec.Cluster, cc.skipIpCheck, uc.hardwareFileName, cc.skipPowerActions).
		WithFluxAddonClient(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig).
		WithWriter().
		WithCAPIManager().
		WithKubectl().
		Build(ctx)
	if err != nil {
		return err
	}
	defer cleanup(ctx, deps, &err)

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
	)

	workloadCluster := &types.Cluster{
		Name:           clusterSpec.Cluster.Name,
		KubeconfigFile: getKubeconfigPath(clusterSpec.Cluster.Name, uc.wConfig),
	}

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: getKubeconfigPath(clusterSpec.Cluster.Name, uc.wConfig),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.ManagementCluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	validationOpts := &validations.Opts{
		Kubectl:           deps.Kubectl,
		Spec:              clusterSpec,
		WorkloadCluster:   workloadCluster,
		ManagementCluster: cluster,
		Provider:          deps.Provider,
	}
	upgradeValidations := upgradevalidations.New(validationOpts)

	err = upgradeCluster.Run(ctx, clusterSpec, cluster, upgradeValidations, uc.forceClean)
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
