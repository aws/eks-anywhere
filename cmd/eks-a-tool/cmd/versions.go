package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/executables"
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Get cluster versions",
	Long:  "Get the versions of images in cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := versions(cmd.Context())
		if err != nil {
			log.Fatalf("Error getting image versions: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionsCmd)
}

func versions(ctx context.Context) error {
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	return executableBuilder.BuildKubectlExecutable().ListCluster(ctx)
}
