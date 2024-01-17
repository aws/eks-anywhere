package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type deletePackageOptions struct {
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	clusterName     string
	bundlesOverride string
}

var delPkgOpts = deletePackageOptions{}

func init() {
	deleteCmd.AddCommand(deletePackageCommand)

	deletePackageCommand.Flags().StringVar(&delPkgOpts.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file to use.")
	deletePackageCommand.Flags().StringVar(&delPkgOpts.clusterName, "cluster", "",
		"Cluster for package deletion.")
	deletePackageCommand.Flags().StringVar(&delPkgOpts.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
	if err := deletePackageCommand.MarkFlagRequired("cluster"); err != nil {
		log.Fatalf("marking cluster flag as required: %s", err)
	}
}

var deletePackageCommand = &cobra.Command{
	Use:          "package [flags]",
	Aliases:      []string{"packages"},
	Short:        "Delete package(s)",
	Long:         "This command is used to delete the curated packages custom resources installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteResources(cmd.Context(), args)
	},
	Args: cobra.MinimumNArgs(1),
}

func deleteResources(ctx context.Context, args []string) error {
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(delPkgOpts.kubeConfig, "")
	if err != nil {
		return err
	}
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(delPkgOpts.bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	err = packages.DeletePackages(ctx, args, kubeConfig, delPkgOpts.clusterName)
	if err != nil {
		return err
	}

	return nil
}
