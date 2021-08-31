package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/pkg/conformance"
)

var conformanceDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Conformance download command",
	Long:  "This command downloads the conformance test suite",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := conformance.Download()
		if err != nil {
			log.Fatalf("Error downloading conformance: %v", err)
		}
		return nil
	},
}

func init() {
	conformanceCmd.AddCommand(conformanceDownloadCmd)
	err := viper.BindPFlags(conformanceDownloadCmd.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
