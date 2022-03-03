package cmd

import (
	"github.com/spf13/cobra"
)

type getPackageBundleOptions struct {
	output string
}

var gpbo = &getPackageBundleOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleCommand)
	getPackageBundleCommand.Flags().StringVarP(&gpbo.output, "output", "o", "", "Specifies the output format (valid option: json, yaml)")
}

var getPackageBundleCommand = &cobra.Command{
	Use:          "packagebundle(s) [flags]",
	Aliases:      []string{"packagebundle", "packagebundles"},
	Short:        "Get packagebundle(s)",
	Long:         "This command is used to display the currently supported packagebundles",
	PreRunE:      preRunGetPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getResources(cmd.Context(), "packagebundles", gpbo.output, args)
	},
}
