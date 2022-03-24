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

const (
	clusterNameFlagName = "cluster-name"
)

var cleanUpVsphereCmd = &cobra.Command{
	Use:          "vsphere",
	Short:        "Clean up e2e vms on vsphere vcenter",
	Long:         "Clean up vms created for e2e testing on vsphere vcenter",
	SilenceUsage: true,
	PreRun:       preRunCleanUpVsphereSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cleanUpVsphereTestResources(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to cleanup e2e vms on vsphere vcenter")
		}
		return nil
	},
}

func preRunCleanUpVsphereSetup(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

var requiredVsphereCleanUpFlags = []string{clusterNameFlagName}

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpVsphereCmd)
	cleanUpVsphereCmd.Flags().StringP(clusterNameFlagName, "n", "", "Cluster name for associated vms")

	for _, flag := range requiredVsphereCleanUpFlags {
		if err := cleanUpVsphereCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func cleanUpVsphereTestResources(ctx context.Context) error {
	clusterName := viper.GetString(clusterNameFlagName)
	err := e2e.CleanUpVsphereTestResources(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("running cleanup for vsphere vcenter vms: %v", err)
	}

	return nil
}
