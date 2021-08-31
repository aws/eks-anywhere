package cmd

import (
	"github.com/spf13/cobra"
)

var integrationTestCmd = &cobra.Command{
	Use:   "e2e",
	Short: "Integration test",
	Long:  "Run integration test on eks-d",
}

func init() {
	rootCmd.AddCommand(integrationTestCmd)
}
