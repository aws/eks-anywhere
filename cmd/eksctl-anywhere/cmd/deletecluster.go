package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteClusterOptions struct {
	fileName     string
	wConfig      string
	forceCleanup bool
}

func (dc *deleteClusterOptions) kubeConfig(clusterName string) string {
	if dc.wConfig == "" {
		return filepath.Join(clusterName, fmt.Sprintf(kubeconfigPattern, clusterName))
	}
	return dc.wConfig
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
}

func (dc *deleteClusterOptions) validate(ctx context.Context, args []string) error {
	if dc.fileName == "" {
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			return fmt.Errorf("please provide either a valid <cluster-name> or -f <config-file>")
		}
		filename := fmt.Sprintf("%s/%s-eks-a-cluster.yaml", clusterName, clusterName)
		if !validations.FileExists(filename) {
			return fmt.Errorf("clusterconfig file %s for cluster: %s not found, please provide the clusterconfig path manually using -f <config-file>", filename, clusterName)
		}
		dc.fileName = filename
	}
	clusterConfig, err := commonValidation(ctx, dc.fileName)
	if err != nil {
		return err
	}
	if !validations.KubeConfigExists(clusterConfig.Name, clusterConfig.Name, dc.wConfig, kubeconfigPattern) {
		return fmt.Errorf("KubeConfig doesn't exists for cluster %s", clusterConfig.Name)
	}
	return nil
}

func (dc *deleteClusterOptions) deleteCluster(ctx context.Context) error {
	clusterSpec, err := cluster.NewSpec(dc.fileName, version.Get())
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).
		WithBootstrapper().
		WithClusterManager().
		WithProvider(dc.fileName, clusterSpec.Cluster, cc.skipIpCheck).
		WithFluxAddonClient(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig).
		WithWriter().
		Build()
	if err != nil {
		return err
	}

	deleteCluster := workflows.NewDelete(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.FluxAddonClient,
	)

	// Initialize Workload cluster type
	workloadCluster := &types.Cluster{
		Name:           clusterSpec.Name,
		KubeconfigFile: dc.kubeConfig(clusterSpec.Name),
	}
	err = deleteCluster.Run(ctx, workloadCluster, clusterSpec, viper.GetBool("force-cleanup"))
	if err == nil {
		deps.Writer.CleanUp()
	}
	return err
}
