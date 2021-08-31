package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the eksctl anywhere version",
	Long:  "This command prints the version of eksctl anywhere",
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
