package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type createSupportBundleOptions struct {
	fileName         string
	wConfig          string
	since            string
	sinceTime        string
	bundleConfig     string
	hardwareFileName string
}

var csbo = &createSupportBundleOptions{}

var supportbundleCmd = &cobra.Command{
	Use:          "support-bundle -f my-cluster.yaml",
	Short:        "Generate a support bundle",
	Long:         "This command is used to create a support bundle to troubleshoot a cluster",
	PreRunE:      preRunSupportBundle,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := csbo.validate(cmd.Context()); err != nil {
			return err
		}
		if err := csbo.createBundle(cmd.Context(), csbo.since, csbo.sinceTime, csbo.bundleConfig); err != nil {
			return fmt.Errorf("failed to create support bundle: %v", err)
		}
		return nil
	},
}

func init() {
	generateCmd.AddCommand(supportbundleCmd)
	supportbundleCmd.Flags().StringVarP(&csbo.sinceTime, "since-time", "", "", "Collect pod logs after a specific datetime(RFC3339) like 2021-06-28T15:04:05Z")
	supportbundleCmd.Flags().StringVarP(&csbo.since, "since", "", "", "Collect pod logs in the latest duration like 5s, 2m, or 3h.")
	supportbundleCmd.Flags().StringVarP(&csbo.bundleConfig, "bundle-config", "", "", "Bundle Config file to use when generating support bundle")
	supportbundleCmd.Flags().StringVarP(&csbo.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	supportbundleCmd.Flags().StringVarP(&csbo.wConfig, "w-config", "w", "", "Kubeconfig file to use when creating support bundle for a workload cluster")
	err := supportbundleCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (csbo *createSupportBundleOptions) validate(ctx context.Context) error {
	clusterConfig, err := commonValidation(ctx, csbo.fileName)
	if err != nil {
		return err
	}

	kubeconfigPath := kubeconfig.FromClusterName(clusterConfig.Name)
	if !validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return kubeconfig.NewMissingFileError(kubeconfigPath)
	}

	return nil
}

func preRunSupportBundle(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (csbo *createSupportBundleOptions) createBundle(ctx context.Context, since, sinceTime, bundleConfig string) error {
	clusterSpec, err := cluster.NewSpecFromClusterConfig(csbo.fileName, version.Get())
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).
		WithProvider(csbo.fileName, clusterSpec.Cluster, cc.skipIpCheck, csbo.hardwareFileName, cc.skipPowerActions).
		WithDiagnosticBundleFactory().
		Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	supportBundle, err := deps.DignosticCollectorFactory.DiagnosticBundle(clusterSpec, deps.Provider, getKubeconfigPath(clusterSpec.Cluster.Name, csbo.wConfig), bundleConfig)
	if err != nil {
		return fmt.Errorf("failed to parse collector: %v", err)
	}

	var sinceTimeValue *time.Time
	sinceTimeValue, err = diagnostics.ParseTimeOptions(since, sinceTime)
	if err != nil {
		return fmt.Errorf("failed parse since time: %v", err)
	}

	err = supportBundle.CollectAndAnalyze(ctx, sinceTimeValue)
	if err != nil {
		return fmt.Errorf("collecting and analyzing bundle: %v", err)
	}

	err = supportBundle.PrintAnalysis()
	if err != nil {
		return fmt.Errorf("printing analysis")
	}

	return nil
}
