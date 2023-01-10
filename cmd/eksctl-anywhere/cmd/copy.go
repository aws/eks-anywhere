package cmd

import (
	"github.com/spf13/cobra"
)

// importCmd represents the import command.
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy resources",
	Long:  "Copy EKS Anywhere resources and artifacts",
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
