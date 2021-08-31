package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Long:  "Use eks-a delete to delete clusters",
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
