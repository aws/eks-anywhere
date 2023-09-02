package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
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
		if uc.wConfig != "" {
			logger.MarkFail(wConfigDeprecationMessage)
			return errors.New("--w-config is deprecated. Use --kubeconfig instead")
		}
		return nil
	},
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().StringVarP(&dc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration, required if <cluster-name> is not provided")
	deleteClusterCmd.Flags().StringVarP(&dc.wConfig, "w-config", "w", "", "Kubeconfig file to use when deleting a workload cluster")
	err := deleteClusterCmd.Flags().MarkDeprecated("w-config", "please use flag --kubeconfig.")
	if err != nil {
		log.Fatalf("Error deprecating flag as required: %v", err)
	}
	deleteClusterCmd.Flags().BoolVar(&dc.forceCleanup, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	hideForceCleanup(deleteClusterCmd.Flags())
	deleteClusterCmd.Flags().StringVar(&dc.clusterKubeconfig, "kubeconfig", "", "kubeconfig file pointing to a management cluster")
	deleteClusterCmd.Flags().StringVar(&dc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
}

func (dc *deleteClusterOptions) validate(ctx context.Context, args []string) error {
	if dc.forceCleanup {
		logger.MarkFail(forceCleanupDeprecationMessageForCreateDelete)
		return errors.New("please remove the --force-cleanup flag")
	}
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

	kubeconfigPath := getKubeconfigPath(clusterConfig.Name, dc.clusterKubeconfig)
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
	var kubeconfigFile string
	if clusterSpec.ManagementCluster == nil {
		kubeconfigFile = kubeconfig.FromClusterName(clusterSpec.Cluster.Name)
		if dc.clusterKubeconfig != "" {
			kubeconfigFile = dc.clusterKubeconfig
		}
		cluster = &types.Cluster{
			Name:               clusterSpec.Cluster.Name,
			KubeconfigFile:     kubeconfigFile,
			ExistingManagement: false,
		}
	} else {
		kubeconfigFile = clusterSpec.ManagementCluster.KubeconfigFile
		if dc.clusterKubeconfig != "" {
			kubeconfigFile = dc.clusterKubeconfig
		}
		cluster = &types.Cluster{
			Name:               clusterSpec.Cluster.Name,
			KubeconfigFile:     kubeconfigFile,
			ExistingManagement: true,
		}
	}

	err = deleteCluster.Run(ctx, cluster, clusterSpec, dc.forceCleanup, dc.clusterKubeconfig)
	cleanup(deps, &err)
	return err
}
