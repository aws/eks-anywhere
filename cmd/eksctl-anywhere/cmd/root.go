package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/logger"
)

var rootCmd = &cobra.Command{
	Use:              "anywhere",
	Short:            "Amazon EKS Anywhere",
	Long:             `Use eksctl anywhere to build your own self-managing cluster on your hardware with the best of Amazon EKS`,
	PersistentPreRun: rootPersistentPreRun,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		outputFilePath := logger.GetOutputFilePath()
		if outputFilePath == "" {
			return
		}

		if err := os.Remove(outputFilePath); err != nil {
			fmt.Printf("Failed to cleanup log file %s: %s", outputFilePath, err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().IntP("verbosity", "v", 0, "Set the log level verbosity")
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatalf("failed to bind flags for root: %v", err)
	}
}

func rootPersistentPreRun(cmd *cobra.Command, args []string) {
	if err := initLogger(); err != nil {
		log.Fatal(err)
	}
}

func initLogger() error {
	outputFilePath := fmt.Sprintf("./eksa-cli-%s.log", time.Now().Format("2006-01-02T15_04_05"))
	if err := logger.InitZap(logger.ZapOpts{
		Level:          viper.GetInt("verbosity"),
		OutputFilePath: outputFilePath,
	}); err != nil {
		return fmt.Errorf("failed init zap logger in root command: %v", err)
	}

	return nil
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}
