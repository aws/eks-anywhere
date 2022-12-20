package cmd

import (
	"github.com/spf13/cobra"
)

// importCmd represents the import command.
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import resources",
	Long:  "Use eksctl anywhere import to import resources, such as images and helm charts",
}

func init() {
	rootCmd.AddCommand(importCmd)
}
