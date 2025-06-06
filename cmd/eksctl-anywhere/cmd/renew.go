package cmd

import (
	"github.com/spf13/cobra"
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "renew resources",
	Long:  "Use eksctl anywhere renew to renew cluster resources",
}

func init() {
	rootCmd.AddCommand(renewCmd)
}
