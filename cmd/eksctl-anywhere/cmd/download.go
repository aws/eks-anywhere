package cmd

import (
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download resources",
	Long:  "Use eksctl anywhere download to download artifacts (manifests, bundles) used by EKS Anywhere",
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
