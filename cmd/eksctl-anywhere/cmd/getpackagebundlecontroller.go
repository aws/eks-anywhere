package cmd

import (
	"github.com/spf13/cobra"
)

type getPackageBundleControllerOptions struct {
	output string
}

var gpbco = &getPackageBundleControllerOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleControllerCommand)
	getPackageBundleControllerCommand.Flags().StringVarP(&gpbco.output, "output", "o", "", "Specifies the output format (valid option: json, yaml)")
}

var getPackageBundleControllerCommand = &cobra.Command{
	Use:          "packagebundlecontroller(s) [flags]",
	Aliases:      []string{"packagebundlecontroller", "packagebundlcontrolleres", "pbc"},
	Short:        "Get packagebundlecontroller(s)",
	Long:         "This command is used to display the currently packagebundlecontrollers",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getResources(cmd.Context(), "packagebundlecontrollers", gpbco.output, args)
	},
}
