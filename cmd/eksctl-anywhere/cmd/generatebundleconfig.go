package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type generateSupportBundleOptions struct {
	fileName              string
	hardwareFileName      string
	tinkerbellBootstrapIP string
}

var gsbo = &generateSupportBundleOptions{}

var generateBundleConfigCmd = &cobra.Command{
	Use:     "support-bundle-config",
	Short:   "Generate support bundle config",
	Long:    "This command is used to generate a default support bundle config yaml",
	PreRunE: bindFlagsToViper,
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
	clusterConfigPath := gsbo.fileName
	if clusterConfigPath == "" {
		return gsbo.generateDefaultBundleConfig(ctx)
	}

	clusterSpec, err := readAndValidateClusterSpec(clusterConfigPath, version.Get())
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	deps, err := dependencies.ForSpec(clusterSpec).
		WithProvider(clusterConfigPath, clusterSpec.Cluster, cc.skipIpCheck, gsbo.hardwareFileName, false, gsbo.tinkerbellBootstrapIP, map[string]bool{}, nil).
		WithDiagnosticBundleFactory().
		Build(ctx)
	if err != nil {
		return nil, err
	}
	defer close(ctx, deps)

	return deps.DignosticCollectorFactory.DiagnosticBundleWorkloadCluster(clusterSpec, deps.Provider, kubeconfig.FromClusterName(clusterSpec.Cluster.Name))
}

func (gsbo *generateSupportBundleOptions) generateDefaultBundleConfig(ctx context.Context) (diagnostics.DiagnosticBundle, error) {
	f := dependencies.NewFactory().WithFileReader()
	deps, err := f.Build(ctx)
	if err != nil {
		return nil, err
	}
	defer close(ctx, deps)

	factory := diagnostics.NewFactory(diagnostics.EksaDiagnosticBundleFactoryOpts{
		AnalyzerFactory:  diagnostics.NewAnalyzerFactory(),
		CollectorFactory: diagnostics.NewDefaultCollectorFactory(deps.FileReader),
	})
	return factory.DiagnosticBundleDefault(), nil
}
