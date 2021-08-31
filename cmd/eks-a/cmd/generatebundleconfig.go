package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	support "github.com/aws/eks-anywhere/pkg/support"
)

var generateBundleConfigCmd = &cobra.Command{
	Use:   "support-bundle-config",
	Short: "Generate support bundle config",
	Long:  "This command is used to generate a default support bundle config yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := support.GenerateBundleConfig()
		if err != nil {
			return fmt.Errorf("failed to generate bunlde config: %v", err)
		}
		return nil
	},
}

func init() {
	generateCmd.AddCommand(generateBundleConfigCmd)
}
