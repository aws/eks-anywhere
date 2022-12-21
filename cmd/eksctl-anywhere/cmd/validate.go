package cmd

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate resource or action",
	Long:  "Use eksctl anywhere validate to validate a resource or action",
}

func init() {
	expCmd.AddCommand(validateCmd)
}
