package cmd

import (
	"github.com/spf13/cobra"
)

var conformanceCmd = &cobra.Command{
	Use:   "conformance",
	Short: "Conformance tests",
	Long:  "Use eks-a-tool conformance to run conformance tests",
}

func init() {
	rootCmd.AddCommand(conformanceCmd)
}
