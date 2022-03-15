package cmd

import (
	"context"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listPackagesCommand)
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Generate a list of curated packages available to install",
	Long:         "This command is used to generate a list of curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPackages(cmd.Context(), args)
	},
}

func listPackages(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig)
	if err != nil {
		return err
	}
	packages, err := curatedpackages.GetPackages(ctx, bundle, kubeConfig)
	curatedpackages.DisplayPackages(ctx, packages)
	return nil
}
