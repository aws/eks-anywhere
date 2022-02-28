package cmd

import (
	"github.com/spf13/cobra"
)

type getPackageOptions struct {
	output string
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)
	getPackageCommand.Flags().StringVarP(&gpo.output, "output", "o", "", "Specifies the output format (valid option: json, yaml)")
}

var getPackageCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunGetPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return GetResources(cmd.Context(), "packages", gpo.output, args)
	},
}
