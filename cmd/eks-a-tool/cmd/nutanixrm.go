package cmd

import (
	"github.com/spf13/cobra"
)

var nutanixRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Nutanix rm commands",
	Long:  "Use eks-a-tool nutanix rm to run nutanix rm utilities",
}

func init() {
	nutanixCmd.AddCommand(nutanixRmCmd)
}
