package cmd

import (
	"github.com/spf13/cobra"
)

var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "experimental commands",
	Long:  "Use eksctl anywhere experimental commands",
}

func init() {
	rootCmd.AddCommand(expCmd)
}
