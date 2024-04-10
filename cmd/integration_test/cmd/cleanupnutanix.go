package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/cleanup"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	endpointFlag     = "endpoint"
	portFlag         = "port"
	insecureFlag     = "insecure"
	ignoreErrorsFlag = "ignoreErrors"
)

var requiredNutanixCleanUpFlags = []string{clusterNameFlagName, endpointFlag}

var cleanUpNutanixCmd = &cobra.Command{
	Use:          "nutanix",
	Short:        "Clean up e2e vms on Nutanix Prism",
	Long:         "Clean up vms created for e2e testing on Nutanix Prism",
	SilenceUsage: true,
	PreRun:       preRunCleanUpNutanixSetup,
	RunE: func(_ *cobra.Command, _ []string) error {
		err := cleanUpNutanixTestResources()
		if err != nil {
			logger.Fatal(err, "Failed to cleanup e2e vms on Nutanix Prism")
		}
		return nil
	},
}

func preRunCleanUpNutanixSetup(cmd *cobra.Command, _ []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpNutanixCmd)

	cleanUpNutanixCmd.Flags().StringP(clusterNameFlagName, "n", "", "Cluster name for associated vms")
	cleanUpNutanixCmd.Flags().StringP(endpointFlag, "e", "", "Nutanix Prism endpoint")
	cleanUpNutanixCmd.Flags().StringP(portFlag, "p", "9440", "Nutanix Prism port")
	cleanUpNutanixCmd.Flags().BoolP(insecureFlag, "k", false, "skip TLS when contacting Prism APIs")
	cleanUpNutanixCmd.Flags().Bool(ignoreErrorsFlag, true, "ignore APIs errors when deleting VMs")

	for _, flag := range requiredNutanixCleanUpFlags {
		if err := cleanUpNutanixCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func cleanUpNutanixTestResources() error {
	clusterName := viper.GetString(clusterNameFlagName)
	err := cleanup.NutanixTestResources(clusterName, viper.GetString(endpointFlag), viper.GetString(portFlag), viper.IsSet(insecureFlag), viper.IsSet(ignoreErrorsFlag))
	if err != nil {
		return fmt.Errorf("running cleanup for Nutanix vms: %v", err)
	}

	return nil
}
