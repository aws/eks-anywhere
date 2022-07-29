package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/cmdvalidations"
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
	ctx := cmd.Context()

	runner := validations.NewRunner()
	runner.Register(cmdvalidations.PackageDockerValidations(ctx)...)
	runner.StoreValidationResults()
	runner.ReportResults()

	return nil
}
