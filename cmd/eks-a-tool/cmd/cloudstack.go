package cmd

import (
	"github.com/spf13/cobra"
)

var cloudstackCmd = &cobra.Command{
	Use:   "cloudstack",
	Short: "CloudStack commands",
	Long:  "Use eks-a-tool cloudstack to run cloudstack utilities",
}

func init() {
	rootCmd.AddCommand(cloudstackCmd)
}
