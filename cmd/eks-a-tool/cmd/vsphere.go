package cmd

import (
	"github.com/spf13/cobra"
)

var vsphereCmd = &cobra.Command{
	Use:   "vsphere",
	Short: "VSphere commands",
	Long:  "Use eks-a-tool vsphere to run vsphere utilities",
}

func init() {
	rootCmd.AddCommand(vsphereCmd)
}
