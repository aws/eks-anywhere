package cmd

import (
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe resources",
	Long:  "Use eksctl anywhere describe to show details of a specific resource or group of resources",
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
