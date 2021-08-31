package cmd

import (
	"github.com/spf13/cobra"
)

var vsphereRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "VSphere rm commands",
	Long:  "Use eks-a-tool vsphere rm to run vsphere rm utilities",
}

func init() {
	vsphereCmd.AddCommand(vsphereRmCmd)
}
