package cmd

import (
	"github.com/spf13/cobra"
)

var vsphereSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup vSphere objects",
	Long:  "Use eksctl anywhere vsphere setup to configure vSphere objects",
}

func init() {
	vsphereCmd.AddCommand(vsphereSetupCmd)
}
