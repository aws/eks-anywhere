package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources",
	Long:  "Use eksctl anywhere list to list images and artifacts used by EKS Anywhere",
}

func init() {
	rootCmd.AddCommand(listCmd)
}
