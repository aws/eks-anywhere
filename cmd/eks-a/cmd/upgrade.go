package cmd

import (
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade resources",
	Long:  "Use eks-a upgrade to upgrade resources, such as clusters",
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
