package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/e2e"
	"github.com/aws/eks-anywhere/pkg/logger"
)

var cleanUpCloudstackCmd = &cobra.Command{
	Use:          "cloudstack",
	Short:        "Clean up e2e vms on cloudstack",
	Long:         "Clean up vms created for e2e testing on cloudstack",
	SilenceUsage: true,
	PreRun:       preRunCleanUpCloudstackSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cleanUpCloudstackTestResources(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to cleanup e2e vms on cloudstack")
		}
		return nil
	},
}

func preRunCleanUpCloudstackSetup(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

var requiredCloudstackCleanUpFlags = []string{clusterNameFlagName}

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpCloudstackCmd)
	cleanUpCloudstackCmd.Flags().StringP(clusterNameFlagName, "n", "", "Cluster name for associated vms")

	for _, flag := range requiredCloudstackCleanUpFlags {
		if err := cleanUpCloudstackCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func cleanUpCloudstackTestResources(ctx context.Context) error {
	clusterName := viper.GetString(clusterNameFlagName)
	err := e2e.CleanUpCloudstackTestResources(ctx, clusterName, false)
	if err != nil {
		return fmt.Errorf("running cleanup for cloudstack vms: %v", err)
	}

	return nil
}
