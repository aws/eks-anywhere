package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the eks-a version",
	Long:  "This command prints the version of eks-a",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion() error {
	fmt.Println(version.Get().GitVersion)
	return nil
}
