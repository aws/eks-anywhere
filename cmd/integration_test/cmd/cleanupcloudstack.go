package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/cleanup"
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

const deleteDuplicateNetworksFlag = "delete-duplicate-networks"

var requiredCloudstackCleanUpFlags = []string{clusterNameFlagName}

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpCloudstackCmd)
	cleanUpCloudstackCmd.Flags().StringP(clusterNameFlagName, "n", "", "Cluster name for associated vms")
	cleanUpCloudstackCmd.Flags().Bool(deleteDuplicateNetworksFlag, false, "Delete duplicate isolated networks")

	for _, flag := range requiredCloudstackCleanUpFlags {
		if err := cleanUpCloudstackCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func cleanUpCloudstackTestResources(ctx context.Context) error {
	clusterName := viper.GetString(clusterNameFlagName)
	deleteDuplicateNetworks := viper.IsSet(deleteDuplicateNetworksFlag)
	err := cleanup.CloudstackTestResources(ctx, clusterName, false, deleteDuplicateNetworks)
	if err != nil {
		return fmt.Errorf("running cleanup for cloudstack vms: %v", err)
	}

	return nil
}
