package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.AddCommand(updateMakefileCmd)
	triggerCmd.AddCommand(updateProwCmd)
	triggerCmd.AddCommand(stageBundleCmd)
	triggerCmd.AddCommand(stageCliCmd)
	triggerCmd.AddCommand(prodBundleCmd)
	triggerCmd.AddCommand(prodCliCmd)
	triggerCmd.AddCommand(createBranchCmd)
	triggerCmd.AddCommand(updateHomebrewCmd)
	triggerCmd.AddCommand(createReleaseCmd)
}
