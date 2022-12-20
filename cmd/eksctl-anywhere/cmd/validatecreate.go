package cmd

import (
	"github.com/spf13/cobra"
)

var validateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Validate create resources",
	Long:  "Use eksctl anywhere validate create to validate the create action on resources, such as cluster",
}

func init() {
	validateCmd.AddCommand(validateCreateCmd)
}
