package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
	Long:  "Use eksctl anywhere create to create resources, such as clusters",
}

func init() {
	rootCmd.AddCommand(createCmd)
}
