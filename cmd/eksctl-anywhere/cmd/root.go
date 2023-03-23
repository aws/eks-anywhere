package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	logsFolder := filepath.Join(".", "eksa-cli-logs")
	err := os.MkdirAll(logsFolder, 0o750)
	if err != nil {
		return fmt.Errorf("failed to create logs folder: %v", err)
	}

	outputFilePath := filepath.Join(".", "eksa-cli-logs", fmt.Sprintf("%s.log", time.Now().Format("2006-01-02T15_04_05")))
	if err = logger.Init(logger.Options{
		Level:          viper.GetInt("verbosity"),
		OutputFilePath: outputFilePath,
	}); err != nil {
		return fmt.Errorf("root cmd: %v", err)
	}

	return nil
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

// RootCmd returns the eksctl-anywhere root cmd.
func RootCmd() *cobra.Command {
	return rootCmd
}
