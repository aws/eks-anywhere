package cmd

import (
	"github.com/spf13/cobra"
)

var cleanUpInstancesCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up e2e resources",
	Long:  "Clean up resources created for e2e testing",
}

func init() {
	integrationTestCmd.AddCommand(cleanUpInstancesCmd)
}
