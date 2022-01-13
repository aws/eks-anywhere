package cmd

import (
	"github.com/spf13/cobra"
)

var upgradePlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Provides information for a resource upgrade",
	Long:  "Use eksctl anywhere upgrade plan to get information for a resource upgrade",
}

func init() {
	upgradeCmd.AddCommand(upgradePlanCmd)
}
