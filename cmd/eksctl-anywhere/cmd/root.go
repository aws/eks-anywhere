package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/console"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/signals"
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
	stdout := logger.NewPausableWriter(os.Stdout)

	if err := initLogger(stdout); err != nil {
		log.Fatal(err)
	}

	signals.On(func() {
		resume := stdout.Pause()

		fmt.Println("Warning: Terminating this operation may leave the cluster in an irrecoverable state.")
		if console.Confirm("Are you sure you want to exit?", os.Stdout, os.Stdin) {
			os.Exit(-1)
		}

		if err := resume(); err != nil {
			// Logging the error may not alert the user because the logger may be borked.
			// We can't terminate the program so instead we're writing to Stdout so the user is
			// aware an issue has happened.
			//
			// We still log the error as other log sinks may be working.
			logger.Error(err, "Resuming stdout logging")
			fmt.Fprintf(os.Stderr, "Failed to resume stdout logging: %s", err)
		}
	}, syscall.SIGINT)
}

func initLogger(consoleWriter io.Writer) error {
	logsFolder := filepath.Join(".", "eksa-cli-logs")
	err := os.MkdirAll(logsFolder, 0o750)
	if err != nil {
		return fmt.Errorf("failed to create logs folder: %v", err)
	}
	filename := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02T15_04_05"))
	logFilePath := filepath.Join(".", "eksa-cli-logs", filename)

	err = logger.Init(logger.Options{
		Level:          viper.GetInt("verbosity"),
		OutputFilePath: logFilePath,
		Console:        consoleWriter,
	})
	if err != nil {
		return fmt.Errorf("root cmd: init zap: %v", err)
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
