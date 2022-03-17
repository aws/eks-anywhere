package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type generateSupportBundleOptions struct {
	fileName         string
	hardwareFileName string
}

var gsbo = &generateSupportBundleOptions{}

var generateBundleConfigCmd = &cobra.Command{
	Use:     "support-bundle-config",
	Short:   "Generate support bundle config",
	Long:    "This command is used to generate a default support bundle config yaml",
	PreRunE: preRunGenerateBundleConfigCmd,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := gsbo.validateCmdInput()
		if err != nil {
			return fmt.Errorf("command input validation failed: %v", err)
		}
		bundle, err := gsbo.generateBundleConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to generate bunlde config: %v", err)
		}
		err = bundle.PrintBundleConfig()
		if err != nil {
			return fmt.Errorf("failed to print bundle config: %v", err)
		}
		return nil
	},
}

func init() {
	generateCmd.AddCommand(generateBundleConfigCmd)
	generateBundleConfigCmd.Flags().StringVarP(&gsbo.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
}

func preRunGenerateBundleConfigCmd(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (gsbo *generateSupportBundleOptions) validateCmdInput() error {
	f := gsbo.fileName
	if f != "" {
		clusterConfigFileExist := validations.FileExists(f)
		if !clusterConfigFileExist {
			return fmt.Errorf("the cluster config file %s does not exist", f)
		}
		_, err := v1alpha1.GetAndValidateClusterConfig(f)
		if err != nil {
			return fmt.Errorf("unable to get cluster config from file: %v", err)
		}
	}
	return nil
}

func (gsbo *generateSupportBundleOptions) generateBundleConfig(ctx context.Context) (diagnostics.DiagnosticBundle, error) {
	f := gsbo.fileName
	if f == "" {
		factory := diagnostics.NewFactory(diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  diagnostics.NewAnalyzerFactory(),
			CollectorFactory: diagnostics.NewDefaultCollectorFactory(),
		})
		return factory.DiagnosticBundleDefault(), nil
	}

	clusterSpec, err := cluster.NewSpecFromClusterConfig(f, version.Get())
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).
		WithProvider(f, clusterSpec.Cluster, cc.skipIpCheck, gsbo.hardwareFileName, cc.skipPowerActions).
		WithDiagnosticBundleFactory().
		Build(ctx)
	if err != nil {
		return nil, err
	}
	defer close(ctx, deps)

	return deps.DignosticCollectorFactory.DiagnosticBundleFromSpec(clusterSpec, deps.Provider, kubeconfig.FromClusterName(clusterSpec.Cluster.Name))
}
