package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

type generatePackageOptions struct {
	directory string
}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().StringVarP(&gepo.directory, "directory", "d", "", "Directory to save generated packages")
	err := generatePackageCommand.MarkFlagRequired("directory")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var gepo = &generatePackageOptions{}

var generatePackageCommand = &cobra.Command{
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
