package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/validations"
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

func setupClusterOptionFlags(cmd *cobra.Command, clusterOpt *clusterOptions) {
	cmd.Flags().StringVarP(&clusterOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	cmd.Flags().StringVar(&clusterOpt.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	cmd.Flags().StringVar(&clusterOpt.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
}

func applyTinkerbellHardwareFlag(cmd *cobra.Command, pathOut *string) {
	cmd.Flags().StringVarP(
		path,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)
}

func checkTinkerbellFlags(cmd *cobra.Command, hardwareCSVPath string) error {
	flag := cmd.Flags().Lookup(TinkerbellHardwareCSVFlagName)

	// If no flag was returned there is a developer error as the flag has been removed
	// from the program rendering it invalid.
	if flag == nil {
		panic("'hardwarefile' flag not configured")
	}

	if !viper.IsSet(TinkerbellHardwareCSVFlagName) || viper.GetString(TinkerbellHardwareCSVFlagName) == "" {
		return fmt.Errorf("required flag \"%v\" not set", TinkerbellHardwareCSVFlagName)
	}

	if !validations.FileExists(hardwareCSVPath) {
		return fmt.Errorf("hardware config file %s does not exist", hardwareCSVPath)
	}

	return nil
}
