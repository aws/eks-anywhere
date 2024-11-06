package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/diagnostics/analyzer"
	"github.com/aws/eks-anywhere/pkg/diagnostics/analyzer/logs"
	"github.com/aws/eks-anywhere/pkg/diagnostics/kubernetes"
)

var debug = debugOptions{}

func init() {
	cmd := &cobra.Command{
		Use:          "debug [flags]",
		Short:        "Debug cluster from a support bundle",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return debug.run(cmd.Context())
		},
	}

	expCmd.AddCommand(cmd)
	cmd.Flags().StringVarP(&debug.supportBundlePath, "support-bundle", "s", "", "Path to the support bundle")

	if err := cmd.MarkFlagRequired("support-bundle"); err != nil {
		log.Fatalf("marking support-bundle flag as required: %s", err)
	}
}

type debugOptions struct {
	supportBundlePath string
}

func (d debugOptions) run(ctx context.Context) error {
	if d.supportBundlePath == "" {
		return fmt.Errorf("support-bundle flag is required")
	}

	fileInfo, err := os.Stat(d.supportBundlePath)
	if err != nil {
		return errors.Wrapf(err, "getting file info for support bundle path %s", d.supportBundlePath)
	}

	if !fileInfo.IsDir() {
		return errors.Errorf("only support-bundle folders are supported, please untar the support bundle and provide the path to the folder")
	}

	clusterResourcesPath := filepath.Join(d.supportBundlePath, "cluster-resources")
	if _, err := os.Stat(clusterResourcesPath); err != nil {
		return errors.Wrapf(err, "getting file info for cluster resources path %s", clusterResourcesPath)
	}

	client := kubernetes.NewBundleReaderForFolder(clusterResourcesPath)
	if err := client.Init(); err != nil {
		return errors.Wrap(err, "initializing bundle reader")
	}

	podLogsPath := filepath.Join(d.supportBundlePath, "logs")
	if _, err := os.Stat(podLogsPath); err != nil {
		return errors.Wrapf(err, "getting file info for logs path %s", podLogsPath)
	}

	a := analyzer.New(client, logs.NewFSReaderForFolder(podLogsPath))
	result, err := a.AnalyzeAll(ctx)
	if err != nil {
		return errors.Wrap(err, "analyzing support bundle")
	}

	printer := analyzer.NewPrinter()
	for _, r := range result {
		printer.Process(r)
	}
	return nil
}
