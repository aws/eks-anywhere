package cmd

import "github.com/spf13/cobra"

var createPackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getResources(cmd.Context(), "packages", gpo.output, args)
	},
}
