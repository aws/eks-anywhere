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
	maxAgeFlagName = "max-age"
	tagFlagName    = "tag"
)

var cleanUpAwsCmd = &cobra.Command{
	Use:          "aws",
	Short:        "Clean up e2e resources on aws",
	Long:         "Clean up resources created for e2e testing on aws",
	SilenceUsage: true,
	PreRun:       preRunCleanUpAwsSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cleanUpAwsTestResources(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to cleanup e2e resources on aws")
		}
		return nil
	},
}

func preRunCleanUpAwsSetup(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

var requiredAwsCleanUpFlags = []string{storageBucketFlagName, maxAgeFlagName, tagFlagName}

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpAwsCmd)
	cleanUpAwsCmd.Flags().StringP(storageBucketFlagName, "s", "", "Name of s3 bucket used for e2e testing")
	cleanUpAwsCmd.Flags().StringP(maxAgeFlagName, "a", "0", "Instance age in seconds after which it should be deleted")
	cleanUpAwsCmd.Flags().StringP(tagFlagName, "t", "", "EC2 instance tag")

	for _, flag := range requiredAwsCleanUpFlags {
		if err := cleanUpAwsCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func cleanUpAwsTestResources(ctx context.Context) error {
	maxAge := viper.GetString(maxAgeFlagName)
	storageBucket := viper.GetString(storageBucketFlagName)
	tag := viper.GetString(tagFlagName)

	err := e2e.CleanUpAwsTestResources(storageBucket, maxAge, tag)
	if err != nil {
		return fmt.Errorf("running cleanup for aws test resources: %v", err)
	}

	return nil
}
