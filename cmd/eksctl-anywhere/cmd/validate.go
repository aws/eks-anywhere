package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/cmdvalidations"
)

type validateOptions struct {
	clusterOptions
	hardwareCSVPath string
}

var valOpt = &validateOptions{}

var validateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Validate configuration",
	Long:         "This command is used to validate eksctl anywhere configurations",
	PreRunE:      preRunValidate,
	SilenceUsage: true,
	RunE:         valOpt.validateCluster,
}

func init() {
	expCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&valOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	validateCmd.Flags().StringVarP(
		&valOpt.hardwareCSVPath,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)

	if err := validateCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunValidate(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (valOpt *validateOptions) validateCluster(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	runner := validations.NewRunner()
	runner.Register(cmdvalidations.PackageDockerValidations(ctx)...)
	runner.StoreValidationResults()

	// Config parse
	clusterConfig, err := cluster.ParseConfigFromFile(valOpt.fileName)
	if err != nil {
		return runner.ExitError(err)
	}

	runner.Register(cmdvalidations.PackageKubeConfigPath(clusterConfig.Cluster.Name)...)
	runner.StoreValidationResults()

	if clusterConfig.Cluster.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		flag := cmd.Flags().Lookup(TinkerbellHardwareCSVFlagName)

		// If no flag was returned there is a developer error as the flag has been removed
		// from the program rendering it invalid.
		if flag == nil {
			runner.ReportResults()
			panic("'hardwarefile' flag not configured")
		}

		if !viper.IsSet(TinkerbellHardwareCSVFlagName) || viper.GetString(TinkerbellHardwareCSVFlagName) == "" {
			return runner.ExitError(fmt.Errorf("required flag \"%v\" not set", TinkerbellHardwareCSVFlagName))
		}

		if !validations.FileExists(cc.hardwareCSVPath) {
			return runner.ExitError(fmt.Errorf("hardware config file %s does not exist", cc.hardwareCSVPath))
		}
	}

	runner.StoreValidationResults()
	runner.ReportResults()

	return nil
}
