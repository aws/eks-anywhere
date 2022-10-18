package cmd

import (
	"github.com/spf13/cobra"
)

var vsphereCmd = &cobra.Command{
	Use:   "vsphere",
	Short: "Utility vsphere operations",
	Long:  "Use eksctl anywhere vsphere to perform utility operations on vsphere",
}

func init() {
	expCmd.AddCommand(vsphereCmd)
}
