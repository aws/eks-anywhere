package cmd

import (
	"github.com/spf13/cobra"
)

var e2eCmd = &cobra.Command{
	Use:   "e2e",
	Short: "e2e test interaction",
	Long:  "Interact with and debug end-to-end tests",
}

func init() {
	rootCmd.AddCommand(e2eCmd)
}
