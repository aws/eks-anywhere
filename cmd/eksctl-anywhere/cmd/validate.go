package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

type validateOptions struct {
	clusterOptions
}

var valOpt = &validateOptions{}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "This command is used to validate eksctl anywhere configurations",
	RunE:  valOpt.validateCluster,
}

func init() {
	expCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&valOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")

	if err := validateCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (valOpt *validateOptions) validateCluster(cmd *cobra.Command, _ []string) error {
	return nil
}
