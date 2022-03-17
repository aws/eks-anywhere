package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteClusterOptions struct {
	clusterOptions
	wConfig          string
	forceCleanup     bool
	hardwareFileName string
}

var dc = &deleteClusterOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:          "cluster (<cluster-name>|-f <config-file>)",
	Short:        "Workload cluster",
	Long:         "This command is used to delete workload clusters created by eksctl anywhere",
	PreRunE:      preRunDeleteCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := dc.validate(cmd.Context(), args); err != nil {
			return err
		}
		if err := dc.deleteCluster(cmd.Context()); err != nil {
			return fmt.Errorf("failed to delete cluster: %v", err)
		}
		return nil
	},
}

func preRunDeleteCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().StringVarP(&dc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration, required if <cluster-name> is not provided")
	deleteClusterCmd.Flags().StringVarP(&dc.wConfig, "w-config", "w", "", "Kubeconfig file to use when deleting a workload cluster")
	deleteClusterCmd.Flags().BoolVar(&dc.forceCleanup, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	deleteClusterCmd.Flags().StringVar(&dc.managementKubeconfig, "kubeconfig", "", "kubeconfig file pointing to a management cluster")
	deleteClusterCmd.Flags().StringVar(&dc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
}

func (dc *deleteClusterOptions) validate(ctx context.Context, args []string) error {
	if dc.fileName == "" {
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			return fmt.Errorf("please provide either a valid <cluster-name> or -f <config-file>")
		}
		filename := fmt.Sprintf("%[1]s/%[1]s-eks-a-cluster.yaml", clusterName)
		if !validations.FileExists(filename) {
			return fmt.Errorf("clusterconfig file %s for cluster: %s not found, please provide the clusterconfig path manually using -f <config-file>", filename, clusterName)
		}
		dc.fileName = filename
	}
	clusterConfig, err := commonValidation(ctx, dc.fileName)
	if err != nil {
		return err
	}

	kubeconfigPath := getKubeconfigPath(clusterConfig.Name, dc.wConfig)
	if !validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return kubeconfig.NewMissingFileError(kubeconfigPath)
	}

	return nil
}

func (dc *deleteClusterOptions) deleteCluster(ctx context.Context) error {
	clusterSpec, err := newClusterSpec(dc.clusterOptions)
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(cc.mountDirs()...).
		WithBootstrapper().
		WithClusterManager(clusterSpec.Cluster).
		WithProvider(dc.fileName, clusterSpec.Cluster, cc.skipIpCheck, dc.hardwareFileName, cc.skipPowerActions).
		WithFluxAddonClient(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig).
		WithWriter().
		Build(ctx)
	if err != nil {
		return err
	}
	defer cleanup(ctx, deps, &err)

	if !features.IsActive(features.CloudStackProvider()) && deps.Provider.Name() == constants.CloudStackProviderName {
		return fmt.Errorf("Error: provider cloudstack is not supported in this release")
	}
	if !features.IsActive(features.TinkerbellProvider()) && deps.Provider.Name() == "tinkerbell" {
		return fmt.Errorf("Error: provider tinkerbell is not supported in this release")
	}

	deleteCluster := workflows.NewDelete(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.FluxAddonClient,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	err = deleteCluster.Run(ctx, cluster, clusterSpec, dc.forceCleanup, dc.managementKubeconfig)
	return err
}
