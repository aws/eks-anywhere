package cmd

import (
	"github.com/spf13/cobra"
)

var cloudstackRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "CloudStack rm commands",
	Long:  "Use eks-a-tool cloudstack rm to run cloudstack rm utilities",
}

func init() {
	cloudstackCmd.AddCommand(cloudstackRmCmd)
}
