package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Long:  "Use eksctl anywhere delete to delete clusters",
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
