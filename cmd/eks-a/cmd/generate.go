package cmd

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate resources",
	Long:  "Use eks-a generate to generate resources, such as clusterconfig yaml",
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
