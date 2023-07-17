package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type deleteClusterOptions struct {
	clusterOptions
	wConfig               string
	forceCleanup          bool
	hardwareFileName      string
	tinkerbellBootstrapIP string
}

var dc = &deleteClusterOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:          "cluster (<cluster-name>|-f <config-file>)",
	Short:        "Workload cluster",
	Long:         "This command is used to delete workload clusters created by eksctl anywhere",
	PreRunE:      bindFlagsToViper,
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
	if err := kubeconfig.ValidateFilename(kubeconfigPath); err != nil {
		return err
	}

	return nil
}

func (dc *deleteClusterOptions) deleteCluster(ctx context.Context) error {
	clusterSpec, err := newClusterSpec(dc.clusterOptions)
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	if err := validations.ValidateAuthenticationForRegistryMirror(clusterSpec); err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := dc.directoriesToMount(clusterSpec, cliConfig)
	if err != nil {
		return err
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, nil).
		WithProvider(dc.fileName, clusterSpec.Cluster, cc.skipIpCheck, dc.hardwareFileName, false, dc.tinkerbellBootstrapIP, map[string]bool{}).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	deleteCluster := workflows.NewDelete(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.GitOpsFlux,
		deps.Writer,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:               clusterSpec.Cluster.Name,
			KubeconfigFile:     kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
			ExistingManagement: false,
		}
	} else {
		cluster = &types.Cluster{
			Name:               clusterSpec.Cluster.Name,
			KubeconfigFile:     clusterSpec.ManagementCluster.KubeconfigFile,
			ExistingManagement: true,
		}
	}

	err = deleteCluster.Run(ctx, cluster, clusterSpec, dc.forceCleanup, dc.managementKubeconfig)
	cleanup(deps, &err)
	return err
}
