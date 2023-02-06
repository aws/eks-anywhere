package cmd

import (
	"github.com/spf13/cobra"
)

var nutanixCmd = &cobra.Command{
	Use:   "nutanix",
	Short: "Nutanix commands",
	Long:  "Use eks-a-tool nutanix to run nutanix utilities",
}

func init() {
	rootCmd.AddCommand(nutanixCmd)
}
