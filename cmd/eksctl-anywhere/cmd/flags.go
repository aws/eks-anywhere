package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	TinkerbellHardwareCSVFlagName        = "hardware-csv"
	TinkerbellHardwareCSVFlagAlias       = "z"
	TinkerbellHardwareCSVFlagDescription = "Path to a CSV file containing hardware data."
	KubeconfigFile                       = "kubeconfig"
)

func bindFlagsToViper(cmd *cobra.Command, args []string) error {
	var err error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err != nil {
			return
		}
		err = viper.BindPFlag(flag.Name, flag)
	})
	return err
}

func applyClusterOptionFlags(flagSet *pflag.FlagSet, clusterOpt *clusterOptions) {
	flagSet.StringVarP(&clusterOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	flagSet.StringVar(&clusterOpt.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	flagSet.StringVar(&clusterOpt.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
}

func applyTinkerbellHardwareFlag(flagSet *pflag.FlagSet, pathOut *string) {
	flagSet.StringVarP(
		pathOut,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)
}

func checkTinkerbellFlags(flagSet *pflag.FlagSet, hardwareCSVPath string, operationType Operation) error {
	flag := flagSet.Lookup(TinkerbellHardwareCSVFlagName)

	// If no flag was returned there is a developer error as the flag has been removed
	// from the program rendering it invalid.
	if flag == nil {
		panic("'hardwarefile' flag not configured")
	}

	if !viper.IsSet(TinkerbellHardwareCSVFlagName) || viper.GetString(TinkerbellHardwareCSVFlagName) == "" {
		if operationType == Create && !viper.IsSet(KubeconfigFile) { // For upgrade and workload cluster create, hardware-csv is an optional flag
			return fmt.Errorf("required flag \"%v\" not set", TinkerbellHardwareCSVFlagName)
		}
		return nil
	}

	if !validations.FileExists(hardwareCSVPath) {
		return fmt.Errorf("hardware config file %s does not exist", hardwareCSVPath)
	}

	return nil
}
