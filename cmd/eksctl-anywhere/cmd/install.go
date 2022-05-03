package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install resources to the cluster",
	Long:  "Use eksctl anywhere install to install artifacts into a cluster",
}

func init() {
	rootCmd.AddCommand(installCmd)
}
