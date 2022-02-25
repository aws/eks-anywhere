package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/dependencies"
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
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:          "cluster -f <cluster-config-file> [flags]",
	Short:        "Create workload cluster",
	Long:         "This command is used to create workload clusters",
	PreRunE:      preRunCreateCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cc.validate(cmd.Context()); err != nil {
			return err
		}
		if err := cc.createCluster(cmd); err != nil {
			return fmt.Errorf("failed to create cluster: %v", err)
		}
		return nil
	},
}

func init() {
	createCmd.AddCommand(createClusterCmd)
	createClusterCmd.Flags().StringVarP(&cc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	if features.IsActive(features.TinkerbellProvider()) {
		createClusterCmd.Flags().StringVarP(&cc.hardwareFileName, "hardwarefile", "w", "", "Filename that contains datacenter hardware information")
	}
	createClusterCmd.Flags().BoolVar(&cc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	createClusterCmd.Flags().BoolVar(&cc.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	createClusterCmd.Flags().StringVar(&cc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	createClusterCmd.Flags().StringVar(&cc.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
	err := createClusterCmd.MarkFlagRequired("filename")
	if err != nil {
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

func (cc *createClusterOptions) validate(ctx context.Context) error {
	clusterConfig, err := commonValidation(ctx, cc.fileName)
	if err != nil {
		return err
	}

	kubeconfigPath := kubeconfig.FromClusterName(clusterConfig.Name)
	if validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return kubeconfig.NewMissingFileError(clusterConfig.Name, kubeconfigPath)
	}

	return nil
}

func (cc *createClusterOptions) createCluster(cmd *cobra.Command) error {
	ctx := cmd.Context()

	clusterSpec, err := newClusterSpec(cc.clusterOptions)
	if err != nil {
		return err
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(cc.mountDirs()...).
		WithBootstrapper().
		WithClusterManager(clusterSpec.Cluster).
		WithProvider(cc.fileName, clusterSpec.Cluster, cc.skipIpCheck, cc.hardwareFileName).
		WithFluxAddonClient(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig).
		WithWriter().
		Build(ctx)
	if err != nil {
		return err
	}
	defer cleanup(ctx, deps, &err)

	if !features.IsActive(features.TinkerbellProvider()) && deps.Provider.Name() == "tinkerbell" {
		return fmt.Errorf("Error: provider tinkerbell is not supported in this release")
	}

	if deps.Provider.Name() == "tinkerbell" {
		flag := cmd.Flags().Lookup("hardwarefile")
		if flag == nil {
			return fmt.Errorf("Something wrong. Flag hardwarefile not set up for provider tinkerbell")
		}
		if !viper.IsSet("hardwarefile") || viper.GetString("hardwarefile") == "" {
			return fmt.Errorf("Error: required flag \"hardwarefile\" not set")
		}
		hardwareConfigFileExist := validations.FileExists(cc.hardwareFileName)
		if !hardwareConfigFileExist {
			return fmt.Errorf("Error: hardware config file %s does not exist", cc.hardwareFileName)
		}
	}

	createCluster := workflows.NewCreate(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.FluxAddonClient,
		deps.Writer,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Name),
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
			Name:           clusterSpec.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Name),
		},
		ManagementCluster: cluster,
		Provider:          deps.Provider,
	}
	createValidations := createvalidations.New(validationOpts)

	return createCluster.Run(ctx, clusterSpec, createValidations, cc.forceClean)
}
