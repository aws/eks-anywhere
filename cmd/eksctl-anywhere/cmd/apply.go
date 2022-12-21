package cmd

import "github.com/spf13/cobra"

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply resources",
	Long:  "Use eksctl anywhere apply to apply resources",
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
