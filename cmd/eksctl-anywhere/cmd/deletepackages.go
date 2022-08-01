package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

func init() {
	deleteCmd.AddCommand(deletePackageCommand)
}

var deletePackageCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
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
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	err = packages.DeletePackages(ctx, args, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
